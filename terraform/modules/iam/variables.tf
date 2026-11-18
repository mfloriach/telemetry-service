variable "name_prefix" {
  description = "Prefix for resource names"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
}

variable "msk_cluster_arn" {
  description = "ARN of the MSK Serverless cluster"
  type        = string
}
