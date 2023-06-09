<!--kubestellar-syncer-0-deploy-florin-start-->
Go to inventry management workspace and find the mailbox workspace name.
```shell
kubectl ws root:imw-1
mbws=`kubectl get SyncTarget florin -o jsonpath="{.metadata.annotations['kcp\.io/cluster']}-mb-{.metadata.uid}"`
echo "mailbox workspace name = $mbws"
```
``` { .bash .no-copy }
Current workspace is "root:imw-1".
mailbox workspace name = vosh9816n2xmpdwm-mb-bb47149d-52d3-4f14-84dd-7b64ac01c97f
```

Go to the mailbox workspace and run the following command to obtain yaml manifests to bootstrap KubeStellar-Syncer.
```shell
kubectl ws root:espw:$mbws
./bin/kubectl-kubestellar-syncer_gen florin --syncer-image quay.io/kubestellar/syncer:v0.2.2 -o florin-syncer.yaml
```
``` { .bash .no-copy }
Current workspace is "root:espw:vosh9816n2xmpdwm-mb-bb47149d-52d3-4f14-84dd-7b64ac01c97f".
Creating service account "kcp-edge-syncer-florin-32uaph9l"
Creating cluster role "kcp-edge-syncer-florin-32uaph9l" to give service account "kcp-edge-syncer-florin-32uaph9l"

 1. write and sync access to the synctarget "kcp-edge-syncer-florin-32uaph9l"
 2. write access to apiresourceimports.

Creating or updating cluster role binding "kcp-edge-syncer-florin-32uaph9l" to bind service account "kcp-edge-syncer-florin-32uaph9l" to cluster role "kcp-edge-syncer-florin-32uaph9l".

Wrote physical cluster manifest to florin-syncer.yaml for namespace "kcp-edge-syncer-florin-32uaph9l". Use

  KUBECONFIG=<pcluster-config> kubectl apply -f "florin-syncer.yaml"

to apply it. Use

  KUBECONFIG=<pcluster-config> kubectl get deployment -n "kcp-edge-syncer-florin-32uaph9l" kcp-edge-syncer-florin-32uaph9l

to verify the syncer pod is running.
Current workspace is "root:espw".
```

Deploy the generated yaml manifest to the target cluster.
```shell
KUBECONFIG=~/.kube/config kubectl --context kind-florin apply -f florin-syncer.yaml
```
``` { .bash .no-copy }
namespace/kcp-edge-syncer-florin-32uaph9l created
serviceaccount/kcp-edge-syncer-florin-32uaph9l created
secret/kcp-edge-syncer-florin-32uaph9l-token created
clusterrole.rbac.authorization.k8s.io/kcp-edge-syncer-florin-32uaph9l created
clusterrolebinding.rbac.authorization.k8s.io/kcp-edge-syncer-florin-32uaph9l created
secret/kcp-edge-syncer-florin-32uaph9l created
deployment.apps/kcp-edge-syncer-florin-32uaph9l created
```
    
Check that the syncer is running, as follows.
```shell
KUBECONFIG=~/.kube/config kubectl --context kind-florin get deploy -A
```
``` { .bash .no-copy }
NAMESPACE                          NAME                               READY   UP-TO-DATE   AVAILABLE   AGE
kcp-edge-syncer-florin-32uaph9l    kcp-edge-syncer-florin-32uaph9l    1/1     1            1           42s
kube-system                        coredns                            2/2     2            2           41m
local-path-storage                 local-path-provisioner             1/1     1            1           41m
```

<!--kubestellar-syncer-0-deploy-florin-end-->