## Description
The repository currently uses a Prow `OWNERS` file, but lacks a GitHub-native `CODEOWNERS` configuration. Adding a `.github/CODEOWNERS` file will automatically assign reviewers based on which files are modified in a pull request. This improves reviewer assignment automation and speeds up the PR review cycle.

## Proposed Changes
- Create a `.github/CODEOWNERS` file at the root.
- Map directories to appropriate maintainers (based on the root `OWNERS`, `hack/OWNERS`, and `docs/to-be-deleted/OWNERS`).
- Explicitly cover key directories:
  - `pkg/binding/`
  - `pkg/status/`
  - `pkg/transport/`
  - `docs/`
  - `monitoring/`
  - `core-chart/`

## Goal
Automate GitHub PR reviewer assignments by mapping file and directory patterns to the project maintainers.
