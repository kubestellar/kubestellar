<!--kubestellar-prep-syncer-start-->
```shell hl_lines="3 7"
KUBECONFIG=ks-core.kubeconfig kubectl kubestellar prep-for-cluster --imw root:imw1 ks-edge-cluster1 \
  env=ks-edge-cluster1 \
  location-group=edge     #add ks-edge-cluster1 and ks-edge-cluster2 to the same group

KUBECONFIG=ks-core.kubeconfig kubectl kubestellar prep-for-cluster --imw root:imw1 ks-edge-cluster2 \
  env=ks-edge-cluster2 \
  location-group=edge     #add ks-edge-cluster1 and ks-edge-cluster2 to the same group
```
<!--kubestellar-prep-syncer-end-->