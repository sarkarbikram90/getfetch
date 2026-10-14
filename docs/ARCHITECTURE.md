# getfetch / KubeJobs — Architecture & Design Decisions

This document outlines the engineering architecture, reliability patterns, and trade-offs of the `getfetch` job intelligence platform.

---

## 1. Architectural Philosophy

### Why ATS APIs over Generic Web Scraping?
Traditional web scrapers scrape unstructured HTML pages across thousands of company career sites. This approach introduces major fragility:
1. **HTML Drift**: Every company redesign breaks DOM selectors (`.job-title`, `.posting-desc`).
2. **Bot Blocking & Cloudflare**: Modern career portals aggressively block headless browsers and IP addresses.
3. **Heavy Compute Overhead**: Rendering single-page applications with Puppeteer/Playwright requires massive CPU and memory.

**Our Approach**: We prioritize direct public ATS API connectors (**Greenhouse**, **Lever**, **Ashby**). 
- Companies use these official endpoints to power their own public job board frontends.
- They return structured JSON payloads with full descriptions in a single lightweight HTTP request.
- Negligible CPU footprint: A single Go binary can crawl 50 company boards in under 5 seconds.

---

## 2. Core Subsystems

```
┌─────────────────────────────────────────────────────────────┐
│                      INGESTION PIPELINE                     │
│                                                             │
│   ┌──────────────┐      ┌──────────────┐                    │
│   │  ATS Source  │ ───► │  HTTP Client │ (429 Backoff &     │
│   │   Registry   │      │ (Resilience) │  Retry-After)      │
│   └──────────────┘      └──────┬───────┘                    │
│                                │                            │
│                         ┌──────▼───────┐                    │
│                         │  Normalizer  │ (HTML Stripper &   │
│                         │  & Cleaner   │  URL Sanitizer)    │
│                         └──────┬───────┘                    │
│                                │                            │
│                         ┌──────▼───────┐                    │
│                         │ Skill/Role   │ (Regex + Snippet   │
│                         │   Matcher    │  Extraction)       │
│                         └──────┬───────┘                    │
│                                │                            │
│                         ┌──────▼───────┐                    │
│                         │ Deduplicator │ (SHA256 Content    │
│                         │ & Reconciler │  Hash & DB Upsert) │
│                         └──────────────┘                    │
└─────────────────────────────────────────────────────────────┘
```

### Deterministic Matching Engine
We deliberately avoid LLMs in the critical ingestion path:
- **Cost**: Processing thousands of long job descriptions through LLMs creates unnecessary recurring inference costs.
- **Latency & Reliability**: Deterministic regex rules evaluate thousands of descriptions in milliseconds without token limits or hallucinations.
- **Explainability**: Every matched role includes a **verifiable evidence snippet** (30-word window around the target keyword).

### Content Hashing & Change Detection
- Computes SHA256 signature of `title | company | normalized_description`.
- If a recruiter modifies the job description, the hash changes, updating the record and timestamp.
- If the content is unchanged, the pipeline only refreshes `last_seen_at = NOW()` and avoids firing duplicate notifications.

### State Reconciliation (Ghost Job Prevention)
- After every crawl cycle for a company, the database reconciles stored jobs against discovered IDs:
  ```sql
  UPDATE jobs SET status = 'closed', updated_at = NOW()
  WHERE source_id = $1 AND status = 'active' AND NOT (external_id = ANY($active_ids))
  ```
- Jobs removed from company career boards are automatically marked `closed`.

---

## 3. Database Schema Design (PostgreSQL 16)

- **Relational Tables**: `sources`, `jobs`, `crawl_runs`, `user_saved_jobs`.
- **Full-Text Search (GIN)**: Native English language dictionary indexing over `to_tsvector('english', title || ' ' || description_text)` provides sub-millisecond keyword searches without needing Elasticsearch.
- **Array Indexes (GIN)**: Fast filtering by overlapping skills (`matched_skills && ARRAY['Kubernetes', 'Go']`).
- **Audit History**: `crawl_runs` table records discovered, matched, new, and closed job counts alongside execution latency for SRE monitoring.

---

## 4. Security & Hardening

1. **SSRF Prevention**: Crawlers are restricted to approved ATS domains (`*.greenhouse.io`, `*.lever.co`, `*.ashbyhq.com`). User-supplied URLs are never crawled directly.
2. **Payload Protection**: HTTP responses are capped with `io.LimitReader(resp.Body, 512KB)` to prevent memory exhaustion attacks.
3. **Least Privilege IAM**: Separate service accounts for Ingestion Worker, REST API, and Cloud Scheduler. Database credentials never committed to Git; managed via Google Secret Manager.
