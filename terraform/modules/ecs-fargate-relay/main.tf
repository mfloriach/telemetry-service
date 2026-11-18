# ECS Fargate relay module: ECS cluster, Fargate task definition, Fargate service (desiredCount=1).

# ECS cluster
resource "aws_ecs_cluster" "this" {
  name = "${var.name_prefix}-relay-cluster"

  tags = {
    Name        = "${var.name_prefix}-relay-cluster"
    Environment = var.name_prefix
  }
}

# Fargate task definition
resource "aws_ecs_task_definition" "relay" {
  family                   = "${var.name_prefix}-telemetry-relay"
  cpu                      = var.cpu
  memory                   = var.memory
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  execution_role_arn       = var.task_execution_role_arn
  task_role_arn            = var.task_role_arn

  container_definitions = jsonencode([
    {
      name      = "telemetry-relay"
      image     = var.telemetry_relay_image_uri
      essential = true
      environment = [
        { name = "TELEMETRY_WS_URL", value = var.telemetry_ws_url },
        { name = "KAFKA_BROKERS", value = var.kafka_bootstrap_brokers },
        { name = "KAFKA_TOPIC", value = var.kafka_topic },
        { name = "APP_ENV", value = var.app_env },
        { name = "LOG_LEVEL", value = var.log_level },
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = "/ecs/${var.name_prefix}/telemetry-relay"
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "relay"
        }
      }
    }
  ])

  tags = {
    Name        = "${var.name_prefix}-telemetry-relay"
    Environment = var.name_prefix
  }
}

# CloudWatch log group for relay
resource "aws_cloudwatch_log_group" "relay" {
  name              = "/ecs/${var.name_prefix}/telemetry-relay"
  retention_in_days = 30

  tags = {
    Name        = "${var.name_prefix}-relay-log-group"
    Environment = var.name_prefix
  }
}

# Fargate service
resource "aws_ecs_service" "relay" {
  name            = "${var.name_prefix}-telemetry-relay-service"
  cluster         = aws_ecs_cluster.this.id
  task_definition = aws_ecs_task_definition.relay.arn
  desired_count   = var.desired_count
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = var.private_subnet_ids
    security_groups  = [var.fargate_relay_security_group_id]
    assign_public_ip = false
  }

  deployment_circuit_breaker {
    enable   = true
    rollback = true
  }

  depends_on = [
    aws_ecs_task_definition.relay,
    aws_cloudwatch_log_group.relay,
  ]

  tags = {
    Name        = "${var.name_prefix}-telemetry-relay-service"
    Environment = var.name_prefix
  }
}

