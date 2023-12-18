<!--delete-contexts-for-kind-and-openshift-clusters-start-->
**important:** delete any existing kubernetes contexts of the clusters you may have created previously
```shell
KUBECONFIG=~/.kube/config kubectl config delete-context ks-core || true
KUBECONFIG=~/.kube/config kubectl config delete-context ks-edge-cluster1 || true
KUBECONFIG=~/.kube/config kubectl config delete-context ks-edge-cluster2 || true
```
<!--delete-contexts-for-kind-and-openshift-clusters-end-->
