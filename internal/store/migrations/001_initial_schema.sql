-- Schema migration: 001_initial_schema.sql

CREATE TABLE IF NOT EXISTS sources (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    source_type TEXT NOT NULL,
    company_name TEXT NOT NULL,
    board_identifier TEXT NOT NULL,
    source_url TEXT NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    crawl_interval_minutes INT NOT NULL DEFAULT 360,
    last_success_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_source_type_board UNIQUE (source_type, board_identifier)
);

CREATE TABLE IF NOT EXISTS jobs (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    source_id BIGINT NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
    external_id TEXT NOT NULL,
    canonical_url TEXT NOT NULL,
    company_name TEXT NOT NULL,
    title TEXT NOT NULL,
    location TEXT NOT NULL DEFAULT '',
    remote_type TEXT NOT NULL DEFAULT 'unknown',
    country TEXT NOT NULL DEFAULT '',
    department TEXT NOT NULL DEFAULT '',
    description_raw TEXT,
    description_text TEXT NOT NULL,
    matched_skills TEXT[] NOT NULL DEFAULT '{}',
    matched_roles TEXT[] NOT NULL DEFAULT '{}',
    snippet_match TEXT NOT NULL DEFAULT '',
    content_hash CHAR(64) NOT NULL,
    posted_at TIMESTAMPTZ,
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status TEXT NOT NULL DEFAULT 'active',
    source_payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_source_external UNIQUE (source_id, external_id)
);

CREATE TABLE IF NOT EXISTS crawl_runs (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    source_id BIGINT REFERENCES sources(id) ON DELETE SET NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    status TEXT NOT NULL,
    jobs_discovered INT NOT NULL DEFAULT 0,
    jobs_matched INT NOT NULL DEFAULT 0,
    jobs_new INT NOT NULL DEFAULT 0,
    jobs_updated INT NOT NULL DEFAULT 0,
    error_message TEXT
);

CREATE TABLE IF NOT EXISTS user_saved_jobs (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    job_id BIGINT NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'saved',
    notes TEXT,
    applied_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_saved_job UNIQUE (job_id)
);

-- Performance and Full-Text Search Indexes
CREATE INDEX IF NOT EXISTS idx_jobs_first_seen ON jobs (first_seen_at DESC);
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs (status);
CREATE INDEX IF NOT EXISTS idx_jobs_content_hash ON jobs (content_hash);
CREATE INDEX IF NOT EXISTS idx_jobs_canonical_url ON jobs (canonical_url);
CREATE INDEX IF NOT EXISTS idx_jobs_matched_skills ON jobs USING GIN (matched_skills);
CREATE INDEX IF NOT EXISTS idx_jobs_matched_roles ON jobs USING GIN (matched_roles);
CREATE INDEX IF NOT EXISTS idx_jobs_fts ON jobs USING GIN (
    to_tsvector('english', coalesce(title, '') || ' ' || coalesce(description_text, ''))
);
