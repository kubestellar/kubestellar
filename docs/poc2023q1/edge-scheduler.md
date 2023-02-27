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
I0214 17:09:54.739592   50397 placement.go:207] "Found APIExport view" exportName="edge.kcp.io" serverURL="https://192.168.1.54:6443/services/apiexport/21rzp91t9ife44tq/edge.kcp.io"
I0214 17:09:54.755028   50397 controller.go:182] "queueing EdgePlacement" reconciler="edge-scheduler" key="kvdk2spgmbix|test-1"
I0214 17:09:54.841368   50397 controller.go:257] "Starting controller" reconciler="edge-scheduler"
I0214 17:09:54.841492   50397 controller.go:282] "processing key" reconciler="edge-scheduler" key="kvdk2spgmbix|test-1"
reconciling EdgePlacement test-1 in Workspace kvdk2spgmbix
```

Go to `\~` and edit or delete the EdgePlacement.
```
kubectl ws \~
kubectl edit edgeplacement test-1
kubectl delete edgeplacement test-1
```

Additional outputs from the edge scheduler should be similar to:
```console
I0214 17:10:12.404611   50397 controller.go:182] "queueing EdgePlacement" reconciler="edge-scheduler" key="kvdk2spgmbix|test-1"
I0214 17:10:12.404634   50397 controller.go:282] "processing key" reconciler="edge-scheduler" key="kvdk2spgmbix|test-1"
reconciling EdgePlacement test-1 in Workspace kvdk2spgmbix
I0214 17:11:21.744670   50397 controller.go:182] "queueing EdgePlacement" reconciler="edge-scheduler" key="kvdk2spgmbix|test-1"
I0214 17:11:21.744695   50397 controller.go:282] "processing key" reconciler="edge-scheduler" key="kvdk2spgmbix|test-1"
I0214 17:11:21.744722   50397 controller.go:308] "object deleted before handled" reconciler="edge-scheduler" key="kvdk2spgmbix|test-1"
```
