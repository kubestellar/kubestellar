<!--create-files-and-contexts-for-kind-clusters-start-->
**important:** rename the kubernetes contexts of the Kind clusters to match their use in this guide
```shell
KUBECONFIG=~/.kube/config kubectl config rename-context kind-ks-core ks-core
KUBECONFIG=~/.kube/config kubectl config rename-context kind-ks-edge-cluster1 ks-edge-cluster1
KUBECONFIG=~/.kube/config kubectl config rename-context kind-ks-edge-cluster2 ks-edge-cluster2
```
<!--create-files-and-contexts-for-kind-clusters-end-->