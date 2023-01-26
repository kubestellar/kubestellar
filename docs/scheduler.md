# Edge Scheduler

This is a component in an architecture not yet shared in detail. It is
an early milestone in edge multicluster, achieving partial
functionality and implemented in a simple crude approach that is
layered on top of transparent multicluster. In this approach the
edge-mc layer maintains a kcp workspace, called a mailbox, per
selected edge destination per EdgePlacement object. Then TMC syncs
between the mailbox workspaces and the edge destinations.

## Development Status

Currently the code is just the barest start.  

There is a start on the scheduler.  It only creates an informer on
EdgePlacement objects and logs what the informer reports.  This
controller is currently a stand-alone process. No leader election.
Not containerized.

Requires commit 4506fdc06406 (from Jan 19) of kcp-dev/kcp.

## Build

Build with ordinary go commands.

### NOTE WELL

Because https://github.com/kcp-dev/edge-mc/issues/136 is not solved,
the CRDs have been hacked by hand to state that they are NOT
namespaced, and `apigen` has been run manually; `make crds` would
overwrite this with the erroneous content.

## Try It

To exercise it, do the following steps.

Start a kcp server.  Do the remaining steps in a separate shell, with
`$KUBECONFIG` set to the admin config for that kcp server.

`kubectl ws root`

`kubectl create -f config/crds`

`kubectl create -f config/exports`

`kubectl ws \~`

`kubectl create -f config/imports`

`kubectl create -f config/samples/test-placement-1.yaml`

`kubectl ws root`

After that, a run of the controller should look like the following.

```
(base) mspreitz@mjs12 edge-mc % go run ./cmd/scheduler
I0120 16:32:14.882891   48037 main.go:127] "Found APIExport view" exportName="edge.kcp.io" serverURL="https://192.168.58.123:6443/services/apiexport/root/edge.kcp.io"
I0120 16:32:14.889551   48037 main.go:84] "Observed add" obj=&{TypeMeta:{Kind:EdgePlacement APIVersion:edge.kcp.io/v1alpha1} ObjectMeta:{Name:test-placement-1 GenerateName: Namespace: SelfLink: UID:69f046bc-1868-42b1-9c66-9c4100fe82fa ResourceVersion:703 Generation:1 CreationTimestamp:2023-01-20 16:31:13 -0500 EST DeletionTimestamp:<nil> DeletionGracePeriodSeconds:<nil> Labels:map[] Annotations:map[kcp.io/cluster:kvdk2spgmbix] OwnerReferences:[] Finalizers:[] ZZZ_DeprecatedClusterName: ManagedFields:[{Manager:kubectl-create Operation:Update APIVersion:edge.kcp.io/v1alpha1 Time:2023-01-20 16:31:13 -0500 EST FieldsType:FieldsV1 FieldsV1:{"f:spec":{".":{},"f:locationSelectors":{}}} Subresource:}]} Spec:{LocationWorkspaceSelector:{MatchLabels:map[] MatchExpressions:[]} LocationSelectors:[{MatchLabels:map[] MatchExpressions:[]}] NamespaceSelector:{MatchLabels:map[] MatchExpressions:[]} NonNamespacedObjects:[]} Status:{SpecGeneration:0 MatchingLocationCount:0}}
```
