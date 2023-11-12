<!--example1-space-manager-start-->

### Create Kind cluster for space management

```shell
kind create cluster --name space-mgt
KUBECONFIG=~/.kube/config kubectl config rename-context kind-space-mgt space-mgt
```

### The space-manager controller

You can get the latest version from GitHub with the following command,
which will get you the default branch (which is named "main"); add `-b
$branch` to the `git` command in order to get a different branch.

```{.bash}
git clone {{ config.repo_url }}
cd kubestellar
```

Use the following commands to build and add the executables to your
`$PATH`.

```shell
cd space-framework
make build
export PATH=$(pwd)/bin:$PATH
```
Next deploy the space framework CRDs in the space management cluster.
```shell
KUBECONFIG=~/.kube/config kubectl --context space-mgt apply -f config/crds/
cd ..
```
Finally, start the space-manager controller.

```shell
space-manager --kubeconfig ~/.kube/config --context space-mgt -v 2 &
```

<!--example1-space-manager-end-->