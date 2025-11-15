# Proposal: Multi-WEC Aggregated Status Enhancement

This document proposes an enhancement to KubeStellar’s status return mechanisms that enables the `.status` field of a workload object in the Workload Description Space (WDS) to reflect aggregated status when the object is downsynced to more than one Workload Execution Cluster (WEC). This expands the existing singleton reported state feature to the multi-WEC case.

## Problem Statement

When a workload is downsynced to multiple WECs, the current controller behavior leaves the `.status` field of the workload object in the WDS empty. Users must inspect `CombinedStatus` objects to understand the per-WEC state. This limits the value of the `.status` field for both human operators and tools such as Argo CD that rely on it for health evaluation.

The status controller lacks a mechanism to merge status data from multiple WECs into a meaningful summary for the WDS workload object.

## Goals

- Introduce an opt-in API for requesting aggregated status return.
- Provide deterministic and reproducible aggregation behavior.
- Support workload kinds for which Argo CD defines health rules.
- Supply general-purpose aggregation rules for other kinds.
- Preserve current behavior unless the user explicitly opts in.
- Maintain backward compatibility for singleton status return.

## Non-Goals

- Full semantic interpretation of arbitrary workload kinds.
- Aggregation of fields not used by health or readiness evaluation.
- Replacement of the `CombinedStatus` mechanism.

## Proposed API Enhancement

Extend the `DownsyncModulation` struct with a new field:

```
wantMultiWECReportedState: bool
```

This field indicates that when more than one WEC is selected, the status controller should aggregate status results from all selected WECs and write the aggregated value to the `.status` field of the workload object in the WDS.

This field parallels the existing `wantSingletonReportedState` option.

The enhancement requires updating both the `BindingPolicy` and `Binding` API types.

## Controller Behavior

If the user sets `wantMultiWECReportedState: true`, the controller will:

1. Identify the WECs selected for the workload object.
2. Retrieve the reported status from each selected WEC.
3. Determine whether to apply singleton copying or aggregation:
   - One WEC → copy status
   - More than one WEC → aggregate status
4. Write the resulting status into the WDS workload object.

The existing rules for singleton status return remain unchanged.

## Aggregation Approach

Aggregation rules depend on the workload kind.

### Workload Kinds with Argo CD Health Rules

For kinds with built-in Argo CD health assessment:
- Identify the subset of fields used by Argo CD.
- Aggregate readiness-related numeric fields using minimum.
- Aggregate conditions using three-value logic.
- Ignore fields not used by Argo CD’s health evaluator.

### Other Workload Kinds

- Numeric fields: use minimum across WECs.
- Boolean fields: false if any false, true only if all true.
- Conditions: three-value logic for truth value, latest timestamp, and reason/message from the newest entry.
- List/map fields: aggregate only when structurally compatible.

These rules aim to provide a practical summarization without inferring unsupported semantics.

## Data Flow

1. WECs publish status into their respective ExecutionSpaces.
2. The status controller watches these updates.
3. For each workload in the WDS, the controller collects the relevant per-WEC statuses.
4. The controller computes the aggregated output.
5. The controller updates the `.status` field of the WDS object.

The existing `CombinedStatus` objects continue to store per-WEC details and remain unaffected.

## Validation Plan

The following cases will be validated:

- **Singleton fallback**: One WEC selected → `.status` matches the WEC's reported status.
- **Multi-WEC aggregation** for:
  - Deployments (varied readiness across WECs)
  - StatefulSets
  - DaemonSets
  - Jobs (handling of `active`, `succeeded`, `failed`)
  - Arbitrary workload kinds without Argo CD rules
- **Inconsistent fields**:
  - Incompatible list/map structures
  - Mixed boolean values
  - Divergent condition timestamps
- **Argo CD integration**:
  - Aggregated `.status` leads to expected health assessments.

## Risks and Mitigations

- **Misinterpretation of numeric fields**  
  Not all numeric fields represent readiness.  
  *Mitigation:* Aggregate only fields covered by Argo CD rules for known kinds; use minimal general aggregation for unknown kinds.

- **Large numbers of WECs**  
  Aggregation cost and update frequency may increase.  
  *Mitigation:* Continue relying on existing rate-limiting and batching.

- **Inconsistent field shapes across WECs**  
  Some fields cannot be merged safely.  
  *Mitigation:* Omit fields that differ structurally.

## Expected Outcome

- Users gain a meaningful `.status` summary for workloads deployed across multiple WECs.
- Argo CD and other tools can evaluate multi-WEC workloads without reading `CombinedStatus` objects.
- Existing behavior remains unchanged for users who do not opt in.

