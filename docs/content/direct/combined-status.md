# Combined Status from WECs

Note on terminology: the general idea that we wish to address is returning reported state about a workload object to its WDS. At the current level of development, we equate reported state with the status section of an object --- while anticipating a more general treatment in the future.

There are two methods of returning reported state: a general one and a special case. The general method returns reported state from any number of WECs. The special case applies when the number of WECs is exactly 1, and returns the reported state into the original object in the WDS.

## Introduction to the General Technique

The general technique for combining reported state from WECs is built upon the following ideas:

1. The way that reported state is combined is specified by the user, in a simple but powerful way modeled on SQL. This is chosen because it is a well worked out set of ideas, is widely known, and is something that we may someday want to use in our implementation. We do not need to support anything like full SQL (for any version of SQL). This proposal only involves one particular pattern of relatively simple SELECT statement, and a different expression language (CEL, which is the most prominent expression language in the Kubernetes milieu).

1. The expressions may reference the content of the workload object as it sits in the WDS as well as the state returned from the WECs.

1. The specification of how to combine reported state is defined in `StatusCollector` objects. These objects are referenced in the `BindingPolicy` objects right next to the criteria for selecting workload objects. This saves users the trouble of having to write selection criteria twice. With the specification being separate rather than embedded, it is possible to have a library of `StatusCollectors` that can be reused across different `BindingPolicy` objects. In the future KubeStellar would provide a library of `StatusCollector` objects that cover convenient use-cases for kubernetes built-in resources such as deployments.

1. The `Binding` objects also hold references to `StatusCollector` objects. Each reference to a workload object is paired with references to all the `StatusCollectors` mentioned in all the `DownsyncObjectTestAndStatusCollection` structs that matched the workload object.

1. The combined reported state appears in a new kind of object, one per (workload object, `Binding` object) pair.

1. A user can request a list without aggregation, possibly after filtering, but certainly with a limit on list length. The expectation is that such a list makes sense only if the length of the list will be modest. For users that want access to the full reported state from each WEC for a large number of WECs, KubeStellar should have an abstraction that gives the users access --- in a functional way, not by making another copy --- to that state (which is already in the mailbox namespaces).

1. The reported state for a given workload object from a given WEC is implicitly augmented with metadata about the WEC and about the end-to-end propagation from WDS to that WEC. This extra information is available just like the regular contents of the object, for use in combining reported state.
    * The specifics of queryable objects and implicit augmentations can be found in `api/control/v1alpha1/types.go` and are specified in [Queryable Objects](#queryable-objects).

1. Errors in the user-supplied expressions are reported in relevant API objects (`StatusCollector` and `CombinedStatus`). For aggregation operations, input rows with expression errors are skipped.

### Relation with SQL

#### Overview of Relation with SQL

To a given workload object, and in the context of a given `Binding` object, the user has bound some `StatusCollector` objects. The meaning of a `StatusCollector` in the context of a (workload object, `Binding` object) pair is analogous to an SQL SELECT statement that does the following things.

1. The SELECT statement has one input table, which has a row per WEC that the Binding says the workload object should go to.

1. The SELECT statement can have a WHERE clause that filters out some of the rows.

1. The SELECT statement either does aggregation or does not. In the case of not doing aggregation, the SELECT statement simply has a collection of named expressions defining the columns of its output.

1. In the case of aggregation, the SELECT statement has the following.

    - An optional `GROUP BY` clause saying how the rows (WECs) are
      grouped to form the inputs for aggregation, in terms of named
      expressions. For convenience here, each of these named
      expressions is implicitly included in the output columns.

    - A collection of named expressions using aggregation functions to define
      additional output columns.

1. The SELECT statement has a LIMIT on the number of rows that it will yield.

#### Detailed Relation with SQL

For a given workload object, `Binding`, and `StatusCollector`: start with a table named `PerWEC`. This table has one primary key column and it holds the name of the WEC that the reported state is from. The dependent columns hold the workload object content from the WDS, the workload object content returned from the WEC, and the augmentations.

There are three forms of `StatusCollector`, equivalent to three forms of SQL statement.

##### Plain selection

When the `StatusCollector` has selection but no "GROUP BY" and no aggregation, this is equivalent to the following form of SELECT statement. The [List of stale WECs example below](#list-of-stale-wecs) is an example of this form.

```sql
SELECT <selected columns>
FROM PerWEC WHERE <filter condition>
LIMIT <limit>
```

##### Aggregation without `GROUP BY`

When there is aggregation but no plain selection and _no_ `GROUP BY`, this is equivalent to the following form of SELECT statement. The [Number of WECs example below](#number-of-wecs) is an example of this form.

```sql
SELECT <aggregation columns>
FROM PerWEC WHERE <filter condition>
LIMIT <limit>
```

##### Aggregation with `GROUP BY`

When there is `GROUP BY` and aggregation but no plain selection, this is equivalent to the following form of SELECT statement. The [Histogram of Pod phase example below](#histogram-of-pod-phase) is an example of this form.

```sql
SELECT <group-by column names>, <aggregation columns>
FROM (SELECT <group-by column 1 expr> AS <group-by column 1 name>,
             ...
             <group-by column N expr> AS <group-by column N name>,
             *
      FROM PerWEC WHERE <filter condition>)
GROUP BY <group-by column names>
LIMIT <limit>
```

When there are N `GROUP BY` columns, the result has a row for each tuple of values (v1, v2, ... v`N`) such that there exists a WEC for which (v1, v2, ... v`N`) are the values of the `GROUP BY` columns. The result has no more rows than that.

## Specification of the general technique

In `types.go` see (a) `StatusCollector`, (b) the references to those
from `DownsyncPolicyClause`, `NamespaceScopeDownsyncClause`, and
`ClusterScopeDownsyncClause`, and (c) `CombinedStatus`.

### Queryable Objects

A CEL expression within a `StatusCollector` can reference the following objects:

1. `inventory`: The inventory object for the workload object:
    - `inventory.name`: The name of the inventory object.

1. `obj`: The workload object from the WDS:
    - All fields of the workload object except the status subresource.

1. `returned`: The reported state from the WEC:
    - `returned.status`: The status section of the object returned from the WEC.

1. `propagation`: Metadata about the end-to-end propagation process:
    - `propagation.lastReturnedUpdateTimestamp`: metav1.Time of last update to any returned state.

## Examples of using the general technique

### Number of WECs

The `StatusCollector` would look like the following.

```yaml
apiVersion: control.kubestellar.io/v1alpha1
kind: StatusCollector
metadata:
  name: count-wecs
spec:
  combinedFields:
     - name: count
       type: COUNT
  limit: 10
```

To specify using that, the `BindingSpec` would reference it from the `statusCollectors` in the relevant `DownsyncPolicyClause`(s). Following is an example.

```yaml
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: example-binding-policy
spec:
  clusterSelectors:
  - matchLabels: {"location-group":"edge"}
  downsync:
  - objectSelectors:
    - matchLabels: {"app.kubernetes.io/name":"nginx"}
    statusCollectors: [ count-wecs ]
```

The analogous SQL statement would look something like the following.

```sql
SELECT COUNT(*) AS count FROM PerWEC LIMIT <something>
```

The table resulting from this would have one column and one row. The one value in this table would be the number of WECs.

Following is an example of a consequent `CombinedStatus` object.

```yaml
apiVersion: control.kubestellar.io/v1alpha1
kind: CombinedStatus
metadata:
  creationTimestamp: "2024-11-07T20:15:27Z"
  generation: 1
  labels:
    status.kubestellar.io/api-group: apps
    status.kubestellar.io/binding-policy: nginx-bindingpolicy
    status.kubestellar.io/name: nginx-deployment
    status.kubestellar.io/namespace: nginx
    status.kubestellar.io/resource: deployments
  name: 0990056b-ccbc-4c46-b0fe-366ef3a2de5e.332d2c17-7b55-44f6-9a6e-21445523c808
  namespace: nginx
  resourceVersion: "604"
  uid: cc167004-073e-4a20-9857-449f692e9643
results:
- columnNames:
  - count
  name: count-wecs
  rows:
  - columns:
    - float: "2"
      type: Number
```

### Histogram of Pod phase

The `spec` of the `StatusCollector` would look like the following.

```yaml
  groupBy:
     - name: phase
       def: returned.status.phase
  combinedFields:
     - name: count
       type: COUNT
```

The analogous SQL statement would look something like the following.

```sql
SELECT phase, COUNT(*) AS count
FROM (SELECT <SQL expression for returned.status.phase> AS phase, *
      FROM PerWEC)
GROUP BY phase
LIMIT <something>
```

The result would have two columns, holding a phase value and a count. The number of rows equals the number of different values of `returned.status.phase` that appear among the WECs. For each row (P, N): P is a phase value that appears in at least one WEC, and N is the number of WECs where the phase value is P.

### Histogram of number of available replicas of a Deployment

This reports, for each number of available replicas, how many WECs have that number. The `spec` of the `CombinedStatus` would look like the following.

```yaml
  groupBy:
     - name: numAvailable
       def: returned.status.availableReplicas
  combinedFields:
     - name: count
       type: COUNT
```

### List of WECs where the Deployment is not as available as desired

The `spec` of the `CombinedStatus` would look like the following.

```yaml
  filter: "obj.spec.replicas != returned.status.availableReplicas"
  select:
     - name: wec
       def: inventory.name
```

### Full status from each WEC with information retrieval time

The `spec` of the `CombinedStatus` would look like the following.
This produces a listing of object status paired with inventory object name.

```yaml
  select:
     - name: wec
       def: inventory.name
     - name: status
       def: returned.status
     - name: retrievalTime
       def: propagation.lastReturnedUpdateTimestamp
```

## Special case for 1 WEC

When a workload object is distributed from a WDS to exactly one WEC,
the reported state from that WEC can be returned into the copy of the
workload object in the WDS. The design of most kinds of Kubernetes API
object implicitly assumes that the object exists and has its defined
effect in only one cluster. That is why it makes sense to return the
reported state from the WEC to the WDS only when the object goes to
exactly one WEC.

As mentioned above, currently KubeStellar equates "reported state"
with the `.status` section of the API object.

The user has to specifically request this last step of `.status`
propagation. This is done in an optional boolean field, named
`wantSingletonReportedState`, in a `DownsyncPolicyClause` in a
`BindingPolicy`. This is one of the three kinds of downsync
modulations that a policy clause can associate with the matching
workload objects. In a `Binding`, this same optional boolean field
appears in `NamespaceScopeDownsyncClause` and
`ClusterScopeDownsyncClause`. If multiple policy clauses in a
`BindingPolicy` match a given workload object, the settings for
`.wantSingletonReportedState` are combined by OR to get the one
boolean value that appears with the reference to the object in the
corresponding `Binding`. If multiple `Binding` objects in one WDS
reference a given workload object, there is another level of
multiplicity to consider. Read on.

For a given workload object and WDS, we say that "singleton status
return is requested" if and only if there exists at least one
BindingPolicy or Binding that has `wantSingletonReportedState==true`
in a clause that matches/references the workload object.

The "qualified WEC set" of a given workload object in a given WDS is
the set of WECs that are associated with that workload object by at
least one BindingPolicy or Binding that has
`wantSingletonReportedState==true` in a clause that matches/references
the workload object.

For a given workload object in a given WDS, while singleton status
return is requested, KubeStellar maintains a label on the object whose
name (key) is `kubestellar.io/executing-count` and whose value is a
string representation of the size of the qualified WEC set of that
object.  While singleton status return is _not_ requested, KubeStellar
suppresses the existence of a label with that name (key).  While
singleton status return is requested _and_ the size of the qualified
WEC set is 1, KubeStellar propagates the object's `.status` from that
WEC to the `.status` section of the object in the WDS.  While either
singleton status return is NOT requested or the size of the qualified
WEC set is NOT 1, there is nothing in the `.status` of the object in
the WDS that was propagated there from a WEC by KubeStellar.