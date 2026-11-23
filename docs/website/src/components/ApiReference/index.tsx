import React from 'react';
import {load} from 'js-yaml';
import openapiYaml from '../../../../openapi.yaml';

type JsonValue = string | number | boolean | null | JsonValue[] | {[key: string]: JsonValue};

type Ref = {$ref: string};

type Schema = {
  type?: string;
  description?: string;
  items?: Schema | Ref;
  properties?: {[name: string]: Schema | Ref};
  required?: string[];
  additionalProperties?: boolean | Schema;
  example?: JsonValue;
};

type SchemaLike = Schema | Ref;

type Parameter = {
  name: string;
  in: string;
  required?: boolean;
  description?: string;
  schema: {
    type?: string;
    minimum?: number;
    maximum?: number;
  };
};

type MediaType = {
  schema: SchemaLike;
  example?: JsonValue;
};

type Response = {
  description: string;
  headers?: {[name: string]: {description?: string}};
  content?: {[mediaType: string]: MediaType};
};

type Operation = {
  operationId?: string;
  summary?: string;
  description?: string;
  parameters?: Parameter[];
  responses: {[status: string]: Response};
};

type PathItem = {
  get?: Operation;
  post?: Operation;
  put?: Operation;
  patch?: Operation;
  delete?: Operation;
};

type OpenAPISpec = {
  openapi: string;
  info: {
    title: string;
    version: string;
    description?: string;
  };
  servers: Array<{url: string; description?: string}>;
  paths: {[path: string]: PathItem};
  components?: {
    schemas?: {[name: string]: Schema};
  };
};

const spec = load(openapiYaml) as OpenAPISpec;

function refName(schema: SchemaLike | undefined): string | undefined {
  if (!schema) {
    return undefined;
  }
  if ('$ref' in schema) {
    return schema.$ref.split('/').pop();
  }
  if (schema.type === 'array' && schema.items) {
    const item = refName(schema.items);
    return item ? `array of ${item}` : 'array';
  }
  return schema.type;
}

function isRef(schema: SchemaLike): schema is Ref {
  return '$ref' in schema;
}

function statusClass(status: string): string {
  if (status.startsWith('2')) {
    return 'api-status api-status--ok';
  }
  if (status.startsWith('4')) {
    return 'api-status api-status--warn';
  }
  return 'api-status api-status--error';
}

function renderParameters(parameters: Parameter[] | undefined): React.ReactNode {
  if (!parameters || parameters.length === 0) {
    return null;
  }

  return (
    <>
      <h4 className="api-subhead">Parameters</h4>
      <table className="api-table">
        <thead>
          <tr>
            <th>Name</th>
            <th>In</th>
            <th>Required</th>
            <th>Type</th>
            <th>Description</th>
          </tr>
        </thead>
        <tbody>
          {parameters.map((parameter) => {
            const constraints: string[] = [];
            if (parameter.schema.minimum !== undefined) {
              constraints.push(`min ${parameter.schema.minimum}`);
            }
            if (parameter.schema.maximum !== undefined) {
              constraints.push(`max ${parameter.schema.maximum}`);
            }
            return (
              <tr key={parameter.name}>
                <td>
                  <code>{parameter.name}</code>
                </td>
                <td>{parameter.in}</td>
                <td>{parameter.required ? 'yes' : 'no'}</td>
                <td>
                  <code>{parameter.schema.type}</code>
                  {constraints.length > 0 ? ` (${constraints.join(', ')})` : null}
                </td>
                <td>{parameter.description}</td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </>
  );
}

function renderResponses(operation: Operation): React.ReactNode {
  return (
    <>
      <h4 className="api-subhead">Responses</h4>
      <table className="api-table">
        <thead>
          <tr>
            <th>Status</th>
            <th>Description</th>
            <th>Schema</th>
            <th>Headers</th>
          </tr>
        </thead>
        <tbody>
          {Object.entries(operation.responses).map(([status, response]) => {
            const media = response.content ? Object.values(response.content)[0] : undefined;
            const schemaName = refName(media?.schema);
            const headers = response.headers ? Object.keys(response.headers).join(', ') : '—';
            return (
              <tr key={status}>
                <td>
                  <span className={statusClass(status)}>{status}</span>
                </td>
                <td>{response.description}</td>
                <td>
                  {schemaName ? (
                    <a href={`#schema-${schemaName.replace(/^array of /, '')}`}>
                      <code>{schemaName}</code>
                    </a>
                  ) : (
                    '—'
                  )}
                </td>
                <td>{headers}</td>
              </tr>
            );
          })}
        </tbody>
      </table>
      {Object.entries(operation.responses).map(([status, response]) => {
        const media = response.content ? Object.values(response.content)[0] : undefined;
        if (!media?.example) {
          return null;
        }
        return (
          <div className="api-example" key={`${status}-example`}>
            <h4 className="api-subhead">{status} example</h4>
            <pre>
              <code>{JSON.stringify(media.example, null, 2)}</code>
            </pre>
          </div>
        );
      })}
    </>
  );
}

function renderSchema(name: string, schema: Schema): React.ReactNode {
  const entries = Object.entries(schema.properties ?? {});
  return (
    <div className="api-schema" key={name} id={`schema-${name}`}>
      <div className="api-schema__name">{name}</div>
      {schema.description ? <p className="api-schema__description">{schema.description}</p> : null}
      {entries.length > 0 ? (
        <table className="api-table">
          <thead>
            <tr>
              <th>Field</th>
              <th>Type</th>
              <th>Required</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            {entries.map(([field, fieldSchema]) => (
              <tr key={field}>
                <td>
                  <code>{field}</code>
                </td>
                <td>
                  <code>{refName(fieldSchema) ?? 'object'}</code>
                </td>
                <td>{schema.required?.includes(field) ? 'yes' : 'no'}</td>
                <td>{!isRef(fieldSchema) ? fieldSchema.description ?? '—' : '—'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      ) : null}
      {schema.type === 'object' && schema.additionalProperties === true ? (
        <p className="api-schema__description">
          Additional properties are allowed — records are passed through unchanged.
        </p>
      ) : null}
    </div>
  );
}

export default function ApiReference(): React.JSX.Element {
  const operations = Object.entries(spec.paths).flatMap(([path, item]) =>
    Object.entries(item).map(([method, operation]) => ({
      path,
      method: method.toUpperCase(),
      operation: operation as Operation,
    })),
  );
  const schemas = spec.components?.schemas ?? {};

  return (
    <div>
      <header className="api-header">
        <div className="api-header__top">
          <h1 className="api-header__title">{spec.info.title}</h1>
          <div className="api-header__meta">
            <span className="api-badge">OpenAPI {spec.openapi}</span>
            <span className="api-badge">v{spec.info.version}</span>
          </div>
        </div>
        {spec.info.description ? <p className="api-header__description">{spec.info.description}</p> : null}
        <div className="api-header__server">
          Base URL: <code>{spec.servers[0]?.url}</code>
        </div>
      </header>

      <h2 className="api-section-title">Endpoints</h2>
      {operations.map(({path, method, operation}) => (
        <section className="api-endpoint" key={`${method}-${path}`}>
          <div className="api-endpoint__head">
            <span className="api-method">{method}</span>
            <code className="api-path">{path}</code>
          </div>
          {operation.summary ? <div className="api-endpoint__summary">{operation.summary}</div> : null}
          {operation.description ? (
            <p className="api-endpoint__description">{operation.description}</p>
          ) : null}
          {renderParameters(operation.parameters)}
          {renderResponses(operation)}
        </section>
      ))}

      <h2 className="api-section-title">Schemas</h2>
      {Object.entries(schemas).map(([name, schema]) => renderSchema(name, schema))}

      <p className="api-note">
        This page is rendered directly from <code>docs/openapi.yaml</code>. Download the{' '}
        <a href="https://github.com/mfloriach/telemetry-service/blob/main/docs/openapi.yaml">
          raw specification
        </a>{' '}
        to generate clients or import it into an API tool.
      </p>
    </div>
  );
}
