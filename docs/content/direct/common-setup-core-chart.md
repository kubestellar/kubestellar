# Common Setup - KubeStellar Core chart

For more details please see [here](./core-chart.md).

This documents explains how to use KubeStellar Core chart to deploy a new instance of KubeStellar
with a choice of user-defined KubeFlex Control Planes (CPs).

The information provided is specific for the following release:

```shell
export KUBESTELLAR_VERSION=0.23.0
export OCM_TRANSPORT_PLUGIN=0.1.11
```

It may also be a good idea to do a bit of cleanup first. See how it is done in the cleanup script for our E2E tests (in `test/e2e/common/cleanup.sh`).

## Create a Kind cluster

The setup of KubStellar via the Core chart requires the existance of at least one cluster
to be used for the deployment of the chart.

For convenience, a new local Kind cluster that satisfies the requirements for KubeStellar setup
and that can be used to exercises the [examples](./examples.md) can be created with the following command:

```shell
bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v${KUBESTELLAR_VERSION}/scripts/create-kind-cluster-with-SSL-passthrough.sh) --name kubeflex --port 9443
```

## Install KubeStellar Core chart

A KubeStellar Core installation compatible with the common setup suitable for Common Setup described in the [examples](examples.md) could be achieved with the following command:

```shell
helm upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart --version $KUBESTELLAR_VERSION \
  --set-json='ITSes=[{"name":"its1"}]' \
  --set-json='WDSes=[{"name":"wds1"}]'
```

Remember to include the option `--set "kubeflex-operator.isOpenShift=true"`, when deploying in an OpenShift cluster.

## Kubeconfig contexts for Control Planes

After the KubeStellar Core has fully deployed, before changing the kubeconfig context, `kflex` CLI can be used to retrieve the Kubernetes client configurations for all the KubeFlex Control Planes and store them as contexts of the current kubeconfig file:

```shell
kubectl config delete-context its1 || true
kflex ctx its1
kubectl config delete-context wds1 || true
kflex ctx wds1
kflex ctx # switch back to the initial context
```

Afterwards the content of a Control Plane `<cpname>` can be accessed by specifing its context:

```shell
kubectl --context <cpname> ...
```

## (optional) Check relevant deployments and statefulsets running in the hosting cluster

Expect to see the `kubestellar-controller-manager` in the `wds1-system` namespace and the
statefulset `vcluster` in the `its1-system` namespace, both fully ready.

```shell
kubectl get deployments,statefulsets --all-namespaces
```

The output should look something like the following:

```shell
NAMESPACE            NAME                                             READY   UP-TO-DATE   AVAILABLE   AGE
ingress-nginx        deployment.apps/ingress-nginx-controller         1/1     1            1           22h
kube-system          deployment.apps/coredns                          2/2     2            2           22h
kubeflex-system      deployment.apps/kubeflex-controller-manager      1/1     1            1           22h
local-path-storage   deployment.apps/local-path-provisioner           1/1     1            1           22h
wds1-system          deployment.apps/kube-apiserver                   1/1     1            1           22m
wds1-system          deployment.apps/kube-controller-manager          1/1     1            1           22m
wds1-system          deployment.apps/kubestellar-controller-manager   1/1     1            1           21m
wds1-system          deployment.apps/transport-controller             1/1     1            1           21m

NAMESPACE         NAME                                   READY   AGE
its1-system       statefulset.apps/vcluster              1/1     11h
kubeflex-system   statefulset.apps/postgres-postgresql   1/1     22h
```
