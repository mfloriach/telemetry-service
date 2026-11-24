output "table_name" {
  value       = aws_dynamodb_table.telemetry.name
  description = "DynamoDB telemetry table name"
}

output "table_arn" {
  value       = aws_dynamodb_table.telemetry.arn
  description = "DynamoDB telemetry table ARN"
}
