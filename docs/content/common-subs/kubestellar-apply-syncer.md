<!--kubestellar-apply-syncer-start-->
```shell
export KUBECONFIG=~/.kube/config

#apply ks-edge-cluster1 syncer
kubectl --context ks-edge-cluster1 apply -f ks-edge-cluster1-syncer.yaml
sleep 3
kubectl --context ks-edge-cluster1 get pods -A | grep kubestellar  #check if syncer deployed to ks-edge-cluster1 correctly

#apply ks-edge-cluster2 syncer
kubectl --context ks-edge-cluster2 apply -f ks-edge-cluster2-syncer.yaml
sleep 3
kubectl --context ks-edge-cluster2 get pods -A | grep kubestellar  #check if syncer deployed to ks-edge-cluster2 correctly
```
<!--kubestellar-apply-syncer-end-->