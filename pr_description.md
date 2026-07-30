## Summary
This PR improves context propagation and options handling in the binding controller's dynamic list/watch functions in `pkg/binding/crd.go`. Specifically, it replaces `context.TODO()` with the context propagated from the controller's `run` method, and forwards `metav1.ListOptions` instead of using default empty options.

## Related issue(s)
Fixes #3922

## Changes
- **Context propagation in `pkg/binding/crd.go`**:
  - Replaced `context.TODO()` with the controller context `ctx` in the `List` and `Watch` function calls within `startInformersForNewAPIResources`.
  - Updated the dynamic client's `List` and `Watch` calls to pass the `options` parameter instead of `metav1.ListOptions{}`.

## Testing
- Verified compilation: `go build ./cmd/controller-manager/...`
- Ran integration tests: `PATH=$PWD/etcd-v3.5.16-darwin-arm64:$PATH go test -v ./test/integration/controller-manager/...`

## Checklist
- [x] Context threaded from parent context in `pkg/binding/crd.go` List/Watch functions
- [x] Passed options parameter properly into List and Watch calls
- [x] Local builds and integration tests passing successfully
