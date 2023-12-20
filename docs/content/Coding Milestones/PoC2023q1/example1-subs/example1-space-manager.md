<!--example1-space-manager-start-->

### Create Kind cluster for space management

```shell
kind create cluster --name sm-mgt --config - <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 443
    hostPort: 9443
    protocol: TCP
EOF

kubectl create -f https://raw.githubusercontent.com/kubestellar/kubestellar/main/example/kind-nginx-ingress-with-SSL-passthrough.yaml
sleep 20
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=360s

KUBECONFIG=~/.kube/config kubectl config rename-context kind-sm-mgt sm-mgt
SM_CONFIG=~/.kube/config
```

The subsequent uses of `$SM_CONFIG` in this example assume that the
current context is still the one just established, "sm-mgt".

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
KUBECONFIG=$SM_CONFIG kubectl apply -f config/crds/
cd ..
```
Finally, start the space-manager controller.

```shell
space-manager --kubeconfig $SM_CONFIG --context sm-mgt -v 4 &> /tmp/space-manager.log &
```

<!--example1-space-manager-end-->
