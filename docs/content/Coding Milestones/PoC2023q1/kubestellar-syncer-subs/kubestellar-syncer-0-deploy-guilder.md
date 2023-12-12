<!--kubestellar-syncer-0-deploy-guilder-start-->
Go to KCS and find the mailbox space name.
```shell
espw_kubeconfig="${PWD}/espw.kubeconfig"
kubectl-kubestellar-space-get_kubeconfig espw --kubeconfig $SM_CONFIG $espw_kubeconfig

pvname=`kubectl --kubeconfig $espw_kubeconfig get synctargets.edge.kubestellar.io | grep guilder | awk '{print $1}'`
stuid=`kubectl --kubeconfig $espw_kubeconfig get synctargets.edge.kubestellar.io $pvname -o jsonpath="{.metadata.uid}"`
mbs_name="imw1-mb-$stuid"
echo "mailbox space name = $mbs_name"
```

``` { .bash .no-copy }
mailbox space name = vosh9816n2xmpdwm-mb-bf1277df-0da9-4a26-b0fc-3318862b1a5e
```

Go to the mailbox space and run the following command to obtain yaml manifests to bootstrap KubeStellar-Syncer.
```shell
mbs_kubeconfig="${MY_KUBECONFIGS}/${mbs_name}.kubeconfig"
kubectl-kubestellar-space-get_kubeconfig ${mbs_name} --kubeconfig $SM_CONFIG $mbs_kubeconfig

./bin/kubectl-kubestellar-syncer_gen --kubeconfig $mbs_kubeconfig guilder --syncer-image quay.io/kubestellar/syncer:v0.2.2 -o guilder-syncer.yaml
```
``` { .bash .no-copy }
Current workspace is "root:vosh9816n2xmpdwm-mb-bf1277df-0da9-4a26-b0fc-3318862b1a5e".
Creating service account "kubestellar-syncer-guilder-wfeig2lv"
Creating cluster role "kubestellar-syncer-guilder-wfeig2lv" to give service account "kubestellar-syncer-guilder-wfeig2lv"

1. write and sync access to the synctarget "kubestellar-syncer-guilder-wfeig2lv"
2. write access to apiresourceimports.

Creating or updating cluster role binding "kubestellar-syncer-guilder-wfeig2lv" to bind service account "kubestellar-syncer-guilder-wfeig2lv" to cluster role "kubestellar-syncer-guilder-wfeig2lv".

Wrote WEC manifest to guilder-syncer.yaml for namespace "kubestellar-syncer-guilder-wfeig2lv". Use

  KUBECONFIG=<workload-execution-cluster-config> kubectl apply -f "guilder-syncer.yaml"

to apply it. Use

  KUBECONFIG=<workload-execution-cluster-config> kubectl get deployment -n "kubestellar-syncer-guilder-wfeig2lv" kubestellar-syncer-guilder-wfeig2lv

to verify the syncer pod is running.
Current workspace is "root:espw".
```

Deploy the generated yaml manifest to the target cluster.
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
    
Check that the syncer is running, as follows.
```shell
KUBECONFIG=~/.kube/config kubectl --context kind-guilder get deploy -A
```
``` { .bash .no-copy }
NAMESPACE                             NAME                                  READY   UP-TO-DATE   AVAILABLE   AGE
kubestellar-syncer-guilder-saaywsu5   kubestellar-syncer-guilder-saaywsu5   1/1     1            1           52s
kube-system                           coredns                               2/2     2            2           35m
local-path-storage                    local-path-provisioner                1/1     1            1           35m
```

<!--kubestellar-syncer-0-deploy-guilder-end-->
