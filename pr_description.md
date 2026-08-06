## Summary
This PR adds a GitHub-native `.github/CODEOWNERS` file to automate reviewer assignments for pull requests. The directory mapping is based on the existing root `OWNERS`, `hack/OWNERS`, and `docs/to-be-deleted/OWNERS` files, and covers key project areas as specified in the issue.

## Related issue(s)
Fixes #3952

## Changes
- Created `.github/CODEOWNERS` file mapping files and directories to the project maintainers.
- Covered key areas:
  - `pkg/binding/`
  - `pkg/status/`
  - `pkg/transport/`
  - `docs/`
  - `monitoring/`
  - `core-chart/`
- Included specific mappings for `/hack/` and `/docs/to-be-deleted/` utilizing the additional maintainers defined in their respective `OWNERS` configurations.

## Testing
- Verified that all Github handles listed exist and are exactly corresponding to the `OWNERS` files.
- Confirmed correct matching rules for `CODEOWNERS` syntax.
