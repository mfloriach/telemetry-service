// Unit rates and fixed quantities transcribed from the prod c3x estimates:
// terraform/option-a-lambda-apigw/env/prod/cost.csv
// terraform/option-b-lambda-functionurl/env/prod/cost.csv
// terraform/option-c-ec2-dynamodb/env/prod/cost.csv
// Refresh these constants when those CSV files are regenerated.
// Sizing prices below are AWS EC2 us-east-1 Linux on-demand list prices, verified
// 2026-11-24 from https://instances.vantage.sh/aws/ec2/<type>. The t3.micro rate
// matches the prod cost.csv; Fargate vCPU/memory rates stay the c3x CSV values.

export type OptionId = 'a' | 'b' | 'c';

export type Ec2Type = 't3.micro' | 't2.micro' | 't3.small' | 't3.medium';

export type LoadInputs = {
  requests: number;
  requestDurationMs: number;
  ecrStorageGb: number;
  logIngestedGb: number;
  logStoredGb: number;
  logScannedGb: number;
  dynamoStorageGb: number;
  dynamoWriteMillions: number;
  dynamoReadMillions: number;
};

export type CostRow = {
  label: string;
  amounts: Record<OptionId, number | null>;
};

export type Estimate = {
  rows: CostRow[];
  totals: Record<OptionId, number>;
};

export type FargateProfile = {
  id: string;
  label: string;
  cpu: number;
  memoryGb: number;
};

export type SizingInputs = {
  ec2Type: Ec2Type;
  ec2Count: number;
  fargateProfileId: string;
  fargateTaskCount: number;
};

export const DEFAULT_INPUTS: LoadInputs = {
  requests: 1_000_000,
  requestDurationMs: 200,
  ecrStorageGb: 5,
  logIngestedGb: 5,
  logStoredGb: 5,
  logScannedGb: 0,
  dynamoStorageGb: 50,
  dynamoWriteMillions: 1,
  dynamoReadMillions: 1,
};

export const OPTION_LABELS: Record<OptionId, string> = {
  a: 'Option A — API Gateway',
  b: 'Option B — Function URL',
  c: 'Option C — EC2 + DynamoDB',
};

export const OPTION_IDS: OptionId[] = ['a', 'b', 'c'];

export const EC2_RATES: Record<Ec2Type, number> = {
  't3.micro': 0.0104,
  't2.micro': 0.0116,
  't3.small': 0.0208,
  't3.medium': 0.0416,
};

export const FARGATE_PROFILES: FargateProfile[] = [
  {id: '0.25-512', label: '0.25 vCPU / 512 MB', cpu: 0.25, memoryGb: 0.5},
  {id: '0.25-1024', label: '0.25 vCPU / 1 GB', cpu: 0.25, memoryGb: 1},
  {id: '0.5-1024', label: '0.5 vCPU / 1 GB', cpu: 0.5, memoryGb: 1},
  {id: '0.5-2048', label: '0.5 vCPU / 2 GB', cpu: 0.5, memoryGb: 2},
  {id: '1-2048', label: '1 vCPU / 2 GB', cpu: 1, memoryGb: 2},
  {id: '1-4096', label: '1 vCPU / 4 GB', cpu: 1, memoryGb: 4},
  {id: '2-4096', label: '2 vCPU / 4 GB', cpu: 2, memoryGb: 4},
  {id: '2-8192', label: '2 vCPU / 8 GB', cpu: 2, memoryGb: 8},
];

export const DEFAULT_SIZING: SizingInputs = {
  ec2Type: 't3.micro',
  ec2Count: 1,
  fargateProfileId: '0.25-512',
  fargateTaskCount: 1,
};

export const HOURS_PER_MONTH = 730; // cost.csv: 730 hours for MSK, NAT, EIP and EC2.
export const NAT_GATEWAY_COUNT = 3; // cost.csv: aws_nat_gateway.this[0..2] and aws_eip.nat[0..2].
export const LAMBDA_MEMORY_GB = 0.5; // cost.csv: 512 MB default; 100000 GB-s at 1M requests of 200 ms.
export const MSK_DATA_GB = 50; // c3x-usage.yml: declared but not priced; MSK is billed by cluster/partition hours.

export const RATES = {
  httpApiPerRequest: 0.000001, // cost.csv: HTTP API requests.
  lambdaPerRequest: 0.0000002, // cost.csv: Lambda Requests.
  lambdaPerGbSecond: 0.0000166667, // cost.csv: Lambda Duration.
  ecrPerGbMonth: 0.1, // cost.csv: ECR Image storage, each repository.
  logIngestPerGb: 0.5, // cost.csv: CloudWatch Data ingested.
  logStorePerGbMonth: 0.03, // cost.csv: CloudWatch Data stored (archive).
  logScanPerGb: 0.005, // cost.csv: CloudWatch Logs Insights queries.
  mskClusterPerHour: 0.75, // cost.csv: MSK Cluster, 730 hours.
  mskPartitionPerHour: 0.0015, // cost.csv: MSK Partitions, 730 partition-hours.
  fargateVcpuPerHour: 0.04048, // cost.csv: Fargate vCPU, 182.5 vCPU-hours.
  fargateMemoryPerGbHour: 0.004445, // cost.csv: Fargate memory, 365 GB-hours.
  natGatewayPerHour: 0.045, // cost.csv: NAT Gateway, 730 hours each.
  eipPerHour: 0.005, // cost.csv: EIP public IPv4 address-hour, 730 hours each.
  dynamoStoragePerGbMonth: 0.25, // cost.csv: DynamoDB Data storage, 50 GB-month.
  dynamoWritePerMillion: 0.000000625, // cost.csv: Write request units, quantity 1 at 1M requests; rounds to 0.00.
  dynamoReadPerMillion: 0.000000125, // cost.csv: Read request units, quantity 1 at 1M requests; rounds to 0.00.
} as const;

export const FIXED_ASSUMPTIONS = [
  `730 hours per month and ${NAT_GATEWAY_COUNT} NAT gateways with EIPs`,
  '512 MB Lambda memory and HTTP API requests at USD 0.000001 each',
  'MSK Serverless priced by cluster and partition hours, not by data volume',
  'Lambda Function URL is a free resource',
  'DynamoDB write and read request units round to 0.00 at the baseline scale',
] as const;

function round(value: number): number {
  return Math.round((value + Number.EPSILON) * 100) / 100;
}

export function estimate(inputs: LoadInputs, sizing: SizingInputs = DEFAULT_SIZING): Estimate {
  const httpApi = round(inputs.requests * RATES.httpApiPerRequest);
  const lambdaRequests = round(inputs.requests * RATES.lambdaPerRequest);
  const durationSeconds = inputs.requestDurationMs / 1000;
  const lambdaDuration = round(
    inputs.requests * durationSeconds * LAMBDA_MEMORY_GB * RATES.lambdaPerGbSecond,
  );
  const lambda = round(lambdaRequests + lambdaDuration);
  const ecr = round(
    round(inputs.ecrStorageGb * RATES.ecrPerGbMonth) * 2,
  );
  const logs = round(
    round(inputs.logIngestedGb * RATES.logIngestPerGb) +
      round(inputs.logStoredGb * RATES.logStorePerGbMonth) +
      round(inputs.logScannedGb * RATES.logScanPerGb),
  );
  const nat = round(NAT_GATEWAY_COUNT * round(HOURS_PER_MONTH * RATES.natGatewayPerHour));
  const eip = round(NAT_GATEWAY_COUNT * round(HOURS_PER_MONTH * RATES.eipPerHour));
  const vpc = round(nat + eip);
  const mskCluster = round(HOURS_PER_MONTH * RATES.mskClusterPerHour);
  const mskPartitions = round(HOURS_PER_MONTH * RATES.mskPartitionPerHour);
  const msk = round(mskCluster + mskPartitions);
  const profile = FARGATE_PROFILES.find(item => item.id === sizing.fargateProfileId) ?? FARGATE_PROFILES[0];
  const fargateVcpu = round(
    HOURS_PER_MONTH * profile.cpu * sizing.fargateTaskCount * RATES.fargateVcpuPerHour,
  );
  const fargateMemory = round(
    HOURS_PER_MONTH * profile.memoryGb * sizing.fargateTaskCount * RATES.fargateMemoryPerGbHour,
  );
  const fargate = round(fargateVcpu + fargateMemory);
  const ec2 = round(HOURS_PER_MONTH * EC2_RATES[sizing.ec2Type] * sizing.ec2Count);
  const dynamoStorage = round(inputs.dynamoStorageGb * RATES.dynamoStoragePerGbMonth);
  const dynamoWrites = round(inputs.dynamoWriteMillions * RATES.dynamoWritePerMillion);
  const dynamoReads = round(inputs.dynamoReadMillions * RATES.dynamoReadPerMillion);
  const dynamodb = round(dynamoStorage + dynamoWrites + dynamoReads);

  const rows: CostRow[] = [
    {label: 'MSK Serverless (cluster + partitions)', amounts: {a: msk, b: msk, c: null}},
    {label: 'DynamoDB (storage + request units)', amounts: {a: null, b: null, c: dynamodb}},
    {label: `EC2 relay (${sizing.ec2Type})`, amounts: {a: null, b: null, c: ec2}},
    {label: 'Fargate relay (vCPU + memory)', amounts: {a: fargate, b: fargate, c: null}},
    {label: 'VPC (3 NAT gateways + EIPs)', amounts: {a: vpc, b: vpc, c: vpc}},
    {label: 'CloudWatch Logs (ingest + storage)', amounts: {a: logs, b: logs, c: logs}},
    {label: 'ECR (two repositories)', amounts: {a: ecr, b: ecr, c: ecr}},
    {label: 'Lambda (requests + duration)', amounts: {a: lambda, b: lambda, c: lambda}},
    {label: 'HTTP API requests', amounts: {a: httpApi, b: null, c: httpApi}},
    {label: 'Function URL', amounts: {a: null, b: 0, c: null}},
  ];

  const totals: Record<OptionId, number> = {a: 0, b: 0, c: 0};
  for (const row of rows) {
    for (const option of OPTION_IDS) {
      const amount = row.amounts[option];
      if (amount === null) {
        continue;
      }
      totals[option] = round(totals[option] + amount);
    }
  }

  return {rows, totals};
}
