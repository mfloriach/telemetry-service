# Proposal

## Why

The AWS deployment overview quotes monthly costs for three topologies, but the figures are static: readers cannot see how load, EC2 sizing or relay sizing changes the answer. The inputs already exist in `c3x-usage.yml` and the rates in the prod `cost.csv` files, so the site can answer "what if" questions in the browser without re-running c3x.

## What Changes

- Add a React `CostCalculator` component under `docs/website/src/components/`, imported into `docs/aws-deployment/overview.mdx` via the existing `@site/` pattern.
- Replace the static "Expected load and estimated cost" table with an editable input panel and a live per-resource cost table for Options A, B and C.
- Expose load inputs (requests, Lambda duration, ECR storage, relay logs, DynamoDB storage and request units) plus sizing inputs: EC2 instance type and count, Fargate task CPU, memory and count. Show MSK data as a fixed assumption because c3x prices MSK by cluster and partition hours.
- Keep one typed TypeScript data module: fixed rates from the prod `cost.csv` files and a small table of us-east-1 on-demand EC2 hourly rates (t3.micro 0.0104, t2.micro 0.0116, t3.small 0.0208, t3.medium 0.0416). No runtime CSV/YAML import and no new npm dependencies.
- Compute two-decimal USD totals that reproduce the documented baseline (673.63 / 672.63 / 136.11), scale linearly with load and sizing, and preserve documented quirks: free Function URL, HTTP API at 0.000001/request, DynamoDB request units rounded to 0.00.
- Add styles to `src/css/custom.css` using `--ifm-*` variables and dark-mode overrides. Verify with `npm run typecheck`, `npm run build` and browser checks.

No breaking changes.

## Capabilities

### New Capabilities
- `documentation-cost-calculator`: Interactive estimation of AWS deployment option costs in the documentation site from reader-editable expected-load and instance-sizing inputs, using fixed rates derived from c3x cost CSVs and published on-demand prices.

### Modified Capabilities

None.

## Impact

- New `docs/website/src/components/CostCalculator/` holding the component and its typed rate data; `docs/website/src/css/custom.css` gains calculator styles.
- `docs/aws-deployment/overview.mdx`: the static cost table is replaced by the live component; surrounding prose and the baseline caveat remain.
- Existing dependencies only; no Go, Kafka, relay or Terraform changes. The Terraform defaults stay at t3.micro and one 0.25 vCPU / 512 MB Fargate task.
- `make tf-estimate`, `c3x-usage.yml` and `cost.csv` remain authoritative for the baseline; the calculator does not invoke c3x, and its rate module must be refreshed when prices change.
- ADR 0010 (Docusaurus for docs) remains valid and unchanged. No new ADR is needed: rate ownership is a website-local data decision, not a tooling or runtime boundary.
