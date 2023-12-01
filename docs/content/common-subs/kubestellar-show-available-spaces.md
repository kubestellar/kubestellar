<!--kubestellar-show-available-spaces-start-->
```shell
KUBECONFIG=~/.kube/config kubectl --context ks-core get secrets kubestellar \
  -o jsonpath='{.data.external\.kubeconfig}' \
  -n kubestellar | base64 -d > ks-core.kubeconfig

KUBECONFIG=~/.kube/config kubectl --context ks-core get spaces -A 
```
<!--kubestellar-show-available-spaces-end-->