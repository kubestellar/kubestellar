<!--placement-translator-create-imw-and-build-kubestellar-start-->
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
<!--placement-translator-create-imw-and-build-kubestellar-end-->