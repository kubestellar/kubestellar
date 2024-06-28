# Combined Status from WECs

Note on terminology: the general idea that we wish to address is returning reported state about a workload object to its WDS. At the current level of development, we equate reported state with the status section of an object --- while anticipating a more general treatment in the future.

There are two methods of returning reported state: a general one and a special case. The general method returns reported state from any number of WECs. The special case applies when the number of WECs is exactly 1, and returns the reported state into the original object in the WDS.
It may make sense to expand the latter into a general feature in the future, accommodating features such as scheduling.
## Introduction to the General Technique

The general technique involves the following ideas.

1. The way that reported state is combined is specified by the user, in a simple but powerful way modeled on SQL. This is chosen because it is a well worked out set of ideas, is widely known, and is something that we may someday want to use in our implementation. We do not need to support anything like full SQL (for any version of SQL). This proposal only involves one particular pattern of relatively simple SELECT statement, and a small subset of the expression language.

1. The specification of how to combine reported state is defined in `StatusCollector` objects. These objects are referenced in the BindingPolicies right next to the criteria for selecting workload objects. This saves users the trouble of having to write selection criteria twice. With the specification being separate, it is possible to have a library of `StatusCollectors` that can be reused across different BindingPolicies. In the future KubeStellar would provide a library of `StatusCollector` objects that cover convenient use-cases for kubernetes built-in resources such as deployments.

1. The combined reported state appears in a new kind of object, one per (workload object, BindingPolicy object) pair.

1. A user can request a list without aggregation, possibly after filtering, but certainly with a limit on list length. The expectation is that such a list makes sense only if the length of the list will be modest. For users that want access to the full reported state from each WEC for a large number of WECs, KubeStellar should have an abstraction that gives the users access --- in a functional way, not by making another copy --- to that state (which is already in the mailbox namespaces).

1. The reported state for a given workload object from a given WEC is implicitly augmented with metadata about the WEC and about the end-to-end propagation from WDS to that WEC. This extra information is available just like the regular contents of the object, for use in combining reported state.

### Relation with SQL

To a given workload object, the user can bind one or more named "status combiners". Each status combiner is analogous to an SQL SELECT statement that does the following things.

1. The SELECT statement has one input table, and that has a row per WEC that the object is prescribed to propagate to.

1. The SELECT statement can have a WHERE clause that filters out some of the rows.

1. The SELECT statement either does aggregation or does not. In the case of not doing aggregation, the SELECT statement simply has a collection of named expressions defining the columns of its output.

1. In the case of aggregation, the SELECT statement has the following.

   - A `GROUP BY` clause saying how the rows (WECs) are grouped to
     form the inputs for aggregation, in terms of named
     expressions. For convenience here, each of these named
     expressions is implicitly included in the output columns.

   - A collection of named expressions using aggregation functions to define
     additional output columns.

1. The SELECT statement has a LIMIT on the number of rows that it will yield.

## Specification of the general technique

See `types.go`.

## Examples of using the general technique

### Number of WECs

The NamedStatusCombiner would look like the following.

```yaml
name: numWECs
combinedFields:
   - name: count
     type: COUNT
```

### List of stale WECs

This produces a list of WECs for which the core has stale information.

The NamedStatusCombiner would look like the following.

```yaml
name: staleOnes
filter:
   op: Path
   path: "$.propagation.stale"
select:
   op: Path
   path: "$.inventory.name"
```

### Histogram of Pod phase

The NamedStatusCombiner would look like the following.

```yaml
name: podPhase
groupBy:
   - name: phase
     def:
        op: Path
        path: "$.status.phase"
combinedFields:
   - name: count
     type: COUNT
```

### Histogram of number of available replicas of a Deployment

This reports, for each number of available replicas, how many WECs have that number.

```yaml
name: availableReplicasHistogram
groupBy:
   - name: numAvailable
     def:
        op: Path
        path: "$.status.availableReplicas"
combinedFields:
   - name: count
     type: COUNT
```

### List of WECs where the Deployment is not fully available

```yaml
name: sadOnes
filter:
   op: Not
   args:
      - op: Equal
        args:
           - op: Path
             path: "$.spec.replicas"
           - op: Path
             path: "$.status.availableReplicas"
select:
   - name: wec
     def:
        op: Path
        path: "$.inventory.name"
```

### Full status from each WEC

This produces a listing of object status paired with inventory object name.

```yaml
name: fullStatus
select:
   - name: wec
     def:
        op: Path
        path: "$.inventory.name"
   - name: status
     def:
        op: Path
        path: "$.status"
```

## Special case for 1 WEC

This proposal refines the meaning of `BindingPolicySpec.WantSingletonReportedState` to request the special case when it applies.
