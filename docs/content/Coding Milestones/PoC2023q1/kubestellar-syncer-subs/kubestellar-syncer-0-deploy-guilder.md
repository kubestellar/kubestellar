<!--kubestellar-syncer-0-deploy-guilder-start-->
Go to inventry management workspace and find the mailbox workspace name.
```shell
kubectl ws root:imw-1
mbws=`kubectl get SyncTarget guilder -o jsonpath="{.metadata.annotations['kcp\.io/cluster']}-mb-{.metadata.uid}"`
echo "mailbox workspace name = $mbws"
```
``` { .bash .no-copy }
Current workspace is "root:imw-1".
mailbox workspace name = vosh9816n2xmpdwm-mb-bf1277df-0da9-4a26-b0fc-3318862b1a5e
```

Go to the mailbox workspace and run the following command to obtain yaml manifests to bootstrap KubeStellar-Syncer.
```shell
kubectl ws root:espw:$mbws
./bin/kubectl-kubestellar-syncer_gen guilder --syncer-image quay.io/kubestellar/syncer:v0.2.2 -o guilder-syncer.yaml
```
``` { .bash .no-copy }
Current workspace is "root:espw:vosh9816n2xmpdwm-mb-bf1277df-0da9-4a26-b0fc-3318862b1a5e".
Creating service account "kcp-edge-syncer-guilder-wfeig2lv"
Creating cluster role "kcp-edge-syncer-guilder-wfeig2lv" to give service account "kcp-edge-syncer-guilder-wfeig2lv"

1. write and sync access to the synctarget "kcp-edge-syncer-guilder-wfeig2lv"
2. write access to apiresourceimports.

Creating or updating cluster role binding "kcp-edge-syncer-guilder-wfeig2lv" to bind service account "kcp-edge-syncer-guilder-wfeig2lv" to cluster role "kcp-edge-syncer-guilder-wfeig2lv".

Wrote physical cluster manifest to guilder-syncer.yaml for namespace "kcp-edge-syncer-guilder-wfeig2lv". Use

  KUBECONFIG=<pcluster-config> kubectl apply -f "guilder-syncer.yaml"

to apply it. Use

  KUBECONFIG=<pcluster-config> kubectl get deployment -n "kcp-edge-syncer-guilder-wfeig2lv" kcp-edge-syncer-guilder-wfeig2lv

to verify the syncer pod is running.
Current workspace is "root:espw".
```

Deploy the generated yaml manifest to the target cluster.
```shell
KUBECONFIG=~/.kube/config kubectl --context kind-guilder apply -f guilder-syncer.yaml
```
``` { .bash .no-copy }
namespace/kcp-edge-syncer-guilder-wfeig2lv created
serviceaccount/kcp-edge-syncer-guilder-wfeig2lv created
secret/kcp-edge-syncer-guilder-wfeig2lv-token created
clusterrole.rbac.authorization.k8s.io/kcp-edge-syncer-guilder-wfeig2lv created
clusterrolebinding.rbac.authorization.k8s.io/kcp-edge-syncer-guilder-wfeig2lv created
secret/kcp-edge-syncer-guilder-wfeig2lv created
deployment.apps/kcp-edge-syncer-guilder-wfeig2lv created
```
    
Check that the syncer is running, as follows.
```shell
KUBECONFIG=~/.kube/config kubectl --context kind-guilder get deploy -A
```
``` { .bash .no-copy }
NAMESPACE                          NAME                               READY   UP-TO-DATE   AVAILABLE   AGE
kcp-edge-syncer-guilder-saaywsu5   kcp-edge-syncer-guilder-saaywsu5   1/1     1            1           52s
kube-system                        coredns                            2/2     2            2           35m
local-path-storage                 local-path-provisioner             1/1     1            1           35m
```

<!--kubestellar-syncer-0-deploy-guilder-end-->