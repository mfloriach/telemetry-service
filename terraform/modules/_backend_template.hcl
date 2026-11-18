# Backend configuration for remote state with DynamoDB locking.
#
# Usage per environment:
#   terraform init -backend-config=env/<ENV>/backend.hcl
#
# Before first use, create the S3 bucket and DynamoDB table (e.g. via the AWS CLI
# or a one-off Terraform run with a local backend):
#
#   aws s3api create-bucket --bucket <bucket> --region <region>
#   aws dynamodb create-table \
#     --table-name <table> \
#     --attribute-definitions AttributeName=LockID,AttributeType=S \
#     --key-schema AttributeName=LockID,KeyType=HASH \
#     --billing-mode PAY_PER_REQUEST

bucket         = "telemetry-terraform-state-<ENV>"
key            = "aws-serverless-terraform/<OPTION>/<ENV>/terraform.tfstate"
region         = "us-east-1"
dynamodb_table = "telemetry-terraform-locks"
encrypt        = true
