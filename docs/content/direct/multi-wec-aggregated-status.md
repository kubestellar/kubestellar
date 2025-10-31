# Proposal: Multi-WEC Aggregated Status Controller Enhancement

**Status:** Proposal  
**Related Issue:** [#2809 – KubeStellar controller logic to map back the status from WECs to READY values in WDSes](https://github.com/kubestellar/kubestellar/issues/2809)

---

## Current Scenario

In the current KubeStellar implementation:

1. **Singleton Enabled (WEC == 1):**  
   The WDS and WEC workload statuses match correctly when `wantSingletonReportedState: true` is enabled.

2. **Singleton Disabled (WEC > 1):**  
   For workloads deployed to multiple WECs, the `.status` field in the WDS remains empty, resulting in non-matching WDS and WEC status.

This leads to limited visibility in ArgoCD, where workloads appear as *OutOfSync* or *Unknown*, even though all WECs are healthy.

**Reference:**  
- [PR #3297](https://github.com/kubestellar/kubestellar/pull/3297)  
- [Multi-cluster example in docs](https://docs.kubestellar.io/release-0.28.0/direct/example-scenarios/#scenario-1-multi-cluster-workload-deployment-with-kubectl)

---

## Goal

Implement a **multi-WEC aggregated status reporting mechanism** in KubeStellar to reflect combined readiness in the WDS.  
This will allow aggregated workload visibility across all WECs and ensure compatibility with ArgoCD’s health checks.

Key objectives:
- Aggregate per-WEC `WorkStatus` objects.
- Update the `.status` field in WDS workload objects.
- Maintain backward compatibility with singleton workloads.
- Provide fallback to existing singleton mechanism when applicable.

---

## Solution Approach

### 1. Extend BindingPolicy CRD

Add a new field `wantMultiWECReportedState`, similar to `wantSingletonReportedState`.

Logic:
- If **enabled** and number of WECs == 1 → use existing singleton mechanism.
- If **enabled** and number of WECs > 1 → invoke new multi-WEC aggregation logic.
- If **disabled**, retain existing empty status behavior.

```yaml
spec:
  wantMultiWECReportedState: true
```

---

### 2. Introduce `handleMultiReportedState()` Function

A new controller function will handle multi-WEC aggregation:

**Responsibilities:**
1. Collect `.status` from all target WECs using existing status collection logic.  
2. Aggregate status data according to defined field rules (numeric, conditions, timestamps, string).  
3. Update the aggregated `.status` back into the WDS workload object.

**Focus:** Aggregation of fields relevant to ArgoCD health interpretation.

---

### 3. Status Aggregation Rules

| Field Type | Aggregation Logic | Description |
|-------------|------------------|--------------|
| **Numeric** | Average or Minimum | Used for fields like replica counts. |
| **Condition** | Group by `type`; aggregate status accordingly. | Ensures consistent boolean conditions across clusters. |
| **Timestamp** | Use latest timestamp. | Reflects the most recent update across all WECs. |
| **String** | Use latest value. | Keeps newest message or reason from clusters. |

---

### 4. Implementation Approach

1. Extend BindingPolicy CRD with `wantMultiWECReportedState`.
2. Add flag detection logic in the status controller.
3. Implement `handleMultiReportedState()` using the aggregation approach above.
4. Use type-specific aggregators for common workload kinds (Deployment, StatefulSet, etc.).
5. Patch `.status` only if a change is detected (deep equality check).

**Code Reference:**
- [ArgoCD Health Logic](https://github.com/argoproj/argo-cd/blob/master/gitops-engine/pkg/health/health.go)
- [Initial Proposal PR](https://github.com/rishi-jat/kubestellar/pull/1)

---

## Example

```yaml
apiVersion: policies.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: multiwec-nginx
spec:
  objectSelector:
    kind: Deployment
    namespace: nginx
    name: nginx-deployment
  targetClusters:
    - cluster1
    - cluster2
  wantMultiWECReportedState: true
```

After aggregation, the corresponding WDS Deployment status:

```yaml
status:
  replicas: 2
  readyReplicas: 2
  availableReplicas: 2
  conditions:
    - type: Ready
      status: "True"
      reason: AllClustersReady
      message: "Deployment ready in cluster1, cluster2"
      lastTransitionTime: "2025-11-01T12:34:56Z"
```

---

## Validation Plan

- Verify aggregated `.status` in WDS after multi-cluster propagation.
- Confirm ArgoCD displays workloads as *Synced* and *Healthy*.
- Validate fallback to singleton logic when only one WEC is bound.
- Add integration tests under `test/e2e/status/` to verify correctness.

---

## Future Work

- Weighted aggregation logic based on cluster priority or capacity.
- Enhanced visualization for per-WEC readiness.
- Unified CRD version for both singleton and multi-WEC status.

---

## References

- [KubeStellar Status Controller](https://github.com/kubestellar/kubestellar/tree/main/pkg/status)
- [ArgoCD Health Evaluation](https://argo-cd.readthedocs.io/en/stable/operator-manual/health/)
- [Issue #2809 – Multi-WEC Aggregation](https://github.com/kubestellar/kubestellar/issues/2809)
- [Related PR #3423](https://github.com/kubestellar/kubestellar/pull/3423)