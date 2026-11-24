data "aws_ami" "amazon_linux" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-2023.*-x86_64"]
  }
}

resource "aws_security_group" "relay" {
  name        = "${var.name_prefix}-ec2-relay-sg"
  description = "Security group for the EC2 telemetry relay"
  vpc_id      = var.vpc_id

  ingress {
    description = "Device WebSocket connections"
    from_port   = var.ws_port
    to_port     = var.ws_port
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    description = "Allow all outbound"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${var.name_prefix}-ec2-relay-sg"
    Environment = var.name_prefix
  }
}

resource "aws_cloudwatch_log_group" "relay" {
  name              = "/ec2/${var.name_prefix}/telemetry-relay"
  retention_in_days = 30

  tags = {
    Name        = "${var.name_prefix}-relay-log-group"
    Environment = var.name_prefix
  }
}

resource "aws_instance" "relay" {
  ami                         = data.aws_ami.amazon_linux.id
  instance_type               = var.instance_type
  subnet_id                   = var.subnet_id
  vpc_security_group_ids      = [aws_security_group.relay.id]
  iam_instance_profile        = var.iam_instance_profile_name
  associate_public_ip_address = true

  user_data_base64 = base64encode(<<-EOT
    #!/bin/bash
    set -euo pipefail
    dnf install -y docker
    systemctl enable --now docker
    aws ecr get-login-password --region ${var.aws_region} | docker login --username AWS --password-stdin ${split("/", var.telemetry_relay_image_uri)[0]}
    docker pull ${var.telemetry_relay_image_uri}
    docker run -d --restart unless-stopped --name telemetry-relay \
      -p ${var.ws_port}:${var.ws_port} \
      -e WS_ADDR=:${var.ws_port} \
      -e DYNAMODB_TABLE=${var.dynamodb_table_name} \
      -e AWS_REGION=${var.aws_region} \
      -e APP_ENV=${var.app_env} \
      -e LOG_LEVEL=${var.log_level} \
      --log-driver awslogs \
      --log-opt awslogs-group=${aws_cloudwatch_log_group.relay.name} \
      --log-opt awslogs-region=${var.aws_region} \
      --log-opt awslogs-stream-prefix=relay \
      ${var.telemetry_relay_image_uri}
  EOT
  )

  tags = {
    Name        = "${var.name_prefix}-ec2-relay"
    Environment = var.name_prefix
  }
}
