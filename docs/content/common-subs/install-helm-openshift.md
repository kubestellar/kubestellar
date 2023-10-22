<!--install-helm-openshift-start-->
```shell
KUBECONFIG=~/.kube/config kubectl config use-context ks-core  
kubectl create namespace kubestellar  

helm repo add kubestellar https://helm.kubestellar.io
helm repo update
helm install kubestellar/kubestellar-core \
  --set clusterType=OpenShift \
  --namespace kubestellar \
  --generate-name
```
<!--install-helm-openshift-end-->