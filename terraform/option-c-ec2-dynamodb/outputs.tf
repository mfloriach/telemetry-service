output "api_gateway_endpoint" {
  value       = aws_apigatewayv2_stage.default.invoke_url
  description = "API Gateway HTTP API endpoint"
}

output "lambda_function_name" {
  value       = aws_lambda_function.telemetry_server.function_name
  description = "Lambda function name"
}

output "dynamodb_table_name" {
  value       = module.dynamodb.table_name
  description = "DynamoDB telemetry table name"
}

output "ec2_instance_id" {
  value       = module.ec2_relay.instance_id
  description = "EC2 relay instance ID"
}

output "ec2_public_ip" {
  value       = module.ec2_relay.public_ip
  description = "Public IP address of the EC2 relay"
}

output "vpc_id" {
  value       = module.vpc.vpc_id
  description = "VPC ID"
}
