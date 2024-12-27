# Flexible Trichotomy of Workload Objects

KubeStellar divides the content of each workload object into one of
the following three categories.

1. **Desired State**. If you think of an API object as a
   request/response inteface between a user and a server, this is what
   the user writes. Typically it is in a field called `.spec` and also
   the `.metadata.labels` and `.metadata.annotations`. In KubeStellar,
   desired state is authored in the WDS and is propagated to the WECs.

1. **Reported State**. In that simple model, this is the response from
   the server. It is typically in a field called `.status`. In
   KubeStellar, reported state is authoried in the WECs and is
   propagated to the core (to be viewed through the [combined
   status](combined-status.md) interfaces).

1. **State That Does Not Propagate**. This is typically state that
   differs from space to space (for "the same object") and is managed
   by space-specific controllers or apiservers. Examples include
   `.metadata.uid` and `.metadata.creationTimestamp`.

1. **Object Identifier**. This is not even "state" but is mentioned
   here for completeness and understanding of what is meant by "the
   same object". KubeStellar considers an object to be identified by
   the combination of four things: name, namespace (taken to be the
   empty string for cluster-scoped objects), API group, and either
   "resource" or "kind". Here "resource" and "kind" are Kubernetes
   jargon for two different but equivalent (when you restrict your
   attention to top-level objects --- i.e., excluding "subresources")
   ways of talking about a sort of object.

KubeStellar has default rules for classifying an object's contents
into these categories, and you can override these rules with
`ContentClassifier` objects.

## Default Classification Rules

1. The apiserver-managed fields do not propagate. These are the
   following fields in `.metadata`: `creationTimestamp`, `generation`,
   `managedFields`, `resourceVersion`, `selfLink` (an obsolescent
   feature of Kubernetes), `uid`.

1. These additional fields in `.metadata` do not propagate:
   `finalizers`, `generateName`,
   `labels["kubectl.kubernetes.io/last-applied-configuration"]`,
   `ownerReferences`.

1. `.status` is reported state.

1. A Kubernetes `Service` object has complicated special rules.

    1. These `.spec` fields do not propagate: `ipFamilies`,
       `externalTrafficPolicy`, `internalTrafficPolicy`,
       `ipFamilyPolicy`, `sessionAffinity`.

    1. `.spec.clusterIP` is desired state if it equals "None",
       otherwise it does not propagate.

    1. For `.spec.clusterIPs`, which is an array of strings, the only
       part that is desired state is the element (if any) "None"; the
       rest does not propagate.

1. For a Kubernetes `Job` object, the following do not propagate:
   `.spec.selector`, `.spec.suspended`,
   `.metadata.annotations["batch.kubernetes.io/job-tracking"]`, and
   the labels in `.metadata` and `.spec.template.metadata` whose name
   is "controller-uid" or "batch.kubernetes.io/controller-uid"

1. Everything else is desired state.

## Custom Classifier

You can override the default classification rules by maintaining
`ContentClassifier` objects in the WDS. Each `CustomClassifier`
applies to a specific kind/resource and API group&version.

The name of the ContentClassifier must be exactly resource.version.apiGroup
(resource.version when the apiGroup is the empty string).
This is modeled on what `kubectl` accepts.
Examples: "pods.v1", "jobs.v1.batch", "networkpolicies.v1.networking.k8s.io".

Currently there are limitations on overrides, as follows.

1. You can not add to the reported state.
2. You can not declare any part of `.status` to be desired state.

Invalid overrides will be reported in `.status.errors` of the
`ContentClassifier` and otherwise ignored.

See the definition of `ContentClassifier` in
`api/control/v1alpha1/types.go` for full details (after this feature
appears in a release we can replace this reference with a URL#fragment
for the definition on pkg.go.dev).

Following is an example of a `CustomClassifier` that says that the
`.status.ready` field of a `Job` does not propagate.

```yaml
apiVersion: control.kubestellar.io/v1alpha1
kind: ContentClassifier
metadata:
  name: jobs.v1.batch
spec:
  dontPropagatePointers:
  - [ "status", "ready" ]
```
