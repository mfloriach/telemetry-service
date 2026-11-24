# Spec Delta

## Purpose

Lets readers of the AWS deployment documentation explore how monthly cost for each option changes when expected load, EC2 instance sizing or Fargate relay sizing changes, without running the cost estimation toolchain.

## ADDED Requirements

### Requirement: Load and sizing inputs are editable in the browser
The calculator SHALL present load and instance-sizing inputs as values a reader can change on the published page and SHALL recalculate without reloading the page or contacting a server. Inputs SHALL cover HTTP and Lambda requests, Lambda duration, ECR storage, relay log ingestion/storage/scanning, DynamoDB storage and request units, EC2 instance type and count, and Fargate task CPU, memory and count. The MSK data volume SHALL be shown as a fixed assumption rather than an editable input.

#### Scenario: Reader changes a load value
- **WHEN** a reader changes the number of monthly requests in the calculator
- **THEN** the calculator shows recalculated costs without navigating away or making a server request

#### Scenario: Reader changes EC2 sizing
- **WHEN** a reader selects a different EC2 instance type or changes the instance count
- **THEN** the EC2 relay charge and Option C total change while load-driven charges keep their values

#### Scenario: Reader changes Fargate sizing
- **WHEN** a reader changes the Fargate task CPU, memory or task count
- **THEN** the Fargate relay charge and the Options A and B totals change

#### Scenario: MSK data volume is not editable
- **WHEN** a reader views the broker/store data assumption
- **THEN** the MSK volume is presented as a fixed figure and the reader cannot change it

### Requirement: Published baseline is reproduced exactly
When all editable inputs hold their documented default values, the calculator SHALL display monthly totals of 673.63 USD for Option A, 672.63 USD for Option B and 136.11 USD for Option C.

#### Scenario: Initial display matches the documentation
- **WHEN** the calculator is first displayed with default inputs
- **THEN** it shows the same per-resource line items and totals as the published cost table

### Requirement: Recalculation updates affected line items and totals
The calculator SHALL recompute per-resource charges and each option total from the current inputs and SHALL leave charges that do not depend on those inputs unchanged.

#### Scenario: Request volume changes
- **WHEN** the reader increases the monthly request count
- **THEN** the HTTP API, Lambda request and DynamoDB request charges and the affected option totals change

#### Scenario: Log ingestion volume changes
- **WHEN** the reader changes the relay log ingestion value
- **THEN** the CloudWatch Logs charge changes for every option while unrelated line items keep their values

### Requirement: Fixed pricing assumptions are disclosed and constant
The calculator SHALL hold the remaining infrastructure assumptions constant and SHALL make them visible to the reader. These SHALL include 730 hours per month, three NAT gateways, 512 MB Lambda memory, a free Function URL, the HTTP API per-request rate, MSK cluster and partition hour pricing, and DynamoDB request-unit charges of 0.00 at the baseline scale.

#### Scenario: Reader reviews assumptions
- **WHEN** a reader opens the calculator
- **THEN** the fixed assumptions are stated alongside the inputs

#### Scenario: No control for other fixed infrastructure
- **WHEN** a reader looks for controls for hours per month, NAT gateway count or Lambda memory
- **THEN** no such controls are available

### Requirement: Costs are displayed consistently
The calculator SHALL display each line item and total as a monthly USD amount with two decimal places and SHALL show the difference between the Option A and Option C totals.

#### Scenario: Formatted totals
- **WHEN** the reader changes any input
- **THEN** all displayed amounts remain monthly USD figures with two decimal places and the A-minus-C difference reflects the new totals

### Requirement: Invalid input does not corrupt results
The calculator SHALL reject non-numeric, negative or non-integer input values and SHALL preserve the last valid calculation while telling the reader the value was not accepted.

#### Scenario: Negative value entered
- **WHEN** a reader enters a negative number for an input
- **THEN** the value is rejected, an explanation is shown and the previously computed totals remain displayed

#### Scenario: Invalid instance count entered
- **WHEN** a reader enters a non-integer instance or task count
- **THEN** the value is rejected, an explanation is shown and the previously computed totals remain displayed

### Requirement: Calculator replaces the static table on the overview page
The published AWS deployment overview SHALL show the calculator where the static expected-load cost table was and SHALL retain the surrounding explanation that the printed figures are a planning baseline.

#### Scenario: Reader visits the overview page
- **WHEN** a reader opens the AWS deployment overview
- **THEN** the expected-load section provides editable inputs and a live cost table in place of the static table, with the baseline caveat still present
