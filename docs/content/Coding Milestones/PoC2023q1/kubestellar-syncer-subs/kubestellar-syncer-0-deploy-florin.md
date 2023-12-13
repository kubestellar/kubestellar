<!--kubestellar-syncer-0-deploy-florin-start-->
Go to KCS and find the mailbox space name.
```shell
pvname=`kubectl --kubeconfig $espw_kubeconfig get synctargets.edge.kubestellar.io | grep florin | awk '{print $1}'`
stuid=`kubectl --kubeconfig $espw_kubeconfig get synctargets.edge.kubestellar.io $pvname -o jsonpath="{.metadata.uid}"`
mbs_name="imw1-mb-$stuid"
echo "mailbox space name = $mbs_name"
```
``` { .bash .no-copy }
mailbox space name = vosh9816n2xmpdwm-mb-bb47149d-52d3-4f14-84dd-7b64ac01c97f
```

Go to the mailbox space and run the following command to obtain yaml manifests to bootstrap KubeStellar-Syncer.
```shell
mbs_kubeconfig="${MY_KUBECONFIGS}/${mbs_name}.kubeconfig"
kubectl-kubestellar-space-get_kubeconfig ${mbs_name} $in_cluster --kubeconfig $SM_CONFIG $mbs_kubeconfig

./bin/kubectl-kubestellar-syncer_gen --kubeconfig $mbs_kubeconfig florin --syncer-image quay.io/kubestellar/syncer:v0.2.2 -o florin-syncer.yaml
```
``` { .bash .no-copy }
Current workspace is "root:vosh9816n2xmpdwm-mb-bb47149d-52d3-4f14-84dd-7b64ac01c97f".
Creating service account "kubestellar-syncer-florin-32uaph9l"
Creating cluster role "kubestellar-syncer-florin-32uaph9l" to give service account "kubestellar-syncer-florin-32uaph9l"

 1. write and sync access to the synctarget "kubestellar-syncer-florin-32uaph9l"
 2. write access to apiresourceimports.

Creating or updating cluster role binding "kubestellar-syncer-florin-32uaph9l" to bind service account "kubestellar-syncer-florin-32uaph9l" to cluster role "kubestellar-syncer-florin-32uaph9l".

Wrote workload execution cluster (WEC) manifest to florin-syncer.yaml for namespace "kubestellar-syncer-florin-32uaph9l". Use

  KUBECONFIG=<workload-execution-cluster-config> kubectl apply -f "florin-syncer.yaml"

to apply it. Use

  KUBECONFIG=<workload-execution-cluster-config> kubectl get deployment -n "kubestellar-syncer-florin-32uaph9l" kubestellar-syncer-florin-32uaph9l

to verify the syncer pod is running.
Current workspace is "root:espw".
```

Deploy the generated yaml manifest to the target cluster.
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
    
Check that the syncer is running, as follows.
```shell
KUBECONFIG=~/.kube/config kubectl --context kind-florin get deploy -A
```
``` { .bash .no-copy }
NAMESPACE                             NAME                                  READY   UP-TO-DATE   AVAILABLE   AGE
kubestellar-syncer-florin-32uaph9l    kubestellar-syncer-florin-32uaph9l    1/1     1            1           42s
kube-system                           coredns                               2/2     2            2           41m
local-path-storage                    local-path-provisioner                1/1     1            1           41m
```

<!--kubestellar-syncer-0-deploy-florin-end-->
