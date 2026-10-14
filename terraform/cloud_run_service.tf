# Cloud Run Service for Go REST API
resource "google_cloud_run_v2_service" "api_service" {
  name     = "${local.app_name}-api"
  location = var.region
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    service_account = google_service_account.api_sa.email

    scaling {
      min_instance_count = 0
      max_instance_count = 5
    }

    volumes {
      name = "cloudsql"
      cloud_sql_instance {
        instances = [google_sql_database_instance.postgres.connection_name]
      }
    }

    containers {
      image = var.api_image

      ports {
        container_port = 8080
      }

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
        name  = "PORT"
        value = "8080"
      }

      startup_probe {
        http_get {
          path = "/healthz"
          port = 8080
        }
        initial_delay_seconds = 2
        period_seconds        = 5
        failure_threshold     = 3
      }

      liveness_probe {
        http_get {
          path = "/healthz"
          port = 8080
        }
        period_seconds = 15
      }
    }
  }

  depends_on = [
    google_project_service.services,
    google_secret_manager_secret_version.database_url_val
  ]
}

# Allow public unauthenticated access to the API
resource "google_cloud_run_service_iam_member" "api_public" {
  service  = google_cloud_run_v2_service.api_service.name
  location = var.region
  role     = "roles/run.invoker"
  member   = "allUsers"
}

# Cloud Run Service for Next.js Web Dashboard
resource "google_cloud_run_v2_service" "web_service" {
  name     = "${local.app_name}-web"
  location = var.region
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    scaling {
      min_instance_count = 0
      max_instance_count = 5
    }

    containers {
      image = var.web_image

      ports {
        container_port = 3000
      }

      resources {
        limits = {
          cpu    = "1000m"
          memory = "512Mi"
        }
      }

      env {
        name  = "NEXT_PUBLIC_API_URL"
        value = "${google_cloud_run_v2_service.api_service.uri}/api/v1"
      }
    }
  }

  depends_on = [
    google_project_service.services,
    google_cloud_run_v2_service.api_service
  ]
}

# Allow public unauthenticated access to the Web Dashboard
resource "google_cloud_run_service_iam_member" "web_public" {
  service  = google_cloud_run_v2_service.web_service.name
  location = var.region
  role     = "roles/run.invoker"
  member   = "allUsers"
}
