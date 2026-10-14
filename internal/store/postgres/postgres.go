package postgres

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bikrams/getfetch/internal/model"
	"github.com/bikrams/getfetch/internal/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var migrationSQL string

// PostgresStore implements store.Store backed by PostgreSQL
type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewStore initializes a new connection pool to PostgreSQL
func NewStore(ctx context.Context, databaseURL string) (*PostgresStore, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres config: %w", err)
	}

	cfg.MaxConns = 25
	cfg.MinConns = 5
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// Ping database to verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return &PostgresStore{pool: pool}, nil
}

// Migrate executes the embedded schema migration
func (s *PostgresStore) Migrate(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, migrationSQL)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

// Close terminates the connection pool
func (s *PostgresStore) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

// GetActiveSources returns all enabled sources
func (s *PostgresStore) GetActiveSources(ctx context.Context) ([]model.Source, error) {
	query := `
		SELECT id, source_type, company_name, board_identifier, source_url,
		       enabled, crawl_interval_minutes, last_success_at, last_error, created_at, updated_at
		FROM sources
		WHERE enabled = TRUE
		ORDER BY company_name ASC
	`
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active sources: %w", err)
	}
	defer rows.Close()

	var sources []model.Source
	for rows.Next() {
		var src model.Source
		err := rows.Scan(
			&src.ID, &src.SourceType, &src.CompanyName, &src.BoardIdentifier, &src.SourceURL,
			&src.Enabled, &src.CrawlIntervalMinutes, &src.LastSuccessAt, &src.LastError,
			&src.CreatedAt, &src.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan source row: %w", err)
		}
		sources = append(sources, src)
	}

	return sources, nil
}

// GetSourceByBoard returns a source by its type and board identifier
func (s *PostgresStore) GetSourceByBoard(ctx context.Context, sourceType, boardIdentifier string) (*model.Source, error) {
	query := `
		SELECT id, source_type, company_name, board_identifier, source_url,
		       enabled, crawl_interval_minutes, last_success_at, last_error, created_at, updated_at
		FROM sources
		WHERE source_type = $1 AND board_identifier = $2
	`
	var src model.Source
	err := s.pool.QueryRow(ctx, query, sourceType, boardIdentifier).Scan(
		&src.ID, &src.SourceType, &src.CompanyName, &src.BoardIdentifier, &src.SourceURL,
		&src.Enabled, &src.CrawlIntervalMinutes, &src.LastSuccessAt, &src.LastError,
		&src.CreatedAt, &src.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query source: %w", err)
	}

	return &src, nil
}

// UpsertSource inserts or updates a source configuration
func (s *PostgresStore) UpsertSource(ctx context.Context, src *model.Source) error {
	query := `
		INSERT INTO sources (source_type, company_name, board_identifier, source_url, enabled, crawl_interval_minutes)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (source_type, board_identifier) DO UPDATE SET
			company_name = EXCLUDED.company_name,
			source_url = EXCLUDED.source_url,
			enabled = EXCLUDED.enabled,
			crawl_interval_minutes = EXCLUDED.crawl_interval_minutes,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`
	return s.pool.QueryRow(ctx, query,
		src.SourceType, src.CompanyName, src.BoardIdentifier, src.SourceURL, src.Enabled, src.CrawlIntervalMinutes,
	).Scan(&src.ID, &src.CreatedAt, &src.UpdatedAt)
}

// UpsertJob persists a job, detecting whether it is newly discovered or modified
func (s *PostgresStore) UpsertJob(ctx context.Context, job *model.Job) (isNew bool, isUpdated bool, err error) {
	var payloadBytes []byte
	if job.SourcePayload != nil {
		payloadBytes, _ = json.Marshal(job.SourcePayload)
	}

	// Check existing record
	var existingID int64
	var existingHash string
	checkQuery := `SELECT id, content_hash FROM jobs WHERE source_id = $1 AND external_id = $2`
	checkErr := s.pool.QueryRow(ctx, checkQuery, job.SourceID, job.ExternalID).Scan(&existingID, &existingHash)

	if errors.Is(checkErr, pgx.ErrNoRows) {
		// Insert new job
		insertQuery := `
			INSERT INTO jobs (
				source_id, external_id, canonical_url, company_name, title, location,
				remote_type, country, department, description_raw, description_text,
				matched_skills, matched_roles, snippet_match, content_hash, posted_at,
				first_seen_at, last_seen_at, status, source_payload
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
				NOW(), NOW(), 'active', $17
			) RETURNING id
		`
		err = s.pool.QueryRow(ctx, insertQuery,
			job.SourceID, job.ExternalID, job.CanonicalURL, job.CompanyName, job.Title, job.Location,
			job.RemoteType, job.Country, job.Department, job.DescriptionRaw, job.DescriptionText,
			job.MatchedSkills, job.MatchedRoles, job.SnippetMatch, job.ContentHash, job.PostedAt,
			payloadBytes,
		).Scan(&job.ID)
		if err != nil {
			return false, false, fmt.Errorf("failed to insert new job: %w", err)
		}
		return true, false, nil
	} else if checkErr != nil {
		return false, false, fmt.Errorf("failed to check existing job: %w", checkErr)
	}

	job.ID = existingID

	if existingHash != job.ContentHash {
		// Content has changed -> update and mark updated
		updateQuery := `
			UPDATE jobs SET
				canonical_url = $1, title = $2, location = $3, remote_type = $4,
				country = $5, department = $6, description_raw = $7, description_text = $8,
				matched_skills = $9, matched_roles = $10, snippet_match = $11,
				content_hash = $12, posted_at = COALESCE($13, posted_at), last_seen_at = NOW(),
				status = 'active', source_payload = $14, updated_at = NOW()
			WHERE id = $15
		`
		_, err = s.pool.Exec(ctx, updateQuery,
			job.CanonicalURL, job.Title, job.Location, job.RemoteType,
			job.Country, job.Department, job.DescriptionRaw, job.DescriptionText,
			job.MatchedSkills, job.MatchedRoles, job.SnippetMatch,
			job.ContentHash, job.PostedAt, payloadBytes, existingID,
		)
		if err != nil {
			return false, false, fmt.Errorf("failed to update changed job: %w", err)
		}
		return false, true, nil
	}

	// Content identical -> touch last_seen_at and keep active
	touchQuery := `UPDATE jobs SET last_seen_at = NOW(), status = 'active' WHERE id = $1`
	_, err = s.pool.Exec(ctx, touchQuery, existingID)
	if err != nil {
		return false, false, fmt.Errorf("failed to touch existing job: %w", err)
	}
	return false, false, nil
}

// GetJobByID retrieves a single job by its primary key
func (s *PostgresStore) GetJobByID(ctx context.Context, id int64) (*model.Job, error) {
	query := `
		SELECT id, source_id, external_id, canonical_url, company_name, title, location,
		       remote_type, country, department, description_raw, description_text,
		       matched_skills, matched_roles, snippet_match, content_hash, posted_at,
		       first_seen_at, last_seen_at, status, created_at, updated_at
		FROM jobs
		WHERE id = $1
	`
	var j model.Job
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&j.ID, &j.SourceID, &j.ExternalID, &j.CanonicalURL, &j.CompanyName, &j.Title, &j.Location,
		&j.RemoteType, &j.Country, &j.Department, &j.DescriptionRaw, &j.DescriptionText,
		&j.MatchedSkills, &j.MatchedRoles, &j.SnippetMatch, &j.ContentHash, &j.PostedAt,
		&j.FirstSeenAt, &j.LastSeenAt, &j.Status, &j.CreatedAt, &j.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get job by id: %w", err)
	}
	return &j, nil
}

// ListJobs returns jobs matching the provided filter criteria
func (s *PostgresStore) ListJobs(ctx context.Context, filter store.JobFilter) ([]model.Job, int, error) {
	var whereClauses []string
	var args []any
	argIdx := 1

	if filter.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	} else {
		whereClauses = append(whereClauses, "status = 'active'")
	}

	if filter.RemoteType != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("remote_type = $%d", argIdx))
		args = append(args, filter.RemoteType)
		argIdx++
	}

	if filter.Country != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("country = $%d", argIdx))
		args = append(args, filter.Country)
		argIdx++
	}

	if len(filter.Skills) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("matched_skills && $%d", argIdx))
		args = append(args, filter.Skills)
		argIdx++
	}

	if len(filter.Roles) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("matched_roles && $%d", argIdx))
		args = append(args, filter.Roles)
		argIdx++
	}

	if filter.Query != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("to_tsvector('english', coalesce(title, '') || ' ' || coalesce(description_text, '')) @@ plainto_tsquery('english', $%d)", argIdx))
		args = append(args, filter.Query)
		argIdx++
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Count total matching
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM jobs %s", whereSQL)
	var totalCount int
	if err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count jobs: %w", err)
	}

	limit := 50
	if filter.Limit > 0 {
		limit = filter.Limit
	}
	offset := 0
	if filter.Offset > 0 {
		offset = filter.Offset
	}

	query := fmt.Sprintf(`
		SELECT id, source_id, external_id, canonical_url, company_name, title, location,
		       remote_type, country, department, description_text,
		       matched_skills, matched_roles, snippet_match, content_hash, posted_at,
		       first_seen_at, last_seen_at, status, created_at, updated_at
		FROM jobs
		%s
		ORDER BY first_seen_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list jobs: %w", err)
	}
	defer rows.Close()

	var jobs []model.Job
	for rows.Next() {
		var j model.Job
		err := rows.Scan(
			&j.ID, &j.SourceID, &j.ExternalID, &j.CanonicalURL, &j.CompanyName, &j.Title, &j.Location,
			&j.RemoteType, &j.Country, &j.Department, &j.DescriptionText,
			&j.MatchedSkills, &j.MatchedRoles, &j.SnippetMatch, &j.ContentHash, &j.PostedAt,
			&j.FirstSeenAt, &j.LastSeenAt, &j.Status, &j.CreatedAt, &j.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan job: %w", err)
		}
		jobs = append(jobs, j)
	}

	return jobs, totalCount, nil
}

// ReconcileSourceJobs marks active jobs not in activeExternalIDs as closed
func (s *PostgresStore) ReconcileSourceJobs(ctx context.Context, sourceID int64, activeExternalIDs []string) (int64, error) {
	if len(activeExternalIDs) == 0 {
		return 0, nil
	}

	query := `
		WITH closed AS (
			UPDATE jobs
			SET status = 'closed', updated_at = NOW()
			WHERE source_id = $1
			  AND status = 'active'
			  AND NOT (external_id = ANY($2))
			RETURNING id
		)
		SELECT COUNT(*) FROM closed
	`
	var closedCount int64
	err := s.pool.QueryRow(ctx, query, sourceID, activeExternalIDs).Scan(&closedCount)
	if err != nil {
		return 0, fmt.Errorf("failed to reconcile closed jobs: %w", err)
	}
	return closedCount, nil
}

// StartCrawlRun records the start of a crawler run
func (s *PostgresStore) StartCrawlRun(ctx context.Context, sourceID int64) (int64, error) {
	query := `
		INSERT INTO crawl_runs (source_id, started_at, status)
		VALUES ($1, NOW(), 'running')
		RETURNING id
	`
	var runID int64
	err := s.pool.QueryRow(ctx, query, sourceID).Scan(&runID)
	if err != nil {
		return 0, fmt.Errorf("failed to start crawl run: %w", err)
	}
	return runID, nil
}

// FinishCrawlRun updates the completion metadata of a crawl run
func (s *PostgresStore) FinishCrawlRun(ctx context.Context, runID int64, status string, discovered, matched, newJobs, updatedJobs int, errMsg *string) error {
	query := `
		UPDATE crawl_runs SET
			completed_at = NOW(),
			status = $1,
			jobs_discovered = $2,
			jobs_matched = $3,
			jobs_new = $4,
			jobs_updated = $5,
			error_message = $6
		WHERE id = $7
	`
	_, err := s.pool.Exec(ctx, query, status, discovered, matched, newJobs, updatedJobs, errMsg, runID)
	if err != nil {
		return fmt.Errorf("failed to update crawl run: %w", err)
	}

	// If successful, update source last_success_at
	if status == model.CrawlRunStatusSuccess {
		updateSourceQuery := `
			UPDATE sources SET
				last_success_at = NOW(),
				last_error = NULL
			WHERE id = (SELECT source_id FROM crawl_runs WHERE id = $1)
		`
		_, _ = s.pool.Exec(ctx, updateSourceQuery, runID)
	} else if errMsg != nil {
		updateSourceErrQuery := `
			UPDATE sources SET
				last_error = $1
			WHERE id = (SELECT source_id FROM crawl_runs WHERE id = $2)
		`
		_, _ = s.pool.Exec(ctx, updateSourceErrQuery, *errMsg, runID)
	}

	return nil
}

// SaveJob inserts or updates a bookmarked job
func (s *PostgresStore) SaveJob(ctx context.Context, jobID int64, status, notes string) error {
	if status == "" {
		status = "saved"
	}
	query := `
		INSERT INTO user_saved_jobs (job_id, status, notes, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (job_id) DO UPDATE SET
			status = EXCLUDED.status,
			notes = EXCLUDED.notes,
			updated_at = NOW()
	`
	_, err := s.pool.Exec(ctx, query, jobID, status, notes)
	if err != nil {
		return fmt.Errorf("failed to save job: %w", err)
	}
	return nil
}

// UnsaveJob removes a job from bookmarks
func (s *PostgresStore) UnsaveJob(ctx context.Context, jobID int64) error {
	query := `DELETE FROM user_saved_jobs WHERE job_id = $1`
	_, err := s.pool.Exec(ctx, query, jobID)
	if err != nil {
		return fmt.Errorf("failed to unsave job: %w", err)
	}
	return nil
}

// GetSavedJobs returns all tracked jobs with full details
func (s *PostgresStore) GetSavedJobs(ctx context.Context) ([]model.SavedJobWithJob, error) {
	query := `
		SELECT s.id, s.job_id, s.status, s.notes, s.applied_at, s.created_at, s.updated_at,
		       j.id, j.source_id, j.external_id, j.canonical_url, j.company_name, j.title,
		       j.location, j.remote_type, j.country, j.department, j.description_text,
		       j.matched_skills, j.matched_roles, j.snippet_match, j.content_hash, j.posted_at,
		       j.first_seen_at, j.last_seen_at, j.status, j.created_at, j.updated_at
		FROM user_saved_jobs s
		JOIN jobs j ON s.job_id = j.id
		ORDER BY s.updated_at DESC
	`
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query saved jobs: %w", err)
	}
	defer rows.Close()

	var result []model.SavedJobWithJob
	for rows.Next() {
		var item model.SavedJobWithJob
		err := rows.Scan(
			&item.ID, &item.JobID, &item.Status, &item.Notes, &item.AppliedAt, &item.CreatedAt, &item.UpdatedAt,
			&item.Job.ID, &item.Job.SourceID, &item.Job.ExternalID, &item.Job.CanonicalURL, &item.Job.CompanyName, &item.Job.Title,
			&item.Job.Location, &item.Job.RemoteType, &item.Job.Country, &item.Job.Department, &item.Job.DescriptionText,
			&item.Job.MatchedSkills, &item.Job.MatchedRoles, &item.Job.SnippetMatch, &item.Job.ContentHash, &item.Job.PostedAt,
			&item.Job.FirstSeenAt, &item.Job.LastSeenAt, &item.Job.Status, &item.Job.CreatedAt, &item.Job.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan saved job: %w", err)
		}
		result = append(result, item)
	}
	return result, nil
}

// GetDashboardStats aggregates overview metrics for the dashboard
func (s *PostgresStore) GetDashboardStats(ctx context.Context) (*model.DashboardStats, error) {
	stats := &model.DashboardStats{
		TopSkills: make(map[string]int),
	}

	// 1. Total active jobs
	_ = s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM jobs WHERE status = 'active'").Scan(&stats.TotalActiveJobs)

	// 2. New jobs in last 24h
	_ = s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM jobs WHERE status = 'active' AND first_seen_at >= NOW() - INTERVAL '24 hours'").Scan(&stats.NewJobsLast24h)

	// 3. Total active sources
	_ = s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM sources WHERE enabled = TRUE").Scan(&stats.TotalSourcesActive)

	// 4. Top skills
	skillsQuery := `
		SELECT skill, COUNT(*) as cnt
		FROM (
			SELECT unnest(matched_skills) AS skill
			FROM jobs
			WHERE status = 'active'
		) t
		GROUP BY skill
		ORDER BY cnt DESC
		LIMIT 10
	`
	rows, err := s.pool.Query(ctx, skillsQuery)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var skill string
			var count int
			if err := rows.Scan(&skill, &count); err == nil {
				stats.TopSkills[skill] = count
			}
		}
	}

	return stats, nil
}
