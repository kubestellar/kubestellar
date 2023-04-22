---
title: "2023q1 PoC Scripts"
linkTitle: "2023q1 PoC Scripts"
weight: 100
---

There are some scripts that automate small steps in the process of
using this PoC.

## Creating SyncTarget/Location pairs

In this PoC, the interface between infrastructure and workload
management is inventory API objects.  Specifically, for each edge
cluster there is a unique pair of SyncTarget and Location objects in a
so-called inventory management workspace.  The following script helps
with making that pair of objects.  Invoke it when your current
workspace is your chosen inventory management workspace.

```console
$ scripts/ensure-location.sh -h
scripts/ensure-location.sh usage: objname labelname=labelvalue...

$ kubectl ws root:imw-1
Current workspace is "root:imw-1".

$ scripts/ensure-location.sh demo1 foo=bar the-word=the-bird
synctarget.workload.kcp.io/demo1 created
location.scheduling.kcp.io/demo1 created
synctarget.workload.kcp.io/demo1 labeled
location.scheduling.kcp.io/demo1 labeled
synctarget.workload.kcp.io/demo1 labeled
location.scheduling.kcp.io/demo1 labeled
```

The above example shows using this script to create a SyncTarget and a
Location named `demo1` with labels `foo=bar` and `the-word=the-bird`.
This was equivalent to the following commands.

```shell
kubectl create -f -<<EOF
apiVersion: workload.kcp.io/v1alpha1
kind: SyncTarget
metadata:
  name: demo1
  labels:
    id: demo1
    foo: bar
    the-word: the-bird
---
apiVersion: scheduling.kcp.io/v1alpha1
kind: Location
metadata:
  name: demo1
  labels:
    foo: bar
    the-word: the-bird
spec:
  resource: {group: workload.kcp.io, version: v1alpha1, resource: synctargets}
  instanceSelector:
    matchLabels: {"id":"demo1"}
EOF
```

This script operates in an idempotent style.  It looks at the current
state and makes whatever changes are needed.  Caveat: it does not cast
a skeptical eye on the spec of a pre-existing Location.

## Removing SyncTarget/Location pairs

The following script undoes whatever remains from a corresponding
usage of `ensure-location.sh`.  Invoke it with the inventory
management workspace current.

```console
$ scripts/remove-location.sh -h
scripts/remove-location.sh usage: objname

$ kubectl ws root:imw-1
Current workspace is "root:imw-1".

$ scripts/remove-location.sh demo1
synctarget.workload.kcp.io "demo1" deleted
location.scheduling.kcp.io "demo1" deleted

$ scripts/remove-location.sh demo1

$ 
```

## Syncer preparation and installation

The syncer runs in each edge cluster and also talks to the
corresponding mailbox workspace.  In order for it to be able to do
that, there is some work to do in the mailbox workspace to create a
ServiceAccount for the syncer to authenticate as and create RBAC
objects to give the syncer the privileges that it needs.  The
following script does those things and also outputs YAML to be used to
install the syncer in the edge cluster.  Invoke this script with the
edge service provider workspace current.

```console
$ scripts/mailbox-prep.sh -h
scripts/mailbox-prep.sh usage: (-o file_pathname | --syncer-image container_image_ref )* synctarget_name

$ kubectl ws root:espw
Current workspace is "root:espw".

$ scripts/mailbox-prep.sh demo1
Current workspace is "root:espw:4yqm57kx0m6mn76c-mb-406c54d1-64ce-4fdc-99b3-cef9c4fc5010" (type root:universal).
Cloning into 'build/syncer-kcp'...
remote: Enumerating objects: 47989, done.
remote: Total 47989 (delta 0), reused 0 (delta 0), pack-reused 47989
Receiving objects: 100% (47989/47989), 18.59 MiB | 12.96 MiB/s, done.
Resolving deltas: 100% (31297/31297), done.
branch 'emc' set up to track 'origin/emc'.
Switched to a new branch 'emc'
hack/verify-go-versions.sh
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build  -ldflags="-X k8s.io/client-go/pkg/version.gitCommit=11d1f3a8 -X k8s.io/client-go/pkg/version.gitTreeState=clean -X k8s.io/client-go/pkg/version.gitVersion=v1.24.3+kcp-v0.0.0-11d1f3a8 -X k8s.io/client-go/pkg/version.gitMajor=1 -X k8s.io/client-go/pkg/version.gitMinor=24 -X k8s.io/client-go/pkg/version.buildDate=2023-04-22T06:00:42Z -X k8s.io/component-base/version.gitCommit=11d1f3a8 -X k8s.io/component-base/version.gitTreeState=clean -X k8s.io/component-base/version.gitVersion=v1.24.3+kcp-v0.0.0-11d1f3a8 -X k8s.io/component-base/version.gitMajor=1 -X k8s.io/component-base/version.gitMinor=24 -X k8s.io/component-base/version.buildDate=2023-04-22T06:00:42Z -extldflags '-static'" -o bin ./cmd/kubectl-kcp
ln -sf kubectl-workspace bin/kubectl-workspaces
ln -sf kubectl-workspace bin/kubectl-ws
Creating service account "kcp-edge-syncer-demo1-28at01r3"
Creating cluster role "kcp-edge-syncer-demo1-28at01r3" to give service account "kcp-edge-syncer-demo1-28at01r3"

 1. write and sync access to the synctarget "kcp-edge-syncer-demo1-28at01r3"
 2. write access to apiresourceimports.

Creating or updating cluster role binding "kcp-edge-syncer-demo1-28at01r3" to bind service account "kcp-edge-syncer-demo1-28at01r3" to cluster role "kcp-edge-syncer-demo1-28at01r3".

Wrote physical cluster manifest to demo1-syncer.yaml for namespace "kcp-edge-syncer-demo1-28at01r3". Use

  KUBECONFIG=<pcluster-config> kubectl apply -f "demo1-syncer.yaml"

to apply it. Use

  KUBECONFIG=<pcluster-config> kubectl get deployment -n "kcp-edge-syncer-demo1-28at01r3" kcp-edge-syncer-demo1-28at01r3

to verify the syncer pod is running.
```

On the first usage this script will `git clone` the repo that has the
source for this plugin and build it locally.  That will require `go`
version 1.19 to be on your `$PATH`.

Once that script has run, the YAML for the objects to create in the
edge cluster is in your chosen output file.  The default for the
output file is the name of the SyncTarget object with "-syncer.yaml"
appended.

Create those objects with a command like the following; adjust as
needed to configure `kubectl` to modify the edge cluster and read your
chosen output file.

```shell
KUBECONFIG=$demo1_kubeconfig kubectl apply -f demo1-syncer.yaml
```
