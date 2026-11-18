variable "name_prefix" {
  description = "Prefix for resource names"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID"
  type        = string
}

variable "private_subnet_ids" {
  description = "Private subnet IDs"
  type        = list(string)
}

variable "fargate_relay_security_group_id" {
  description = "Fargate relay security group ID"
  type        = string
}

variable "task_execution_role_arn" {
  description = "ECS task execution role ARN"
  type        = string
}

variable "task_role_arn" {
  description = "ECS task role ARN"
  type        = string
}

variable "telemetry_relay_image_uri" {
  description = "Container image URI for telemetry-relay"
  type        = string
}

variable "telemetry_ws_url" {
  description = "Upstream WebSocket URL for the relay"
  type        = string
}

variable "kafka_bootstrap_brokers" {
  description = "MSK Kafka bootstrap brokers"
  type        = string
}

variable "kafka_topic" {
  description = "Kafka topic name"
  type        = string
  default     = "telemetry.raw"
}

variable "app_env" {
  description = "Application environment"
  type        = string
  default     = "development"
}

variable "log_level" {
  description = "Log level"
  type        = string
  default     = "info"
}

variable "desired_count" {
  description = "Number of Fargate tasks"
  type        = number
  default     = 1
}

variable "cpu" {
  description = "CPU units for Fargate task"
  type        = string
  default     = "256"
}

variable "memory" {
  description = "Memory in MB for Fargate task"
  type        = string
  default     = "512"
}
