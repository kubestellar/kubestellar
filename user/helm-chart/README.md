# Deploy **KubeStellar** service in a cluster using Helm

Table of contests:
- [Deploy **KubeStellar** service in a cluster using Helm](#deploy-kubestellar-service-in-a-cluster-using-helm)
  - [Deploy **KubeStellar** in a **Kubernetes** cluster (**Kind** cluster)](#deploy-kubestellar-in-a-kubernetes-cluster-kind-cluster)
  - [Deploy **KubeStellar** in an **OpenShift** cluster](#deploy-kubestellar-in-an-openshift-cluster)
  - [Accessing **KubeStellar** after deployment](#accessing-kubestellar-after-deployment)

## Deploy **KubeStellar** in a **Kubernetes** cluster (**Kind** cluster)

[Create a **Kind** cluster with the `extraPortMappings` for port `1024` and an **nginx** ingreass with SSL passthrough.](../yaml/README.md)

Deploy **KubeStellar** `stable` in a `kubestellar` namespace, with a specific host name `kubestellar.svc.cluster.local` and a `1024` port number:

```shell
helm install kubestellar . --set HOSTNAME="kubestellar.svc.cluster.local" --set EXTERNAL_PORT=1024
```

## Deploy **KubeStellar** in an **OpenShift** cluster

Deploy **KubeStellar** `stable` in a `kubestellar` namespace, in an **OpenShift** cluster, letting the cluster decide the route assigned to **KubeStellar** on the default port `443`:

```shell
helm install kubestellar . --set clusterType=OpenShift
```

## Accessing **KubeStellar** after deployment

The `kubestellar-server` deployment, holds its access kubeconfigs in a `kubestellar` secret in the `kubestellar` namespace, which it manages using a `kubestellar-role`. Additionally, the role allows the pod to get its ingress/route to put in the `external.kubeconfig`.

After the deployment has completed, **KubeStellar** `admin.kubeconfig` can be obtained in two ways:

- the `kubestellar` secret in the `kubestellar` namespace;
- directly from the `kubestellar` pod in the `kubestellar` namespace at the location `/home/kubestellar/.kcp/external.kubeconfig`.
