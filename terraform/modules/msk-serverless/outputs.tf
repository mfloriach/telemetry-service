output "msk_cluster_arn" {
  value       = aws_msk_serverless_cluster.this.arn
  description = "ARN of the MSK Serverless cluster"
}

output "msk_cluster_name" {
  value       = aws_msk_serverless_cluster.this.cluster_name
  description = "Name of the MSK Serverless cluster"
}

output "msk_bootstrap_brokers" {
  value       = aws_msk_serverless_cluster.this.bootstrap_brokers_tls
  description = "TLS bootstrap brokers for the MSK cluster"
  sensitive   = false
}
