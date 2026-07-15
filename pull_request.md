## Summary
This PR adds Prometheus metrics for Binding Controller Operations, improving observability and monitoring. It registers counters for reconciliations and resource operations, a gauge for matched workload objects, and histograms for latency tracking. Additionally, it provides a Grafana dashboard for visualization.

## Related issue(s)
Fixes #3910

## Changes
- **Binding Controller metrics**:
  - `kubestellar_binding_controller_reconciliation_duration_seconds`: Histogram measuring the reconciliation duration for binding controller operations by resource (`binding`, `bindingpolicy`, `crd`, `workload`).
  - `kubestellar_binding_controller_binding_resolution_latency_seconds`: Histogram measuring the latency of binding resolution inside `syncBinding`.
  - `kubestellar_binding_controller_binding_policy_reconciliations_total`: CounterVec tracking the total number of reconciliations by status (`success`, `failure`).
  - `kubestellar_binding_controller_binding_operations_total`: CounterVec tracking the total number of binding resource operations (`create`, `update`, `delete`).
  - `kubestellar_binding_controller_workload_objects_matched`: GaugeVec tracking the current count of workload objects matched per binding policy.
- **Controller-Manager integration**:
  - Registered binding controller metrics in the legacy registry on startup.
- **Grafana Dashboard**:
  - Added `monitoring/grafana/BindingControllerMonitoring.json` dashboard.

## Testing
- Verified compilation: `go build ./cmd/controller-manager/...`
- Ran integration tests: `go test -v ./test/integration/controller-manager`

## Checklist
- [x] Counters and histograms defined and registered for binding controller operations
- [x] Reconciliation duration and binding resolution latency metrics implemented
- [x] Grafana dashboard JSON created in `monitoring/grafana/`
- [x] Local builds and integration tests passing successfully
