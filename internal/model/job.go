package model

import (
	"time"
)

// Source types supported by getfetch
const (
	SourceTypeGreenhouse = "greenhouse"
	SourceTypeLever      = "lever"
	SourceTypeAshby      = "ashby"
	SourceTypeCustomHTML = "custom_html"
)

// Remote work classifications
const (
	RemoteTypeRemote  = "remote"
	RemoteTypeHybrid  = "hybrid"
	RemoteTypeOnsite  = "onsite"
	RemoteTypeUnknown = "unknown"
)

// Job listing status
const (
	JobStatusActive   = "active"
	JobStatusClosed   = "closed"
	JobStatusArchived = "archived"
)

// Crawl run status
const (
	CrawlRunStatusRunning = "running"
	CrawlRunStatusSuccess = "success"
	CrawlRunStatusFailed  = "failed"
)

// Source represents an ATS job board or career page to be polled
type Source struct {
	ID                   int64      `json:"id"`
	SourceType           string     `json:"source_type"`           // greenhouse, lever, ashby
	CompanyName          string     `json:"company_name"`          // e.g. "GitLab"
	BoardIdentifier      string     `json:"board_identifier"`      // e.g. "gitlab"
	SourceURL            string     `json:"source_url"`            // original careers board url
	Enabled              bool       `json:"enabled"`
	CrawlIntervalMinutes int        `json:"crawl_interval_minutes"`
	LastSuccessAt        *time.Time `json:"last_success_at,omitempty"`
	LastError            *string    `json:"last_error,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// RawJob represents an unparsed/pre-matched job fetched from an ATS adapter
type RawJob struct {
	ExternalID     string         `json:"external_id"`
	CanonicalURL   string         `json:"canonical_url"`
	CompanyName    string         `json:"company_name"`
	Title          string         `json:"title"`
	Location       string         `json:"location"`
	RemoteType     string         `json:"remote_type"`
	Country        string         `json:"country"`
	Department     string         `json:"department"`
	DescriptionRaw string         `json:"description_raw"`
	PostedAt       *time.Time     `json:"posted_at,omitempty"`
	SourcePayload  map[string]any `json:"source_payload,omitempty"`
}

// Job represents a normalized, skill-matched, deduplicated job listing
type Job struct {
	ID              int64          `json:"id"`
	SourceID        int64          `json:"source_id"`
	ExternalID      string         `json:"external_id"`
	CanonicalURL    string         `json:"canonical_url"`
	CompanyName     string         `json:"company_name"`
	Title           string         `json:"title"`
	Location        string         `json:"location"`
	RemoteType      string         `json:"remote_type"`
	Country         string         `json:"country"`
	Department      string         `json:"department"`
	DescriptionRaw  string         `json:"description_raw,omitempty"`
	DescriptionText string         `json:"description_text"`
	MatchedSkills   []string       `json:"matched_skills"`
	MatchedRoles    []string       `json:"matched_roles"`
	SnippetMatch    string         `json:"snippet_match"`
	ContentHash     string         `json:"content_hash"`
	PostedAt        *time.Time     `json:"posted_at,omitempty"`
	FirstSeenAt     time.Time      `json:"first_seen_at"`
	LastSeenAt      time.Time      `json:"last_seen_at"`
	Status          string         `json:"status"`
	SourcePayload   map[string]any `json:"source_payload,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// MatchResult contains the extracted evaluation metrics from the matcher
type MatchResult struct {
	MatchedSkills []string `json:"matched_skills"`
	MatchedRoles  []string `json:"matched_roles"`
	SnippetMatch  string   `json:"snippet_match"`
	RemoteType    string   `json:"remote_type"`
	IsTargetMatch bool     `json:"is_target_match"` // true if qualifies as K8s/SRE role
}

// CrawlRun logs crawler execution metadata for SRE observability
type CrawlRun struct {
	ID             int64      `json:"id"`
	SourceID       int64      `json:"source_id"`
	StartedAt      time.Time  `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	Status         string     `json:"status"`
	JobsDiscovered int        `json:"jobs_discovered"`
	JobsMatched    int        `json:"jobs_matched"`
	JobsNew        int        `json:"jobs_new"`
	JobsUpdated    int        `json:"jobs_updated"`
	ErrorMessage   *string    `json:"error_message,omitempty"`
}

// SavedJob represents a bookmarked or tracked job application
type SavedJob struct {
	ID        int64      `json:"id"`
	JobID     int64      `json:"job_id"`
	Status    string     `json:"status"` // 'saved', 'applied', 'interviewing', 'rejected', 'archived'
	Notes     string     `json:"notes"`
	AppliedAt *time.Time `json:"applied_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// SavedJobWithJob combines saved metadata with the full job listing
type SavedJobWithJob struct {
	SavedJob
	Job Job `json:"job"`
}

// DashboardStats provides overview metrics for the web dashboard
type DashboardStats struct {
	TotalActiveJobs    int            `json:"total_active_jobs"`
	NewJobsLast24h     int            `json:"new_jobs_last_24h"`
	TotalSourcesActive int            `json:"total_sources_active"`
	TopSkills          map[string]int `json:"top_skills"`
}
