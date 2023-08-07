# Deploy **KubeStellar** service in a cluster using Helm

Table of contests:
- [Deploy **KubeStellar** service in a cluster using Helm](#deploy-kubestellar-service-in-a-cluster-using-helm)
  - [Deploy **KubeStellar** in a **Kubernetes** cluster (**Kind** cluster)](#deploy-kubestellar-in-a-kubernetes-cluster-kind-cluster)
  - [Deploy **KubeStellar** in an **OpenShift** cluster](#deploy-kubestellar-in-an-openshift-cluster)

## Deploy **KubeStellar** in a **Kubernetes** cluster (**Kind** cluster)

[Create a **Kind** cluster with the `extraPortMappings` for port `1024` and an **nginx** ingreass with SSL passthrough.](../yaml/README.md)

Deploy **KubeStellar** `stable` in a `kubestellar` namespace:

```shell
helm install kubestellar . --values values-kubernets.yaml
```

In particular the [values-kubernets.yaml](./values-kubernets.yaml) sets:

```yaml
clusterType: Kubernetes # OpenShift or Kubernetes
HOSTNAME: "kubestellar.svc.cluster.local" # leave it empty to let the cluster set it up
EXTERNAL_HOSTNAME: "" # an empty string will let the container infer its ingress/route
EXTERNAL_PORT: 1024
```

The `kubestellar-server` deployment, holds its access kubeconfigs in a `kubestellar` secret in the `kubestellar` namespace, which it manages using a `kubestellar-role`. Additionally, the role allows the pod to get its ingress/route to put it in the `external.kubeconfig`.

After the deployment has completed, **KubeStellar** `admin.kubeconfig` can be in two ways:

- the `kubestellar` secret in the `kubestellar` namespace;
- directly from the `kubestellar` pod in the `kubestellar` namespace at the location `/home/kubestellar/.kcp/external.kubeconfig`.

## Deploy **KubeStellar** in an **OpenShift** cluster

[Create a **Kind** cluster with the `extraPortMappings` for port `1024` and an **nginx** ingreass with SSL passthrough.](../yaml/README.md)

Deploy **KubeStellar** `stable` in a `kubestellar` namespace:

```shell
helm install kubestellar . --values values-openshift.yaml
```

In particular the [values-openshift.yaml](./values-openshift.yaml) sets:

```yaml
clusterType: OpenShift # OpenShift or Kubernetes
HOSTNAME: "" # leave it empty to let the cluster set it up
EXTERNAL_HOSTNAME: "" # an empty string will let the container infer its ingress/route
EXTERNAL_PORT: 443
```

The `kubestellar-server` deployment, holds its access kubeconfigs in a `kubestellar` secret in the `kubestellar` namespace, which it manages using a `kubestellar-role`. Additionally, the role allows the pod to get its ingress/route to put it in the `external.kubeconfig`.

After the deployment has completed, **KubeStellar** `admin.kubeconfig` can be in two ways:

- the `kubestellar` secret in the `kubestellar` namespace;
- directly from the `kubestellar` pod in the `kubestellar` namespace at the location `/home/kubestellar/.kcp/external.kubeconfig`.
