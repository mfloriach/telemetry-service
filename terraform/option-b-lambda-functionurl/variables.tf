variable "project_name" {
  description = "Project name used for resource naming"
  type        = string
  default     = "telemetry"
}

variable "environment" {
  description = "Deployment environment (dev, stage, prod)"
  type        = string
  validation {
    condition     = contains(["dev", "stage", "prod"], var.environment)
    error_message = "Environment must be one of: dev, stage, prod."
  }
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "availability_zones" {
  description = "List of availability zones"
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b", "us-east-1c"]
}

variable "vpc_cidr_block" {
  description = "VPC CIDR block"
  type        = string
  default     = "10.0.0.0/16"
}

variable "enable_nat_gateway" {
  description = "Whether to create NAT gateways"
  type        = bool
  default     = true
}

variable "kafka_topic" {
  description = "Kafka topic name"
  type        = string
  default     = "telemetry.raw"
}

variable "kafka_version" {
  description = "MSK Kafka version"
  type        = string
  default     = "3.8"
}

variable "lambda_memory_mb" {
  description = "Lambda memory in MB"
  type        = number
  default     = 512
}

variable "lambda_timeout_seconds" {
  description = "Lambda timeout in seconds"
  type        = number
  default     = 30
}

variable "log_level" {
  description = "Log level"
  type        = string
  default     = "info"
}

variable "telemetry_ws_url" {
  description = "Upstream WebSocket URL for the relay"
  type        = string
}

variable "relay_desired_count" {
  description = "Number of Fargate relay tasks"
  type        = number
  default     = 1
}
