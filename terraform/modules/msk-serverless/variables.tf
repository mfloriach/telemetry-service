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

variable "msk_security_group_id" {
  description = "MSK security group ID"
  type        = string
}

variable "kafka_version" {
  description = "MSK Kafka version"
  type        = string
  default     = "3.8"
}
