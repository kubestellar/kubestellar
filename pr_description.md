## Summary
This PR adds comprehensive unit tests for the `pkg/status` package and the `pkg/status/aggregation` package. Specifically, it introduces unit tests for the CEL evaluator, CombinedStatus resolution and mapping, and WEC status aggregation.

## Related issue(s)
Fixes #3938

## Changes
- **CEL Evaluator tests**: Added `pkg/status/cel-evaluator_test.go` to test parser validation, syntax checking, and type checking as well as evaluation of expressions referencing `obj`, `returned`, `inventory`, and `propagation`.
- **CombinedStatus Resolution tests**: Added `pkg/status/combinedstatus-resolution_test.go` to test the updating of collection destinations, mapping of status collectors, evaluation of workstatus fields (select, group-by, count/sum/avg/min/max aggregation), and `compareCombinedStatus` logic.
- **Aggregation tests**: Added `pkg/status/aggregation/aggregation_test.go` to test `GetMin`, `GetMax`, and kind-specific aggregation logic for `Deployment`, `ReplicaSet`, and `DaemonSet` workloads.

## Testing
- Ran `go test -v ./pkg/status/...` to verify all status package and aggregation unit tests pass.
- Ran `go fmt ./pkg/status/... && go vet ./pkg/status/...` to verify formatting and validation checks.

## Checklist
- [x] Unit tests for CEL evaluator added in `cel-evaluator_test.go`
- [x] Unit tests for CombinedStatus resolution added in `combinedstatus-resolution_test.go`
- [x] Unit tests for WEC status aggregation added in `aggregation/aggregation_test.go`
- [x] Local unit tests run and passed successfully
