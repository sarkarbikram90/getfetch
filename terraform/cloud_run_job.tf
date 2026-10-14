# Cloud Run Job for Periodic Ingestion Crawler
resource "google_cloud_run_v2_job" "worker_job" {
  name     = "${local.app_name}-worker-job"
  location = var.region

  template {
    template {
      service_account = google_service_account.worker_sa.email
      timeout         = "900s" # 15 minutes max execution
      max_retries     = 1

      volumes {
        name = "cloudsql"
        cloud_sql_instance {
          instances = [google_sql_database_instance.postgres.connection_name]
        }
      }

      containers {
        image = var.worker_image

        volume_mounts {
          name       = "cloudsql"
          mount_path = "/cloudsql"
        }

        resources {
          limits = {
            cpu    = "1000m"
            memory = "512Mi"
          }
        }

        # Environment variables and secrets
        env {
          name = "DATABASE_URL"
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.database_url.secret_id
              version = "latest"
            }
          }
        }

        env {
          name  = "NOTIFY_EMAIL"
          value = var.notify_email
        }

        env {
          name = "RESEND_API_KEY"
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.resend_api_key.secret_id
              version = "latest"
            }
          }
        }

        env {
          name = "WEBHOOK_URL"
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.webhook_url.secret_id
              version = "latest"
            }
          }
        }
      }
    }
  }

  depends_on = [
    google_project_service.services,
    google_secret_manager_secret_version.database_url_val
  ]
}
