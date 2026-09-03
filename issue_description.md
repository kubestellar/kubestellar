## Description
The `pkg/status` package (~12 files, ~125KB) is one of the largest packages in the KubeStellar codebase with zero unit tests. It includes complex CEL expression evaluation, combined status resolution, and multi-WEC aggregation logic. We need to add comprehensive unit tests to ensure stability, prevent regressions, and increase code coverage.

## Proposed Changes
- Add unit tests for the CEL evaluator with various expression patterns (`cel-evaluator_test.go`).
- Test `CombinedStatus` resolution with mock workload statuses (`combinedstatus-resolution_test.go`).
- Test multi-WEC status aggregation logic (`aggregation_test.go` in `pkg/status/aggregation`).
- Use test fixtures/mocking to simulate realistic workload status scenarios.

## Goal
Establish a robust testing suite for the status package and its subcomponents (CEL evaluator, CombinedStatus resolution, and WEC status aggregation) with high coverage.
