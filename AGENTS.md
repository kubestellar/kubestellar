# Contributor Agent Notes

Use this file as a quick orientation layer. The project rules remain in `CONTRIBUTING.md`; when there is a conflict, follow `CONTRIBUTING.md`.

## Contribution Flow

- Work from a topic branch, not `main`.
- Prefer small, issue-linked patches. Check for existing open PRs before starting an issue.
- Claim an issue with `/assign` when you begin active work, and ask for clarification on the issue if the expected behavior or scope is unclear.
- Sign off commits with `git commit -s` to satisfy the DCO requirement.
- Keep public comments direct and specific: state what was tested, what remains uncertain, or what input is needed.

## Local Validation

- Run focused Go tests for touched packages first, for example `go test ./pkg/transport/...`.
- Use `git diff --check` before committing.
- For docs changes, verify the source file is in the active docs tree before editing; `docs/CLAUDE.md` notes current console-doc locations.

## Transport Controller Notes

- Generic transport code lives in `pkg/transport/generic/`.
- CustomTransform cache and invalidation logic lives in `pkg/transport/generic/custom_transform_collection.go`.
- Wrapped object propagation decisions are in `pkg/transport/generic/generic_transport_controller.go`.
- When changing propagation skip/update logic, add focused unit coverage in `pkg/transport/generic/generic_transport_controller_test.go`.
