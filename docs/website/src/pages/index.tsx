import React from 'react';
import Link from '@docusaurus/Link';
import Layout from '@theme/Layout';

type Feature = {
  title: string;
  description: string;
  to: string;
};

const features: Feature[] = [
  {
    title: 'System architecture',
    description:
      'How the relay, Kafka, ring buffer and HTTP API fit together, including context and container diagrams.',
    to: '/docs/architecture',
  },
  {
    title: 'API reference',
    description:
      'The rendered OpenAPI contract for GET /telemetry/latest — parameters, responses and schemas.',
    to: '/docs/api',
  },
  {
    title: 'AWS deployment',
    description:
      'Two serverless topologies explained: API Gateway plus Lambda, or a IAM-authenticated Lambda Function URL.',
    to: '/docs/aws-deployment/overview',
  },
  {
    title: 'Capability specs',
    description:
      'Canonical OpenSpec requirements and scenarios for ingest, relay, buffering and structured logging.',
    to: '/specs/satellite-telemetry/spec',
  },
  {
    title: 'Decision records',
    description:
      'The context and trade-offs behind key choices, from the WebSocket client to Kafka and logging.',
    to: '/docs/adr/',
  },
  {
    title: 'Instability handling',
    description:
      'Expected behaviour when the upstream telemetry feed disconnects, drops messages or restarts.',
    to: '/docs/instability',
  },
];

export default function Home(): React.JSX.Element {
  return (
    <Layout
      title="Documentation"
      description="Documentation and capability specs for the telemetry server">
      <header className="hero-home">
        <div className="hero-home__inner">
          <span className="hero-home__eyebrow">Telemetry server</span>
          <h1 className="hero-home__title">Documentation, clearly laid out.</h1>
          <p className="hero-home__subtitle">
            Architecture, the HTTP API contract, deployment options and capability specs — all in one place.
          </p>
          <div className="hero-home__actions">
            <Link className="button button--primary button--lg" to="/docs/architecture">
              Read the docs
            </Link>
            <Link className="button button--secondary button--lg" to="/docs/api">
              API reference
            </Link>
          </div>
        </div>
      </header>

      <main className="section-home">
        <h2 className="section-home__heading">Explore</h2>
        <p className="section-home__intro">Pick a topic to get started.</p>
        <div className="feature-grid">
          {features.map((feature) => (
            <Link className="feature-card" key={feature.to} to={feature.to}>
              <div className="feature-card__title">{feature.title}</div>
              <div className="feature-card__body">{feature.description}</div>
            </Link>
          ))}
        </div>
      </main>
    </Layout>
  );
}
