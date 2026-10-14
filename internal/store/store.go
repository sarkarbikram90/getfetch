package store

import (
	"context"

	"github.com/bikrams/getfetch/internal/model"
)

// JobFilter defines query parameters for searching and filtering jobs
type JobFilter struct {
	Query      string   // full text search keywords
	Skills     []string // filter by matched skills (e.g. Kubernetes, Terraform)
	Roles      []string // filter by matched roles (e.g. SRE, Platform)
	RemoteType string   // remote, hybrid, onsite
	Country    string   // India, US, etc.
	Status     string   // active, closed
	Limit      int
	Offset     int
}

// Store defines the database contract for jobs, sources, and crawler metrics
type Store interface {
	// Source operations
	GetActiveSources(ctx context.Context) ([]model.Source, error)
	GetSourceByBoard(ctx context.Context, sourceType, boardIdentifier string) (*model.Source, error)
	UpsertSource(ctx context.Context, src *model.Source) error

	// Job operations
	UpsertJob(ctx context.Context, job *model.Job) (isNew bool, isUpdated bool, err error)
	GetJobByID(ctx context.Context, id int64) (*model.Job, error)
	ListJobs(ctx context.Context, filter JobFilter) ([]model.Job, int, error)

	// CrawlRun audit operations
	StartCrawlRun(ctx context.Context, sourceID int64) (int64, error)
	FinishCrawlRun(ctx context.Context, runID int64, status string, discovered, matched, newJobs, updatedJobs int, errMsg *string) error

	// Migrations & Maintenance
	Migrate(ctx context.Context) error
	Close()
}
