import React, {useState} from 'react';
import {
  DEFAULT_INPUTS,
  DEFAULT_SIZING,
  EC2_RATES,
  FARGATE_PROFILES,
  FIXED_ASSUMPTIONS,
  MSK_DATA_GB,
  OPTION_IDS,
  OPTION_LABELS,
  estimate,
  type Ec2Type,
  type LoadInputs,
  type SizingInputs,
} from './rates';

type InputKey = keyof LoadInputs;
type CountKey = 'ec2Count' | 'fargateTaskCount';

type Field = {
  key: InputKey;
  label: string;
};

const FIELDS: Field[] = [
  {key: 'requests', label: 'GET /telemetry/latest requests per month'},
  {key: 'requestDurationMs', label: 'Lambda request duration (ms)'},
  {key: 'ecrStorageGb', label: 'ECR image storage per repository (GB)'},
  {key: 'logIngestedGb', label: 'Relay logs ingested (GB)'},
  {key: 'logStoredGb', label: 'Relay logs stored (GB)'},
  {key: 'logScannedGb', label: 'Relay logs scanned (GB)'},
  {key: 'dynamoStorageGb', label: 'DynamoDB data stored (GB)'},
  {key: 'dynamoWriteMillions', label: 'DynamoDB writes (millions)'},
  {key: 'dynamoReadMillions', label: 'DynamoDB reads (millions)'},
];

function initialText(): Record<InputKey, string> {
  const result = {} as Record<InputKey, string>;
  for (const field of FIELDS) {
    result[field.key] = String(DEFAULT_INPUTS[field.key]);
  }
  return result;
}

function isEc2Type(value: string): value is Ec2Type {
  return Object.prototype.hasOwnProperty.call(EC2_RATES, value);
}

function usd(value: number): string {
  return value.toFixed(2);
}

function amount(value: number | null): string {
  return value === null ? '—' : usd(value);
}

export default function CostCalculator(): React.JSX.Element {
  const [text, setText] = useState<Record<InputKey, string>>(initialText);
  const [valid, setValid] = useState<LoadInputs>(DEFAULT_INPUTS);
  const [errors, setErrors] = useState<Partial<Record<InputKey, string>>>({});
  const [sizing, setSizing] = useState<SizingInputs>(DEFAULT_SIZING);
  const [countText, setCountText] = useState<Record<CountKey, string>>({
    ec2Count: String(DEFAULT_SIZING.ec2Count),
    fargateTaskCount: String(DEFAULT_SIZING.fargateTaskCount),
  });
  const [countErrors, setCountErrors] = useState<Partial<Record<CountKey, string>>>({});

  const {rows, totals} = estimate(valid, sizing);

  function handleChange(key: InputKey, value: string): void {
    setText(previous => ({...previous, [key]: value}));

    const parsed = Number(value);
    if (value.trim() === '' || !Number.isFinite(parsed) || parsed < 0) {
      setErrors(previous => ({...previous, [key]: 'Enter a non-negative number.'}));
      return;
    }

    setErrors(previous => ({...previous, [key]: undefined}));
    setValid(previous => ({...previous, [key]: parsed}));
  }

  function handleCountChange(key: CountKey, value: string): void {
    setCountText(previous => ({...previous, [key]: value}));

    const parsed = Number(value);
    if (value.trim() === '' || !Number.isInteger(parsed) || parsed < 0) {
      setCountErrors(previous => ({...previous, [key]: 'Enter a non-negative whole number.'}));
      return;
    }

    setCountErrors(previous => ({...previous, [key]: undefined}));
    setSizing(previous => ({...previous, [key]: parsed}));
  }

  return (
    <section className="cost-calculator">
      <p className="cost-calculator__intro">
        Change the expected load and instance sizing below to recalculate the monthly estimates in your
        browser. No request is sent and the page does not reload.
      </p>

      <fieldset className="cost-calculator__inputs">
        <legend>Expected load</legend>
        <div className="cost-calculator__fields">
          {FIELDS.map(field => (
            <div className="cost-calculator__field" key={field.key}>
              <label htmlFor={`cost-${field.key}`}>{field.label}</label>
              <input
                id={`cost-${field.key}`}
                type="text"
                inputMode="decimal"
                value={text[field.key]}
                aria-invalid={errors[field.key] !== undefined}
                aria-describedby={errors[field.key] ? `cost-${field.key}-error` : undefined}
                onChange={event => handleChange(field.key, event.target.value)}
              />
              {errors[field.key] && (
                <span className="cost-calculator__error" id={`cost-${field.key}-error`} role="alert">
                  {errors[field.key]}
                </span>
              )}
            </div>
          ))}
        </div>
      </fieldset>

      <fieldset className="cost-calculator__inputs">
        <legend>Instance sizing</legend>
        <div className="cost-calculator__fields">
          <div className="cost-calculator__field">
            <label htmlFor="cost-ec2Type">EC2 instance type</label>
            <select
              id="cost-ec2Type"
              value={sizing.ec2Type}
              onChange={event => {
                const value = event.target.value;
                if (isEc2Type(value)) {
                  setSizing(previous => ({...previous, ec2Type: value}));
                }
              }}
            >
              {Object.entries(EC2_RATES).map(([type, rate]) => (
                <option key={type} value={type}>
                  {type} — USD {rate.toFixed(4)}/hour
                </option>
              ))}
            </select>
          </div>

          <div className="cost-calculator__field">
            <label htmlFor="cost-ec2Count">EC2 instance count</label>
            <input
              id="cost-ec2Count"
              type="text"
              inputMode="numeric"
              value={countText.ec2Count}
              aria-invalid={countErrors.ec2Count !== undefined}
              aria-describedby={countErrors.ec2Count ? 'cost-ec2Count-error' : undefined}
              onChange={event => handleCountChange('ec2Count', event.target.value)}
            />
            {countErrors.ec2Count && (
              <span className="cost-calculator__error" id="cost-ec2Count-error" role="alert">
                {countErrors.ec2Count}
              </span>
            )}
          </div>

          <div className="cost-calculator__field">
            <label htmlFor="cost-fargateProfileId">Fargate task size</label>
            <select
              id="cost-fargateProfileId"
              value={sizing.fargateProfileId}
              onChange={event => setSizing(previous => ({...previous, fargateProfileId: event.target.value}))}
            >
              {FARGATE_PROFILES.map(profile => (
                <option key={profile.id} value={profile.id}>
                  {profile.label}
                </option>
              ))}
            </select>
          </div>

          <div className="cost-calculator__field">
            <label htmlFor="cost-fargateTaskCount">Fargate relay task count</label>
            <input
              id="cost-fargateTaskCount"
              type="text"
              inputMode="numeric"
              value={countText.fargateTaskCount}
              aria-invalid={countErrors.fargateTaskCount !== undefined}
              aria-describedby={countErrors.fargateTaskCount ? 'cost-fargateTaskCount-error' : undefined}
              onChange={event => handleCountChange('fargateTaskCount', event.target.value)}
            />
            {countErrors.fargateTaskCount && (
              <span className="cost-calculator__error" id="cost-fargateTaskCount-error" role="alert">
                {countErrors.fargateTaskCount}
              </span>
            )}
          </div>
        </div>
      </fieldset>

      <div className="cost-calculator__fixed">
        <p>Fixed assumptions</p>
        <ul>
          <li>
            MSK data: {MSK_DATA_GB} GB ingested, stored and scanned. This is not editable because c3x prices
            MSK by cluster and partition hours, so the data volume does not change the total.
          </li>
          {FIXED_ASSUMPTIONS.map(assumption => (
            <li key={assumption}>{assumption}</li>
          ))}
        </ul>
      </div>

      <div className="cost-calculator__table-wrapper">
        <table className="cost-calculator__table">
          <thead>
            <tr>
              <th scope="col">Monthly cost (USD)</th>
              {OPTION_IDS.map(option => (
                <th scope="col" key={option}>
                  {OPTION_LABELS[option]}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map(row => (
              <tr key={row.label}>
                <th scope="row">{row.label}</th>
                {OPTION_IDS.map(option => (
                  <td key={option}>{amount(row.amounts[option])}</td>
                ))}
              </tr>
            ))}
            <tr className="cost-calculator__total">
              <th scope="row">Total</th>
              {OPTION_IDS.map(option => (
                <td key={option} aria-live="polite">
                  {usd(totals[option])}
                </td>
              ))}
            </tr>
          </tbody>
        </table>
      </div>

      <p className="cost-calculator__difference">
        Option C is <strong>USD {usd(totals.a - totals.c)}/month cheaper</strong> than Option A at this load.
      </p>
    </section>
  );
}
