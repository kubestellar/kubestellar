<!--quickstart-2-apache-example-deployment-c-onboarding-clusters-start-->
### c. Onboarding the clusters

Let's begin by onboarding the `florin` cluster:

```shell
kubectl ws root
kubectl kubestellar prep-for-cluster --imw root:example-imw florin env=prod
```

which should yield something like:

```console
Current workspace is "root:example-imw".
synctarget.workload.kcp.io/florin created
location.scheduling.kcp.io/florin created
synctarget.workload.kcp.io/florin labeled
location.scheduling.kcp.io/florin labeled
Current workspace is "root:example-imw".
Current workspace is "root:espw".
Current workspace is "root:espw:9nemli4rpx83ahnz-mb-c44d04db-ae85-422c-9e12-c5e7865bf37a" (type root:universal).
Creating service account "kcp-edge-syncer-florin-1yi5q9c4"
Creating cluster role "kcp-edge-syncer-florin-1yi5q9c4" to give service account "kcp-edge-syncer-florin-1yi5q9c4"

 1. write and sync access to the synctarget "kcp-edge-syncer-florin-1yi5q9c4"
 2. write access to apiresourceimports.

Creating or updating cluster role binding "kcp-edge-syncer-florin-1yi5q9c4" to bind service account "kcp-edge-syncer-florin-1yi5q9c4" to cluster role "kcp-edge-syncer-florin-1yi5q9c4".

Wrote physical cluster manifest to florin-syncer.yaml for namespace "kcp-edge-syncer-florin-1yi5q9c4". Use

  KUBECONFIG=<pcluster-config> kubectl apply -f "florin-syncer.yaml"

to apply it. Use

  KUBECONFIG=<pcluster-config> kubectl get deployment -n "kcp-edge-syncer-florin-1yi5q9c4" kcp-edge-syncer-florin-1yi5q9c4

to verify the syncer pod is running.
Current workspace is "root:example-imw".
Current workspace is "root".
```

An edge syncer manifest yaml file was created in your current director: `florin-syncer.yaml`. The default for the output file is the name of the SyncTarget object with “-syncer.yaml” appended.

Now let's deploy the edge syncer to the `florin` edge cluster:

  
```shell
kubectl --context kind-florin apply -f florin-syncer.yaml
```

which should yield something like:

```console
namespace/kcp-edge-syncer-florin-1yi5q9c4 created
serviceaccount/kcp-edge-syncer-florin-1yi5q9c4 created
secret/kcp-edge-syncer-florin-1yi5q9c4-token created
clusterrole.rbac.authorization.k8s.io/kcp-edge-syncer-florin-1yi5q9c4 created
clusterrolebinding.rbac.authorization.k8s.io/kcp-edge-syncer-florin-1yi5q9c4 created
secret/kcp-edge-syncer-florin-1yi5q9c4 created
deployment.apps/kcp-edge-syncer-florin-1yi5q9c4 created
```

Optionally, check that the edge syncer pod is running:

```shell
kubectl --context kind-florin get pods -A
```

which should yield something like:

```console
NAMESPACE                         NAME                                               READY   STATUS    RESTARTS   AGE
kcp-edge-syncer-florin-1yi5q9c4   kcp-edge-syncer-florin-1yi5q9c4-77cb588c89-xc5qr   1/1     Running   0          12m
kube-system                       coredns-565d847f94-92f4k                           1/1     Running   0          58m
kube-system                       coredns-565d847f94-kgddm                           1/1     Running   0          58m
kube-system                       etcd-florin-control-plane                          1/1     Running   0          58m
kube-system                       kindnet-p8vkv                                      1/1     Running   0          58m
kube-system                       kube-apiserver-florin-control-plane                1/1     Running   0          58m
kube-system                       kube-controller-manager-florin-control-plane       1/1     Running   0          58m
kube-system                       kube-proxy-jmxsg                                   1/1     Running   0          58m
kube-system                       kube-scheduler-florin-control-plane                1/1     Running   0          58m
local-path-storage                local-path-provisioner-684f458cdd-kw2xz            1/1     Running   0          58m

``` 

Now, let's onboard the `guilder` cluster:

```shell
kubectl ws root
kubectl kubestellar prep-for-cluster --imw root:example-imw guilder env=prod extended=si
```

Apply the created edge syncer manifest:

```shell
kubectl --context kind-guilder apply -f guilder-syncer.yaml
```
<!--quickstart-2-apache-example-deployment-c-onboarding-clusters-end-->