# Database URL secret stored in Google Secret Manager
resource "google_secret_manager_secret" "database_url" {
  secret_id = "${local.app_name}-database-url"

  replication {
    auto {}
  }

  depends_on = [google_project_service.services]
}

resource "google_secret_manager_secret_version" "database_url_val" {
  secret = google_secret_manager_secret.database_url.id

  # Connects via Cloud SQL unix socket: /cloudsql/INSTANCE_CONNECTION_NAME
  secret_data = "postgres://${google_sql_user.db_user.name}:${random_password.db_password.result}@localhost/${google_sql_database.database.name}?host=/cloudsql/${google_sql_database_instance.postgres.connection_name}&sslmode=disable"
}

# Resend API Key Secret
resource "google_secret_manager_secret" "resend_api_key" {
  secret_id = "${local.app_name}-resend-key"

  replication {
    auto {}
  }

  depends_on = [google_project_service.services]
}

resource "google_secret_manager_secret_version" "resend_api_key_val" {
  secret      = google_secret_manager_secret.resend_api_key.id
  secret_data = var.resend_api_key != "" ? var.resend_api_key : "dummy_placeholder"
}

# Slack / Discord Webhook URL Secret
resource "google_secret_manager_secret" "webhook_url" {
  secret_id = "${local.app_name}-webhook-url"

  replication {
    auto {}
  }

  depends_on = [google_project_service.services]
}

resource "google_secret_manager_secret_version" "webhook_url_val" {
  secret      = google_secret_manager_secret.webhook_url.id
  secret_data = var.webhook_url != "" ? var.webhook_url : "https://placeholder"
}
