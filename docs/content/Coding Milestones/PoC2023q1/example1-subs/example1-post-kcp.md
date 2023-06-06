<!--example1-post-kcp-start-->
### Create an inventory management workspace.

Use the following commands.

```shell
kubectl ws root
kubectl ws create imw-1 --enter
```

### Get edge-mc

Download and build or install
[edge-mc](https://github.com/kcp-dev/edge-mc), according to your
preference.  That is, either (a) `git clone` the repo and then `make
build` to populate its `bin` directory, or (b) fetch the binary
archive appropriate for your machine from a release and unpack it
(creating a `bin` directory).  In the following exhibited command
lines, the commands described as "edge-mc commands" and the commands
that start with `kubectl kubestellar` rely on the edge-mc `bin` directory
being on the `$PATH`.  Alternatively you could invoke them with
explicit pathnames.  The kubectl plugin lines use fully specific
executables (e.g., `kubectl kubestellar prep-for-syncer` corresponds to
`bin/kubectl-kubestellar-prep_for_syncer`).

```shell
cd ../KubeStellar
make build
export PATH=$(pwd)/bin:$PATH
```
### Create SyncTarget and Location objects to represent the florin and guilder clusters

Use the following two commands. They label both florin and guilder
with `env=prod`, and also label guilder with `extended=si`.

```shell
kubectl kubestellar ensure location florin  loc-name=florin  env=prod
kubectl kubestellar ensure location guilder loc-name=guilder env=prod extended=si
```

Those two script invocations are equivalent to creating the following
four objects.

```yaml
apiVersion: workload.kcp.io/v1alpha1
kind: SyncTarget
metadata:
  name: florin
  labels:
    id: florin
    loc-name: florin
    env: prod
---
apiVersion: scheduling.kcp.io/v1alpha1
kind: Location
metadata:
  name: florin
  labels:
    loc-name: florin
    env: prod
spec:
  resource: {group: workload.kcp.io, version: v1alpha1, resource: synctargets}
  instanceSelector:
    matchLabels: {id: florin}
---
apiVersion: workload.kcp.io/v1alpha1
kind: SyncTarget
metadata:
  name: guilder
  labels:
    id: guilder
    loc-name: guilder
    env: prod
    extended: si
---
apiVersion: scheduling.kcp.io/v1alpha1
kind: Location
metadata:
  name: guilder
  labels:
    loc-name: guilder
    env: prod
    extended: si
spec:
  resource: {group: workload.kcp.io, version: v1alpha1, resource: synctargets}
  instanceSelector:
    matchLabels: {id: guilder}
```

That script also deletes the Location named `default`, which is not
used in this PoC, if it shows up.

### Create the edge service provider workspace

Use the following commands.

```shell
kubectl ws root
kubectl ws create espw --enter
```
<!--example1-post-kcp-end-->