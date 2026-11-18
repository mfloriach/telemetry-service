# Validation Runbook

This document describes how to validate the Terraform deployment for both options across dev/stage/prod environments.

## Prerequisites

- Terraform >= 1.8.0 installed
- AWS credentials configured (`aws configure` or environment variables)
- Docker installed (for building container images)
- S3 bucket and DynamoDB table created for remote state (see `modules/_backend_template.hcl`)

## Setup

1. Create the state bucket and DynamoDB lock table (one-time):

   ```bash
   aws s3api create-bucket --bucket telemetry-terraform-state-dev --region us-east-1
   aws s3api create-bucket --bucket telemetry-terraform-state-stage --region us-east-1
   aws s3api create-bucket --bucket telemetry-terraform-state-prod --region us-east-1

   aws dynamodb create-table \
     --table-name telemetry-terraform-locks \
     --attribute-definitions AttributeName=LockID,AttributeType=S \
     --key-schema AttributeName=LockID,KeyType=HASH \
     --billing-mode PAY_PER_REQUEST
   ```

2. Build and push container images to ECR:

   ```bash
   # Build telemetry-server
   docker build -t <account-id>.dkr.ecr.<region>.amazonaws.com/telemetry-telemetry-server:latest .
   docker push <account-id>.dkr.ecr.<region>.amazonaws.com/telemetry-telemetry-server:latest

   # Build telemetry-relay (relay target)
   docker build --target relay-runtime -t <account-id>.dkr.ecr.<region>.amazonaws.com/telemetry-telemetry-relay:latest .
   docker push <account-id>.dkr.ecr.<region>.amazonaws.com/telemetry-telemetry-relay:latest
   ```

## Option A: Lambda + API Gateway HTTP API

### Initialize and validate (per environment)

```bash
cd terraform/option-a-lambda-apigw/env/dev
terraform init -backend-config=backend.hcl
terraform validate
terraform plan -var-file=terraform.tfvars
```

Repeat for `stage` and `prod` environments.

### Deploy

```bash
terraform apply -var-file=terraform.tfvars
```

### Verify

1. Get the API Gateway endpoint from Terraform outputs:

   ```bash
   terraform output api_gateway_endpoint
   ```

2. Test the telemetry endpoint:

   ```bash
   curl "https://<api-gateway-id>.execute-api.us-east-1.amazonaws.com/dev/telemetry/latest?n=10"
   ```

3. Expected response: JSON array of telemetry records.

## Option B: Lambda + Lambda Function URL

### Initialize and validate (per environment)

```bash
cd terraform/option-b-lambda-functionurl/env/dev
terraform init -backend-config=backend.hcl
terraform validate
terraform plan -var-file=terraform.tfvars
```

### Deploy

```bash
terraform apply -var-file=terraform.tfvars
```

### Verify

1. Get the Function URL from Terraform outputs:

   ```bash
   terraform output function_url
   ```

2. Test the telemetry endpoint (requires AWS SigV4 signing for IAM auth).

## Relay (Fargate) Verification

Both options deploy the same Fargate relay service. Verify:

1. Check ECS service is running:

   ```bash
   aws ecs describe-services --cluster <cluster-name> --services <service-name>
   ```

2. Check CloudWatch logs for relay:

   ```bash
   aws logs describe-log-streams --log-group-name /ecs/<env>/telemetry-relay
   ```

## Integration Checks

1. Lambda → MSK: Confirm Lambda can reach MSK brokers (check VPC security groups allow Lambda SG → MSK SG on port 9098).
2. Fargate → Upstream WebSocket: Confirm relay can reach the upstream WebSocket URL.
3. Fargate → MSK: Confirm relay can publish to MSK (check Fargate task role has kafka-cluster permissions).
4. End-to-end: Publish a test message via the upstream WebSocket, confirm it appears in `GET /telemetry/latest`.

## Rollback

To rollback an environment:

```bash
terraform apply -var-file=terraform.tfvars -target=aws_lambda_function.telemetry_server
```

Or destroy and re-apply:

```bash
terraform destroy -var-file=terraform.tfvars
terraform apply -var-file=terraform.tfvars
```

## Known Limitations

- Terraform CLI and AWS credentials are required for `terraform plan`/`terraform apply` — not available in this environment.
- MSK Serverless requires IAM auth; ensure Kafka clients are configured for SASL/IAM.
- Lambda container images require the Lambda Web Adapter or LRIC for HTTP handler compatibility (not yet added to Dockerfile).
