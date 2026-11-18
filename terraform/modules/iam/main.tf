# IAM module: Lambda execution role, Fargate task/execution roles, MSK IAM auth policies.

# Lambda execution role (for Lambda to write logs to CloudWatch and pull image from ECR)
resource "aws_iam_role" "lambda_execution" {
  name = "${var.name_prefix}-lambda-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      },
    ]
  })

  tags = {
    Name        = "${var.name_prefix}-lambda-execution-role"
    Environment = var.name_prefix
  }
}

resource "aws_iam_role_policy_attachment" "lambda_execution_basic" {
  role       = aws_iam_role.lambda_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "lambda_execution_ecr" {
  role       = aws_iam_role.lambda_execution.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
}

resource "aws_iam_role_policy_attachment" "lambda_execution_vpc" {
  role       = aws_iam_role.lambda_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}

# MSK IAM auth policy for Lambda (connect + describe + read/write data)
resource "aws_iam_role_policy" "lambda_msk" {
  name = "${var.name_prefix}-lambda-msk-policy"
  role = aws_iam_role.lambda_execution.name
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "kafka-cluster:Connect",
          "kafka-cluster:DescribeCluster",
          "kafka-cluster:DescribeClusterV2",
        ]
        Resource = var.msk_cluster_arn
      },
      {
        Effect = "Allow"
        Action = [
          "kafka-cluster:ReadData*",
          "kafka-cluster:WriteData*",
        ]
        Resource = "${var.msk_cluster_arn}/topic/*"
      },
    ]
  })
}

# Fargate task execution role (for ECS agent to pull images, write logs)
resource "aws_iam_role" "fargate_task_execution" {
  name = "${var.name_prefix}-fargate-task-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      },
    ]
  })

  tags = {
    Name        = "${var.name_prefix}-fargate-task-execution-role"
    Environment = var.name_prefix
  }
}

resource "aws_iam_role_policy_attachment" "fargate_task_execution_basic" {
  role       = aws_iam_role.fargate_task_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# Fargate task role (for the relay container to access MSK with IAM auth)
resource "aws_iam_role" "fargate_task" {
  name = "${var.name_prefix}-fargate-task-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      },
    ]
  })

  tags = {
    Name        = "${var.name_prefix}-fargate-task-role"
    Environment = var.name_prefix
  }
}

resource "aws_iam_role_policy" "fargate_task_msk" {
  name = "${var.name_prefix}-fargate-task-msk-policy"
  role = aws_iam_role.fargate_task.name
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "kafka-cluster:Connect",
          "kafka-cluster:DescribeCluster",
          "kafka-cluster:DescribeClusterV2",
        ]
        Resource = var.msk_cluster_arn
      },
      {
        Effect = "Allow"
        Action = [
          "kafka-cluster:ReadData*",
          "kafka-cluster:WriteData*",
        ]
        Resource = "${var.msk_cluster_arn}/topic/*"
      },
    ]
  })
}

