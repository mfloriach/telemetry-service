terraform {
  backend "s3" {}
}

locals {
  name_prefix = "${var.project_name}-${var.environment}-option-c"
}

module "vpc" {
  source = "../modules/vpc"

  name_prefix        = local.name_prefix
  aws_region         = var.aws_region
  availability_zones = var.availability_zones
  cidr_block         = var.vpc_cidr_block
  enable_nat_gateway = var.enable_nat_gateway
}

module "ecr" {
  source = "../modules/ecr"

  name_prefix = local.name_prefix
}

module "dynamodb" {
  source = "../modules/dynamodb"

  name_prefix = local.name_prefix
}

module "iam_dynamodb" {
  source = "../modules/iam-dynamodb"

  name_prefix        = local.name_prefix
  aws_region         = var.aws_region
  dynamodb_table_arn = module.dynamodb.table_arn
}

module "ec2_relay" {
  source = "../modules/ec2-relay"

  name_prefix               = local.name_prefix
  aws_region                = var.aws_region
  vpc_id                    = module.vpc.vpc_id
  subnet_id                 = module.vpc.public_subnet_ids[0]
  iam_instance_profile_name = module.iam_dynamodb.ec2_relay_instance_profile_name
  telemetry_relay_image_uri = "${module.ecr.telemetry_relay_repository_url}:latest"
  dynamodb_table_name       = module.dynamodb.table_name
  app_env                   = var.environment
  log_level                 = var.log_level
  instance_type             = var.instance_type
}

resource "aws_security_group" "lambda" {
  name        = "${local.name_prefix}-lambda-sg"
  description = "Security group for the telemetry-server Lambda"
  vpc_id      = module.vpc.vpc_id

  ingress {
    description = "Allow inbound from API Gateway"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    description = "Allow outbound to DynamoDB and other AWS services"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${local.name_prefix}-lambda-sg"
    Environment = var.environment
  }
}

resource "aws_lambda_function" "telemetry_server" {
  function_name = "${local.name_prefix}-telemetry-server"
  role          = module.iam_dynamodb.lambda_execution_role_arn
  package_type  = "Image"
  image_uri     = "${module.ecr.telemetry_server_repository_url}:latest"
  vpc_config {
    subnet_ids         = module.vpc.private_subnet_ids
    security_group_ids = [aws_security_group.lambda.id]
  }
  environment {
    variables = {
      DYNAMODB_TABLE = module.dynamodb.table_name
      HTTP_ADDR      = ":8080"
      APP_ENV        = var.environment
      LOG_LEVEL      = var.log_level
    }
  }
  memory_size = var.lambda_memory_mb
  timeout     = var.lambda_timeout_seconds

  tags = {
    Name        = "${local.name_prefix}-telemetry-server"
    Environment = var.environment
  }
}

resource "aws_lambda_permission" "api_gateway" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.telemetry_server.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.http_api.execution_arn}/*/*"
}

resource "aws_apigatewayv2_api" "http_api" {
  name          = "${local.name_prefix}-http-api"
  protocol_type = "HTTP"

  cors_configuration {
    allow_origins = ["*"]
    allow_methods = ["GET"]
    allow_headers = ["*"]
    max_age       = 300
  }

  tags = {
    Name        = "${local.name_prefix}-http-api"
    Environment = var.environment
  }
}

resource "aws_apigatewayv2_integration" "lambda_integration" {
  api_id                 = aws_apigatewayv2_api.http_api.id
  integration_type       = "AWS_PROXY"
  integration_uri        = aws_lambda_function.telemetry_server.invoke_arn
  payload_format_version = "2.0"
}

resource "aws_apigatewayv2_route" "telemetry_latest" {
  api_id    = aws_apigatewayv2_api.http_api.id
  route_key = "GET /telemetry/latest"
  target    = "integrations/${aws_apigatewayv2_integration.lambda_integration.id}"
}

resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.http_api.id
  name        = var.environment
  auto_deploy = true

  tags = {
    Name        = "${local.name_prefix}-api-stage"
    Environment = var.environment
  }
}
