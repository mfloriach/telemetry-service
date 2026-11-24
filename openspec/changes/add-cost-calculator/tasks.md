# Tasks

## 1. Rate data and calculation

- [x] 1.1 Create `docs/website/src/components/CostCalculator/rates.ts` with typed constants for unit rates, fixed infrastructure quantities and default inputs, naming the prod `cost.csv` files as the source; verify `npm run typecheck` succeeds.
- [x] 1.2 Implement line-item calculation that rounds each charge to two decimals before summing option totals; verify the baseline totals equal 673.63, 672.63 and 136.11 USD and that A minus B equals 1.00.
- [x] 1.3 Document each fixed assumption (730 hours, three NAT gateways, one relay task, 512 MB Lambda, free Function URL, MSK cluster and partition charges, DynamoDB request-unit rounding) beside its constant; verify every value can be traced to a row in the named CSV without reading Terraform.

## 2. Calculator component

- [x] 2.1 Create `docs/website/src/components/CostCalculator/index.tsx` holding editable values as strings with derived numeric results; verify `npm run typecheck` succeeds.
- [x] 2.2 Add input controls for HTTP and Lambda requests, Lambda duration, ECR storage, relay log ingestion/storage/scanning and DynamoDB storage/write/read units, with non-negative validation and inline feedback; verify a negative or non-numeric value is rejected and the last valid totals remain displayed.
- [x] 2.3 Render the live per-resource line items, two-decimal monthly USD totals for Options A, B and C, the A-minus-C difference and the fixed-assumption list with MSK data shown as non-editable; verify all appear with the documented default values.

## 3. Page integration and styling

- [x] 3.1 Add calculator classes to `docs/website/src/css/custom.css` using `--ifm-*` variables and `[data-theme='dark']` overrides; verify the input panel and table are readable in both light and dark mode using `npm run start`.
- [x] 3.2 Replace the static cost table in `docs/aws-deployment/overview.mdx` with the `@site/src/components/CostCalculator` import and element, keeping the surrounding prose and baseline caveat; verify the published page shows exactly one cost table and no duplicated figures.

## 4. Integration verification

- [x] 4.1 Run `npm run typecheck` and `npm run build`; verify both exit successfully with no new dependencies added.
- [x] 4.2 In the built or dev-served overview page, check the baseline totals, a doubled request count, a changed log ingestion value and an invalid input; verify each observed result matches the corresponding delta spec scenario.
- [x] 4.3 Observe browser network activity while changing an input; verify no request is made and the page does not reload.

## 5. Configurable EC2 and Fargate sizing

- [x] 5.1 Add the EC2 hourly rate table (t3.micro 0.0104, t2.micro 0.0116, t3.small 0.0208, t3.medium 0.0416) and the curated Fargate CPU/memory pairs to `rates.ts`, recording the source and verification date; verify `npm run typecheck` succeeds.
- [x] 5.2 Update `estimate` so EC2 cost is `730 × rate × count` and Fargate cost is `730 × cpu × count × 0.04048 + 730 × memoryGB × count × 0.004445`; verify baseline totals stay 673.63, 672.63 and 136.11 and that changing to t2.micro or two relay tasks moves only the expected line items.
- [x] 5.3 Add EC2 type and count and Fargate CPU/memory and count controls with non-negative integer validation for counts; verify an invalid or non-integer count is rejected with an inline message and the last valid totals remain displayed.
- [x] 5.4 Update the fixed-assumption disclosure and page prose so EC2 type, EC2 count and Fargate sizing are no longer described as fixed, while MSK data remains non-editable; verify the visible list matches the revised delta spec.
- [x] 5.5 Run `npm run typecheck` and `npm run build`, then check baseline, load-change, EC2-change, Fargate-change and invalid-input scenarios in the browser; verify no dynamic network requests are made.

## Workflow follow-up

- Archive the change after implementation and verification are complete and the project's review requirements are satisfied.
- Propose a Conventional Commits message and validate it with commitlint before presenting it; do not create the commit unless explicitly requested.
