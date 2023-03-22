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

The placement translator is a work in progress.  Thus far, only the
"what resolver" part is implemented.  This document shows a brief
exercise of that.

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
      --allclusters-cluster string       The name of the kubeconfig cluster to use for all clusters
      --allclusters-context string       The name of the kubeconfig context to use for all clusters (default "system:admin")
      --allclusters-kubeconfig string    Path to the kubeconfig file to use for all clusters
      --allclusters-user string          The name of the kubeconfig user to use for all clusters
      --espw-cluster string              The name of the kubeconfig cluster to use for edge service provider workspace
      --espw-context string              The name of the kubeconfig context to use for edge service provider workspace
      --espw-kubeconfig string           Path to the kubeconfig file to use for edge service provider workspace
      --espw-user string                 The name of the kubeconfig user to use for edge service provider workspace
      --server-bind-address ipport       The IP address with port at which to serve /metrics and /debug/pprof/ (default :10204)
```

## Try It

The nascent placement translator can be exercised following the
scenario in [example1](../example1).  You do not need the scheduler nor
the mailbox controller.

When you get to the step of "Populate the edge service provider
workspace", it suffices to do the following.

```console
$ kubectl ws root:edge
$ kubectl create -f config/crds
$ kubectl create -f config/exports
```

At the start of Stage 3 of the exercise, just run the placement
translator from the command line.  It should look something like the
following.

```console
$ go run ./cmd/placement-translator     
I0321 01:51:26.085119   78562 shared_informer.go:282] Waiting for caches to sync for placement-translator
I0321 01:51:26.186396   78562 shared_informer.go:289] Caches are synced for placement-translator
I0321 01:51:26.186481   78562 shared_informer.go:282] Waiting for caches to sync for what-resolver
I0321 01:51:26.186492   78562 shared_informer.go:289] Caches are synced for what-resolver
I0321 01:51:26.389049   78562 main.go:119] "Receive" key="13bz53escegchz86:edge-placement-c" val=map[{APIGroup: Resource:namespaces Name:commonstuff}:{APIVersion:v1 IncludeNamespaceObject:false}]
I0321 01:51:26.390641   78562 main.go:119] "Receive" key="1b859usu9u5oultp:edge-placement-s" val=map[{APIGroup: Resource:namespaces Name:specialstuff}:{APIVersion:v1 IncludeNamespaceObject:false}]
```

Those `"Receive"` log entries show what the "what resolver" is
reporting.  It reports mappings from `ExternalName` of an
`EdgePlacement` object to the workload parts that that `EdgePlacement`
says to downsync.

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
