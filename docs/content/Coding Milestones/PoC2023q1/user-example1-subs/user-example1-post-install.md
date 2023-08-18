<!--user-example1-post-install-start-->
#### Initialize the KubeStellar platform as bare processes

In this step KubeStellar creates and populates the Edge Service
Provider Workspace (ESPW), which exports the KubeStellar API, and also
augments the `root:compute` workspace from kcp TMC as needed here.
Those augmentation consists of adding authorization to update the
relevant `/status` and `/scale` subresources (missing in kcp TMC) and
extending the supported subset of the Kubernetes API for managing
containerized workloads from the four resources built into kcp TMC
(`Deployment`, `Pod`, `Service`, and `Ingress`) to the other ones that
are namespaced and are meaningful in KubeStellar.

```shell
kubestellar init
```

### Deploy kcp and KubeStellar as a workload in a Kubernetes cluster

First you will need to get a build of KubeStellar.  See
[above](../#get-kubestellar) and **NOTE WELL**: as yet there is no
release of KubeStellar that supports this style of deployment, you
will have to get the latest code from github and `make build`.

To do the deployment and prepare to use it you will be using [the
commands defined for
that](../../commands/#deployment-into-a-kubernetes-cluster).  These
require your shell to be in a state where `kubectl` manipulates the
hosting cluster (the Kubernetes cluster into which you want to deploy
kcp and KubeStellar), either by virtue of having set your `KUBECONFIG`
envar appropriately or putting the relevant contents in
`~/.kube/config` or by passing `--kubeconfig` explicitly on the
following command lines.

Use the [kubectl kubestellar deploy
command](../../commands/#deploy-to-cluster) to do the deployment.

Then use the [kubectl kubestellar get-external-kubeconfig
command](../../commands/#fetch-kubeconfig-for-external-clients) to put
into a file the kubeconfig that you will use as a user of kcp and
KubeStellar.  Do not overwrite the kubeconfig file for your hosting
cluster.  But _do_ update your `KUBECONFIG` envar setting or remember
to pass the new file with `--kubeconfig` on the command lines when
using kcp or KubeStellar.


### Create an inventory management workspace.
```shell
kubectl ws root
kubectl ws create imw-1 
```
### Create SyncTarget and Location objects to represent the ren and stimpy clusters

Use the following two commands. They label both ren and stimpy
with `env=prod`, and also label stimpy with `extended=si`.

```shell
kubectl ws root:imw-1
kubectl kubestellar ensure location ren  loc-name=ren  env=prod
kubectl kubestellar ensure location stimpy loc-name=stimpy env=prod extended=si
echo "decribe the ren location object"
kubectl describe location.edge.kcp.io ren
```

Those two script invocations are equivalent to creating the following
four objects.

```yaml
apiVersion: edge.kcp.io/v1alpha1
kind: SyncTarget
metadata:
  name: ren
  labels:
    id: ren
    loc-name: ren
    env: prod
---
apiVersion: edge.kcp.io/v1alpha1
kind: Location
metadata:
  name: ren
  labels:
    loc-name: ren
    env: prod
spec:
  resource: {group: edge.kcp.io, version: v1alpha1, resource: synctargets}
  instanceSelector:
    matchLabels: {id: ren}
---
apiVersion: edge.kcp.io/v1alpha1
kind: SyncTarget
metadata:
  name: stimpy
  labels:
    id: stimpy
    loc-name: stimpy
    env: prod
    extended: si
---
apiVersion: edge.kcp.io/v1alpha1
kind: Location
metadata:
  name: stimpy
  labels:
    loc-name: stimpy
    env: prod
    extended: si
spec:
  resource: {group: edge.kcp.io, version: v1alpha1, resource: synctargets}
  instanceSelector:
    matchLabels: {id: stimpy}
```

That script also deletes the Location named `default`, which is not
used in this PoC, if it shows up.

<!--user-example1-post-install-end-->
