variable "aws_region" {
  type        = string
  description = "AWS region for the deployment."
}

variable "environment" {
  type        = string
  description = "Deployment environment name."
  validation {
    condition     = can(regex("^[a-z0-9-]+$", var.environment))
    error_message = "environment must contain lowercase letters, numbers, and hyphens."
  }
}

variable "app_name" {
  type        = string
  description = "Application name."
  validation {
    condition     = can(regex("^[a-z0-9-]+$", var.app_name))
    error_message = "app_name must contain lowercase letters, numbers, and hyphens."
  }
}

variable "app_port" {
  type        = number
  default     = 8080
  validation {
    condition     = var.app_port > 0 && var.app_port < 65536
    error_message = "app_port must be a valid TCP port."
  }
}

variable "github_repository" {
  type        = string
  description = "GitHub repository in owner/repository form."
  validation {
    condition     = can(regex("^[^/]+/[^/]+$", var.github_repository))
    error_message = "github_repository must be owner/repository."
  }
}

variable "github_branch" {
  type        = string
  default     = "main"
}
