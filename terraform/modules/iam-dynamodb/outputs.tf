output "lambda_execution_role_arn" {
  value       = aws_iam_role.lambda_execution.arn
  description = "Lambda execution role ARN"
}

output "ec2_relay_instance_profile_name" {
  value       = aws_iam_instance_profile.ec2_relay.name
  description = "Instance profile name for the EC2 relay"
}
