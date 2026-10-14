package engine

import (
	"context"
	"log/slog"
	"time"

	"github.com/bikrams/getfetch/internal/adapter"
	"github.com/bikrams/getfetch/internal/matcher"
	"github.com/bikrams/getfetch/internal/model"
	"github.com/bikrams/getfetch/internal/normalizer"
	"github.com/bikrams/getfetch/internal/store"
	"github.com/bikrams/getfetch/internal/telemetry"
)

// CrawlSummary captures metrics for a single source crawl
type CrawlSummary struct {
	Source      model.Source
	Discovered  int
	Matched     int
	New         int
	Updated     int
	Duration    time.Duration
	MatchedJobs []model.Job
	Error       error
}

// Engine coordinates adapters, normalizers, matchers, and storage
type Engine struct {
	store    store.Store
	registry *adapter.Registry
	matcher  *matcher.Matcher
	logger   *slog.Logger
}

// NewEngine creates an initialized crawler engine
func NewEngine(s store.Store, reg *adapter.Registry, m *matcher.Matcher, logger *slog.Logger) *Engine {
	if logger == nil {
		logger = slog.Default()
	}
	if m == nil {
		m = matcher.NewMatcher()
	}
	return &Engine{
		store:    s,
		registry: reg,
		matcher:  m,
		logger:   logger,
	}
}

// CrawlSource executes an ingestion cycle for a single source
func (e *Engine) CrawlSource(ctx context.Context, source model.Source, dryRun bool) CrawlSummary {
	startTime := time.Now()
	summary := CrawlSummary{
		Source:      source,
		MatchedJobs: []model.Job{},
	}

	e.logger.Info("Starting source crawl",
		"company", source.CompanyName,
		"type", source.SourceType,
		"board", source.BoardIdentifier,
		"dry_run", dryRun,
	)

	var runID int64
	var err error
	if e.store != nil && !dryRun && source.ID > 0 {
		runID, err = e.store.StartCrawlRun(ctx, source.ID)
		if err != nil {
			e.logger.Error("Failed to start crawl run record", "err", err)
		}
	}

	ad, err := e.registry.Get(source.SourceType)
	if err != nil {
		summary.Error = err
		summary.Duration = time.Since(startTime)
		e.recordFailure(source, runID, err, summary)
		return summary
	}

	// Fetch raw jobs
	fetchStart := time.Now()
	rawJobs, err := ad.FetchJobs(ctx, source)
	fetchDuration := time.Since(fetchStart).Seconds()
	telemetry.CrawlerFetchDurationSeconds.WithLabelValues(source.CompanyName).Observe(fetchDuration)

	if err != nil {
		telemetry.CrawlerFetchTotal.WithLabelValues(source.CompanyName, "error").Inc()
		telemetry.CrawlerErrorsTotal.WithLabelValues(source.CompanyName, "fetch_failed").Inc()
		summary.Error = err
		summary.Duration = time.Since(startTime)
		e.recordFailure(source, runID, err, summary)
		return summary
	}

	telemetry.CrawlerFetchTotal.WithLabelValues(source.CompanyName, "success").Inc()
	summary.Discovered = len(rawJobs)
	telemetry.JobsDiscoveredTotal.WithLabelValues(source.CompanyName).Add(float64(len(rawJobs)))

	// Process and match each job
	for _, raw := range rawJobs {
		descText := normalizer.StripHTML(raw.DescriptionRaw)
		canonicalURL := normalizer.NormalizeURL(raw.CanonicalURL)

		matchRes := e.matcher.Evaluate(raw.Title, raw.Location, descText, raw.RemoteType)
		if !matchRes.IsTargetMatch {
			continue
		}

		summary.Matched++
		telemetry.JobsMatchedTotal.WithLabelValues(source.CompanyName).Inc()

		contentHash := normalizer.ComputeContentHash(raw.Title, source.CompanyName, descText)

		country := raw.Country
		if country == "" && e.matcher.IsIndiaLocation(raw.Location, "") {
			country = "India"
		}

		job := model.Job{
			SourceID:        source.ID,
			ExternalID:      raw.ExternalID,
			CanonicalURL:    canonicalURL,
			CompanyName:     source.CompanyName,
			Title:           raw.Title,
			Location:        raw.Location,
			RemoteType:      matchRes.RemoteType,
			Country:         country,
			Department:      raw.Department,
			DescriptionRaw:  raw.DescriptionRaw,
			DescriptionText: descText,
			MatchedSkills:   matchRes.MatchedSkills,
			MatchedRoles:    matchRes.MatchedRoles,
			SnippetMatch:    matchRes.SnippetMatch,
			ContentHash:     contentHash,
			PostedAt:        raw.PostedAt,
			Status:          model.JobStatusActive,
			SourcePayload:   raw.SourcePayload,
		}

		summary.MatchedJobs = append(summary.MatchedJobs, job)

		if e.store != nil && !dryRun {
			isNew, isUpdated, err := e.store.UpsertJob(ctx, &job)
			if err != nil {
				e.logger.Error("Failed to upsert job", "external_id", job.ExternalID, "err", err)
				telemetry.CrawlerErrorsTotal.WithLabelValues(source.CompanyName, "db_upsert").Inc()
				continue
			}

			if isNew {
				summary.New++
				telemetry.JobsNewTotal.WithLabelValues(source.CompanyName).Inc()
			} else if isUpdated {
				summary.Updated++
				telemetry.JobsUpdatedTotal.WithLabelValues(source.CompanyName).Inc()
			}
		}
	}

	summary.Duration = time.Since(startTime)

	if e.store != nil && !dryRun && runID > 0 {
		_ = e.store.FinishCrawlRun(ctx, runID, model.CrawlRunStatusSuccess, summary.Discovered, summary.Matched, summary.New, summary.Updated, nil)
	}

	e.logger.Info("Source crawl completed",
		"company", source.CompanyName,
		"discovered", summary.Discovered,
		"matched", summary.Matched,
		"new", summary.New,
		"updated", summary.Updated,
		"duration", summary.Duration.Round(time.Millisecond),
	)

	return summary
}

func (e *Engine) recordFailure(source model.Source, runID int64, err error, summary CrawlSummary) {
	e.logger.Error("Source crawl failed", "company", source.CompanyName, "err", err)
	if e.store != nil && runID > 0 {
		errMsg := err.Error()
		_ = e.store.FinishCrawlRun(context.Background(), runID, model.CrawlRunStatusFailed, summary.Discovered, summary.Matched, summary.New, summary.Updated, &errMsg)
	}
}
