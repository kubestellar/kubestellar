# KubeStellar Multi-WEC Status Aggregation Proposal

**Status:** Design for Review | **Issue:** #2809

---

## 1. Problem Statement

**Current Behavior:**

* Singleton workloads (nWECs == 1) propagate `.status` correctly in WDS objects.
* Multi-WEC deployments (nWECs > 1) leave WDS `.status` empty, preventing users and tools like ArgoCD from seeing true readiness.

**Goal:**

* Aggregate per-WEC `WorkStatus` and update `.status` in WDS workload objects to reflect **true multi-cluster readiness**.
* Maintain **backward compatibility** with singleton workloads.
* Avoid conflicts with other controllers (e.g., ArgoCD, Helm).
* Ensure **deterministic and observable behavior** across multiple BindingPolicies.

---

## 2. Key Principles

1. **Reuse Existing Patterns**

   * Collect WorkStatus slices per WDS object using the CombinedStatus slice/map approach.
   * Reuse `updateObjectStatus()` from SingletonStatus for safe `.status` patching.

2. **Predefined Field-Type Aggregation**

   * Rules are **field-type based**, no user-defined expressions initially.
   * Configurable for future analytics (sum, average, quorum).

3. **Direct WDS Update**

   * Aggregated status is pushed **directly into WDS workload objects**, maintaining a single source of truth.

4. **Opt-In**

   * Aggregation triggered via `wantMultiReportedState: true` per BindingPolicy/StatusCollector.
   * Singleton behavior remains unchanged for nWECs == 1.

5. **Deterministic Multi-BindingPolicy Handling**

   * Conflicting policies fall back to **conservative defaults** (`min()` for numeric, AND for negative conditions, OR for positive conditions).
   * `.status.kubestellar/policyBreakdown` records contributing policies and modes for observability.

---

## 3. Slice/Map Data Gathering

```go
c.workStatusToObject.ReadInverse().ContGet(wObjID, func(wsONSet sets.Set[cache.ObjectName]) {
    for wsON := range wsONSet {
        ws := c.workStatusLister.Get(wsON)
        workStatuses = append(workStatuses, WorkStatusData{
            ClusterName: ws.ClusterName,
            Status:      ws.Status,
            UpdateTime:  ws.LastUpdateTime,
        })
    }
})
```

* Slice mirrors CombinedStatus:
  * `.status` map from each WEC
  * Cluster metadata
  * Last update timestamp
* Stale WorkStatus (`LastUpdateTime` older than configurable threshold, e.g., 30s) is ignored during aggregation.

---

## 4. Aggregation Rules

### A. Numeric Fields

| Field                                                       | Aggregation             | Notes                                           |
| ----------------------------------------------------------- | ----------------------- | ----------------------------------------------- |
| replicas, readyReplicas, availableReplicas, updatedReplicas | min(values) across WECs | Conservative; prevents over-reporting in ArgoCD |
| Optional override                                           | sum / average / quorum  | For analytics/reporting                         |

**Example:** 3 clusters `[1, 0, 1]` → Output `0` (reflects weakest cluster)

### B. Conditions

* Group conditions by `Type`
* Aggregation logic:

| Type                        | Aggregation                                          |
| --------------------------- | ---------------------------------------------------- |
| Available / Ready / Healthy | OR / majority → healthy if any cluster ready         |
| Progressing / Updating      | OR / majority → shows ongoing work                   |
| Degraded / Failed / Error   | AND / conservative → unhealthy if any cluster failed |
| Unknown                     | OR → positive by default                             |

* Reason: pick latest/worst-case across WECs
* Message: include most informative message from degraded cluster

**Example:**

* Available: `[True, False, True]` → `False`
* Progressing: `[False, True, False]` → `True`

### C. Timestamps

* `lastTransitionTime`: max across WECs
* `observedGeneration`: max across WECs
* `ObjectMeta.Generation`: ignored for health

### D. Other Fields

* Copy from dominant WEC or leave untouched
* Nested maps, strings, or CRD-specific fields: no aggregation; log warning if ambiguous

---

## 5. Controller Integration

```go
func (c *Controller) syncWorkloadObject(ctx context.Context, wObjID util.ObjectIdentifier) error {
    singleRequested, nWECs := c.bindingPolicyResolver.GetSingletonReportedStateRequestForObject(wObjID)
    multiRequested := c.bindingPolicyResolver.GetMultiReportedStateRequestForObject(wObjID)

    if singleRequested && nWECs == 1 {
        return c.handleSingleton(...) // existing path
    }
    if multiRequested && nWECs > 1 {
        return c.handleMultiWECs(...) // slice -> aggregator -> updateObjectStatus
    }
    return c.updateObjectStatus(ctx, wObjID, nil, c.listers) // clear status
}
```

* `handleMultiWECs()` collects the WorkStatus slice, calls `processSlice()` for aggregation, and patches `.status` safely.
* Rate-limiting ensures performance at scale.

---

## 6. Aggregator Architecture

* Pluggable per-kind aggregators:
  * DeploymentAggregator
  * StatefulSetAggregator
  * DaemonSetAggregator
  * GenericAggregator fallback → `.status.kubestellar/combined`
* Aggregation mode configurable per BindingPolicy / StatusCollector (`min` default)
* Optional overrides for sum/average/quorum

---

## 7. Safety & Observability

* Patch only if `.status` changed (deep-equal check)
* Use MergePatch to avoid conflicts with other controllers
* Per-WEC breakdown stored in `.status.kubestellar/combined`
* Rate-limit updates to reduce API server churn
* Log warnings for ambiguous fields and conflicting policies

---

## 8. Expected Outcomes

* Accurate multi-cluster health in ArgoCD/Flux
* Meaningful READY counts in `kubectl get`
* Operators can debug per-cluster status
* Singleton workloads remain unaffected
* Flexible aggregation modes for reporting or analytics

---

## 9. Known Risks & Mitigations

| Risk                               | Impact | Mitigation                                     |
| ---------------------------------- | ------ | ---------------------------------------------- |
| Transient degraded state (WEC lag) | Medium | Ignore stale WorkStatus, document expected lag |
| Scale-up/down misreporting         | Medium | Optional override (sum/average/quorum)         |
| Non-numeric / non-condition fields | Low    | Copy from dominant WEC; log warning            |
| Conflicts with other controllers   | High   | MergePatch + rate-limiting                     |
| Multiple BindingPolicies           | Medium | Deterministic mode, log policy breakdown       |
| Condition message ambiguity        | Medium | Pick latest/worst-case; log for audit          |
| ObjectMeta.Generation              | Low    | Max stored for reference; not used for health  |

---

## 10. Summary

* Default numeric aggregation: **min()**
* Condition aggregation: **OR for positive, AND for negative**
* Timestamps: **max**
* `.status` patched directly in WDS workload objects (singleton path untouched)
* Configurable overrides per CR for sum/average/quorum
* Fully aligned with CombinedStatus slice/map mechanism, safe, ArgoCD-compatible
* Deterministic multi-BindingPolicy handling, transient state handling, and performance rate-limiting addressed

---