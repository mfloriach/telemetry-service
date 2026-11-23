# Proposal

## Why

Documentation currently lives as scattered Markdown under `docs/` and `openspec/specs/`, with no rendered or searchable view and no way to see spec edits live. This change adds a Docusaurus site that publishes both trees and gives contributors a single-command preview with hot reload.

## What Changes

- Add a Docusaurus site in a new `docs/website/` directory, with its own `package.json` and TypeScript configuration.
- Configure Docusaurus to source content from `../docs` and `../openspec/specs` without duplicating files into the site.
- Provide `npm run start` for a live-reloading preview of both trees and `npm run build` for a production static build.
- Document the setup in `README.md` and pin the Node version with `.nvmrc`.
- Add a CI step that verifies the site builds; no hosting or deployment pipeline is included.
- Record the tooling choice in a new ADR.

No breaking changes.

## Capabilities

### New Capabilities
- `documentation-site`: Rendering the repository's Markdown documentation and canonical specs as a browsable site, including a preview mode that reloads when source files change.

### Modified Capabilities

None.

## Impact

- New `docs/website/` directory containing Docusaurus configuration, sidebars, and scripts; new `.nvmrc` at the repo root.
- New Node.js toolchain dependency (npm, TypeScript, Docusaurus); no Go, Kafka, WebSocket, or runtime architecture changes.
- `README.md` gains local preview/build instructions; CI gains a build check.
- Existing ADRs remain unchanged; ADR 0010 records the documentation tooling decision.
