output "msk_cluster_arn" {
  value       = aws_msk_serverless_cluster.this.arn
  description = "ARN of the MSK Serverless cluster"
}

output "msk_cluster_name" {
  value       = aws_msk_serverless_cluster.this.cluster_name
  description = "Name of the MSK Serverless cluster"
}

output "msk_bootstrap_brokers" {
  value       = aws_msk_serverless_cluster.this.bootstrap_brokers_sasl_iam
  description = "SASL/IAM bootstrap brokers for the MSK Serverless cluster"
  sensitive   = false
}
