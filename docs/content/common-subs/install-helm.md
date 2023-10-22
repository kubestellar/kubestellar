<!--install-helm-start-->
```shell hl_lines="6"
kubectl create namespace kubestellar  

helm repo add kubestellar https://helm.kubestellar.io
helm repo update
helm install kubestellar/kubestellar-core \
  --set EXTERNAL_HOSTNAME="kubestellar.core" \
  --set EXTERNAL_PORT={{ config.ks_kind_port_num }} \
  --namespace kubestellar --generate-name
```
<!--install-helm-end-->