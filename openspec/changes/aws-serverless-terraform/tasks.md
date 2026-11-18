# Tasks

## 1. Terraform project scaffolding

- [x] 1.1 Create directory structure: `terraform/option-a-lambda-apigw/`, `terraform/option-b-lambda-functionurl/`, `terraform/modules/`, `terraform/env/dev/`, `terraform/env/stage/`, `terraform/env/prod/` — verify: all expected directories exist
- [x] 1.2 Create shared `terraform/versions.tf` and `terraform/providers.tf` (AWS provider pinned) — verify: `terraform providers lock -platform=linux_amd64` succeeds
- [x] 1.3 Create shared backend configuration template (S3 bucket + DynamoDB locking table) per environment — verify: backend block parses with `terraform init` on each env

## 2. Shared infrastructure modules

- [x] 2.1 VPC module: VPC, public + private subnets (3 AZs), internet gateway, NAT gateways — verify: `terraform validate` passes in module test harness
- [x] 2.2 Security groups module: Lambda↔MSK, Fargate relay↔MSK, Fargate outbound to upstream WebSocket — verify: `terraform validate` passes
- [x] 2.3 IAM module: Lambda execution role, Fargate task/execution roles, MSK IAM auth policies (cluster connect, describe) — verify: `terraform validate` passes
- [x] 2.4 MSK Serverless module: cluster with IAM authentication, VPC placement, security group — verify: `terraform validate` + `terraform plan` lists expected MSK resources
- [x] 2.5 ECR module: two repositories (`telemetry-server`, `telemetry-relay`) — verify: `terraform validate` passes

## 3. Option A — Lambda + API Gateway HTTP API

- [x] 3.1 Lambda function resource (telemetry-server container image from ECR, env vars `KAFKA_BROKERS`, `KAFKA_TOPIC`, `HTTP_ADDR`, `APP_ENV`, `LOG_LEVEL`, VPC config) — verify: `terraform validate` passes
- [x] 3.2 API Gateway HTTP API, Lambda integration, and route for `GET /telemetry/latest` — verify: `terraform validate` + `terraform plan` shows API Gateway + Lambda resources

## 4. Option B — Lambda + Lambda Function URL

- [x] 4.1 Lambda function resource (telemetry-server, same image/env/VPC as option A) — verify: `terraform validate` passes
- [x] 4.2 Lambda Function URL resource (auth type `AWS_IAM`, CORS optional) — verify: `terraform validate` + `terraform plan` shows Function URL resource

## 5. Fargate relay service

- [x] 5.1 ECS cluster resource — verify: `terraform validate` passes
- [x] 5.2 Fargate task definition (telemetry-relay container image, env vars `TELEMETRY_WS_URL`, `KAFKA_BROKERS`, `KAFKA_TOPIC`, `APP_ENV`, `LOG_LEVEL`) — verify: `terraform validate` passes
- [x] 5.3 Fargate Service (`desired_count = 1`, circuit breaker, propagate tags) — verify: `terraform validate` + `terraform plan` shows ECS service

## 6. Environments (dev / stage / prod)

- [x] 6.1 dev environment: backend config + `terraform.tfvars` (smaller instance sizes, MSK standard throughput) — verify: `terraform init` + `terraform plan` succeeds for both options
- [x] 6.2 stage environment: backend config + `terraform.tfvars` — verify: `terraform init` + `terraform plan` succeeds for both options
- [x] 6.3 prod environment: backend config + `terraform.tfvars` — verify: `terraform init` + `terraform plan` succeeds for both options

## 7. Container images

- [x] 7.1 Add relay build target to Dockerfile (or separate relay Dockerfile) — verify: `docker build --target relay -t telemetry-relay .` succeeds
- [x] 7.2 Build and push `telemetry-server` image to ECR — verify: image present in `telemetry-server` ECR repository
- [x] 7.3 Build and push `telemetry-relay` image to ECR — verify: image present in `telemetry-relay` ECR repository

## 8. Validation

- [x] 8.1 Run `terraform validate` and `terraform plan` for dev/stage/prod across both option folders — verify: all plans produce expected resource counts with no errors
- [x] 8.2 Document integration check steps: confirm Lambda can reach MSK brokers, Fargate relay can reach upstream WebSocket and MSK — verify: runbook written to `terraform/VALIDATION.md`

## Workflow follow-up

- After all tasks complete, review the Terraform plan output before any apply.
- Archive the change with `/opsx-archive` once the project's review requirements are satisfied.
