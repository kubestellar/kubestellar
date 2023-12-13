<!--example1-post-provider-start-->
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

``` {.bash}
curl -SL -o kubestellar.tar.gz "https://github.com/kubestellar/kubestellar/releases/download/${kubestellar_version}/kubestellar_${kubestellar_version}_${os_type}_${arch_type}.tar.gz
tar xzf kubestellar.tar.gz
export PATH=$PWD/bin:$PATH
```

##### Get from GitHub

You can get the latest version from GitHub with the following command,
which will get you the default branch (which is named "main"); add `-b
$branch` to the `git` command in order to get a different branch.

``` {.bash}
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

#### Get binaries of kube-bind and dex
The command below makes kube-bind binaries and dex binary available in `$PATH`.

```shell
rm -rf kube-bind
git clone https://github.com/waltforme/kube-bind.git && \
pushd kube-bind && \
mkdir bin && \
IGNORE_GO_VERSION=1 go build -o ./bin/example-backend ./cmd/example-backend/main.go && \
git checkout origin/syncmore && \
IGNORE_GO_VERSION=1 go build -o ./bin/konnector ./cmd/konnector/main.go && \
git checkout origin/autobind && \
IGNORE_GO_VERSION=1 go build -o ./bin/kubectl-bind ./cmd/kubectl-bind/main.go && \
export PATH=$(pwd)/bin:$PATH && \
popd && \
git clone https://github.com/dexidp/dex.git && \
pushd dex && \
IGNORE_GO_VERSION=1 make build && \
export PATH=$(pwd)/bin:$PATH && \
popd
```

#### Initialize the KubeStellar platform as bare processes

In this step KubeStellar creates and populates the KubeStellar Core
Space (KCS) (formerly called the Edge Service Provider Workspace
(ESPW)), which exports the KubeStellar API.

```shell
KUBECONFIG=$SM_CONFIG kubestellar -X init
```

### Deploy kcp and KubeStellar as a workload in a Kubernetes cluster

(This style of deployment requires release v0.6 or later of KubeStellar.)

You need a Kubernetes cluster; see [the documentation for `kubectl kubestellar deploy`](../../commands/#deployment-into-a-kubernetes-cluster) for more information.

You will need a domain name that, on each of your clients, resolves to
an IP address that the client can use to open a TCP connection to the
Ingress controller's listening socket.

You will need the kcp `kubectl` plugins.  See [the "Start kcp" section
above](../#start-kcp) for instructions on how to get all of the kcp
executables.

You will need to get a build of KubeStellar.  See
[above](../#get-kubestellar).

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
using kcp or KubeStellar. For example, you might use the following
commands to fetch and start using that kubeconfig file; the first
assumes that you deployed the core into a Kubernetes namespace named
"kubestellar".

``` {.bash}
kubectl kubestellar get-external-kubeconfig -n kubestellar -o kcs.kubeconfig
export KUBECONFIG=$(pwd)/kcs.kubeconfig
```

Note that you now care about two different kubeconfig files: the one
that you were using earlier, which holds the contexts for your `kind`
clusters, and the one that you just fetched and started using for
working with the KubeStellar interface. The remainder of this document
assumes that your `kind` cluster contexts are in `~/.kube/config` and
the current context is for the KubeStellar hosting cluster.

The following variable will be used in later commands to indicate that
they are _not_ being invoked from within the hosting cluster (see [doc
on "in-cluster"](../../commands/#in-cluster)).

``` {.bash}
in_cluster=""
```

### Create SyncTarget and Location objects to represent the florin and guilder clusters

Use the following two commands to put inventory objects in the IMW at
`root:imw1` that was automatically created during deployment of
KubeStellar. They label both florin and guilder with `env=prod`, and
also label guilder with `extended=yes`.

```shell
IMW1_KUBECONFIG="${MY_KUBECONFIGS}/imw1.kubeconfig"
kubectl-kubestellar-space-get_kubeconfig imw1 --kubeconfig $SM_CONFIG $in_cluster $IMW1_KUBECONFIG
KUBECONFIG=$IMW1_KUBECONFIG kubectl kubestellar ensure location florin  loc-name=florin  env=prod
KUBECONFIG=$IMW1_KUBECONFIG kubectl kubestellar ensure location guilder loc-name=guilder env=prod extended=yes
echo "describe the florin location object"
KUBECONFIG=$IMW1_KUBECONFIG kubectl describe location.edge.kubestellar.io florin
```

Those two script invocations are equivalent to creating the following
four objects plus the kcp `APIBinding` objects that import the
definition of the KubeStellar API.

```yaml
apiVersion: edge.kubestellar.io/v2alpha1
kind: SyncTarget
metadata:
  name: florin
  labels:
    id: florin
    loc-name: florin
    env: prod
---
apiVersion: edge.kubestellar.io/v2alpha1
kind: Location
metadata:
  name: florin
  labels:
    loc-name: florin
    env: prod
spec:
  resource: {group: edge.kubestellar.io, version: v2alpha1, resource: synctargets}
  instanceSelector:
    matchLabels: {id: florin}
---
apiVersion: edge.kubestellar.io/v2alpha1
kind: SyncTarget
metadata:
  name: guilder
  labels:
    id: guilder
    loc-name: guilder
    env: prod
    extended: yes
---
apiVersion: edge.kubestellar.io/v2alpha1
kind: Location
metadata:
  name: guilder
  labels:
    loc-name: guilder
    env: prod
    extended: yes
spec:
  resource: {group: edge.kubestellar.io, version: v2alpha1, resource: synctargets}
  instanceSelector:
    matchLabels: {id: guilder}
```

That script also deletes the Location named `default`, which is not
used in this PoC, if it shows up.

<!--example1-post-provider-end-->
