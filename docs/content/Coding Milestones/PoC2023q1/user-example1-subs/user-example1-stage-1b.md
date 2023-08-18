<!--user-example1-stage-1b-start-->
### Connect to the stimpy workload execution cluster with its mailbox workspace

The following command will (a) create, in the mailbox workspace for
stimpy, an identity and authorizations for the edge syncer and (b)
write a file containing YAML for deploying the syncer in the stimpy
cluster.

```shell
kubectl kubestellar prep-for-syncer --imw root:imw-1 stimpy
```
``` { .bash .no-copy }
Current workspace is "root:imw-1".
Current workspace is "root:espw".
Current workspace is "root:espw:1t82bk54r6gjnzsp-mb-f0a82ab1-63f4-49ea-954d-3a41a35a9f1c" (type root:universal).
Creating service account "kubestellar-syncer-stimpy-wfeig2lv"
Creating cluster role "kubestellar-syncer-stimpy-wfeig2lv" to give service account "kubestellar-syncer-stimpy-wfeig2lv"

 1. write and sync access to the synctarget "kubestellar-syncer-stimpy-wfeig2lv"
 2. write access to apiresourceimports.

Creating or updating cluster role binding "kubestellar-syncer-stimpy-wfeig2lv" to bind service account "kubestellar-syncer-stimpy-wfeig2lv" to cluster role "kubestellar-syncer-stimpy-wfeig2lv".

Wrote physical cluster manifest to stimpy-syncer.yaml for namespace "kubestellar-syncer-stimpy-wfeig2lv". Use

  KUBECONFIG=<workload-execution-cluster-config> kubectl apply -f "stimpy-syncer.yaml"

to apply it. Use

  KUBECONFIG=<workload-execution-cluster-config> kubectl get deployment -n "kubestellar-syncer-stimpy-wfeig2lv" kubestellar-syncer-stimpy-wfeig2lv

to verify the syncer pod is running.
Current workspace is "root:espw".
```

The file written was, as mentioned in the output,
`stimpy-syncer.yaml`.  Next `kubectl apply` that to the stimpy
cluster.  That will look something like the following; adjust as
necessary to make kubectl manipulate **your** stimpy cluster.

```shell
KUBECONFIG=~/.kube/config kubectl --context kind-stimpy apply -f stimpy-syncer.yaml
```
``` { .bash .no-copy }
namespace/kubestellar-syncer-stimpy-wfeig2lv created
serviceaccount/kubestellar-syncer-stimpy-wfeig2lv created
secret/kubestellar-syncer-stimpy-wfeig2lv-token created
clusterrole.rbac.authorization.k8s.io/kubestellar-syncer-stimpy-wfeig2lv created
clusterrolebinding.rbac.authorization.k8s.io/kubestellar-syncer-stimpy-wfeig2lv created
secret/kubestellar-syncer-stimpy-wfeig2lv created
deployment.apps/kubestellar-syncer-stimpy-wfeig2lv created
```

You might check that the syncer is running, as follows.

```shell
KUBECONFIG=~/.kube/config kubectl --context kind-stimpy get deploy -A
```
``` { .bash .no-copy }
NAMESPACE                          NAME                               READY   UP-TO-DATE   AVAILABLE   AGE
kubestellar-syncer-stimpy-saaywsu5   kubestellar-syncer-stimpy-saaywsu5   1/1     1            1           52s
kube-system                        coredns                            2/2     2            2           35m
local-path-storage                 local-path-provisioner             1/1     1            1           35m
```

### Connect ren edge cluster with its mailbox workspace

Do the analogous stuff for the ren cluster.

```shell
kubectl kubestellar prep-for-syncer --imw root:imw-1 ren
```
``` { .bash .no-copy }
Current workspace is "root:imw-1".
Current workspace is "root:espw".
Current workspace is "root:espw:1t82bk54r6gjnzsp-mb-1a045336-8178-4026-8a56-5cd5609c0ec1" (type root:universal).
Creating service account "kubestellar-syncer-ren-32uaph9l"
Creating cluster role "kubestellar-syncer-ren-32uaph9l" to give service account "kubestellar-syncer-ren-32uaph9l"

 1. write and sync access to the synctarget "kubestellar-syncer-ren-32uaph9l"
 2. write access to apiresourceimports.

Creating or updating cluster role binding "kubestellar-syncer-ren-32uaph9l" to bind service account "kubestellar-syncer-ren-32uaph9l" to cluster role "kubestellar-syncer-ren-32uaph9l".

Wrote physical cluster manifest to ren-syncer.yaml for namespace "kubestellar-syncer-ren-32uaph9l". Use

  KUBECONFIG=<workload-execution-cluster-config> kubectl apply -f "ren-syncer.yaml"

to apply it. Use

  KUBECONFIG=<workload-execution-cluster-config> kubectl get deployment -n "kubestellar-syncer-ren-32uaph9l" kubestellar-syncer-ren-32uaph9l

to verify the syncer pod is running.
Current workspace is "root:espw".
```

And deploy the syncer in the ren cluster.

```shell
KUBECONFIG=~/.kube/config kubectl --context kind-ren apply -f ren-syncer.yaml 
```
``` { .bash .no-copy }
namespace/kubestellar-syncer-ren-32uaph9l created
serviceaccount/kubestellar-syncer-ren-32uaph9l created
secret/kubestellar-syncer-ren-32uaph9l-token created
clusterrole.rbac.authorization.k8s.io/kubestellar-syncer-ren-32uaph9l created
clusterrolebinding.rbac.authorization.k8s.io/kubestellar-syncer-ren-32uaph9l created
secret/kubestellar-syncer-ren-32uaph9l created
deployment.apps/kubestellar-syncer-ren-32uaph9l created
```
<!--user-example1-stage-1b-end-->
