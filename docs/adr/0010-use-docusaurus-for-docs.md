# ADR-0010: Use Docusaurus for the documentation site

**Date**: 2026-11-23
**Status**: proposed
**Deciders**: repository owner

## Context

Documentation is spread across `docs/` (architecture, ADRs, instability notes, OpenAPI spec) and `openspec/specs/` (canonical capability specs). There is no rendered view, and contributors have no quick way to preview changes to either tree. The site must read both directories in place so the Markdown remains the single source of truth, and edits must be visible during local development without a manual copy step. The Go runtime architecture is unaffected; this is a tooling decision.

## Decision

We use Docusaurus in a `docs/website/` directory with two `@docusaurus/plugin-content-docs` instances whose paths point at `..` (the `docs/` tree) and `../../openspec/specs`. `npm run start` provides hot reload across both trees and `npm run build` produces a static build for CI verification. No deployment pipeline is part of this decision.

## Alternatives Considered

### Alternative 1: MkDocs
- **Pros**: Lightweight, simple configuration, good Markdown support.
- **Cons**: Python toolchain added to a Go/Node-free repository; multiple doc roots need extra plugins; weaker theming and search story.
- **Why not**: Docusaurus fits the desired TypeScript/npm toolchain and handles multiple doc instances natively.

### Alternative 2: Rely on GitHub's Markdown rendering
- **Pros**: No new tooling or build step.
- **Cons**: No local preview or hot reload, no cross-document navigation, specs and docs cannot be browsed as one site.
- **Why not**: Does not satisfy the live-preview requirement.

### Alternative 3: Copy generated Markdown into a site directory on each change
- **Pros**: Works with any static site generator.
- **Cons**: Introduces duplicate files, sync scripts, and drift risk.
- **Why not**: Reading the source trees directly keeps a single source of truth.

## Consequences

### Positive
- One command previews documentation and canonical specs together with hot reload.
- Markdown files stay authoritative; no duplication or copy step.
- Static build can be verified in CI without running application infrastructure.

### Negative
- Adds a Node.js/npm toolchain and `node_modules` to a Go repository.
- In-progress `openspec/changes/*` artifacts are intentionally not published.

### Risks
- Docusaurus watch mode may not follow files outside the site root — mitigated by using two configured doc-plugin paths rather than symlinks and verifying hot reload in a smoke check.
- Upstream Docusaurus upgrades may change plugin path behaviour — mitigated by pinning the version and running `npm run build` in CI.
