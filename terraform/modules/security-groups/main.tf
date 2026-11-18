# Security groups module: Lambda↔MSK, Fargate relay↔MSK, Fargate outbound to upstream WebSocket.

# MSK cluster security group (inbound from Lambda and Fargate relay on 9098 TLS)
resource "aws_security_group" "msk" {
  name        = "${var.name_prefix}-msk-sg"
  description = "Security group for MSK Serverless cluster"
  vpc_id      = var.vpc_id

  ingress {
    description     = "Allow Lambda to MSK over TLS (9098)"
    from_port       = 9098
    to_port         = 9098
    protocol        = "tcp"
    security_groups = [aws_security_group.lambda.id]
  }

  ingress {
    description     = "Allow Fargate relay to MSK over TLS (9098)"
    from_port       = 9098
    to_port         = 9098
    protocol        = "tcp"
    security_groups = [aws_security_group.fargate_relay.id]
  }

  egress {
    description = "Allow all outbound"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${var.name_prefix}-msk-sg"
    Environment = var.name_prefix
  }
}

# Lambda security group (inbound from API Gateway/ALB, outbound to MSK)
resource "aws_security_group" "lambda" {
  name        = "${var.name_prefix}-lambda-sg"
  description = "Security group for telemetry-server Lambda"
  vpc_id      = var.vpc_id

  ingress {
    description = "Allow inbound from API Gateway / Function URL"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    description = "Allow outbound to MSK and other AWS services"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${var.name_prefix}-lambda-sg"
    Environment = var.name_prefix
  }
}

# Fargate relay security group (inbound none - outbound only to upstream WebSocket and MSK)
resource "aws_security_group" "fargate_relay" {
  name        = "${var.name_prefix}-fargate-relay-sg"
  description = "Security group for telemetry-relay Fargate task"
  vpc_id      = var.vpc_id

  ingress {
    description = "Allow inbound from ALB or none (outbound-only workload)"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    description = "Allow outbound to upstream WebSocket and MSK"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${var.name_prefix}-fargate-relay-sg"
    Environment = var.name_prefix
  }
}

