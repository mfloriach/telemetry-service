output "ecs_cluster_id" {
  value       = aws_ecs_cluster.this.id
  description = "ECS cluster ID"
}

output "ecs_service_name" {
  value       = aws_ecs_service.relay.name
  description = "ECS service name"
}

output "task_definition_arn" {
  value       = aws_ecs_task_definition.relay.arn
  description = "Fargate task definition ARN"
}
