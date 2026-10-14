package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bikrams/getfetch/internal/adapter"
	"github.com/bikrams/getfetch/internal/adapter/ashby"
	"github.com/bikrams/getfetch/internal/adapter/greenhouse"
	"github.com/bikrams/getfetch/internal/adapter/lever"
	"github.com/bikrams/getfetch/internal/engine"
	"github.com/bikrams/getfetch/internal/matcher"
	"github.com/bikrams/getfetch/internal/model"
	"github.com/bikrams/getfetch/internal/store/postgres"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	var (
		dryRun      = flag.Bool("dry-run", false, "Run crawler in dry-run mode without database connection")
		seedsPath   = flag.String("seeds", "seeds/sources.json", "Path to initial seed sources JSON")
		sourceFlag  = flag.String("source", "", "Filter crawling to a specific company or board identifier")
		typeFlag    = flag.String("type", "", "Filter crawling to a specific ATS type (greenhouse, lever, ashby)")
		dbURL       = flag.String("db", "", "PostgreSQL database connection URL (defaults to DATABASE_URL env)")
		metricsPort = flag.Int("metrics-port", 9090, "Port to serve Prometheus /metrics (set to 0 to disable)")
		concurrency = flag.Int("concurrency", 3, "Number of concurrent sources to process")
		loop        = flag.Bool("loop", false, "Run continuously on a periodic schedule instead of exiting")
	)
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	fmt.Println("===================================================================")
	fmt.Println("  getfetch / KubeJobs — Platform & SRE Ingestion Engine")
	fmt.Println("===================================================================")

	// Context with graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Optional Prometheus metrics server
	if *metricsPort > 0 {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})
		srv := &http.Server{
			Addr:    fmt.Sprintf(":%d", *metricsPort),
			Handler: mux,
		}
		go func() {
			logger.Info("Starting Prometheus metrics server", "addr", srv.Addr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Warn("Metrics server failed", "err", err)
			}
		}()
		defer srv.Shutdown(context.Background())
	}

	// Setup ATS adapters
	registry := adapter.NewRegistry()
	registry.Register(greenhouse.NewAdapter(nil))
	registry.Register(lever.NewAdapter(nil))
	registry.Register(ashby.NewAdapter(nil))

	// Setup database store (unless dry-run)
	var store *postgres.PostgresStore
	databaseURL := *dbURL
	if databaseURL == "" {
		databaseURL = os.Getenv("DATABASE_URL")
	}

	if !*dryRun {
		if databaseURL == "" {
			logger.Info("No DATABASE_URL configured; running in --dry-run mode")
			*dryRun = true
		} else {
			var err error
			store, err = postgres.NewStore(ctx, databaseURL)
			if err != nil {
				logger.Error("Failed to connect to PostgreSQL", "err", err)
				os.Exit(1)
			}
			defer store.Close()

			logger.Info("Connected to PostgreSQL, applying schema migrations...")
			if err := store.Migrate(ctx); err != nil {
				logger.Error("Failed to apply database migrations", "err", err)
				os.Exit(1)
			}
			logger.Info("PostgreSQL migrations applied successfully")
		}
	}

	// Load seed sources
	sources, err := loadSources(*seedsPath, store, ctx, *dryRun)
	if err != nil {
		logger.Error("Failed to load sources", "err", err)
		os.Exit(1)
	}

	// Filter sources if flags were provided
	var targetSources []model.Source
	for _, src := range sources {
		if *sourceFlag != "" && !strings.EqualFold(src.BoardIdentifier, *sourceFlag) && !strings.EqualFold(src.CompanyName, *sourceFlag) {
			continue
		}
		if *typeFlag != "" && !strings.EqualFold(src.SourceType, *typeFlag) {
			continue
		}
		targetSources = append(targetSources, src)
	}

	if len(targetSources) == 0 {
		logger.Warn("No sources matched the given criteria")
		return
	}

	logger.Info("Loaded sources for crawling", "total", len(targetSources), "concurrency", *concurrency)

	eng := engine.NewEngine(store, registry, matcher.NewMatcher(), logger)

	runCycle := func() {
		runIngestionCycle(ctx, eng, targetSources, *concurrency, *dryRun)
	}

	runCycle()

	if *loop {
		ticker := time.NewTicker(4 * time.Hour)
		defer ticker.Stop()
		logger.Info("Running in continuous schedule mode; next crawl in 4 hours")
		for {
			select {
			case <-ctx.Done():
				logger.Info("Shutting down worker gracefully...")
				return
			case <-ticker.C:
				logger.Info("Triggering scheduled crawl run...")
				runCycle()
			}
		}
	}
}

func loadSources(seedsPath string, s *postgres.PostgresStore, ctx context.Context, dryRun bool) ([]model.Source, error) {
	data, err := os.ReadFile(seedsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read seeds file: %w", err)
	}

	var seedList []model.Source
	if err := json.Unmarshal(data, &seedList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal seeds json: %w", err)
	}

	if dryRun || s == nil {
		return seedList, nil
	}

	// Upsert seeds into database
	for i := range seedList {
		if err := s.UpsertSource(ctx, &seedList[i]); err != nil {
			slog.Warn("Failed to seed source into database", "company", seedList[i].CompanyName, "err", err)
		}
	}

	// Fetch active sources from database
	return s.GetActiveSources(ctx)
}

func runIngestionCycle(ctx context.Context, eng *engine.Engine, sources []model.Source, concurrency int, dryRun bool) {
	start := time.Now()
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	var (
		mu              sync.Mutex
		totalDiscovered int
		totalMatched    int
		totalNew        int
		totalUpdated    int
		allMatchedJobs  []model.Job
	)

	for _, src := range sources {
		if ctx.Err() != nil {
			break
		}

		wg.Add(1)
		sem <- struct{}{}

		go func(source model.Source) {
			defer wg.Done()
			defer func() { <-sem }()

			summary := eng.CrawlSource(ctx, source, dryRun)

			mu.Lock()
			totalDiscovered += summary.Discovered
			totalMatched += summary.Matched
			totalNew += summary.New
			totalUpdated += summary.Updated
			allMatchedJobs = append(allMatchedJobs, summary.MatchedJobs...)
			mu.Unlock()
		}(src)
	}

	wg.Wait()

	duration := time.Since(start).Round(time.Millisecond)
	fmt.Println()
	fmt.Println("-------------------------------------------------------------------")
	fmt.Printf("Crawling Summary: %d Discovered | %d Matched | %d New | %d Updated\n",
		totalDiscovered, totalMatched, totalNew, totalUpdated)
	fmt.Printf("Total Duration: %s\n", duration)
	fmt.Println("-------------------------------------------------------------------")

	if dryRun && len(allMatchedJobs) > 0 {
		fmt.Printf("\n=== MATCHED KUBERNETES & PLATFORM JOBS (%d) ===\n\n", len(allMatchedJobs))
		for i, j := range allMatchedJobs {
			fmt.Printf("[%d] %s — %s\n", i+1, j.Title, j.CompanyName)
			fmt.Printf("    Location:    %s (Remote: %s, Country: %s)\n", j.Location, j.RemoteType, j.Country)
			fmt.Printf("    Skills:      %s\n", strings.Join(j.MatchedSkills, ", "))
			if len(j.MatchedRoles) > 0 {
				fmt.Printf("    Roles:       %s\n", strings.Join(j.MatchedRoles, ", "))
			}
			fmt.Printf("    Match Proof: %s\n", j.SnippetMatch)
			fmt.Printf("    URL:         %s\n\n", j.CanonicalURL)
		}
	}
}
