---
title: "Edge Scheduler"
date: 2023-03-27
weight: 4
description: >
---

{{% pageinfo %}}
The edge scheduler monitors the EdgePlacement, Location, and SyncTarget objects and maintains the results of matching.
{{% /pageinfo %}}

#### Steps to try the edge scheduler

open a terminal window and clone the latest kcp-edge source:
```console
git clone https://github.com/kcp-dev/edge-mc
```

clone the v0.11.0 branch kcp source:
```console
git clone -b v0.11.0 https://github.com/kcp-dev/kcp.git
```

run kcp (kcp will spit out tons of information and stay running in this terminal window)
```console
kcp start
```

open another terminal window and point `$KUBECONFIG` to the admin kubeconfig for that kcp server.
```console
export KUBECONFIG=~/kcp/.kcp/admin.kubeconfig
```

Use workspace `root:edge` as the edge service provider workspace.
```console
$ kubectl ws root
$ kubectl ws create edge --enter
```

Install CRDs and APIExport.
```console
$ kubectl apply -f config/crds/ -f config/exports/
```

Use user home workspace (`~`) as the workload management workspace.
```console
$ kubectl ws ~
```

Bind APIs.
```console
$ kubectl apply -f config/imports/
```

Go to `root:edge` workspace and run the edge scheduler.
```console
$ kubectl ws root:edge
go run cmd/scheduler/main.go --kcp-kubeconfig=<path to kcp admin kubeconfig> -v <verbosity (default 2)>
```

The outputs from the edge scheduler should be similar to:
```console
I0327 17:14:42.222112   51241 scheduler.go:243] "Found APIExport view" exportName="edge.kcp.io" serverURL="https://192.168.1.54:6443/services/apiexport/291lkbsqq181xfng/edge.kcp.io"
I0327 17:14:42.225075   51241 scheduler.go:243] "Found APIExport view" exportName="scheduling.kcp.io" serverURL="https://192.168.1.54:6443/services/apiexport/root/scheduling.kcp.io"
I0327 17:14:42.226954   51241 scheduler.go:243] "Found APIExport view" exportName="workload.kcp.io" serverURL="https://192.168.1.54:6443/services/apiexport/root/workload.kcp.io"
I0327 17:14:42.528573   51241 controller.go:201] "starting controller" controller="edge-scheduler"
```

Use workspace `root:compute` as the inventory management workspace.
```console
$ kubectl ws root:compute
```

Create two Locations and two SyncTargets.
```console
$ kubectl create -f config/samples/location_prod.yaml
$ kubectl create -f config/samples/location_dev.yaml
$ kubectl create -f config/samples/synctarget_prod.yaml
$ kubectl create -f config/samples/synctarget_dev.yaml
```

Note that kcp automatically creates a Location `default`. So there are 3 Locations and 2 SyncTargets in `root:compute`.
```console
$ kubectl get locations,synctargets
NAME                                 RESOURCE      AVAILABLE   INSTANCES   LABELS   AGE
location.scheduling.kcp.io/default   synctargets   0           2                    2m12s
location.scheduling.kcp.io/dev       synctargets   0           1                    2m39s
location.scheduling.kcp.io/prod      synctargets   0           1                    3m13s

NAME                              AGE
synctarget.workload.kcp.io/dev    110s
synctarget.workload.kcp.io/prod   2m12s
```

Go to user home workspace, and create an EdgePlacement `test-1`.
```console
$ kubectl ws ~
$ kubectl create -f config/samples/edgeplacement_test-1.yaml
```

The edge scheduler maintains a SinglePlacementSlice for an EdgePlacement in the same workspace.
```console
$ kubectl get sps test-1 -oyaml
apiVersion: edge.kcp.io/v1alpha1
destinations:
- cluster: f3il38atqno12hfd
  locationName: prod
  syncTargetName: prod
  syncTargetUID: 8c0a7003-ad18-4bf0-90a0-b1d74cda2437
- cluster: f3il38atqno12hfd
  locationName: dev
  syncTargetName: dev
  syncTargetUID: dc490a42-e8f1-4930-a142-6c0ba8fd39d5
- cluster: f3il38atqno12hfd
  locationName: default
  syncTargetName: prod
  syncTargetUID: 8c0a7003-ad18-4bf0-90a0-b1d74cda2437
- cluster: f3il38atqno12hfd
  locationName: default
  syncTargetName: dev
  syncTargetUID: dc490a42-e8f1-4930-a142-6c0ba8fd39d5
kind: SinglePlacementSlice
metadata:
  annotations:
    kcp.io/cluster: kvdk2spgmbix
  creationTimestamp: "2023-03-27T21:37:29Z"
  generation: 1
  name: test-1
  ownerReferences:
  - apiVersion: edge.kcp.io/v1alpha1
    kind: EdgePlacement
    name: test-1
    uid: 0c68724e-6d11-4cff-bd0a-8fa32c86caa9
  resourceVersion: "877"
  uid: 45ec86d7-bdf8-4c2d-bc02-087073a1ac17
```
EdgePlacement `test-1` selects all the 3 Locations in `root:compute`.

Create a more specific EdgePlacement which selects Locations labeled by `env: dev`.
```console
$ kubectl create -f config/samples/edgeplacement_dev.yaml
```

The corresponding SinglePlacementSlice has a shorter list of `destinations`:
```console
$ kubectl get sps dev -oyaml
apiVersion: edge.kcp.io/v1alpha1
destinations:
- cluster: f3il38atqno12hfd
  locationName: dev
  syncTargetName: dev
  syncTargetUID: dc490a42-e8f1-4930-a142-6c0ba8fd39d5
kind: SinglePlacementSlice
metadata:
  annotations:
    kcp.io/cluster: kvdk2spgmbix
  creationTimestamp: "2023-03-27T21:40:52Z"
  generation: 1
  name: dev
  ownerReferences:
  - apiVersion: edge.kcp.io/v1alpha1
    kind: EdgePlacement
    name: dev
    uid: 6e9d608d-12cd-47cc-8887-3695199259ba
  resourceVersion: "880"
  uid: 9b8de087-21bc-4585-99cb-e6c03ba0a8ae
```

Feel free to change the Locations, SyncTargets, and EdgePlacements and see how the edge scheduler reacts.
