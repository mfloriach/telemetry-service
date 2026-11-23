# Spec Delta

## Purpose

Provides one browsable view of the repository's documentation and canonical capability specs, with a local preview that reflects source edits without a manual rebuild step.

## ADDED Requirements

### Requirement: Documentation pages are published
The system SHALL make every Markdown document in the repository's documentation tree available as a navigable page in the published site.

#### Scenario: Existing documents are available
- **WHEN** a Markdown document exists in the documentation tree
- **THEN** the published site includes a corresponding navigable page

#### Scenario: Nested documents are grouped
- **WHEN** documents are organized in subdirectories
- **THEN** the published site groups their pages in a matching navigation hierarchy

### Requirement: Canonical specs are published
The system SHALL include the canonical capability specs in the published site as browsable pages.

#### Scenario: Capability spec is published
- **WHEN** a capability spec exists in the canonical spec tree
- **THEN** the published site includes a page showing that spec's content

#### Scenario: Published specs follow source edits
- **WHEN** a canonical spec file is edited and the site is rebuilt
- **THEN** the corresponding published page reflects the edited content

### Requirement: In-progress change artifacts are excluded
The system SHALL exclude in-progress change artifacts from the published site.

#### Scenario: Change artifact exists
- **WHEN** a change contains proposal, design, task, or delta spec files
- **THEN** none of those files are published as site pages

### Requirement: Preview reflects source changes
The system SHALL provide a preview mode that updates published pages when documentation or canonical spec source files change, without a manual rebuild or copy step.

#### Scenario: Documentation source changes during preview
- **WHEN** a documentation Markdown file is edited while preview is running
- **THEN** the affected page updates in the preview

#### Scenario: Spec source changes during preview
- **WHEN** a canonical spec file is edited while preview is running
- **THEN** the affected spec page updates in the preview

### Requirement: Static build is independently verifiable
The system SHALL produce a static site build that completes successfully from source files and reports failure when the build cannot complete.

#### Scenario: Successful build
- **WHEN** all source documents are valid
- **THEN** the build completes with a success status and produces static site output

#### Scenario: Failed build is reported
- **WHEN** the build cannot complete
- **THEN** the build exits with a failure status that an automated check can detect
