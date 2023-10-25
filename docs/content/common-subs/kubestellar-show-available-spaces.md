<!--kubestellar-show-available-spaces-start-->
```shell
KUBECONFIG=~/.kube/config kubectl --context ks-core get secrets kubestellar \
  -o jsonpath='{.data.external\.kubeconfig}' \
  -n kubestellar | base64 -d > ks-core.kubeconfig

KUBECONFIG=ks-core.kubeconfig kubectl ws --context root tree
```
<!--kubestellar-show-available-spaces-end-->