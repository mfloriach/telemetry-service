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

resource "aws_iam_role_policy" "lambda_dynamodb_read" {
  name = "${var.name_prefix}-lambda-dynamodb-read-policy"
  role = aws_iam_role.lambda_execution.name
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:Query",
          "dynamodb:GetItem",
          "dynamodb:BatchGetItem",
          "dynamodb:DescribeTable",
        ]
        Resource = [var.dynamodb_table_arn, "${var.dynamodb_table_arn}/index/*"]
      },
    ]
  })
}

resource "aws_iam_role" "ec2_relay" {
  name = "${var.name_prefix}-ec2-relay-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      },
    ]
  })

  tags = {
    Name        = "${var.name_prefix}-ec2-relay-role"
    Environment = var.name_prefix
  }
}

resource "aws_iam_role_policy_attachment" "ec2_relay_ecr" {
  role       = aws_iam_role.ec2_relay.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
}

resource "aws_iam_role_policy" "ec2_relay_dynamodb_write" {
  name = "${var.name_prefix}-ec2-relay-dynamodb-write-policy"
  role = aws_iam_role.ec2_relay.name
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:PutItem",
          "dynamodb:BatchWriteItem",
          "dynamodb:DescribeTable",
        ]
        Resource = var.dynamodb_table_arn
      },
    ]
  })
}

resource "aws_iam_role_policy" "ec2_relay_logs" {
  name = "${var.name_prefix}-ec2-relay-logs-policy"
  role = aws_iam_role.ec2_relay.name
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents",
        ]
        Resource = "arn:aws:logs:${var.aws_region}:*:log-group:/ec2/${var.name_prefix}/telemetry-relay:*"
      },
    ]
  })
}

resource "aws_iam_instance_profile" "ec2_relay" {
  name = "${var.name_prefix}-ec2-relay-profile"
  role = aws_iam_role.ec2_relay.name
}
