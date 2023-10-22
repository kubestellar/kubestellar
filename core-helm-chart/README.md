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

[Create a **Kind** cluster with the `extraPortMappings` for port `1024` and an **nginx** ingress with SSL passthrough.](https://docs.kubestellar.io/main/Coding%20Milestones/PoC2023q1/environments/dev-env/#hosting-kubestellar-in-a-kind-cluster)

Deploy **KubeStellar** with a specific host name `my-long-app-name.aregion.some.cloud.com` and a `1024` port, matching **Kind** ingress port above:

```shell
kubectl create namespace kubestellar
helm install kubestellar/kubestellar-core . \
  --set EXTERNAL_HOSTNAME="kubestellar.core" \
  --set EXTERNAL_PORT=1024 \
  --namespace kubestellar \
  --generate-name
```

Use `--namespace` argument to specify an optional user-defined namespace for the deployment of **KubeStellar**, *e.g.* `--namespace kubestellar`.

**important:** Add 'kubestellar.core' to your /etc/hosts file with the local network IP address (e.g., 192.168.x.y) where your ks-core Kind cluster is running. DO NOT use 127.0.0.1 because the edge-cluster1 and edge-cluster2 kind clusters map 127.0.0.1 to their local kubernetes cluster, not the ks-core kind cluster.

## Deploy **KubeStellar** in an **OpenShift** cluster

Use `--namespace` argument to specify an optional user-defined namespace for the deployment of **KubeStellar**, *e.g.* `--namespace kubestellar`.

As an alternative, one could also create a new project and install the Helm chart in that project, *e.g.* `oc new-project kubestellar`.

Deploy **KubeStellar** in an **OpenShift** cluster, letting the cluster decide the route assigned to **KubeStellar** on the default port `443`:

```shell
helm install kubestellar/kubestellar-core . \
  --set clusterType=OpenShift \
  --namespace kubestellar
```

## Wait for **KubeStellar** to be ready

```shell
echo -n 'Waiting for KubeStellar to be ready'
while ! kubectl exec $(kubectl get pod --selector=app=kubestellar -o jsonpath='{.items[0].metadata.name}' -n kubestellar) -c init -- ls /home/kubestellar/ready &> /dev/null; do
    sleep 10
    echo -n "."
done
echo "KubeStellar is now ready to take requests"
```

## Check **KubeStellar** logs
<!--check-log-start-->
The logs of each runtime container in the **KubeStellar** application pods can be access this way:

```shell
kubectl logs $(kubectl get pod --selector=app=kubestellar -o jsonpath='{.items[0].metadata.name}' -n kubestellar) -n kubestellar -c kcp
kubectl logs $(kubectl get pod --selector=app=kubestellar -o jsonpath='{.items[0].metadata.name}' -n kubestellar) -n kubestellar -c init
kubectl logs $(kubectl get pod --selector=app=kubestellar -o jsonpath='{.items[0].metadata.name}' -n kubestellar) -n kubestellar -c mailbox-controller
kubectl logs $(kubectl get pod --selector=app=kubestellar -o jsonpath='{.items[0].metadata.name}' -n kubestellar) -n kubestellar -c where-resolver
kubectl logs $(kubectl get pod --selector=app=kubestellar -o jsonpath='{.items[0].metadata.name}' -n kubestellar) -n kubestellar -c placement-translator
```
<!--check-log-end-->
## Access **KubeStellar** after deployment

The `kubestellar` deployment, secures it's access kubeconfigs in a `kubestellar` secret.

After the deployment has completed, for you to access **KubeStellar** from the host OS, the `external.kubeconfig` must be extracted from the `kubestellar` secret:

```shell
kubectl get secrets kubestellar -o jsonpath='{.data.external\.kubeconfig}' -n kubestellar | base64 -d > ks-core.kubeconfig
```

**NOTE:** currently, the `external.kubeconfig` needs to be retrieved from the `kubestellar` secret after each restart/recreation of the **KubeStellar** pod.

## Obtain **kcp** and/or **KubeStellar** plugins/executables

If matching plugins/executables are already available locally, then this step is unnecessary.

```
brew tap kubestellar/kubestellar
brew install kcp_cli
brew install kubestellar_cli
```
