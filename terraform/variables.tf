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
  description = "Container port exposed by the backend."

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
  type    = string
  default = "main"
}

variable "public_subnet_ids" {
  type        = set(string)
  description = "At least two public subnet IDs in distinct Availability Zones for the internet-facing Express service."

  validation {
    condition     = length(var.public_subnet_ids) >= 2
    error_message = "public_subnet_ids must contain at least two public subnets in distinct Availability Zones."
  }
}

variable "initial_image_tag" {
  type        = string
  description = "Existing immutable ECR image tag used when the Express service is first created."

  validation {
    condition     = can(regex("^sha-[0-9a-f]{40}$", var.initial_image_tag))
    error_message = "initial_image_tag must be a Git commit tag in sha-<40 lowercase hexadecimal characters> form."
  }
}

variable "cors_allowed_origins" {
  type        = list(string)
  default     = []
  description = "Browser origins allowed to call the backend. An empty value retains the backend's local-development default."

  validation {
    condition     = alltrue([for origin in var.cors_allowed_origins : can(regex("^https?://[^/?#]+$", origin))])
    error_message = "cors_allowed_origins entries must be HTTP(S) origins without a path."
  }
}
