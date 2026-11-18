output "telemetry_server_repository_url" {
  value       = aws_ecr_repository.telemetry_server.repository_url
  description = "ECR repository URL for telemetry-server image"
}

output "telemetry_server_repository_arn" {
  value       = aws_ecr_repository.telemetry_server.arn
  description = "ECR repository ARN for telemetry-server"
}

output "telemetry_relay_repository_url" {
  value       = aws_ecr_repository.telemetry_relay.repository_url
  description = "ECR repository URL for telemetry-relay image"
}

output "telemetry_relay_repository_arn" {
  value       = aws_ecr_repository.telemetry_relay.arn
  description = "ECR repository ARN for telemetry-relay"
}
