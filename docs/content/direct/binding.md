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
object selection predicate and also three kinds of information that
modulate the downsync. Note that each such clause must have at least
one field specifying part of the workload selection predicate.

For more definitional details about a `BindingPolicy`, see [the API reference](https://pkg.go.dev/github.com/kubestellar/kubestellar@v{{ config.ks_latest_release }}/api/control/v1alpha1#BindingPolicy){# readers of the unrendered sources should see [the Go source](../../../api/control/v1alpha1/types.go) instead #}.

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
