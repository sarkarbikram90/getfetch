# getfetch / KubeJobs

An automated, high-precision job discovery and intelligence platform focusing on **Kubernetes, Platform Engineering, SRE, and Cloud Infrastructure** positions.

Instead of unreliable, fragile HTML scraping, `getfetch` uses direct public ATS API connectors (**Greenhouse**, **Lever**, **Ashby**) to poll published career boards, normalizes job data, runs a deterministic skill-matching engine, extracts match proof snippets, deduplicates via SHA256 content hashes, reconciles closed postings, dispatches email/webhook digests, and serves a full **REST API** with a modern **Next.js & TypeScript dashboard**.

---

## Architecture Overview

```
                      ┌──────────────────────┐
                      │   Public ATS APIs    │
                      │ Greenhouse/Lever/Ash │
                      └──────────┬───────────┘
                                 │
                     ┌───────────▼───────────┐
                     │   Ingestion Worker    │  (cmd/worker)
                     │ - Rate limiter & 429  │
                     │ - HTML & URL cleaner  │
                     │ - Skill/Role matcher  │
                     │ - Content hasher      │
                     │ - State reconciler    │
                     └─────┬───────────┬─────┘
                           │           │
                 Alerts    │           │ Persist & Deduplicate
        (Resend/Webhooks)  │           │
                           ▼           ▼
                   ┌──────────┐  ┌──────────────┐
                   │  Slack / │  │  PostgreSQL  │ (Full-text GIN,
                   │  Email   │  │   Database   │  Skills & Roles GIN)
                   └──────────┘  └──────┬───────┘
                                        │
                               ┌────────▼────────┐
                               │   Go REST API   │ (cmd/api)
                               │ - Jobs & Search │
                               │ - Saved Tracker │
                               │ - Source Health │
                               │ - Prometheus    │
                               └────────┬────────┘
                                        │
                               ┌────────▼────────┐
                               │     Next.js     │ (web/)
                               │  Web Dashboard  │
                               └─────────────────┘
```

---

## Features

- **ATS Connectors**: Native API integration for **Greenhouse**, **Lever**, and **Ashby** job boards.
- **Resilience & Politeness**: Automatic exponential backoff with jitter and `Retry-After` header handling on HTTP 429 and transient 5xx server errors.
- **Deterministic Matcher**:
  - Skills: Kubernetes / K8s, Helm, Terraform, Docker, ArgoCD, Flux, Istio, Envoy, Prometheus, Grafana, OpenTelemetry, Cilium, eBPF, Linux, GKE, EKS, AKS, AWS, GCP, Azure, Go, Python, Rust.
  - Roles: Site Reliability Engineer (SRE), Platform Engineer, DevOps Engineer, Cloud Engineer, Infrastructure Engineer, Production Engineer.
  - Context & Snippet Extraction: Extracts a 30-word snippet around matching keywords as verifiable proof.
  - Work Mode & Location: Detects Remote, Hybrid, On-site, and India-specific locations (Bengaluru, Hyderabad, Pune, Mumbai, etc.).
- **URL & HTML Normalization**: Strips tracking tokens (`utm_*`, `gh_src`, `lever-source`), canonicalizes URLs, strips HTML tags cleanly without mashing words.
- **Deduplication & Change Detection**: SHA256 content hashing detects title/JD updates across recrawls without triggering duplicate alerts.
- **State Reconciliation**: Automatically marks jobs as `closed` in PostgreSQL when they are no longer published on subsequent crawls.
- **Go REST API Server** (`cmd/api`):
  - `GET /api/v1/jobs` (full-text search, skills, roles, remote, location filtering, pagination)
  - `GET /api/v1/jobs/{id}`
  - `POST /api/v1/jobs/{id}/save` & `DELETE /api/v1/jobs/{id}/save` (personal application tracker)
  - `GET /api/v1/saved` (bookmarked roles)
  - `GET /api/v1/sources` (monitored boards)
  - `GET /api/v1/stats` (overview counts & top skills)
  - `GET /healthz` & `GET /metrics`
- **Next.js Web Dashboard** (`web/`):
  - Fast, clean, dark-themed UI built with Tailwind CSS & TypeScript.
  - Real-time debounced full-text search.
  - Multi-tag filtering for Skills, Target Roles, Remote/Hybrid modes, and India locations.
  - Personal **Application Tracker** (`Saved`, `Applied`, `Interviewing`, `Rejected`).
  - **Monitored Sources Directory** with live health checks.
- **Alerts & Daily Digests**:
  - **Email Digests (Resend)**: Sends responsive, modern HTML digests with match pills, snippets, and direct links.
  - **Webhooks**: Sends real-time notifications to Slack or Discord incoming webhook URLs.
- **SRE Observability**: Built-in Prometheus metrics (`/metrics`) tracking fetch latency, discovery counts, match rates, notification delivery, and errors.

---

## Quickstart

### 1. Run Everything via Docker Compose (Single Command)

```bash
docker compose -f deployments/docker-compose.yml up -d
```

This starts:
- **PostgreSQL 16**: `localhost:5432`
- **Go REST API**: `http://localhost:8080`
- **Next.js Dashboard**: `http://localhost:3000`
- **Prometheus**: `http://localhost:9091`
- **Ingestion Worker**: Polling company boards in the background

---

### 2. Local Development (Step-by-Step)

#### Step 1: Start Database
```bash
docker compose -f deployments/docker-compose.yml up -d postgres
```

#### Step 2: Ingest Jobs into Database
```bash
export DATABASE_URL="postgres://postgres:postgrespassword@localhost:5432/getfetch?sslmode=disable"
go run ./cmd/worker
```

*(Or test in dry-run mode without database: `go run ./cmd/worker -dry-run -source gitlab`)*

#### Step 3: Start Go REST API
```bash
go run ./cmd/api -port 8080
```

#### Step 4: Start Next.js Frontend
```bash
cd web
npm run dev
```

Open **`http://localhost:3000`** in your browser to explore jobs and track applications!

---

## REST API Reference

| Endpoint | Method | Description |
| :--- | :--- | :--- |
| `/api/v1/jobs` | `GET` | List matching jobs with filtering (`q`, `skills`, `roles`, `remote`, `country`, `limit`, `offset`) |
| `/api/v1/jobs/{id}` | `GET` | Retrieve full job details by ID |
| `/api/v1/jobs/{id}/save` | `POST` | Bookmark or update application status (`saved`, `applied`, `interviewing`, `rejected`) |
| `/api/v1/jobs/{id}/save` | `DELETE`| Remove job from personal tracker |
| `/api/v1/saved` | `GET` | List all tracked applications |
| `/api/v1/sources` | `GET` | List registered ATS company boards & crawl health |
| `/api/v1/stats` | `GET` | Overview metrics (active jobs, today's jobs, top skills) |
| `/healthz` | `GET` | Liveness health check |
| `/metrics` | `GET` | Prometheus observability metrics |

---

## Running Tests

Run the full Go test suite:

```bash
go test -v ./...
```

Test the Next.js frontend production build:

```bash
cd web && npm run build
```
