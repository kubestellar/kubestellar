# Deploy **KubeStellar** in a cluster using Helm

Table of contents:

- [Deploy **KubeStellar** in a cluster using Helm](#deploy-kubestellar-in-a-cluster-using-helm)
  - [Deploy **KubeStellar** in a **Kubernetes** cluster (**Kind** cluster)](#deploy-kubestellar-in-a-kubernetes-cluster-kind-cluster)
  - [Deploy **KubeStellar** in an **OpenShift** cluster](#deploy-kubestellar-in-an-openshift-cluster)
  - [Wait for **KubeStellar** to be ready](#wait-for-kubestellar-to-be-ready)
  - [Check **KubeStellar** logs](#check-kubestellar-logs)
  - [Access **KubeStellar** after deployment](#access-kubestellar-after-deployment)
  - [Obtain **kcp** and/or **KubeStellar** plugins/executables](#obtain-kcp-andor-kubestellar-pluginsexecutables)

## Deploy **KubeStellar** in a **Kubernetes** cluster (**Kind** cluster)

[Create a **Kind** cluster with the `extraPortMappings` for port `1119` and an **nginx** ingress with SSL passthrough.](https://docs.kubestellar.io/main/Coding%20Milestones/PoC2023q1/environments/dev-env/#hosting-kubestellar-in-a-kind-cluster)

Deploy **KubeStellar** with a specific host name `my-long-app-name.aregion.some.cloud.com` and a `1119` port, matching **Kind** ingress port above:

```shell
helm install kubestellar . \
  --set EXTERNAL_HOSTNAME="my-long-app-name.aregion.some.cloud.com" \
  --set EXTERNAL_PORT=1119
```

Use `--namespace` argument to specify an optional user-defined namespace for the deployment of **KubeStellar**, *e.g.* `--namespace kubestellar`.

## Deploy **KubeStellar** in an **OpenShift** cluster

Use `--namespace` argument to specify an optional user-defined namespace for the deployment of **KubeStellar**, *e.g.* `--namespace kubestellar`.

As an alternative, one could also create a new project and install the Helm chart in that project, *e.g.* `oc new-project kubestellar`.

Deploy **KubeStellar** in an **OpenShift** cluster, letting the cluster decide the route assigned to **KubeStellar** on the default port `443`:

```shell
helm install kubestellar . \
  --set clusterType=OpenShift
```

## Wait for **KubeStellar** to be ready

```shell
echo -n 'Waiting for KubeStellar to be ready'
while ! kubectl exec $(kubectl get pod --selector=app=kubestellar -o jsonpath='{.items[0].metadata.name}') -c init -- ls /home/kubestellar/ready &> /dev/null; do
    sleep 10
    echo -n "."
done
echo "Ready!"
```

## Check **KubeStellar** logs
<!--check-log-start-->
The logs of each runtime container in the **KubeStellar** application pods can be access this way:

```shell
kubectl logs $(kubectl get pod --selector=app=kubestellar -o jsonpath='{.items[0].metadata.name}') -c kcp
kubectl logs $(kubectl get pod --selector=app=kubestellar -o jsonpath='{.items[0].metadata.name}') -c init
kubectl logs $(kubectl get pod --selector=app=kubestellar -o jsonpath='{.items[0].metadata.name}') -c mailbox-controller
kubectl logs $(kubectl get pod --selector=app=kubestellar -o jsonpath='{.items[0].metadata.name}') -c where-resolver
kubectl logs $(kubectl get pod --selector=app=kubestellar -o jsonpath='{.items[0].metadata.name}') -c placement-translator
```
<!--check-log-end-->
## Access **KubeStellar** after deployment

The `kubestellar` deployment, holds its access kubeconfigs in a `kubestellar` secret.

After the deployment has completed, in order to access **KubeStellar** from the host OS, the `external.kubeconfig` must be extracted from the `kubestellar` secret:

```shell
kubectl get secrets kubestellar -o 'go-template={{index .data "external.kubeconfig"}}' | base64 --decode > admin.kubeconfig
```

**NOTE:** currently, the `external.kubeconfig` needs to be retrieved from the `kubestellar` secret after each restart/recreation of the **KubeStellar** pod.

## Obtain **kcp** and/or **KubeStellar** plugins/executables

If matching plugins/executables are already available locally, then this step is unnecessary.

If the host OS and cluster share the same OS type and architecture, then the plugins/executables can be extracted from the container image.

Obtaining the **kcp** plugins (this copy preserves links):

```shell
kubectl exec $(kubectl get pod --selector=app=kubestellar -o jsonpath='{.items[0].metadata.name}') -c init -- tar cf - /home/kubestellar/kcp/bin | tar xf - --strip-components=2
```

Obtaining the **KubeStellar** plugins/executables:

```shell
kubectl cp $(kubectl get pod --selector=app=kubestellar -o jsonpath='{.items[0].metadata.name}'):/home/kubestellar/bin kubestellar/bin -c init
```

Add the plugins and executables to the PATH:

```shell
export PATH=$PATH:$PWD/kcp/bin:$PWD/kubestellar/bin
```

If the host OS and cluster do not share the same OS type and architecture, then compatible plugins must be obtained from a corresponding **kcp** release at https://github.com/kcp-dev/kcp/releases and **KubeStellar** release at https://github.com/kubestellar/kubestellar/releases.
