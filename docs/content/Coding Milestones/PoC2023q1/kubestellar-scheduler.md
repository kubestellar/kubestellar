## Required Packages:
  - [GO v1.19](https://gist.github.com/jniltinho/8758e15a9ef80a189fce) - You will need GO to compile and run kcp and the KubeStellar scheduler.  Currently kcp requires go version 1.19.

### Steps to try the scheduler

#### Pull the kcp and KubeStellar source code, build the kubectl-ws binary, and start kcp
open a terminal window(1) and clone the latest KubeStellar source:
{%
   include-markdown "kubestellar-scheduler-subs/kubestellar-scheduler-0-pull-kcp-and-kuberstellar-source-and-start-kcp.md"
   start="<!--kubestellar-scheduler-0-pull-kcp-and-kuberstellar-source-and-start-kcp-start-->"
   end="<!--kubestellar-scheduler-0-pull-kcp-and-kuberstellar-source-and-start-kcp-end-->"
%}

#### Create the Edge Service Provider Workspace (ESPW) and populate it with CRDs and APIs
open another terminal window(2) and point `$KUBECONFIG` to the admin kubeconfig for the kcp server and include the location of kubectl-ws in `$PATH`.
{%
   include-markdown "kubestellar-scheduler-subs/kubestellar-scheduler-1-export-kubeconfig-and-path-for-kcp.md"
   start="<!--kubestellar-scheduler-1-export-kubeconfig-and-path-for-kcp-start-->"
   end="<!--kubestellar-scheduler-1-export-kubeconfig-and-path-for-kcp-end-->"
%}

{%
   include-markdown "kubestellar-scheduler-subs/kubestellar-scheduler-2-ws-root-and-ws-create-edge.md"
   start="<!--kubestellar-scheduler-2-ws-root-and-ws-create-edge-start-->"
   end="<!--kubestellar-scheduler-2-ws-root-and-ws-create-edge-end-->"
%}

Install CRDs and APIExport.
{%
   include-markdown "kubestellar-scheduler-subs/kubestellar-scheduler-exports.md"
   start="<!--kubestellar-scheduler-exports-start-->"
   end="<!--kubestellar-scheduler-exports-end-->"
%}

#### Create the Workload Management Workspace (WMW) and bind it to the ESPW APIs
{%
   include-markdown "kubestellar-scheduler-subs/kubestellar-scheduler-imports.md"
   start="<!--kubestellar-scheduler-imports-start-->"
   end="<!--kubestellar-scheduler-imports-end-->"
%}

#### Run the KubeStellar Scheduler against the ESPW
Go to the `root:espw` workspace and run the edge scheduler.
{%
   include-markdown "kubestellar-scheduler-subs/kubestellar-scheduler-process-start.md"
   start="<!--kubestellar-scheduler-process-start-start-->"
   end="<!--kubestellar-scheduler-process-start-end-->"
%}
The outputs from the edge scheduler should be similar to:
``` { .bash .no-copy }
I0605 10:53:00.156100   29786 scheduler.go:212] "Found APIExport view" exportName="edge.kcp.io" serverURL="https://192.168.1.13:6443/services/apiexport/jxch2kyb3c1h6bac/edge.kcp.io"
I0605 10:53:00.157874   29786 scheduler.go:212] "Found APIExport view" exportName="scheduling.kcp.io" serverURL="https://192.168.1.13:6443/services/apiexport/root/scheduling.kcp.io"
I0605 10:53:00.159242   29786 scheduler.go:212] "Found APIExport view" exportName="workload.kcp.io" serverURL="https://192.168.1.13:6443/services/apiexport/root/workload.kcp.io"
I0605 10:53:00.261128   29786 controller.go:201] "starting controller" controller="kubestellar-scheduler"
```

#### Create the Inventory Management Workspace (IMW) and populate it with locations and synctargets
open another terminal window(3) and point `$KUBECONFIG` to the admin kubeconfig for the kcp server and include the location of kubectl-ws in $PATH.
```shell
cd ../kcp
export KUBECONFIG=$(pwd)/.kcp/admin.kubeconfig
export PATH=$(pwd)/bin:$PATH
```

Use workspace `root:compute` as the Inventory Management Workspace (IMW).
```shell
kubectl ws root:compute
```

Create two Locations and two SyncTargets.
```shell
kubectl create -f ../KubeStellar/config/samples/location_prod.yaml
kubectl create -f ../KubeStellar/config/samples/location_dev.yaml
kubectl create -f ../KubeStellar/config/samples/synctarget_prod.yaml
kubectl create -f ../KubeStellar/config/samples/synctarget_dev.yaml
sleep 5
```

Note that kcp automatically creates a Location `default`. So there are 3 Locations and 2 SyncTargets in `root:compute`.
```shell
kubectl get locations,synctargets
```
``` { .bash .no-copy }
NAME                                 RESOURCE      AVAILABLE   INSTANCES   LABELS   AGE
location.scheduling.kcp.io/default   synctargets   0           2                    2m12s
location.scheduling.kcp.io/dev       synctargets   0           1                    2m39s
location.scheduling.kcp.io/prod      synctargets   0           1                    3m13s

NAME                              AGE
synctarget.workload.kcp.io/dev    110s
synctarget.workload.kcp.io/prod   2m12s
```

#### Create some EdgePlacements in the WMW
Go to Workload Management Workspace (WMW) and create an EdgePlacement `all2all`.
```shell
kubectl ws \~
kubectl create -f ../KubeStellar/config/samples/edgeplacement_all2all.yaml
sleep 3
```

The scheduler maintains a SinglePlacementSlice for an EdgePlacement in the same workspace.
```shell
kubectl get sps all2all -oyaml
```
``` { .bash .no-copy }
apiVersion: edge.kcp.io/v1alpha1
destinations:
- cluster: 1yotsgod0d2p3xa5
  locationName: prod
  syncTargetName: prod
  syncTargetUID: 13841ffd-33f2-4cf4-9114-6156f73aa5c8
- cluster: 1yotsgod0d2p3xa5
  locationName: dev
  syncTargetName: dev
  syncTargetUID: ea5492ec-44af-4173-a4ca-9c5cd59afcb1
- cluster: 1yotsgod0d2p3xa5
  locationName: default
  syncTargetName: dev
  syncTargetUID: ea5492ec-44af-4173-a4ca-9c5cd59afcb1
- cluster: 1yotsgod0d2p3xa5
  locationName: default
  syncTargetName: prod
  syncTargetUID: 13841ffd-33f2-4cf4-9114-6156f73aa5c8
kind: SinglePlacementSlice
metadata:
  annotations:
    kcp.io/cluster: kvdk2spgmbix
  creationTimestamp: "2023-06-05T14:55:20Z"
  generation: 1
  name: all2all
  ownerReferences:
  - apiVersion: edge.kcp.io/v1alpha1
    kind: EdgePlacement
    name: all2all
    uid: 31915018-6a25-4f01-943e-b8a0a0ed35ba
  resourceVersion: "875"
  uid: a2b8224d-5feb-40a1-adb2-67c07965f13b
```
EdgePlacement `all2all` selects all the 3 Locations in `root:compute`.

Create a more specific EdgePlacement which selects Locations labeled by `env: dev`.
```shell
kubectl create -f ../KubeStellar/config/samples/edgeplacement_dev.yaml
sleep 3
```

The corresponding SinglePlacementSlice has a shorter list of `destinations`:
```shell
kubectl get sps dev -oyaml
```
``` { .bash .no-copy }
apiVersion: edge.kcp.io/v1alpha1
destinations:
- cluster: 1yotsgod0d2p3xa5
  locationName: dev
  syncTargetName: dev
  syncTargetUID: ea5492ec-44af-4173-a4ca-9c5cd59afcb1
kind: SinglePlacementSlice
metadata:
  annotations:
    kcp.io/cluster: kvdk2spgmbix
  creationTimestamp: "2023-06-05T14:57:00Z"
  generation: 1
  name: dev
  ownerReferences:
  - apiVersion: edge.kcp.io/v1alpha1
    kind: EdgePlacement
    name: dev
    uid: 1ac4b7f5-5521-4b5a-a0fa-cc2ec87b458b
  resourceVersion: "877"
  uid: c9c0c2fc-d721-4c73-9788-e10711bad23a
```

Feel free to change the Locations, SyncTargets, and EdgePlacements and see how the scheduler reacts.

Your next step is to deliver a workload to a mailbox (that represents an edge location).  Go here to take the next step... (TBD)
