# Multi-WEC Aggregated Status

This page describes how KubeStellar returns status information to workload objects in a Workload Description Space (WDS) when those objects are downsynced to more than one Workload Execution Cluster (WEC). This mechanism complements the existing singleton reported state feature and allows the `.status` field of the workload object in the WDS to reflect an aggregated result from multiple WECs.

## Overview

KubeStellar propagates workload objects from a WDS to one or more WECs according to the bindings specified by the user. When a workload is downsynced to exactly one WEC, users may request singleton reported state, which copies the `.status` field from that WEC to the corresponding object in the WDS.

When a workload is downsynced to multiple WECs, the default behavior leaves the `.status` field of the WDS object empty. Users can inspect the `CombinedStatus` objects for full per-cluster details, but the workload object itself does not show a summarized readiness or health view.

The Multi-WEC Aggregated Status option allows users to request status aggregation in these multi-cluster cases. The aggregated status is written into the `.status` field of the workload object in the WDS using deterministic rules.

## Enabling Multi-WEC Status Reporting

Status return is configured in the `DownsyncModulation` section of a `BindingPolicy` or `Binding`. To request aggregated status from multiple WECs, set:

```yaml
wantMultiWECReportedState: true
```

This option parallels `wantSingletonReportedState` but applies in the case where more than one WEC is selected.

Example:

```yaml
apiVersion: policies.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: multiwec-nginx
spec:
  clusterSelectors:
    - matchLabels:
        location: edge
  downsync:
    - objectSelectors:
        - matchLabels:
            app.kubernetes.io/name: nginx-multi
      wantMultiWECReportedState: true
```

For field definitions, see the DownsyncModulation API:
https://github.com/kubestellar/kubestellar/blob/v0.29.0/api/control/v1alpha1/types.go#L138-L167

## Behavior Summary

The effect of status return options depends on the number of selected WECs.

| `wantSingletonReportedState` | `wantMultiWECReportedState` | WEC count | Result |
|:--:|:--:|:--:|:--|
| T | F | 1 | Copy status from the single WEC. |
| T | F | 0 or >1 | Clear `.status`. |
| F | T | 1 | Copy status from the single WEC. |
| F | T | >1 | Aggregate status from all WECs. |
| T | T | 1 | Copy status from the single WEC. |
| T | T | >1 | Aggregate status from all WECs (multi-WEC takes precedence). |
| F | F | any | Leave `.status` empty. |

**Legend:** T = true, F = false

When both flags are enabled, the multi-WEC aggregation behavior takes precedence if more than one WEC is selected.

## Aggregated Status Semantics

The `.status` field of a Kubernetes workload object is defined for a single cluster. When aggregating from multiple WECs, KubeStellar applies approximation rules. The aggregation logic distinguishes two broad categories:

- Workload kinds for which Argo CD defines built-in health assessment.
- All other kinds.

For Argo CD’s health rules, see:
https://argo-cd.readthedocs.io/en/stable/operator-manual/health/

### Workload Kinds with Argo CD Health Rules

Only fields evaluated by Argo CD’s health logic are aggregated. These include readiness-related numeric fields and conditions.

For Deployments, StatefulSets, ReplicaSets, and DaemonSets:

- Readiness-related numeric fields (such as `readyReplicas`, `availableReplicas`, and their equivalents) are aggregated using the minimum across WECs.
- Conditions are aggregated using the condition rules described below.

For Jobs:

- Numeric fields used by Argo CD’s health assessment (`active`, `succeeded`, `failed`) are aggregated using minimum.
- Conditions are aggregated using the same condition rules.

### General Aggregation Rules (Other Kinds)

For workload kinds not evaluated by Argo CD:

- Numeric fields are aggregated using the minimum value.
- Boolean fields are aggregated as `false` if any WEC reports `false`, `true` only if all report `true`.
- Conditions follow the condition rules described below.
- Timestamps from condition entries use the latest `lastTransitionTime`.
- `reason` and `message` fields for a condition are taken from the condition entry with the latest `lastTransitionTime`.
- List and map fields are aggregated only when their structure matches across all WECs. Fields with incompatible shapes are omitted.

These rules avoid making assumptions about workload semantics while still providing a useful summary.

## Condition Aggregation

Conditions use uniform rules across all workload kinds.

### Truth Value

| WEC condition values | Aggregated value |
| -- | -- |
| Any `False` | `False` |
| All `True` | `True` |
| Mix of `True` and `Unknown` | `Unknown` |
| All `Unknown` | `Unknown` |

### Timestamps

The aggregated condition uses the most recent `lastTransitionTime` across all WECs.

### Reason and Message

The `reason` and `message` fields come from the condition entry whose `lastTransitionTime` is the most recent.

## Example

Consider a Deployment that is downsynced to two WECs, each reporting the Deployment as fully available. With Multi-WEC status reporting enabled, the WDS object contains:

```yaml
status:
  replicas: 3
  readyReplicas: 3
  availableReplicas: 3
  conditions:
    - type: Available
      status: "True"
      lastTransitionTime: "2025-11-01T12:34:56Z"
      reason: MinimumReplicasAvailable
      message: Deployment has minimum availability.
```

This aggregated status supports tools such as Argo CD in evaluating the workload’s health directly from the WDS object.

## Relation to Combined Status

This feature updates only the `.status` field of the workload object in the WDS. The `CombinedStatus` mechanism continues to provide detailed per-WEC information and remains available for inspection and debugging.

See:  
[Combining reported state](combined-status.md)

## Limitations

- Aggregation is an approximation because workload `.status` fields are cluster-specific.
- Only the fields used in Argo CD’s health rules are aggregated for the workload kinds supported by Argo CD.
- For other kinds, aggregation is best-effort and avoids semantic assumptions.
- Larger numbers of WECs increase status reporting volume; rate-limiting and batching apply.

## See Also

- [Binding](binding.md)
- [Transforming desired state](transforming.md)
- [Combining reported state](combined-status.md)
- [Example scenarios](example-scenarios.md)
- DownsyncModulation API reference: https://github.com/kubestellar/kubestellar/blob/v0.29.0/api/control/v1alpha1/types.go#L138-L167
- Argo CD health checks: https://argo-cd.readthedocs.io/en/stable/operator-manual/health/
