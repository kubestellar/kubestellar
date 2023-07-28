<!--example1-stage-1b-start-->
### Connect guilder edge cluster with its mailbox workspace

The following command will (a) create, in the mailbox workspace for
guilder, an identity and authorizations for the edge syncer and (b)
write a file containing YAML for deploying the syncer in the guilder
cluster.

```shell
kubectl kcp bind apiexport root:espw:edge.kcp.io
kubectl kubestellar prep-for-syncer --imw root:imw-1 guilder
```
``` { .bash .no-copy }
Current workspace is "root:imw-1".
Current workspace is "root:espw".
Current workspace is "root:espw:1t82bk54r6gjnzsp-mb-f0a82ab1-63f4-49ea-954d-3a41a35a9f1c" (type root:universal).
Creating service account "kubestellar-syncer-guilder-wfeig2lv"
Creating cluster role "kubestellar-syncer-guilder-wfeig2lv" to give service account "kubestellar-syncer-guilder-wfeig2lv"

 1. write and sync access to the synctarget "kubestellar-syncer-guilder-wfeig2lv"
 2. write access to apiresourceimports.

Creating or updating cluster role binding "kubestellar-syncer-guilder-wfeig2lv" to bind service account "kubestellar-syncer-guilder-wfeig2lv" to cluster role "kubestellar-syncer-guilder-wfeig2lv".

Wrote physical cluster manifest to guilder-syncer.yaml for namespace "kubestellar-syncer-guilder-wfeig2lv". Use

  KUBECONFIG=<workload-execution-cluster-config> kubectl apply -f "guilder-syncer.yaml"

to apply it. Use

  KUBECONFIG=<workload-execution-cluster-config> kubectl get deployment -n "kubestellar-syncer-guilder-wfeig2lv" kubestellar-syncer-guilder-wfeig2lv

to verify the syncer pod is running.
Current workspace is "root:espw".
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
NAMESPACE                          NAME                               READY   UP-TO-DATE   AVAILABLE   AGE
kubestellar-syncer-guilder-saaywsu5   kubestellar-syncer-guilder-saaywsu5   1/1     1            1           52s
kube-system                        coredns                            2/2     2            2           35m
local-path-storage                 local-path-provisioner             1/1     1            1           35m
```

### Connect florin edge cluster with its mailbox workspace

Do the analogous stuff for the florin cluster.

```shell
kubectl kubestellar prep-for-syncer --imw root:imw-1 florin
```
``` { .bash .no-copy }
Current workspace is "root:imw-1".
Current workspace is "root:espw".
Current workspace is "root:espw:1t82bk54r6gjnzsp-mb-1a045336-8178-4026-8a56-5cd5609c0ec1" (type root:universal).
Creating service account "kubestellar-syncer-florin-32uaph9l"
Creating cluster role "kubestellar-syncer-florin-32uaph9l" to give service account "kubestellar-syncer-florin-32uaph9l"

 1. write and sync access to the synctarget "kubestellar-syncer-florin-32uaph9l"
 2. write access to apiresourceimports.

Creating or updating cluster role binding "kubestellar-syncer-florin-32uaph9l" to bind service account "kubestellar-syncer-florin-32uaph9l" to cluster role "kubestellar-syncer-florin-32uaph9l".

Wrote physical cluster manifest to florin-syncer.yaml for namespace "kubestellar-syncer-florin-32uaph9l". Use

  KUBECONFIG=<workload-execution-cluster-config> kubectl apply -f "florin-syncer.yaml"

to apply it. Use

  KUBECONFIG=<workload-execution-cluster-config> kubectl get deployment -n "kubestellar-syncer-florin-32uaph9l" kubestellar-syncer-florin-32uaph9l

to verify the syncer pod is running.
Current workspace is "root:espw".
```

And deploy the syncer in the florin cluster.

```shell
KUBECONFIG=~/.kube/config kubectl --context kind-florin apply -f florin-syncer.yaml 
```
``` { .bash .no-copy }
namespace/kubestellar-syncer-florin-32uaph9l created
serviceaccount/kubestellar-syncer-florin-32uaph9l created
secret/kubestellar-syncer-florin-32uaph9l-token created
clusterrole.rbac.authorization.k8s.io/kubestellar-syncer-florin-32uaph9l created
clusterrolebinding.rbac.authorization.k8s.io/kubestellar-syncer-florin-32uaph9l created
secret/kubestellar-syncer-florin-32uaph9l created
deployment.apps/kubestellar-syncer-florin-32uaph9l created
```
<!--example1-stage-1b-end-->