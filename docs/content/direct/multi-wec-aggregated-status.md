# Proposal: Multi-WEC Aggregated Status Controller Enhancement

**Status:** Proposal  
**Related Issue:** [#2809 – KubeStellar controller logic to map back the status from WECs to READY values in WDSes](https://github.com/kubestellar/kubestellar/issues/2809)

---

## Current Scenario

In the current KubeStellar implementation:

### Case 1 — Single WEC and `wantSingletonReportedState` is true  
The WDS workload `.status` is propagated from the single WEC.

### Case 2 — Multiple WECs and `wantSingletonReportedState` is false  
For workloads deployed to multiple WECs, the `.status` field in the WDS remains empty, resulting in no aggregated status propagated.

This leads to limited visibility in ArgoCD, where workloads appear as *OutOfSync* or *Unknown*, even though all WECs are healthy.

**Reference:**  
- [PR #3297](https://github.com/kubestellar/kubestellar/pull/3297)  
- [Multi-cluster example in docs](https://docs.kubestellar.io/release-0.28.0/direct/example-scenarios/#scenario-1-multi-cluster-workload-deployment-with-kubectl)

---

## Goal

KubeStellar already collects per-WEC status in `WorkStatus` objects. This enhancement propagates that aggregated information into the `.status` field of the original workload object in the WDS, enabling ArgoCD and similar tools to observe combined readiness and health across all WECs.

Key objectives:  
- Maintain backward compatibility when `wantMultiWECReportedState` is false, preserving existing behavior.  
- When `wantMultiWECReportedState` is enabled, provide fallback to the singleton mechanism if only one WEC is bound.  
- Aggregate per-WEC `WorkStatus` objects and update the `.status` field in WDS workload objects with combined readiness and health status.

---

## Solution Approach

### 1. Extend BindingPolicy CRD

The configuration bit `wantMultiWECReportedState` is added under each `DownsyncModulation` rule (not directly under `BindingPolicy`), consistent with the canonical definition available at [./combined-status/#special-case-for-1-wec](./combined-status/#special-case-for-1-wec).

Example:

```yaml
apiVersion: policies.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: multiwec-nginx
spec:
  clusterSelectors:
  - matchLabels: {"location": "edge"}
  downsync:
  - objectSelectors:
    - matchLabels: {"app.kubernetes.io/name": "nginx-multi"}
    wantMultiWECReportedState: true
```

This example follows the [singleton status example](https://docs.kubestellar.io/release-0.28.0/direct/example-scenarios/#scenario-4-singleton-status) but replaces `wantSingletonReportedState` with `wantMultiWECReportedState`.

---

### 2. Introduce `handleMultiWECReportedState()` Function (Planned)

A new controller function, `handleMultiWECReportedState()`, will implement multi-WEC aggregated status reporting using per-kind aggregation logic for common workload kinds such as Deployment, StatefulSet, DaemonSet, ReplicaSet, Job, CronJob, and Pod.

Both `BindingPolicy` and `Binding` inherit the `DownsyncModulation` extension where this flag resides.

---

### 3. Status Aggregation Rules

The aggregation logic depends on the **workload object kind**.

#### Case 1 – Known Workload Kinds

For workload kinds recognized by ArgoCD, the controller applies predetermined field aggregation rules consistent with ArgoCD’s native health evaluation. These include:

- **Deployment**  
  - `replicas`, `updatedReplicas`, `availableReplicas`, `readyReplicas`: use `min()` across clusters to reflect the least ready state.  
  - `conditions`: apply three-value logic per condition type:  
    - `False` if any cluster reports `False`.  
    - `True` if all clusters report `True`.  
    - Otherwise, `Unknown`.  
  - `observedGeneration`: use the minimum observed generation.  
  - `unavailableReplicas`: use `max()` to represent the worst state.

- **StatefulSet**  
  - `replicas`, `readyReplicas`, `currentReplicas`: use `min()`.  
  - `updatedReplicas`: use `min()`.  
  - `conditions`: three-value logic as above.

- **DaemonSet**  
  - `currentNumberScheduled`, `desiredNumberScheduled`, `numberAvailable`, `numberReady`: use `min()`.  
  - `conditions`: three-value logic.

- **ReplicaSet**  
  - `replicas`, `readyReplicas`, `availableReplicas`: use `min()`.  
  - `conditions`: three-value logic.

- **Job**  
  - `active`, `succeeded`, `failed`: use `min()` for all numeric fields to represent the least-ready state across clusters.  
  - `conditions`: three-value logic.

- **CronJob**  
  - `lastScheduleTime`: use the latest timestamp across clusters.  
  - `conditions`: three-value logic.

- **Pod**  
  - `phase`: aggregate by prioritizing `Failed` > `Unknown` > `Running` > `Succeeded` > `Pending`.  
  - `conditions`: three-value logic.

**Three-value logic for Conditions:**  
- `False` if any cluster reports `False` for a condition type.  
- `True` if all clusters report `True`.  
- Otherwise, `Unknown`.

**Timestamps:**  
- For all condition timestamps (`lastTransitionTime`), take the most recent timestamp across clusters.

**Strings (Reason, Message):**  
- For each condition, take the `Reason` and `Message` from the condition with the most recent `lastTransitionTime`.

- **Slices and maps:**  
  - Match items by keys (e.g., condition `type`) before per-field aggregation.

#### Case 2 – Unknown Workload Kinds

For workload kinds not recognized by ArgoCD, the controller performs generic per-field aggregation using the same logical rules as above but constrained for general applicability:

- **Numeric fields:** use `min()` to represent the least-ready state across clusters.  
- **Conditions:** apply the same three-value logic (False if any False, True if all True, otherwise Unknown).  
- **Timestamps:** take the latest `lastTransitionTime`.  
- **Strings:** pick the most recent `Reason` and `Message` values.  
- **Slices or maps:** aggregate per key before applying the above rules.

This ensures consistency across arbitrary resource kinds without making unsupported assumptions about summation or scaling of numeric fields.

---

### 4. Implementation Approach

1. Extend BindingPolicy CRD with `wantMultiWECReportedState` under `DownsyncModulation`.
2. Add flag detection logic in the status controller.
3. Implement `handleMultiWECReportedState()` (planned) using per-kind aggregation logic for common workload kinds.
4. Update the workload `.status` only when the aggregated result differs from the current value (deep equality check).

**Note:** The current function [singletonstatus.go at commit 7d8a2f4](https://github.com/kubestellar/kubestellar/blob/7d8a2f4/pkg/status/singletonstatus.go#L64) copies `.status` directly from one WEC when only one `WorkStatus` object exists.

---

## Example

After aggregation, the corresponding WDS Deployment status:

```yaml
status:
  replicas: 2
  readyReplicas: 2
  availableReplicas: 2
  conditions:
    - type: Available
      status: "True"
      reason: AggregatedAllClustersReady
      message: "Ready if all clusters report Ready; False if any cluster reports False; Unknown otherwise"
      lastTransitionTime: "2025-11-01T12:34:56Z"
```

In this example, `reason` and `message` are synthesized summaries to indicate overall cluster results; they are not field-level merges.

# Note: Only fields relevant to ArgoCD's built-in health checks are aggregated.  
# Reference: https://github.com/argoproj/argo-cd/blob/master/gitops-engine/pkg/health/health_deployment.go

---

## Validation Plan

- Verify aggregated `.status` in WDS after multi-cluster propagation.  
- Confirm ArgoCD displays workloads as *Synced* and *Healthy*.  
- Validate fallback to singleton logic when only one WEC is bound.  
- Add end-to-end tests under `test/e2e/status/` to verify correctness.

---

## References

- [KubeStellar Status Controller](https://github.com/kubestellar/kubestellar/tree/main/pkg/status)  
- ArgoCD health evaluation logic from [`health.go`](https://github.com/argoproj/argo-cd/blob/master/gitops-engine/pkg/health/health.go) and [ArgoCD Health Documentation](https://argo-cd.readthedocs.io/en/stable/operator-manual/health/)  
- [Issue #2809 – Multi-WEC Aggregated Status Reporting](https://github.com/kubestellar/kubestellar/issues/2809)  
- [Related PR #3297 – Add Comprehensive Singleton Status Report for Franco](https://github.com/kubestellar/kubestellar/pull/3297)  
- [Related PR #3423](https://github.com/kubestellar/kubestellar/pull/3423)