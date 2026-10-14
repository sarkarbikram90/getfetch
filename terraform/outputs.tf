output "api_url" {
  description = "Public URL of the Go REST API"
  value       = google_cloud_run_v2_service.api_service.uri
}

output "web_url" {
  description = "Public URL of the Next.js Web Dashboard"
  value       = google_cloud_run_v2_service.web_service.uri
}

output "cloudsql_connection_name" {
  description = "Cloud SQL instance connection name"
  value       = google_sql_database_instance.postgres.connection_name
}

output "worker_job_name" {
  description = "Name of the Cloud Run Ingestion Job"
  value       = google_cloud_run_v2_job.worker_job.name
}
