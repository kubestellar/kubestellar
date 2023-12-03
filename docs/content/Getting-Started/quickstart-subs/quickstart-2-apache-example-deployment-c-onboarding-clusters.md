<!--quickstart-2-apache-example-deployment-c-onboarding-clusters-start-->
KubeStellar will have created an Inventory Management space (IMW)
for the user to put inventory objects in, describing the user's
clusters. The IMW that is automatically created for the user is at
`imw1`.

Let's begin by onboarding the `florin` cluster:

```shell
KUBECONFIG=$SM_CONFIG kubectl kubestellar prep-for-cluster --imw root:imw1 florin env=prod
```

which should yield something like:

``` { .sh .no-copy }
synctarget.edge.kubestellar.io/florin created
location.edge.kubestellar.io/florin created
synctarget.edge.kubestellar.io/florin labeled
location.edge.kubestellar.io/florin labeled
Creating service account "kubestellar-syncer-florin-1yi5q9c4"
Creating cluster role "kubestellar-syncer-florin-1yi5q9c4" to give service account "kubestellar-syncer-florin-1yi5q9c4"

 1. write and sync access to the synctarget "kubestellar-syncer-florin-1yi5q9c4"
 2. write access to apiresourceimports.

Creating or updating cluster role binding "kubestellar-syncer-florin-1yi5q9c4" to bind service account "kubestellar-syncer-florin-1yi5q9c4" to cluster role "kubestellar-syncer-florin-1yi5q9c4".

Wrote workload execution cluster (WEC) manifest to florin-syncer.yaml for namespace "kubestellar-syncer-florin-1yi5q9c4". Use

  KUBECONFIG=<workload-execution-cluster-config> kubectl apply -f "florin-syncer.yaml"

to apply it. Use

  KUBECONFIG=<workload-execution-cluster-config> kubectl get deployment -n "kubestellar-syncer-florin-1yi5q9c4" kubestellar-syncer-florin-1yi5q9c4

to verify the syncer pod is running.
```

An edge syncer manifest yaml file was created in your current directory: `florin-syncer.yaml`. The default for the output file is the name of the SyncTarget object with “-syncer.yaml” appended.

Now let's deploy the edge syncer to the `florin` edge cluster:

  
```shell
kubectl --context kind-florin apply -f florin-syncer.yaml
```

which should yield something like:

``` { .sh .no-copy }
namespace/kubestellar-syncer-florin-1yi5q9c4 created
serviceaccount/kubestellar-syncer-florin-1yi5q9c4 created
secret/kubestellar-syncer-florin-1yi5q9c4-token created
clusterrole.rbac.authorization.k8s.io/kubestellar-syncer-florin-1yi5q9c4 created
clusterrolebinding.rbac.authorization.k8s.io/kubestellar-syncer-florin-1yi5q9c4 created
secret/kubestellar-syncer-florin-1yi5q9c4 created
deployment.apps/kubestellar-syncer-florin-1yi5q9c4 created
```

Optionally, check that the edge syncer pod is running:

```shell
kubectl --context kind-florin get pods -A
```

which should yield something like:

``` { .sh .no-copy }
NAMESPACE                            NAME                                                  READY   STATUS    RESTARTS   AGE
kubestellar-syncer-florin-1yi5q9c4   kubestellar-syncer-florin-1yi5q9c4-77cb588c89-xc5qr   1/1     Running   0          12m
kube-system                          coredns-565d847f94-92f4k                              1/1     Running   0          58m
kube-system                          coredns-565d847f94-kgddm                              1/1     Running   0          58m
kube-system                          etcd-florin-control-plane                             1/1     Running   0          58m
kube-system                          kindnet-p8vkv                                         1/1     Running   0          58m
kube-system                          kube-apiserver-florin-control-plane                   1/1     Running   0          58m
kube-system                          kube-controller-manager-florin-control-plane          1/1     Running   0          58m
kube-system                          kube-proxy-jmxsg                                      1/1     Running   0          58m
kube-system                          kube-scheduler-florin-control-plane                   1/1     Running   0          58m
local-path-storage                   local-path-provisioner-684f458cdd-kw2xz               1/1     Running   0          58m

``` 

Now, let's onboard the `guilder` cluster:

```shell
KUBECONFIG=$SM_CONFIG kubectl kubestellar prep-for-cluster --imw root:imw1 guilder env=prod extended=yes
```

Apply the created edge syncer manifest:

```shell
kubectl --context kind-guilder apply -f guilder-syncer.yaml
```
<!--quickstart-2-apache-example-deployment-c-onboarding-clusters-end-->
