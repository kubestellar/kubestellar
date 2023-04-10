---
title: "Placement Translator"
date: 2023-03-21
weight: 4
description: >
---

{{% pageinfo %}}
The placement translator runs in the edge service provider workspace and translates EMC placement problems to TMC placement problems.
{{% /pageinfo %}}

## Status

The placement translator is a work in progress.  The functionality of
maintaining syncer config objects is starting to work, but is still
rough around the edges.  The functionality of propagating objects from
workload management workspace to mailbox workspace is not there yet.
This document includes a brief exercise of the current code.

## Additional Design Details

The placement translator maintains one `SyncerConfig` object in each
mailbox workspace.  That object is named `the-one`.  Other
`SyncerConfig` objects may exist; the placement translator ignores
them.

## Usage

The placement translator needs two kube client configurations.  One
points to the edge service provider workspace and provides authority
to write into the mailbox workspaces.  The other points to the kcp
server base (i.e., does not identify a particular logical cluster nor
`*`) and is authorized to read all clusters.  In the kubeconfig
created by `kcp start` the latter is satisfied by the context named
`system:admin`.

The command line flags, beyond the basics, are as follows.  For a
string parameter, if no default is explicitly stated then the default
is the empty string, which usually means "not specified here".  For
both kube client configurations, the usual rules apply: first consider
command line parameters, then `$KUBECONFIG`, then `~/.kube/config`.

```console
      --allclusters-cluster string       The name of the kubeconfig cluster to use for access to all clusters
      --allclusters-context string       The name of the kubeconfig context to use for access to all clusters (default "system:admin")
      --allclusters-kubeconfig string    Path to the kubeconfig file to use for access to all clusters
      --allclusters-user string          The name of the kubeconfig user to use for access to all clusters
      --espw-cluster string              The name of the kubeconfig cluster to use for access to the edge service provider workspace
      --espw-context string              The name of the kubeconfig context to use for access to the edge service provider workspace
      --espw-kubeconfig string           Path to the kubeconfig file to use for access to the edge service provider workspace
      --espw-user string                 The name of the kubeconfig user to use for access to the edge service provider workspace
      --server-bind-address ipport       The IP address with port at which to serve /metrics and /debug/pprof/ (default :10204)
```

## Try It

The nascent placement translator can be exercised following the
scenario in [example1](../example1).  You will need to run the
scheduler and mailbox controller long enough for them to create what
this scenario calls for, but they can be terminated after that.

When you get to the step of "Populate the edge service provider
workspace", it suffices to do the following.

```console
$ kubectl ws root:espw
$ kubectl create -f config/crds
$ kubectl create -f config/exports
```

Continue to follow the steps until the start of Stage 3 of the
exercise.  Because the mailbox controller does not yet install the
needed `APIBinding` objects into the mailbox workspaces, you will have
to do that by hand.  In each mailbox workspace, do the following.

```shell
kubectl create -f - <<EOF
apiVersion: apis.kcp.io/v1alpha1
kind: APIBinding
metadata:
  name: bind-edge
spec:
  reference:
    export:
      name: edge.kcp.io
      path: root:espw
EOF
```

Next make sure you run `kubectl ws root:espw` to enter the edge
service provider workspace, then just run the placement translator
from the command line.  It should start out like the following and
continue with more log messages (possibly including some complaints,
which do not necessarily indicate real problems because the subsequent
success is not logged so profligately).

```console
$ go run ./cmd/placement-translator

I0410 01:05:03.920954   27116 shared_informer.go:282] Waiting for caches to sync for placement-translator
I0410 01:05:04.024390   27116 shared_informer.go:289] Caches are synced for placement-translator
I0410 01:05:04.024800   27116 shared_informer.go:282] Waiting for caches to sync for what-resolver
I0410 01:05:04.024814   27116 shared_informer.go:289] Caches are synced for what-resolver
I0410 01:05:04.025051   27116 shared_informer.go:282] Waiting for caches to sync for where-resolver
I0410 01:05:04.025067   27116 shared_informer.go:289] Caches are synced for where-resolver
I0410 01:05:04.026845   27116 main.go:157] "Put" map="where" key="24oofcs9mbj9j4nd:edge-placement-c" val="[&{{SinglePlacementSlice edge.kcp.io/v1alpha1} {edge-placement-c    00f4c242-41bc-48f4-800f-2b41e42c94ea 1489 4 2023-04-09 19:19:59 -0400 EDT <nil> <nil> map[] map[kcp.io/cluster:24oofcs9mbj9j4nd] [{edge.kcp.io/v1alpha1 EdgePlacement edge-placement-c 14e26ba1-c255-421a-8a21-bf1e72e91873 <nil> <nil>}] []  [{scheduler Update edge.kcp.io/v1alpha1 2023-04-09 19:19:59 -0400 EDT FieldsV1 {\"f:destinations\":{},\"f:metadata\":{\"f:ownerReferences\":{\".\":{},\"k:{\\\"uid\\\":\\\"14e26ba1-c255-421a-8a21-bf1e72e91873\\\"}\":{}}}} }]} [{2yv5njwiqbnakwd7 location-g sync-target-g 39afc608-de3c-4faa-bbca-d72cd42d6c6c} {2yv5njwiqbnakwd7 location-f sync-target-f 1dffad3f-8783-420b-95b8-1da7c854302e}]}]"
I0410 01:05:04.027469   27116 main.go:157] "Put" map="where" key="2rdfbafqhpt3cekf:edge-placement-s" val="[&{{SinglePlacementSlice edge.kcp.io/v1alpha1} {edge-placement-s    1ddd6510-d98a-430b-a130-319f151c1da3 1487 1 2023-04-09 19:19:59 -0400 EDT <nil> <nil> map[] map[kcp.io/cluster:2rdfbafqhpt3cekf] [{edge.kcp.io/v1alpha1 EdgePlacement edge-placement-s 6d459327-59a5-4874-b670-fa254aaeb79c <nil> <nil>}] []  [{scheduler Update edge.kcp.io/v1alpha1 2023-04-09 19:19:59 -0400 EDT FieldsV1 {\"f:destinations\":{},\"f:metadata\":{\"f:ownerReferences\":{\".\":{},\"k:{\\\"uid\\\":\\\"6d459327-59a5-4874-b670-fa254aaeb79c\\\"}\":{}}}} }]} [{2yv5njwiqbnakwd7 location-g sync-target-g 39afc608-de3c-4faa-bbca-d72cd42d6c6c}]}]"
I0410 01:05:04.229392   27116 main.go:157] "Put" map="what" key="24oofcs9mbj9j4nd:edge-placement-c" val=map[{APIGroup: Resource:namespaces Name:commonstuff}:{APIVersion:v1 IncludeNamespaceObject:false}]
I0410 01:05:04.231417   27116 main.go:157] "Put" map="what" key="2rdfbafqhpt3cekf:edge-placement-s" val=map[{APIGroup: Resource:namespaces Name:specialstuff}:{APIVersion:v1 IncludeNamespaceObject:false}]
```

The "Put" log entries with `map="what"` show what the "what resolver" is
reporting.  This reports mappings from `ExternalName` of an
`EdgePlacement` object to the workload parts that that `EdgePlacement`
says to downsync.

The "Put" log entries with `map="where"` show the
`SinglePlacementSlice` objects associated with each `EdgePlacement`.


Next, using a separate shell, examine the SyncerConfig objects in the
mailbox workspaces.  Make sure to use the same kubeconfig as you use
to run the placement translator, or any other that is pointed at the
edge service provider workspace. The following with switch the focus
to mailbox workspace(s).

You can get a listing of mailbox workspaces, while in the edge service
provider workspace, as follows.

```console
$ kubectl get Workspace
NAME                                                       TYPE        REGION   PHASE   URL                                                     AGE
2yv5njwiqbnakwd7-mb-1dffad3f-8783-420b-95b8-1da7c854302e   universal            Ready   https://192.168.58.123:6443/clusters/2s895m084fo48yyv   5h50m
2yv5njwiqbnakwd7-mb-39afc608-de3c-4faa-bbca-d72cd42d6c6c   universal            Ready   https://192.168.58.123:6443/clusters/27eh2l0czvz5eqa4   5h50m
```

Next switch to one of the mailbox workspaces (in my case I picked the
one for the guilder cluster) and examine the `SyncerConfig` object.
That should look like the following.

```console
$ kubectl ws 2yv5njwiqbnakwd7-mb-39afc608-de3c-4faa-bbca-d72cd42d6c6c
Current workspace is "root:espw:2yv5njwiqbnakwd7-mb-39afc608-de3c-4faa-bbca-d72cd42d6c6c" (type root:universal).

$ kubectl get syncerconfigs the-one -o yaml
apiVersion: edge.kcp.io/v1alpha1
kind: SyncerConfig
metadata:
  annotations:
    kcp.io/cluster: 27eh2l0czvz5eqa4
  creationTimestamp: "2023-04-10T05:05:04Z"
  generation: 2
  name: the-one
  resourceVersion: "1641"
  uid: abb39edc-b3f8-4b10-aa49-888b98ca5110
spec:
  namespaceScope:
    namespaces:
    - commonstuff
    - specialstuff
    resources:
    - apiVersion: v1
      group: ""
      resource: resourcequotas
    - apiVersion: v1
      group: ""
      resource: events
    - apiVersion: v1
      group: networking.k8s.io
      resource: ingresses
    - apiVersion: v1
      group: ""
      resource: serviceaccounts
    - apiVersion: v1
      group: ""
      resource: configmaps
    - apiVersion: v1
      group: ""
      resource: secrets
    - apiVersion: v1
      group: ""
      resource: pods
    - apiVersion: v1
      group: rbac.authorization.k8s.io
      resource: roles
    - apiVersion: v1
      group: rbac.authorization.k8s.io
      resource: rolebindings
    - apiVersion: v1
      group: ""
      resource: services
    - apiVersion: v1
      group: apps
      resource: deployments
    - apiVersion: v1
      group: coordination.k8s.io
      resource: leases
    - apiVersion: v1alpha1
      group: edge.kcp.io
      resource: customizers
    - apiVersion: v1
      group: events.k8s.io
      resource: events
    - apiVersion: v1
      group: ""
      resource: limitranges
status: {}
```


At this point you might veer off from the example sceario and try
tweaking things.  For example, try deleting an EdgePlacement as
follows.

```console
$ kubectl ws root:work-c
Current workspace is "root:work-c".
$ kubectl delete EdgePlacement edge-placement-c
edgeplacement.edge.kcp.io "edge-placement-c" deleted
```

That will cause the placement translator to log updates, as follows.

```
I0410 01:20:15.024734   27116 main.go:157] "Put" map="what" key="24oofcs9mbj9j4nd:edge-placement-c" val=map[]
I0410 01:20:15.102275   27116 main.go:161] "Delete" map="where" key="24oofcs9mbj9j4nd:edge-placement-c"
```
