output "msk_security_group_id" {
  value       = aws_security_group.msk.id
  description = "MSK security group ID"
}

output "lambda_security_group_id" {
  value       = aws_security_group.lambda.id
  description = "Lambda security group ID"
}

output "fargate_relay_security_group_id" {
  value       = aws_security_group.fargate_relay.id
  description = "Fargate relay security group ID"
}
