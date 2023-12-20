<!--example1-stage-1b-start-->
### Connect guilder edge cluster with its mailbox space

The following command will (a) create, in the mailbox space for
guilder, an identity and authorizations for the edge syncer and (b)
write a file containing YAML for deploying the syncer in the guilder
cluster.

```shell
kubectl kubestellar prep-for-syncer --imw imw1 $in_cluster guilder
```
``` { .bash .no-copy }
Creating service account "kubestellar-syncer-kube-bind-sx6pl-guilder-2laxc4df"
Creating cluster role "kubestellar-syncer-kube-bind-sx6pl-guilder-2laxc4df" to give service account "kubestellar-syncer-kube-bind-sx6pl-guilder-2laxc4df"

 1. write and sync access to the synctarget "kubestellar-syncer-kube-bind-sx6pl-guilder-2laxc4df"
 2. write access to apiresourceimports.

Creating or updating cluster role binding "kubestellar-syncer-kube-bind-sx6pl-guilder-2laxc4df" to bind service account "kubestellar-syncer-kube-bind-sx6pl-guilder-2laxc4df" to cluster role "kubestellar-syncer-kube-bind-sx6pl-guilder-2laxc4df".

Wrote workload execution cluster manifest to guilder-syncer.yaml for namespace "kubestellar-syncer-kube-bind-sx6pl-guilder-2laxc4df". Use

  KUBECONFIG=<wec-config> kubectl apply -f "guilder-syncer.yaml"

to apply it. Use

  KUBECONFIG=<wec-config> kubectl get deployment -n "kubestellar-syncer-kube-bind-sx6pl-guilder-2laxc4df" kubestellar-syncer-kube-bind-sx6pl-guilder-2laxc4df

to verify the syncer pod is running.
```

The file written was, as mentioned in the output,
`guilder-syncer.yaml`.  Next `kubectl apply` that to the guilder
cluster.  That will look something like the following; adjust as
necessary to make kubectl manipulate **your** guilder cluster.

```shell
KUBECONFIG=~/.kube/config kubectl --context kind-guilder apply -f guilder-syncer.yaml
```
``` { .bash .no-copy }
namespace/kubestellar-syncer-guilder-wfeig2lv created
serviceaccount/kubestellar-syncer-guilder-wfeig2lv created
secret/kubestellar-syncer-guilder-wfeig2lv-token created
clusterrole.rbac.authorization.k8s.io/kubestellar-syncer-guilder-wfeig2lv created
clusterrolebinding.rbac.authorization.k8s.io/kubestellar-syncer-guilder-wfeig2lv created
secret/kubestellar-syncer-guilder-wfeig2lv created
deployment.apps/kubestellar-syncer-guilder-wfeig2lv created
```

You might check that the syncer is running, as follows.

```shell
KUBECONFIG=~/.kube/config kubectl --context kind-guilder get deploy -A
```
``` { .bash .no-copy }
NAMESPACE                                             NAME                                                  READY   UP-TO-DATE   AVAILABLE   AGE
kube-system                                           coredns                                               2/2     2            2           4m1s
kubestellar-syncer-kube-bind-sx6pl-guilder-2laxc4df   kubestellar-syncer-kube-bind-sx6pl-guilder-2laxc4df   1/1     1            1           0s
local-path-storage                                    local-path-provisioner                                1/1     1            1           3m58s
```

### Connect florin edge cluster with its mailbox space

Do the analogous stuff for the florin cluster.

```shell
kubectl kubestellar prep-for-syncer --imw imw1 $in_cluster florin
```
``` { .bash .no-copy }
Creating service account "kubestellar-syncer-kube-bind-sx6pl-florin-1pa812aj"
Creating cluster role "kubestellar-syncer-kube-bind-sx6pl-florin-1pa812aj" to give service account "kubestellar-syncer-kube-bind-sx6pl-florin-1pa812aj"

 1. write and sync access to the synctarget "kubestellar-syncer-kube-bind-sx6pl-florin-1pa812aj"
 2. write access to apiresourceimports.

Creating or updating cluster role binding "kubestellar-syncer-kube-bind-sx6pl-florin-1pa812aj" to bind service account "kubestellar-syncer-kube-bind-sx6pl-florin-1pa812aj" to cluster role "kubestellar-syncer-kube-bind-sx6pl-florin-1pa812aj".

Wrote workload execution cluster manifest to florin-syncer.yaml for namespace "kubestellar-syncer-kube-bind-sx6pl-florin-1pa812aj". Use

  KUBECONFIG=<wec-config> kubectl apply -f "florin-syncer.yaml"

to apply it. Use

  KUBECONFIG=<wec-config> kubectl get deployment -n "kubestellar-syncer-kube-bind-sx6pl-florin-1pa812aj" kubestellar-syncer-kube-bind-sx6pl-florin-1pa812aj

to verify the syncer pod is running.
```

And deploy the syncer in the florin cluster.

```shell
KUBECONFIG=~/.kube/config kubectl --context kind-florin apply -f florin-syncer.yaml 
```
``` { .bash .no-copy }
namespace/kubestellar-syncer-kube-bind-sx6pl-florin-1pa812aj created
serviceaccount/kubestellar-syncer-kube-bind-sx6pl-florin-1pa812aj created
secret/kubestellar-syncer-kube-bind-sx6pl-florin-1pa812aj-token created
clusterrole.rbac.authorization.k8s.io/kubestellar-syncer-kube-bind-sx6pl-florin-1pa812aj created
clusterrolebinding.rbac.authorization.k8s.io/kubestellar-syncer-kube-bind-sx6pl-florin-1pa812aj created
secret/kubestellar-syncer-kube-bind-sx6pl-florin-1pa812aj created
deployment.apps/kubestellar-syncer-kube-bind-sx6pl-florin-1pa812aj created
```
<!--example1-stage-1b-end-->
