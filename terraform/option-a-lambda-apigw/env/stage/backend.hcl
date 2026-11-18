bucket         = "telemetry-terraform-state-stage"
key            = "aws-serverless-terraform/option-a/stage/terraform.tfstate"
region         = "us-east-1"
dynamodb_table = "telemetry-terraform-locks"
encrypt        = true
