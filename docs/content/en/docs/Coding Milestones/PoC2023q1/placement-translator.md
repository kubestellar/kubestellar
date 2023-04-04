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

The placement translator is a work in progress.  Thus far, the only
substantial part that has been implemented is the "what resolver".
The "where resolver", which merely reads what the scheduler has
writen, is also implemented.  This document includes  a brief exercise of
the current code.

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
scenario in [example1](../example1).  You do not need the scheduler nor
the mailbox controller.

When you get to the step of "Populate the edge service provider
workspace", it suffices to do the following.

```console
$ kubectl ws root:espw
$ kubectl create -f config/crds
$ kubectl create -f config/exports
```

Continue to follow the steps until the start of Stage 3 of the exercise.
First make sure you run `kubectl ws root:espw` to enter the service provider
workspace, then just run the placement translator from the command line.  
It should look something like the following.

```console
$ go run ./cmd/placement-translator

I0330 02:02:14.936347   70807 shared_informer.go:282] Waiting for caches to sync for placement-translator

I0330 02:02:15.037858   70807 shared_informer.go:289] Caches are synced for placement-translator

I0330 02:02:15.038104   70807 shared_informer.go:282] Waiting for caches to sync for what-resolver

I0330 02:02:15.038121   70807 shared_informer.go:289] Caches are synced for what-resolver

I0330 02:02:15.038162   70807 shared_informer.go:282] Waiting for caches to sync for where-resolver

I0330 02:02:15.038171   70807 shared_informer.go:289] Caches are synced for where-resolver

I0330 02:02:15.038848   70807 main.go:123] "Receive" key="2fx5g4vkijmda9ci:edge-placement-c" val="[&{{SinglePlacementSlice edge.kcp.io/v1alpha1} {edge-placement-c    6d4c01e1-10cd-48df-881a-29a13b65627e 1103 1 2023-03-30 01:51:15 -0400 EDT <nil> <nil> map[] map[kcp.io/cluster:2fx5g4vkijmda9ci] [{edge.kcp.io/v1alpha1 EdgePlacement edge-placement-c cfd3b279-1c2f-42e4-9100-88217a1cc62c <nil> <nil>}] []  [{main Update edge.kcp.io/v1alpha1 2023-03-30 01:51:15 -0400 EDT FieldsV1 {\"f:destinations\":{},\"f:metadata\":{\"f:ownerReferences\":{\".\":{},\"k:{\\\"uid\\\":\\\"cfd3b279-1c2f-42e4-9100-88217a1cc62c\\\"}\":{}}}} }]} []}]"

I0330 02:02:15.038931   70807 main.go:123] "Receive" key="315mw7yqh3mzqhnd:edge-placement-s" val="[&{{SinglePlacementSlice edge.kcp.io/v1alpha1} {edge-placement-s    b483cf30-f2ee-4128-936f-3927947c53e7 1104 1 2023-03-30 01:46:27 -0400 EDT <nil> <nil> map[] map[kcp.io/cluster:315mw7yqh3mzqhnd] [{edge.kcp.io/v1alpha1 EdgePlacement edge-placement-s cd897d34-3ed1-4d8b-8429-b39109cb543c <nil> <nil>}] []  [{main Update edge.kcp.io/v1alpha1 2023-03-30 01:46:27 -0400 EDT FieldsV1 {\"f:destinations\":{},\"f:metadata\":{\"f:ownerReferences\":{\".\":{},\"k:{\\\"uid\\\":\\\"cd897d34-3ed1-4d8b-8429-b39109cb543c\\\"}\":{}}}} }]} []}]"

I0330 02:02:15.243791   70807 main.go:123] "Receive" key="2fx5g4vkijmda9ci:edge-placement-c" val=map[{APIGroup: Resource:namespaces Name:commonstuff}:{APIVersion:v1 IncludeNamespaceObject:false}]

I0330 02:02:15.243929   70807 main.go:123] "Receive" key="315mw7yqh3mzqhnd:edge-placement-s" val=map[{APIGroup: Resource:namespaces Name:specialstuff}:{APIVersion:v1 IncludeNamespaceObject:false}]
```

The shorter `"Receive"` log entries show what the "what resolver" is
reporting.  This report mappings from `ExternalName` of an
`EdgePlacement` object to the workload parts that that `EdgePlacement`
says to downsync.

The longer `"Receive"` log entries show the `SinglePlacementSlice`
objects associated with each `EdgePlacement`.


At this point you might veer off from the example sceario and try
tweaking things.  For example, try deleting an EdgePlacement as
follows.

```console
$ kubectl ws root:work-c
Current workspace is "root:work-c".
$ kubectl delete EdgePlacement edge-placement-c
edgeplacement.edge.kcp.io "edge-placement-c" deleted
```

That will cause the placement translator to log a new mapping, as
follows.

```
I0321 02:11:49.381442   78678 main.go:119] "Receive" key="13bz53escegchz86:edge-placement-c" val=map[]
```
