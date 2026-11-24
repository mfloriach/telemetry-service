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

variable "subnet_id" {
  description = "Public subnet ID for the relay instance"
  type        = string
}

variable "iam_instance_profile_name" {
  description = "Instance profile name granting ECR, DynamoDB and CloudWatch access"
  type        = string
}

variable "telemetry_relay_image_uri" {
  description = "Container image URI for telemetry-relay"
  type        = string
}

variable "dynamodb_table_name" {
  description = "DynamoDB table the relay writes to"
  type        = string
}

variable "app_env" {
  description = "Application environment"
  type        = string
}

variable "log_level" {
  description = "Log level"
  type        = string
}

variable "instance_type" {
  description = "EC2 instance type for the relay"
  type        = string
  default     = "t3.micro"
}

variable "ws_port" {
  description = "Port the relay WebSocket server listens on"
  type        = number
  default     = 9000
}
