# Tasks

## 1. Toolchain setup

- [x] 1.1 Add `.nvmrc` with the supported Node version and `docs/website/package.json` with pinned Docusaurus, TypeScript, and `start`, `build`, `typecheck` scripts — verify `nvm use && npm install` from `docs/website/` succeeds and creates `package-lock.json`.
- [x] 1.2 Add `docs/website/tsconfig.json` and the minimal Docusaurus TypeScript entry files — verify `npm run typecheck` exits zero.

## 2. Content sources and navigation

- [x] 2.1 Configure two `@docusaurus/plugin-content-docs` instances (`docs` → `../docs`, `specs` → `../openspec/specs`) — verify `npm run build` succeeds and the output contains a page for a known doc and a known canonical spec.
- [x] 2.2 Define a sidebar for each content instance that mirrors its source directory structure — verify nested documents and capability specs appear under the expected sidebar groups in the built output.
- [x] 2.3 Restrict the content plugins to the two source trees so `openspec/changes/` is never published — verify no proposal, design, task, or delta-spec filenames from a change directory appear in the build output.

## 3. Preview and hot reload

- [x] 3.1 Add `npm run start` backed by the Docusaurus dev server using both content instances — verify the server starts and serves one page from each tree over HTTP.
- [x] 3.2 Smoke-test reload for the documentation tree — verify editing a Markdown file under `docs/` while preview runs updates the page, then revert the edit.
- [x] 3.3 Smoke-test reload for the spec tree — verify editing a Markdown file under `openspec/specs/` while preview runs updates the page, then revert the edit.

## 4. Documentation

- [x] 4.1 Add a "Documentation site" section to `README.md` covering the Node version, install, preview, and build commands — verify each documented command runs as written from `docs/website/`.

## 5. CI and integration verification

- [x] 5.1 Add a CI job that runs `npm ci && npm run build` in `docs/website/` with path filters for `docs/website/`, `docs/`, and `openspec/specs/` — verify the job triggers on a test push touching each path and exits zero.
- [x] 5.2 Run the full local verification sequence (`npm ci`, `npm run typecheck`, preview smoke checks, `npm run build`) — verify every command exits zero and both trees are present in the build output.
- [x] 5.3 Verify the runtime is untouched — verify `go build ./cmd/...`, `go test -race ./...`, and `golangci-lint run ./...` pass and `git diff --name-only` lists no changes under `cmd/`, `internal/`, or `pkg/`.
