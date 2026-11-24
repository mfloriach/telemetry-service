variable "name_prefix" {
  description = "Prefix for resource names"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
}

variable "dynamodb_table_arn" {
  description = "ARN of the DynamoDB telemetry table"
  type        = string
}
