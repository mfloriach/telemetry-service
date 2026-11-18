# ECR module: two repositories (telemetry-server, telemetry-relay).

resource "aws_ecr_repository" "telemetry_server" {
  name                 = "${var.name_prefix}-telemetry-server"
  image_tag_mutability = "MUTABLE"
  image_scanning_configuration {
    scan_on_push = true
  }

  tags = {
    Name        = "${var.name_prefix}-telemetry-server"
    Environment = var.name_prefix
  }
}

resource "aws_ecr_repository" "telemetry_relay" {
  name                 = "${var.name_prefix}-telemetry-relay"
  image_tag_mutability = "MUTABLE"
  image_scanning_configuration {
    scan_on_push = true
  }

  tags = {
    Name        = "${var.name_prefix}-telemetry-relay"
    Environment = var.name_prefix
  }
}

