bucket         = "telemetry-terraform-state-prod"
key            = "aws-serverless-terraform/option-a/prod/terraform.tfstate"
region         = "us-east-1"
dynamodb_table = "telemetry-terraform-locks"
encrypt        = true
