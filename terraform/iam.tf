# Service Account for Cloud Run Ingestion Worker
resource "google_service_account" "worker_sa" {
  account_id   = "${local.app_name}-worker-sa"
  display_name = "getfetch Ingestion Worker Service Account"
  depends_on   = [google_project_service.services]
}

# Service Account for Cloud Run API Server
resource "google_service_account" "api_sa" {
  account_id   = "${local.app_name}-api-sa"
  display_name = "getfetch REST API Service Account"
  depends_on   = [google_project_service.services]
}

# Service Account for Cloud Scheduler
resource "google_service_account" "scheduler_sa" {
  account_id   = "${local.app_name}-scheduler-sa"
  display_name = "getfetch Cloud Scheduler Service Account"
  depends_on   = [google_project_service.services]
}

# Grant Cloud SQL Client permission to Worker and API
resource "google_project_iam_member" "worker_cloudsql" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.worker_sa.email}"
}

resource "google_project_iam_member" "api_cloudsql" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.api_sa.email}"
}

# Grant Secret Manager Secret Accessor to Worker and API
resource "google_project_iam_member" "worker_secrets" {
  project = var.project_id
  role    = "roles/secretmanager.secretAccessor"
  member  = "serviceAccount:${google_service_account.worker_sa.email}"
}

resource "google_project_iam_member" "api_secrets" {
  project = var.project_id
  role    = "roles/secretmanager.secretAccessor"
  member  = "serviceAccount:${google_service_account.api_sa.email}"
}

# Allow Cloud Scheduler to invoke Cloud Run Job
resource "google_cloud_run_v2_job_iam_member" "scheduler_invoker" {
  name     = google_cloud_run_v2_job.worker_job.name
  location = var.region
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.scheduler_sa.email}"
}
