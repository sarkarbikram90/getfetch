# Generate secure random password for database
resource "random_password" "db_password" {
  length  = 24
  special = false
}

# Cloud SQL PostgreSQL 16 Instance
resource "google_sql_database_instance" "postgres" {
  name             = "${local.app_name}-db-${var.environment}"
  database_version = "POSTGRES_16"
  region           = var.region
  deletion_protection = false # Set to true for production

  settings {
    tier = var.db_tier

    ip_configuration {
      ipv4_enabled = true # Enabled for Cloud SQL Auth proxy / Cloud Run Unix socket connection
    }

    backup_configuration {
      enabled                        = true
      point_in_time_recovery_enabled = false
      start_time                     = "03:00"
    }

    maintenance_window {
      day  = 7 # Sunday
      hour = 4
    }

    database_flags {
      name  = "cloudsql.enable_pgaudit"
      value = "on"
    }
  }

  depends_on = [google_project_service.services]
}

# Application Database
resource "google_sql_database" "database" {
  name     = "getfetch"
  instance = google_sql_database_instance.postgres.name
}

# Application Database User
resource "google_sql_user" "db_user" {
  name     = "getfetch_app"
  instance = google_sql_database_instance.postgres.name
  password = random_password.db_password.result
}
