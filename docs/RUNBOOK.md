# getfetch / KubeJobs — SRE Operational Runbook

This runbook defines Service Level Objectives (SLOs), observability metrics, and step-by-step incident response procedures.

---

## 1. Service Level Objectives (SLOs)

| Metric | Target | Measurement Period | SLI Source |
| :--- | :--- | :--- | :--- |
| **Source Freshness** | 99% of active sources refreshed every 6 hours | Rolling 7 days | `rate(crawler_fetch_total{status="success"}[6h])` |
| **API Availability** | 99.9% 2xx/3xx HTTP responses | Rolling 30 days | `sum(rate(http_requests_total{status=~"2..|3.."}[5m])) / sum(rate(http_requests_total[5m]))` |
| **API Latency** | 95% of queries completed < 100ms | Rolling 30 days | `histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))` |
| **Alert Delivery** | 99.5% digest dispatch success | Rolling 30 days | `rate(notifications_sent_total[24h]) / (rate(notifications_sent_total[24h]) + rate(notifications_failed_total[24h]))` |

---

## 2. Key Prometheus Metrics

- `crawler_fetch_total{source, status}`: Total HTTP requests made to ATS sources.
- `crawler_fetch_duration_seconds{source}`: Latency histogram for fetching listings.
- `jobs_discovered_total{source}`: Raw postings parsed.
- `jobs_matched_total{source}`: Roles meeting Kubernetes / Platform criteria.
- `jobs_new_total{source}`: Newly discovered jobs persisted.
- `jobs_updated_total{source}`: Existing jobs with JD updates.
- `crawler_errors_total{source, error_type}`: Failures during fetch, parse, or database commit.
- `notifications_sent_total{channel}`: Email digests and webhook alerts sent.
- `notifications_failed_total{channel}`: Notification dispatch failures.

---

## 3. Incident Playbooks

### Playbook 1: High 429 Rate Limiting from ATS Sources
**Alert**: `HighCrawlerFetchErrors` or repeated 429 logs in worker.
**Symptoms**: Crawler logs show `HTTP 429 Too Many Requests`.
**Diagnosis**:
1. Check if concurrency is set too high:
   ```bash
   kubectl -n getfetch logs -l app=getfetch-worker --tail=100
   ```
2. Verify if a specific source is rate-limiting by checking `crawler_errors_total{source="..."}`.
**Mitigation**:
- Reduce concurrency: adjust `-concurrency 2` or `-concurrency 1`.
- Verify `httputil.Client` is respecting `Retry-After` headers and exponential backoff.
- Distribute polling schedule by adjusting Cloud Scheduler or K8s CronJob intervals.

---

### Playbook 2: Schema Break / Zero Jobs Discovered
**Symptoms**: A previously healthy board suddenly returns 0 jobs or 404.
**Diagnosis**:
1. Check if the company changed their board identifier (e.g., from `gitlab` to `gitlab-remote`):
   ```bash
   curl -I "https://boards-api.greenhouse.io/v1/boards/<board>/jobs"
   ```
2. Check `sources.last_error` in the database:
   ```sql
   SELECT company_name, board_identifier, last_error FROM sources WHERE last_error IS NOT NULL;
   ```
**Mitigation**:
- Update `sources` table with the corrected `board_identifier`.
- If the ATS provider updated their API schema, update the corresponding adapter package (`internal/adapter/<provider>`) and deploy a patched image.

---

### Playbook 3: Notification Delivery Failures
**Alert**: `NotificationDeliveryFailing`
**Symptoms**: `notifications_failed_total` counter is increasing; users report missing daily digests.
**Diagnosis**:
1. Inspect worker logs for `Failed to send notification digest`:
   ```bash
   kubectl -n getfetch logs -l app=getfetch-worker | grep "Failed to send notification"
   ```
2. Check Resend API key status or Slack webhook URL validity:
   ```bash
   curl -X POST "https://api.resend.com/emails" \
     -H "Authorization: Bearer $RESEND_API_KEY" \
     -H "Content-Type: application/json" \
     -d '{"from":"alerts@kubejobs.dev","to":["test@example.com"],"subject":"Test","text":"Ping"}'
   ```
**Mitigation**:
- Rotate `RESEND_API_KEY` in Google Secret Manager or Kubernetes secret.
- Check Resend account sending quotas or verified domain DNS records (SPF, DKIM, DMARC).

---

### Playbook 4: Database Connection Pool Saturation
**Alert**: `DatabaseConnectionPoolHigh`
**Symptoms**: REST API latency spikes, worker logs show `pool exhausted`.
**Diagnosis**:
```sql
SELECT count(*), state FROM pg_stat_activity GROUP BY state;
```
**Mitigation**:
- Scale max pool size in `internal/store/postgres/postgres.go` (`cfg.MaxConns`).
- For Cloud SQL, increase machine tier (`db_tier` in Terraform) or enable PgBouncer connection pooling.
