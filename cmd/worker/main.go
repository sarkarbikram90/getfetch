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
	"github.com/bikrams/getfetch/internal/notify"
	"github.com/bikrams/getfetch/internal/store/postgres"
	"github.com/bikrams/getfetch/internal/telemetry"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	DryRun      bool
	SeedsPath   string
	SourceFlag  string
	TypeFlag    string
	DBURL       string
	MetricsPort int
	Concurrency int
	Loop        bool
	Interval    time.Duration
	NotifyEmail string
	ResendKey   string
	FromEmail   string
	WebhookURL  string
}

func main() {
	var cfg Config
	flag.BoolVar(&cfg.DryRun, "dry-run", false, "Run crawler in dry-run mode without database connection")
	flag.StringVar(&cfg.SeedsPath, "seeds", "seeds/sources.json", "Path to initial seed sources JSON")
	flag.StringVar(&cfg.SourceFlag, "source", "", "Filter crawling to a specific company or board identifier")
	flag.StringVar(&cfg.TypeFlag, "type", "", "Filter crawling to a specific ATS type (greenhouse, lever, ashby)")
	flag.StringVar(&cfg.DBURL, "db", "", "PostgreSQL database connection URL (defaults to DATABASE_URL env)")
	flag.IntVar(&cfg.MetricsPort, "metrics-port", 9090, "Port to serve Prometheus /metrics (set to 0 to disable)")
	flag.IntVar(&cfg.Concurrency, "concurrency", 3, "Number of concurrent sources to process")
	flag.BoolVar(&cfg.Loop, "loop", false, "Run continuously on a periodic schedule instead of exiting")
	flag.DurationVar(&cfg.Interval, "interval", 4*time.Hour, "Schedule polling interval when --loop is active")
	flag.StringVar(&cfg.NotifyEmail, "notify-email", "", "Recipient email address for digest alerts (or NOTIFY_EMAIL env)")
	flag.StringVar(&cfg.ResendKey, "resend-key", "", "Resend API key for email delivery (or RESEND_API_KEY env)")
	flag.StringVar(&cfg.FromEmail, "from-email", "KubeJobs Alerts <onboarding@resend.dev>", "Sender email address (or FROM_EMAIL env)")
	flag.StringVar(&cfg.WebhookURL, "webhook-url", "", "Slack/Discord incoming webhook URL (or WEBHOOK_URL env)")
	flag.Parse()

	// Environment variable fallbacks
	if cfg.DBURL == "" {
		cfg.DBURL = os.Getenv("DATABASE_URL")
	}
	if cfg.NotifyEmail == "" {
		cfg.NotifyEmail = os.Getenv("NOTIFY_EMAIL")
	}
	if cfg.ResendKey == "" {
		cfg.ResendKey = os.Getenv("RESEND_API_KEY")
	}
	if envFrom := os.Getenv("FROM_EMAIL"); envFrom != "" {
		cfg.FromEmail = envFrom
	}
	if cfg.WebhookURL == "" {
		cfg.WebhookURL = os.Getenv("WEBHOOK_URL")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	fmt.Println("===================================================================")
	fmt.Println("  getfetch / KubeJobs — Platform & SRE Ingestion Engine")
	fmt.Println("===================================================================")

	// Context with graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Optional Prometheus metrics server
	if cfg.MetricsPort > 0 {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})
		srv := &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.MetricsPort),
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
	if !cfg.DryRun {
		if cfg.DBURL == "" {
			logger.Info("No DATABASE_URL configured; running in --dry-run mode")
			cfg.DryRun = true
		} else {
			var err error
			store, err = postgres.NewStore(ctx, cfg.DBURL)
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
	sources, err := loadSources(cfg.SeedsPath, store, ctx, cfg.DryRun)
	if err != nil {
		logger.Error("Failed to load sources", "err", err)
		os.Exit(1)
	}

	// Filter sources if flags were provided
	var targetSources []model.Source
	for _, src := range sources {
		if cfg.SourceFlag != "" && !strings.EqualFold(src.BoardIdentifier, cfg.SourceFlag) && !strings.EqualFold(src.CompanyName, cfg.SourceFlag) {
			continue
		}
		if cfg.TypeFlag != "" && !strings.EqualFold(src.SourceType, cfg.TypeFlag) {
			continue
		}
		targetSources = append(targetSources, src)
	}

	if len(targetSources) == 0 {
		logger.Warn("No sources matched the given criteria")
		return
	}

	logger.Info("Loaded sources for crawling", "total", len(targetSources), "concurrency", cfg.Concurrency)

	eng := engine.NewEngine(store, registry, matcher.NewMatcher(), logger)

	// Setup notifiers
	var notifiers []notify.Notifier
	if cfg.ResendKey != "" && cfg.NotifyEmail != "" {
		notifiers = append(notifiers, notify.NewResendNotifier(cfg.ResendKey, cfg.FromEmail, nil))
		logger.Info("Email digest alerts configured via Resend", "to", cfg.NotifyEmail)
	}
	if cfg.WebhookURL != "" {
		notifiers = append(notifiers, notify.NewWebhookNotifier(cfg.WebhookURL, nil))
		logger.Info("Webhook alerts configured")
	}

	runCycle := func() {
		runIngestionCycle(ctx, eng, targetSources, cfg, notifiers, logger)
	}

	runCycle()

	if cfg.Loop {
		ticker := time.NewTicker(cfg.Interval)
		defer ticker.Stop()
		logger.Info("Running in continuous schedule mode", "next_crawl_in", cfg.Interval)
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

	for i := range seedList {
		if err := s.UpsertSource(ctx, &seedList[i]); err != nil {
			slog.Warn("Failed to seed source into database", "company", seedList[i].CompanyName, "err", err)
		}
	}

	return s.GetActiveSources(ctx)
}

func runIngestionCycle(ctx context.Context, eng *engine.Engine, sources []model.Source, cfg Config, notifiers []notify.Notifier, logger *slog.Logger) {
	start := time.Now()
	sem := make(chan struct{}, cfg.Concurrency)
	var wg sync.WaitGroup

	var (
		mu              sync.Mutex
		totalDiscovered int
		totalMatched    int
		totalNew        int
		totalUpdated    int
		totalClosed     int
		allMatchedJobs  []model.Job
		newlyFoundJobs  []model.Job
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

			summary := eng.CrawlSource(ctx, source, cfg.DryRun)

			mu.Lock()
			totalDiscovered += summary.Discovered
			totalMatched += summary.Matched
			totalNew += summary.New
			totalUpdated += summary.Updated
			totalClosed += summary.Closed
			allMatchedJobs = append(allMatchedJobs, summary.MatchedJobs...)

			// In dry-run mode, all matched jobs are treated as newly discovered for preview
			if cfg.DryRun {
				newlyFoundJobs = append(newlyFoundJobs, summary.MatchedJobs...)
			} else if summary.New > 0 {
				for _, j := range summary.MatchedJobs {
					newlyFoundJobs = append(newlyFoundJobs, j)
				}
			}
			mu.Unlock()
		}(src)
	}

	wg.Wait()

	duration := time.Since(start).Round(time.Millisecond)
	fmt.Println()
	fmt.Println("-------------------------------------------------------------------")
	fmt.Printf("Crawling Summary: %d Discovered | %d Matched | %d New | %d Updated | %d Closed\n",
		totalDiscovered, totalMatched, totalNew, totalUpdated, totalClosed)
	fmt.Printf("Total Duration: %s\n", duration)
	fmt.Println("-------------------------------------------------------------------")

	if cfg.DryRun && len(allMatchedJobs) > 0 {
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

	// Dispatch alerts if notifiers configured and jobs discovered
	if len(newlyFoundJobs) > 0 && len(notifiers) > 0 {
		logger.Info("Dispatching digest alerts for new roles", "count", len(newlyFoundJobs))
		payload := notify.DigestPayload{
			Recipient:   cfg.NotifyEmail,
			GeneratedAt: time.Now(),
			NewJobs:     newlyFoundJobs,
		}

		for _, n := range notifiers {
			if err := n.SendDigest(ctx, payload); err != nil {
				logger.Error("Failed to send notification digest", "notifier", n.Name(), "err", err)
				telemetry.NotificationsFailedTotal.WithLabelValues(n.Name()).Inc()
			} else {
				logger.Info("Notification digest dispatched successfully", "notifier", n.Name())
				telemetry.NotificationsSentTotal.WithLabelValues(n.Name()).Inc()
			}
		}
	}
}
