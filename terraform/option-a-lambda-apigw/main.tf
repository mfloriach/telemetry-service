# Option A: telemetry-server on Lambda behind API Gateway HTTP API.
# Relay runs on Fargate (shared module pattern replicated here).

terraform {
  backend "s3" {}
}

locals {
  name_prefix = "${var.project_name}-${var.environment}-option-a"
}

module "vpc" {
  source = "../modules/vpc"

  name_prefix        = local.name_prefix
  aws_region         = var.aws_region
  availability_zones = var.availability_zones
  cidr_block         = var.vpc_cidr_block
  enable_nat_gateway = var.enable_nat_gateway
}

module "security_groups" {
  source = "../modules/security-groups"

  name_prefix        = local.name_prefix
  vpc_id             = module.vpc.vpc_id
  private_subnet_ids = module.vpc.private_subnet_ids
}

module "msk_serverless" {
  source = "../modules/msk-serverless"

  name_prefix           = local.name_prefix
  aws_region            = var.aws_region
  vpc_id                = module.vpc.vpc_id
  private_subnet_ids    = module.vpc.private_subnet_ids
  msk_security_group_id = module.security_groups.msk_security_group_id
  kafka_version         = var.kafka_version
}

module "iam" {
  source = "../modules/iam"

  name_prefix     = local.name_prefix
  aws_region      = var.aws_region
  aws_account_id  = data.aws_caller_identity.current.account_id
  msk_cluster_arn = module.msk_serverless.msk_cluster_arn
}

module "ecr" {
  source = "../modules/ecr"

  name_prefix = local.name_prefix
}

module "ecs_fargate_relay" {
  source = "../modules/ecs-fargate-relay"

  name_prefix                     = local.name_prefix
  aws_region                      = var.aws_region
  vpc_id                          = module.vpc.vpc_id
  private_subnet_ids              = module.vpc.private_subnet_ids
  fargate_relay_security_group_id = module.security_groups.fargate_relay_security_group_id
  task_execution_role_arn         = module.iam.fargate_task_execution_role_arn
  task_role_arn                   = module.iam.fargate_task_role_arn
  telemetry_relay_image_uri       = "${module.ecr.telemetry_relay_repository_url}:latest"
  telemetry_ws_url                = var.telemetry_ws_url
  kafka_bootstrap_brokers         = module.msk_serverless.msk_bootstrap_brokers
  kafka_topic                     = var.kafka_topic
  app_env                         = var.environment
  log_level                       = var.log_level
  desired_count                   = var.relay_desired_count
}

data "aws_caller_identity" "current" {}

# Lambda function: telemetry-server container image
resource "aws_lambda_function" "telemetry_server" {
  function_name = "${local.name_prefix}-telemetry-server"
  role          = module.iam.lambda_execution_role_arn
  package_type  = "Image"
  image_uri     = "${module.ecr.telemetry_server_repository_url}:latest"
  vpc_config {
    subnet_ids         = module.vpc.private_subnet_ids
    security_group_ids = [module.security_groups.lambda_security_group_id]
  }
  environment {
    variables = {
      KAFKA_BROKERS = module.msk_serverless.msk_bootstrap_brokers
      KAFKA_TOPIC   = var.kafka_topic
      HTTP_ADDR     = ":8080"
      APP_ENV       = var.environment
      LOG_LEVEL     = var.log_level
    }
  }
  memory_size = var.lambda_memory_mb
  timeout     = var.lambda_timeout_seconds

  tags = {
    Name        = "${local.name_prefix}-telemetry-server"
    Environment = var.environment
  }
}

# Lambda permission for API Gateway to invoke
resource "aws_lambda_permission" "api_gateway" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.telemetry_server.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.http_api.execution_arn}/*/*"
}

# API Gateway HTTP API
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

# API Gateway integration (Lambda)
resource "aws_apigatewayv2_integration" "lambda_integration" {
  api_id                 = aws_apigatewayv2_api.http_api.id
  integration_type       = "AWS_PROXY"
  integration_uri        = aws_lambda_function.telemetry_server.invoke_arn
  payload_format_version = "2.0"
}

# API Gateway route: GET /telemetry/latest
resource "aws_apigatewayv2_route" "telemetry_latest" {
  api_id    = aws_apigatewayv2_api.http_api.id
  route_key = "GET /telemetry/latest"
  target    = "integrations/${aws_apigatewayv2_integration.lambda_integration.id}"
}

# API Gateway stage
resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.http_api.id
  name        = var.environment
  auto_deploy = true

  tags = {
    Name        = "${local.name_prefix}-api-stage"
    Environment = var.environment
  }
}

