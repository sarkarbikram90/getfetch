variable "project_id" {
  description = "Google Cloud Project ID"
  type        = string
}

variable "region" {
  description = "Google Cloud primary region"
  type        = string
  default     = "asia-south1" # Mumbai / India or change to us-central1
}

variable "environment" {
  description = "Deployment environment (e.g. prod, staging, dev)"
  type        = string
  default     = "prod"
}

variable "worker_image" {
  description = "Container image URI for getfetch-worker"
  type        = string
  default     = "gcr.io/getfetch/worker:latest"
}

variable "api_image" {
  description = "Container image URI for getfetch-api"
  type        = string
  default     = "gcr.io/getfetch/api:latest"
}

variable "web_image" {
  description = "Container image URI for getfetch-web"
  type        = string
  default     = "gcr.io/getfetch/web:latest"
}

variable "crawl_schedule" {
  description = "Cron schedule expression for Cloud Scheduler job trigger"
  type        = string
  default     = "0 */4 * * *" # Every 4 hours
}

variable "notify_email" {
  description = "Default recipient email address for daily digest alerts"
  type        = string
  default     = ""
}

variable "resend_api_key" {
  description = "Resend API key for sending email digests"
  type        = string
  sensitive   = true
  default     = ""
}

variable "webhook_url" {
  description = "Slack or Discord incoming webhook URL for instant job alerts"
  type        = string
  sensitive   = true
  default     = ""
}

variable "db_tier" {
  description = "Cloud SQL machine tier"
  type        = string
  default     = "db-f1-micro" # Lightweight & low cost for MVP; upgrade to db-custom-1-3840 for heavy workloads
}
