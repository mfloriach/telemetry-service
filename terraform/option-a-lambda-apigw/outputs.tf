output "api_gateway_endpoint" {
  value       = aws_apigatewayv2_stage.default.invoke_url
  description = "API Gateway HTTP API endpoint URL"
}

output "lambda_function_name" {
  value       = aws_lambda_function.telemetry_server.function_name
  description = "Lambda function name"
}

output "msk_bootstrap_brokers" {
  value       = module.msk_serverless.msk_bootstrap_brokers
  description = "MSK bootstrap brokers"
}

output "vpc_id" {
  value       = module.vpc.vpc_id
  description = "VPC ID"
}

output "ecs_relay_service_name" {
  value       = module.ecs_fargate_relay.ecs_service_name
  description = "Fargate relay ECS service name"
}
