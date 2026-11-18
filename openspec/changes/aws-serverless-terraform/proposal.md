# Proposal

## Why

The telemetry system (relay + server binaries) currently runs locally with Docker Compose Kafka. There is no AWS deployment path. Deploying to AWS serverless enables production-grade, multi-environment operation (dev/stage/prod) with managed Kafka and serverless compute.

## What Changes

- Add Terraform under `terraform/` in two option folders:
  - `option-a-lambda-apigw`: telemetry-server on Lambda behind API Gateway HTTP API; relay on Fargate
  - `option-b-lambda-functionurl`: telemetry-server on Lambda behind Lambda Function URL; relay on Fargate
- Each option supports dev/stage/prod environments with isolated state
- Shared modules: VPC, MSK Serverless (Kafka), ECR (container images), IAM, security groups
- Relay runs as a Fargate Service (long-running WebSocket→Kafka publisher; not Lambda-suitable)
- telemetry-server runs as Lambda (container image, HTTP API on port 8080)
- No Go code changes; infra-only

## Capabilities

- New Capabilities: none — infrastructure provisioning, no new system behavior
- Modified Capabilities: none — existing relay and server capability requirements unchanged

This change is infrastructure/tooling only. Per OpenSpec rules, `skip_specs: true` applies since no spec-level behavior changes.

## Impact

- New `terraform/` directory (not part of Go module)
- `cmd/relay` needs a separate Docker build target (current Dockerfile only builds `telemetry-server`)
- MSK Serverless requires TLS + IAM/SASL auth; current Kafka clients use PLAINTEXT — config/auth gap to resolve before apply
- No changes to `cmd/telemetry-server`, `internal/*`, or existing tests (consistent with ADR 0006 boundary)
