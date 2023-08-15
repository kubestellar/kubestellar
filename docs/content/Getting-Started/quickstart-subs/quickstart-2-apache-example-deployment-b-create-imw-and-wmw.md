<!--quickstart-2-apache-example-deployment-b-create-imw-and-wmw-start-->
The above use of `kind` has knocked kcp's `kubectl ws` plugin off kilter, as the latter uses the local kubeconfig to store its state about the "current" and "previous" workspaces.  Get it back on track with the following command.

```shell
kubectl config use-context root
```

IMWs are used by KubeStellar to store inventory objects (`SyncTargets` and `Locations`). Create an IMW named `example-imw` with the following command:

```shell
kubectl ws root
kubectl ws create example-imw
```

WMWs are used by KubeStellar to store workload descriptions and `EdgePlacement` objects. Create an WMW named `example-wmw` in a `my-org` workspace with the following commands:

```shell
kubectl ws root
kubectl ws create my-org --enter
kubectl kubestellar ensure wmw example-wmw
```

A WMW does not have to be created before the edge cluster is on-boarded; the WMW only needs to be created before content is put in it.
<!--quickstart-2-apache-example-deployment-b-create-imw-and-wmw-end-->
