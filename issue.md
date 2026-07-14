## Description
The binding controller (`pkg/binding/binding-controller.go`) performs core work (resolving BindingPolicies, creating Bindings) but exposes no Prometheus metrics for observability. The metrics infrastructure already exists in `pkg/metrics`.

## Proposed Changes
- Add counters for: BindingPolicy reconciliations (success/failure), Binding creations/updates/deletions, workload objects matched per BindingPolicy
- Add histograms for: reconciliation duration, binding resolution latency
- Use the existing `pkg/metrics` framework
- Add Grafana dashboard JSON for the new metrics in `monitoring/grafana/`

## Goal
Improve observability and monitoring of the binding controller's reconciliation loop and resource operations.
