## Description
The project currently has a `.pre-commit-config.yaml` file and a GitHub Actions spellcheck workflow, but the spellcheck only runs in CI. Integrating spellcheck into local pre-commit hooks will improve the developer experience by catching typos in documentation before they are committed and pushed.

## Proposed Changes
- Add the `pyspelling` hook to `.pre-commit-config.yaml`.
- Ensure it shares the same dictionary/ignore list as the CI workflow.
- Fix invalid configuration keys in `.github/spellcheck/.spellcheck.yml` and add support for fenced code blocks.
- Update `CONTRIBUTING.md` with instructions for installing prerequisites (`aspell`, `pre-commit`) and setting up local pre-commit hooks.
- Fix existing typos in documentation files.

## Goal
Enable local spellchecking via pre-commit hooks to catch typos before they are pushed to the remote repository.
