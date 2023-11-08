<!--install-helm-test-start-->
```shell hl_lines="7"
KUBECONFIG=~/.kube/config kubectl config use-context ks-core  
KUBECONFIG=~/.kube/config kubectl create namespace kubestellar  

helm repo add kubestellar https://helm.kubestellar.io
helm repo update
KUBECONFIG=~/.kube/config helm install ./core-helm-chart \
  --set EXTERNAL_HOSTNAME="kubestellar.core" \
  --set EXTERNAL_PORT={{ config.ks_kind_port_num }} \
  --set CONTROLLER_VERBOSITY=4 \
  --set image.tag=$EXTRA_CORE_TAG \
  --namespace kubestellar \
  --generate-name
```
<!--install-helm-test-end-->