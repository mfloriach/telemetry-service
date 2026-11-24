output "instance_id" {
  value       = aws_instance.relay.id
  description = "EC2 relay instance ID"
}

output "public_ip" {
  value       = aws_instance.relay.public_ip
  description = "Public IP address of the EC2 relay"
}

output "log_group_name" {
  value       = aws_cloudwatch_log_group.relay.name
  description = "CloudWatch log group for the relay container"
}
