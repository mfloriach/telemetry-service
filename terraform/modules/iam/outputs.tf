output "lambda_execution_role_arn" {
  value       = aws_iam_role.lambda_execution.arn
  description = "Lambda execution role ARN"
}

output "fargate_task_execution_role_arn" {
  value       = aws_iam_role.fargate_task_execution.arn
  description = "Fargate task execution role ARN"
}

output "fargate_task_role_arn" {
  value       = aws_iam_role.fargate_task.arn
  description = "Fargate task role ARN"
}
