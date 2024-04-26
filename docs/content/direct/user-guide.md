# User Guide

Status of this document: it is the barest of a start. Much more needs to be written.

This document is for users of a release. Examples of using the latest stable release are in [the examples document](examples.md). This document adds information not conveyed in the examples.

## Using a pre-existing cluster as the hosting cluster

See [Using an existing hosting cluster](./hosting-cluster.md)

## When everything is not on the same machine

Thus far we can only say how to handle this when the hosting cluster is OpenShift. The problem is getting URLs that work from everywhere. OpenShift is a hosted product, your clusters have domain names that are resolvable from everywhere. In other words, if you use an OpenShift cluster as your hosting cluster then this problem is already solved.

## WEC-independent workload object transformation

KubeStellar does some transformation of workload objects on their way from WDS to WEC. First, there are transformations that are independent of the destination; these are described in this section. Second, there is customization to the WEC, described [later](#rule-based-customization).

The WEC-independent transformations are removal of certain content.

There are three categories of these transformations, as follows. They are applied in this order.

1. Transformations that are built into KubeStellar and apply to all workload objects.
1. Transformations that are built into KubeStellar and apply to specific kinds of workload objects.
1. Transformations that are configured by control objects and apply to specific kinds of workload objects.

### Transformations for all workload objects

The following are applied to every workload object.

1. Remove the following fields from `metadata`: `managedFields`, `finalizers`, `generation`, `ownerReferences`, `selfLink`, `resourceVersion`, `UID`, `generateName`.
1. Remove the annotation named `kubectl.kubernetes.io/last-applied-configuration`.
1. Remove the `status`.

## Built-in transformations of specific kinds of workload object

In a `Service` (core API group) object:

1. remove the following fields from `spec`: `ipFamilies`, `externalTrafficPolicy`, `internalTrafficPolicy`, `ipFamilyPolicy`, `sessionAffinity`. Also remove the `nodePort` field from every port unless the annotation `kubestellar.io/annotations/preserve=nodeport` is present;

1. in the `spec` remove the field `clusterIP` unless it is present with value "None".

1. in the `spec`: if the field `clusterIPs` (which holds an array of strings) is present and those strings include "None" then keep it present holding only "None", otherwise remove that field if it is present.

In a `Job` (API group `batch`) object, remove the following things.

1. `spec.selector`

1. `spec.suspended`

1. In `metadata`, the annotation named `batch.kubernetes.io/job-tracking`

1. In `metadata` _and_ in `spec.template.metadata`, the labels named `controller-uid` or `batch.kubernetes.io/controller-uid`.

### Configured transformation of workload objects

The user can configure additional transformations of workload objects by putting `CustomTransform` (in the `control.kubestellar.io` API group) objects in the WDS. Each `CustomTransform` object binds to certain workload objects and specifies certain transformations.

Currently the binding is simply by naming the workload object's API group and "resource" name in the `CustomTransform`'s `spec`. The transformations from all of the bound `CustomTransform` objects are applied to the workload object. There should be at most one `CustomTransform` object that specifies a given API group and resource.

Currently the only available transformations are removals of specified content. The content to be removed is identified by a small subset of JSONPath. In this subset, the only available operators are `.someName` and `["some string"]`. The allowed names and strings are as specified in [RFC 9535](https://www.rfc-editor.org/rfc/rfc9535.html), except that only double-quoted strings are allowed.

For example, the following `CustomTransform` object says to remove the `spec` field named `suspend` from `Job` objects (in the API group `batch`).

```yaml
kind: CustomTransform
metadata:
  name: example
spec:
  apiGroup: batch
  resource: jobs
  remove:
  - "$.spec.suspend"
```


## Rule-based customization

KubeStellar can distribute one workload object to multiple WECs, and it is common for users to need some customization to each WEC. By _rule based_ we mean that the customization is not expressed via one or more literal expressions but rather can refer to _properties_ of each WEC by property name. As KubeStellar distributes or transports a workload object from WDS to a WEC, the object can be transformed in a way that depends on those properties.

At its current level of development, KubeStellar has a simple but limited way to specify rule-based customization, called "template expansion".

### Template Expansion

Template expansion is an optional feature that a user can request on an object-by-object basis. The way to request this feature on an object is to put the following annotation on the object.

```yaml
    control.kubestellar.io/expand-templates: "true"
```

The customization that template expansion does when distributing an object from a WDS to a WEC is applied independently to each leaf string of the object and is based on the "text/template" standard pacakge of Go. The string is parsed as a template and then replaced with the result of expanding the template. Errors from this process are reported in the status field of the Binding object involved. Errors during template expansion usually produce broken YAML, in which case no corresponding object will be created in the WEC.

The data used when expanding the template are properties of the WEC. These properties are collected from the following four sources, which are listed in decreasing order of precedence.

1. The ConfigMap object, if any, that is in the namespace named "customization-properties" in the ITS and has the same name as the inventory object for the WEC. In particular, the ConfigMap string and binary data items whose name is valid as a [Go language identifier](https://go.dev/ref/spec#Identifiers) supply properties.
1. The annotations of the inventory item for the WEC supply properties if the annotation's name (AKA key) is valid as a Go language identifier.
1. The labels of the inventory item for the WEC supply properties if the label's name (AKA key) is valid as a Go language identifier.
1. There is a pre-defined property whose name is "clusterName" and whose value is the name of the inventory item (i.e., the ManagedCluster object) for the WEC.

A Binding object's `status` section has a field holding a slice of error message strings reporting user errors that arose the last time the transport controller processed that Binding, along with the `observedGeneration` reporting the `metadata.generation` that was processed. For each workload object that the Binding references: if template expansion reports errors for any destinations, the errors reported for the first such destination are included in the Binding object's status.

Any failure in any template expansion for a given Binding suppresses propagation of desired state from that Binding; the previosly propagated desired state from that Binding, if any, remains in place in the WEC.

Template expansion can only be applied when and where the un-expanded leaf strings pass the validation that the WDS applies, and can only epxress substring replacements.

For example, consider the following example workload object.

```yaml
apiVersion: logging.openshift.io/v1
kind: ClusterLogForwarder
metadata:
  name: instance
  namespace: openshift-logging
  annotations:
    control.kubestellar.io/expand-templates: "true"
spec:
  outputs:
    - name: remote-loki
      type: loki
      url: "https://my.loki.server.com/{\u007B .clusterName }}-{\u007B.clusterHash}}"
...
```

(Note: "{\u007B" is JSON for a string consisting of two consecutive left curly brackets --- which mkdocs does not have a way to quote inside a fenced code block.)

The following ConfigMap in the ITS provides a value for the `clusterHash` property.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: customization-properties
  name: virgo
data:
  clusterHash: 1001-dead-beef
...
```

When distributed to the virgo WEC, that ClusterLogForwarder would say the following.

```yaml
...
      url: "https://my.loki.server.com/virgo-1001-dead-beef"
...
```
