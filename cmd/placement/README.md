This is the kcp-scheduling-placement controller separated from kcp-dev/kcp.
The purpose of this work is documented in the Summary of [this PR](https://github.com/kcp-dev/edge-mc/pull/58).
It works with [this snapshot](https://github.com/kcp-dev/kcp/tree/4506fdc064060b3fe82e1082533f9798b36ba7a5) of kcp.

#### Run the controller
Two steps should be taken to run the controller.

First, disable the kcp-scheduling-placement and kcp-workload-placement controllers in kcp.
There are many ways to disable them. One way is to insert
```go
	if true {
		fmt.Printf("%v controller skipped\n", ControllerName)
		return nil
	}
```
into pkg/reconciler/scheduling/placement/placement_reconcile.go.reconcile(),
and insert
```go
	if true {
		fmt.Printf("%v controller skipped\n", ControllerName)
		return false, nil
	}
```
into pkg/reconciler/workload/placement/placement_reconcile.go.reconcile(),
along with the required import statements for "fmt".
After that, start kcp.

Second,
```console
go run cmd/placement/main.go --kcp-kubeconfig=<path to kcp admin.kubeconfig> -v <verbosity (default 2)>
```

#### A short demo
The short demo shows the Placement State Machine described in this [document](https://docs.google.com/document/d/1AzyjuyjNIDVAXEGHslaggltIQ9Cs8pLCzc9Ma_RBmuM/edit#heading=h.vmt32rdidje6), or in this [code block](https://github.com/kcp-dev/kcp/blob/fb4d4a42373ea4da001b8c88396eabaf6f825be1/pkg/apis/scheduling/v1alpha1/types_placement.go#L123-L134).

In the `root:compute` workspace, create Location `foo`:
```console
kubectl create -f config/samples/location_foo.yaml
```

Switch to a user's workspace, say `root:my-org:dev`, create Placement `dev`:
```console
kubectl create -f config/samples/placement_dev.yaml
```
The status of Placement `dev` should show `phase: Pending`.

`kubectl edit` the `spec.locationSelectors` of the Placement `dev`, from `env: prod` to `env: dev`.
The status of Placement `dev` should show (1) `phase: Unbound`, (2) the `foo` Location selected by Placement `dev`.

Label namespace `default` in the user's workspace:
```
kubectl label ns default env=dev
```
The status of the placement should show `phase: Bound`.

Remove the `env=dev` label from the `default` namespace, the state machine should transit to `Unbound`.

Revert the edits of the Placement `dev`:
```console
kubectl apply -f config/samples/placement_dev.yaml
```
The state machines should transit back to `Pending`.

Delete the Placement `dev` from the user's workspace,
and delete the Location `foo` in the `root:compute` workspace.
