# Binding workload with WEC

This document is about associating WECs with workload objects. The
primary concept is sometimes called "downsync", which confusingly
refers to both the propagation and [transformation](transforming.md)
of desired state from core to WECs _and_ the propagation and
summarization of reported state from WECs to core.

## Binding Basics

The user controls downsync primarily through API objects of kinds
`BindingPolicy` and `Binding`. These go in a WDS and associate
workload objects in that WDS with WECs, along with adding some
modulations on how downsync is done.

`BindingPolicy` is a higher level concept than `Binding`. KubeStellar
has a controller that translates each `BindingPolicy` to a
`Binding`. A user _could_ eschew the `BindingPolicy` and directly
maintain a `Binding` object or let a different controller maintain the
`Binding` object (TODO: check that this is true). The `Binding` object
shows which workload objects and which WECs matched the predicates in
the `BindingPolicy` and so is also useful as feedback to the user
about that.

## BindingPolicy

The `spec` of a `BindingPolicy` has two predicates that (1) identify a
subset of the WECs in the inventory of the ITS associated with the WDS
and (2) identify a subset of the workload objects in the WDS. The
primary function of the `BindingPolicy` is to assert the desired
association between (1) and (2). A `BindingPolicy` can also add some
modulations on how those workload objects are downsynced to/from those
WECs.

The WEC-selecting predicate is an array of label selectors in
`spec.clusterSelectors`. These label selectors test the labels of the
inventory objects describing the WECs. The bound WECs are the ones
whose inventory object passes at least one of the the label selectors
in `spec.clusterSelectors`.

The workload object selection predicate is in `spec.downsync`, which
holds a list of `DownsyncPolicyClause`s; each includes both a workload
object selection predicate and also two kinds of information that
modulate the downsync.

The workload object selection part of a `DownsyncPolicyClause`
consists of the following fields, all of which are optional. A
workload object matches a `DownsyncPolicyClause` if that object
matches all of these selection fields; the default value for each
field specifies a test that every object passes. Note, however, that
not every API object in a WDS is considered to be a _workload_
object. The KubeStellar control objects and `CombinedStatus` objects
are not considered to be workload objects. (TODO: also talk about the
other objects that never match).

- `apiGroup`: an optional string that identifies the API group of the
  object. As is usual in Kubernetes, the empty string identifies the
  _core_ API group. Omitting this string means to match every API
  group.

- `resources`: an array of strings containing names of Kubernetes
  "resources". **BEWARE**: The Kubernetes world is inconsistent in its
  usage of the word "resource". In much --- but not all --- of the
  code, the word "resource" is roughly analogous to a "kind" of object
  --- and specifically _not_ an individual object. In some of the code
  and much of the documentation, the word "resource" refers to an
  individual object. Some authors hold with the wider web world, in
  which the word "resource" means anything with a URI/URL (Kubernetes
  object kinds _and_ individuals both have URLs). In this position in
  KubeStellar we take the first view: by "resource" here we mean the
  Kubernetes concept that is roughly analogous to "kind" and the names
  are lowercase and plural. Examples of resources are `namespaces` and
  `deployments`. KubeStellar also allows this list to include `*`
  meaning to match every resource; in this case the list should _only_
  contain that value. The empty array --- which is the default value
  --- is treated specially, it matches every resource.

- `namespaces`: an array of strings containing names of
  namespaces. Again, a value of `*` matches every namespace and should
  be the only thing in the array if it is present. `*` also matches
  cluster-scoped objects.  The empty array --- which is the default
  value --- is treated specially, it matches every object.

- `namespaceSelectors`: an array of label selectors that test the
  labels of a `Namespace` object --- specifically, the `Namespace`
  object of the namespace of the workload object being tested. For
  example, when a `Deployment` object has namespace `foo`, this test
  is applied to the `Namespace` object named `foo`. The `Namespace`
  object --- and thus the workload object having that namespace ---
  passes this test if the `Namespace`'s labels pass _any_ of the label
  selectors in this array. Again, the empty list is the default value
  and is given the exceptional meaning of matching everything.

- `objectSelectors`: another array of label selectors, this one
  applying to the labels of the workload object in question. Again,
  the workload object passes this test if the object's labels match
  _any_ of the label selectors. Again, the empty list is the default
  value and is given the exceptional meaning of matching everything.

- `objectNames`: an array of object names. A workload object passes
  this test if its name is in this array, or this array contains `*`,
  or this array is empty (which is the default value).

A `DownsyncPolicyClause` also contains the following two fields to
modulate the downsync.

- `statusCollectors`: an array of references to `StatusCollector`
  objects that specify how status from the WECs is to be combined and
  reported in `CombinedStatus` objects in the WDS. See [Combining
  returned status](combined-status.md) for more details.

- `createOnly`: a boolean that, when true, suppresses on-going
  maintenance of the object in the WEC. When this boolean is `true`,
  KubeStellar will only (a) create the object in the WEC if the
  object's presence is desired but the object is not yet present there
  and (b) delete the object from the WEC when the object's presence is
  not desired but the object currently exists there. This is a blunt
  instrument for coping with objects that that have conflicting
  authorities writing to the WDS and to the WEC. `Job` objects are an
  example. The application arm of OCM --- which KubeStellar currently
  uses for transport --- normally insists that the whole of the `spec`
  in the WEC should _equal_ the whole of the `spec` in the
  WDS. However, a Kubernetes cluster (such as the WEC) normally adds
  content (e.g., the pod selector, which the user should not supply)
  to the `spec`. Without the `createOnly` option, there would be a
  controller fight between OCM transport and the WEC's normal
  behavior; this controller fight is known to wedge the OCM
  transport's handling associated with the relevant `Binding`. _With_
  `createOnly`, the controller fight is called off: the OCM transport
  will _not_ modify the `spec` of a `Job` object that already exists
  in the WEC. **NOTE**: `createOnly: true` is not fully supported yet
  --- DO NOT RELY ON IT.

When multiple `DownsyncPolicyClause` of one `BindingPolicy` match a
given workload object, their modulations are combined as follows when
composing the corresponding `Binding` object.

- Their `StatusCollector` reference sets are combined by union.
- Their `createOnly` booleans are combined by OR.

The `spec` of a `BindingPolicy` also holds an optional boolean
`wantSingletonReportedState`. Setting this `true` indicates both (a)
an expectation that only one WEC matches the WEC selection predicate
and (b) for each selected workload object, the reported state from the
one WEC should be propagated back to the workload object in the
WDS. **NOTE**: regarding the questions of what this setting means when
multiple `BindingPolicy` objects select a given workload object and/or
when multiple WECs are selected, the implementation is in flux and the
user should not present such situations.

Following is an example of a `BindingPolicy` object, used in the
end-to-end test of `createOnly` functionality.

```yaml
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: nginx
spec:
  clusterSelectors:
  - matchLabels:
      location-group: edge
  downsync:
  - objectSelectors:
    - matchLabels:
        app.kubernetes.io/name: nginx
    resources:
    - namespaces
  - createOnly: true
    objectSelectors:
    - matchLabels:
        app.kubernetes.io/name: nginx
    resources:
    - deployments
```

## Binding

TODO: write this
