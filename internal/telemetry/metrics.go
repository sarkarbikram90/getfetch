package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// CrawlerFetchTotal counts total requests made to ATS sources
	CrawlerFetchTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "crawler_fetch_total",
			Help: "Total number of HTTP fetch requests executed by the crawler",
		},
		[]string{"source", "status"},
	)

	// CrawlerFetchDurationSeconds measures latency of fetching jobs from ATS sources
	CrawlerFetchDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "crawler_fetch_duration_seconds",
			Help:    "Latency in seconds for source fetch operations",
			Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30},
		},
		[]string{"source"},
	)

	// JobsDiscoveredTotal counts raw job postings extracted from sources
	JobsDiscoveredTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "jobs_discovered_total",
			Help: "Total raw job postings fetched across all sources",
		},
		[]string{"source"},
	)

	// JobsMatchedTotal counts postings that matched target skill criteria
	JobsMatchedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "jobs_matched_total",
			Help: "Total job postings matching Kubernetes/SRE skill criteria",
		},
		[]string{"source"},
	)

	// JobsNewTotal counts first-time seen jobs inserted into the database
	JobsNewTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "jobs_new_total",
			Help: "Total brand new jobs discovered and persisted",
		},
		[]string{"source"},
	)

	// JobsUpdatedTotal counts existing jobs with description or status changes
	JobsUpdatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "jobs_updated_total",
			Help: "Total existing jobs updated in the database",
		},
		[]string{"source"},
	)

	// CrawlerErrorsTotal counts failures encountered during crawling
	CrawlerErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "crawler_errors_total",
			Help: "Total errors during fetch, parse, or database operations",
		},
		[]string{"source", "error_type"},
	)
)
