# Option B: telemetry-server on Lambda behind Lambda Function URL.
# Relay runs on Fargate (shared module pattern replicated here).

terraform {
  backend "s3" {}
}

locals {
  name_prefix = "${var.project_name}-${var.environment}-option-b"
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

# Lambda Function URL (IAM auth)
resource "aws_lambda_function_url" "http_endpoint" {
  function_name      = aws_lambda_function.telemetry_server.function_name
  authorization_type = "AWS_IAM"

  cors {
    allow_origins = ["*"]
    allow_methods = ["GET"]
    allow_headers = ["*"]
    max_age       = 300
  }
}

