<!--example1-post-kcp-start-->
#### Get KubeStellar

You will need a local copy of KubeStellar.  You can either use the
pre-built archive (containing executables and config files) from a
release or get any desired version from GitHub and build.

##### Use pre-built archive

Fetch the archive for your operating system and instruction set
architecture as follows, in which `$kubestellar_version` is your
chosen release of KubeStellar (see [the releases on
GitHub](https://github.com/kubestellar/kubestellar/releases)) and
`$os_type` and `$arch_type` are chosen according to the list of
"assets" for your chosen release.

```{.base}
curl -SL -o kubestellar.tar.gz "https://github.com/kubestellar/kubestellar/releases/download/${kubestellar_version}/kubestellar_${kubestellar_version}_${os_type}_${arch_type}.tar.gz
tar xzf kubestellar.tar.gz
export PATH=$PWD/bin:$PATH
```

##### Get from GitHub

You can get the latest version from GitHub with the following command,
which will get you the default branch (which is named "main"); add `-b
$branch` to get a different one.

```{.base}
git clone {{ config.repo_url }}
cd kubestellar
```

Use the following commands to build and add the executables to your
`$PATH`.

```shell
make build
export PATH=$(pwd)/bin:$PATH
```

In the following exhibited command lines, the commands described as
"KubeStellar commands" and the commands that start with `kubectl
kubestellar` rely on the KubeStellar `bin` directory being on the
`$PATH`.  Alternatively you could invoke them with explicit pathnames.
The kubectl plugin lines use fully specific executables (e.g.,
`kubectl kubestellar prep-for-syncer` corresponds to
`bin/kubectl-kubestellar-prep_for_syncer`).

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
hosting cluster (the Kubernetes cluser into which you want to deploy
kcp and KubeStellar), either by virtue of having set your `KUBECONFIG`
envar appropriately or putting the relevant contents in
`~/.kube/config` or by passing `--kubeconfig` explicitly on the
following command lines.

Use the [kubectl kubestellar deploy
command](../../commands/#deploy-to-cluster) to do the deployment.

Then use the [kubectl kubestellat get-external-kubeconfig
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
### Create SyncTarget and Location objects to represent the florin and guilder clusters

Use the following two commands. They label both florin and guilder
with `env=prod`, and also label guilder with `extended=si`.

```shell
kubectl ws root:imw-1
kubectl kubestellar ensure location florin  loc-name=florin  env=prod
kubectl kubestellar ensure location guilder loc-name=guilder env=prod extended=si
echo "decribe the florin location object"
kubectl describe location.edge.kcp.io florin
```

Those two script invocations are equivalent to creating the following
four objects.

```yaml
apiVersion: edge.kcp.io/v1alpha1
kind: SyncTarget
metadata:
  name: florin
  labels:
    id: florin
    loc-name: florin
    env: prod
---
apiVersion: edge.kcp.io/v1alpha1
kind: Location
metadata:
  name: florin
  labels:
    loc-name: florin
    env: prod
spec:
  resource: {group: edge.kcp.io, version: v1alpha1, resource: synctargets}
  instanceSelector:
    matchLabels: {id: florin}
---
apiVersion: edge.kcp.io/v1alpha1
kind: SyncTarget
metadata:
  name: guilder
  labels:
    id: guilder
    loc-name: guilder
    env: prod
    extended: si
---
apiVersion: edge.kcp.io/v1alpha1
kind: Location
metadata:
  name: guilder
  labels:
    loc-name: guilder
    env: prod
    extended: si
spec:
  resource: {group: edge.kcp.io, version: v1alpha1, resource: synctargets}
  instanceSelector:
    matchLabels: {id: guilder}
```

That script also deletes the Location named `default`, which is not
used in this PoC, if it shows up.

<!--example1-post-kcp-end-->
