<!--kubestellar-prep-syncer-test-start-->
```shell hl_lines="4 9"
KUBECONFIG=ks-core.kubeconfig kubectl kubestellar prep-for-cluster --imw imw1 ks-edge-cluster1 \
  --syncer-image ko.local/syncer:test \
  env=ks-edge-cluster1 \
  location-group=edge     #add ks-edge-cluster1 and ks-edge-cluster2 to the same group

KUBECONFIG=ks-core.kubeconfig kubectl kubestellar prep-for-cluster --imw imw1 ks-edge-cluster2 \
  --syncer-image ko.local/syncer:test \
  env=ks-edge-cluster2 \
  location-group=edge     #add ks-edge-cluster1 and ks-edge-cluster2 to the same group
```
<!--kubestellar-prep-syncer-test-end-->