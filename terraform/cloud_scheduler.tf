# Cloud Scheduler triggers the Ingestion Job periodically
resource "google_cloud_scheduler_job" "crawler_trigger" {
  name             = "${local.app_name}-crawler-trigger"
  description      = "Triggers getfetch ATS job ingestion crawler on schedule"
  schedule         = var.crawl_schedule
  time_zone        = "Etc/UTC"
  attempt_deadline = "300s"

  http_target {
    http_method = "POST"
    uri         = "https://${var.region}-run.googleapis.com/v2/projects/${var.project_id}/locations/${var.region}/jobs/${google_cloud_run_v2_job.worker_job.name}:run"

    oauth_token {
      service_account_email = google_service_account.scheduler_sa.email
    }
  }

  depends_on = [
    google_project_service.services,
    google_cloud_run_v2_job.worker_job,
    google_cloud_run_v2_job_iam_member.scheduler_invoker
  ]
}
