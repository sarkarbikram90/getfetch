# getfetch / KubeJobs

An automated, high-precision job discovery and intelligence engine focusing on **Kubernetes, Platform Engineering, SRE, and Cloud Infrastructure** positions.

Instead of unreliable, fragile HTML scraping, `getfetch` uses direct public ATS API connectors (**Greenhouse**, **Lever**, **Ashby**) to poll published career boards, normalizes job data, runs a deterministic skill-matching engine, extracts match proof snippets, deduplicates via SHA256 content hashes, and persists listings to PostgreSQL with full-text search.

---

## Features

- **ATS Connectors**: Native API integration for **Greenhouse**, **Lever**, and **Ashby** job boards.
- **Deterministic Matcher**:
  - Skills: Kubernetes / K8s, Helm, Terraform, Docker, ArgoCD, Flux, Istio, Envoy, Prometheus, Grafana, OpenTelemetry, Cilium, eBPF, Linux, GKE, EKS, AKS, AWS, GCP, Azure, Go, Python, Rust.
  - Roles: Site Reliability Engineer (SRE), Platform Engineer, DevOps Engineer, Cloud Engineer, Infrastructure Engineer, Production Engineer.
  - Context & Snippet Extraction: Extracts a 30-word snippet around matching keywords as verifiable proof.
  - Work Mode & Location: Detects Remote, Hybrid, On-site, and India-specific locations (Bengaluru, Hyderabad, Pune, Mumbai, etc.).
- **URL & HTML Normalization**: Strips tracking tokens (`utm_*`, `gh_src`, `lever-source`), canonicalizes URLs, strips HTML tags cleanly without mashing words.
- **Deduplication**: Content hashing (SHA256) detects changes across recrawls without creating duplicate job alerts.
- **SRE Observability**: Built-in Prometheus metrics (`/metrics`) tracking fetch latency, discovery counts, match rates, and errors.
- **PostgreSQL 16**: Full-text search (GIN index), GIN array indexes on skills and roles, JSONB source payloads.

---

## Quickstart

### 1. Test Ingestion in Dry-Run Mode (No DB Needed)

You can run the worker immediately without setting up a database:

```bash
# Dry-run against all seed companies
go run ./cmd/worker -dry-run

# Test a single company board
go run ./cmd/worker -dry-run -source gitlab
go run ./cmd/worker -dry-run -source replit
go run ./cmd/worker -dry-run -source palantir

# Test by ATS provider
go run ./cmd/worker -dry-run -type ashby
```

### 2. Run with PostgreSQL & Docker Compose

Start PostgreSQL 16 and Prometheus:

```bash
docker compose -f deployments/docker-compose.yml up -d
```

Run the ingestion worker to populate the database:

```bash
export DATABASE_URL="postgres://postgres:postgrespassword@localhost:5432/getfetch?sslmode=disable"
go run ./cmd/worker
```

### 3. Run Unit Tests

```bash
go test -v ./...
```

---

## CLI Options

| Flag | Description | Default |
| :--- | :--- | :--- |
| `-dry-run` | Runs crawling without persisting to database | `false` |
| `-seeds` | Path to JSON file with seed companies | `seeds/sources.json` |
| `-source` | Filter crawling to a specific company/board | `""` (all) |
| `-type` | Filter by ATS type (`greenhouse`, `lever`, `ashby`) | `""` (all) |
| `-db` | PostgreSQL connection URL | `DATABASE_URL` env |
| `-metrics-port`| Port to expose Prometheus `/metrics` | `9090` |
| `-concurrency` | Number of concurrent boards to fetch | `3` |
| `-loop` | Run continuously with a 4-hour polling interval | `false` |

---

## Project Structure

```
getfetch/
├── cmd/
│   └── worker/                 # Ingestion worker CLI
├── internal/
│   ├── adapter/                # ATS connectors (Greenhouse, Lever, Ashby)
│   ├── engine/                 # Ingestion pipeline orchestrator
│   ├── matcher/                # Deterministic skill/role matcher & snippet extractor
│   ├── model/                  # Domain models (Job, Source, CrawlRun)
│   ├── normalizer/             # HTML stripping, URL canonicalization, content hashing
│   ├── store/                  # Storage interface & PostgreSQL repository
│   └── telemetry/              # Prometheus metrics
├── seeds/
│   └── sources.json            # Curated seed list of top tech companies
├── deployments/
│   ├── docker-compose.yml      # Local Postgres & Prometheus
│   ├── Dockerfile.worker       # Multi-stage container build
│   └── prometheus.yml          # Prometheus scrape config
└── Makefile
```
