# MSK Serverless module: cluster with IAM authentication, VPC placement, security group.

resource "aws_msk_serverless_cluster" "this" {
  cluster_name = "${var.name_prefix}-msk-serverless"

  vpc_config {
    subnet_ids         = var.private_subnet_ids
    security_group_ids = [var.msk_security_group_id]
  }

  client_authentication {
    sasl {
      iam {
        enabled = true
      }
    }
  }

  tags = {
    Name        = "${var.name_prefix}-msk-serverless"
    Environment = var.name_prefix
  }
}

