### Edge scheduler
The edge scheduler monitors the EdgePlacement, Location, and SyncTarget objects and maintains the results of matching.

It starts from the kcp-scheduling-placement controller forked from kcp-dev/kcp, now being refactored towards the goal above.

It works with [this snapshot](https://github.com/kcp-dev/kcp/tree/4506fdc064060b3fe82e1082533f9798b36ba7a5) of kcp.

#### A short demo
Start a kcp server with [this snapshot](https://github.com/kcp-dev/kcp/tree/4506fdc064060b3fe82e1082533f9798b36ba7a5).

Point `$KUBECONFIG` to the admin kubeconfig for that kcp server.

Workspace `root:edge` is used as the edge service provider workspace.
```console
kubectl ws root
kubectl ws create edge --enter
```

Install CRDs and APIExport.
```console
kubectl apply -f config/crds/ -f config/exports/
```

User home workspace is used as the workload management workspace.
```console
kubectl ws \~
```

Bind APIs and create an EdgePlacement CR.
```console
kubectl apply -f config/imports/
kubectl apply -f config/samples/edgeplacement_test-1.yaml
```

Go to `root:edge` workspace and run the edge scheduler.
```console
kubectl ws root:edge
go run cmd/scheduler/main.go --kcp-kubeconfig=<path to kcp admin kubeconfig> -v <verbosity (default 2)>
```

The outputs from the edge scheduler should be similar to:
```console
I0302 11:28:38.595692   96096 scheduler.go:211] "Found APIExport view" exportName="edge.kcp.io" serverURL="https://192.168.1.54:6443/services/apiexport/wmtp3o8lb1n1c5uj/edge.kcp.io"
I0302 11:28:38.601037   96096 controller.go:132] "queueing EdgePlacement" reconciler="edge-scheduler" key="kvdk2spgmbix|test-1"
I0302 11:28:38.697604   96096 controller.go:159] "Starting controller" reconciler="edge-scheduler"
I0302 11:28:38.697723   96096 controller.go:184] "processing key" reconciler="edge-scheduler" key="kvdk2spgmbix|test-1"
reconciling EdgePlacement test-1 in Workspace kvdk2spgmbix
I0302 11:28:38.807572   96096 controller.go:226] "creating SinglePlacementSlice" reconciler="edge-scheduler" key="kvdk2spgmbix|test-1" edgeplacement.workspace="kvdk2spgmbix" edgeplacement.namespace="" edgeplacement.name="test-1" edgeplacement.apiVersion="edge.kcp.io/v1alpha1" workspace="kvdk2spgmbix" name="test-1"
I0302 11:28:38.813981   96096 controller.go:245] "created SinglePlacementSlice" reconciler="edge-scheduler" key="kvdk2spgmbix|test-1" edgeplacement.workspace="kvdk2spgmbix" edgeplacement.namespace="" edgeplacement.name="test-1" edgeplacement.apiVersion="edge.kcp.io/v1alpha1" workspace="kvdk2spgmbix" name="test-1"
```

The edge scheduler ensures the existence of a SinglePlacementSlice for an EdgePlacement in the same workspace.
```console
kubectl ws \~
kubectl get epl,sps
```

The outputs of `kubectl get epl,sps` should be similar to:
```console
NAME                               AGE
edgeplacement.edge.kcp.io/test-1   7m10s

NAME                                      AGE
singleplacementslice.edge.kcp.io/test-1   5m54s
```

Edit then delete the EdgePlacement.
```
kubectl edit edgeplacement test-1
kubectl delete edgeplacement test-1
```

Additional outputs from the edge scheduler should be similar to:
```console
I0302 11:38:21.224376   96096 controller.go:132] "queueing EdgePlacement" reconciler="edge-scheduler" key="kvdk2spgmbix|test-1"
I0302 11:38:21.224423   96096 controller.go:184] "processing key" reconciler="edge-scheduler" key="kvdk2spgmbix|test-1"
reconciling EdgePlacement test-1 in Workspace kvdk2spgmbix
I0302 11:38:35.956562   96096 controller.go:132] "queueing EdgePlacement" reconciler="edge-scheduler" key="kvdk2spgmbix|test-1"
I0302 11:38:35.956586   96096 controller.go:184] "processing key" reconciler="edge-scheduler" key="kvdk2spgmbix|test-1"
```
