## Description
The codebase has multiple instances of `context.TODO()` in the manually-written binding controller code (`pkg/binding/crd.go`). This prevents proper context propagation, tracing, and cancellation support.

## Proposed Changes
- Replace `context.TODO()` usage in dynamic client `List` and `Watch` calls with the properly propagated context.
- Ensure list/watch options are passed through correctly instead of using empty options.

## Goal
Improve context propagation and cancellation support for the dynamic informers within the binding controller.
