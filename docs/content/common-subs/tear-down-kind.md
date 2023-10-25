<!--tear-down-kind-start-->
to remove what you just installed:

{%
    include-markdown "../common-subs/brew-remove.md"   
    start="<!--brew-remove-start-->"
    end="<!--brew-remove-end-->"
%}
```shell
kind delete cluster --name ks-core
kind delete cluster --name ks-edge-cluster1
kind delete cluster --name ks-edge-cluster2
```
{%
    include-markdown "../common-subs/delete-contexts-for-kind-and-openshift-clusters.md"
    start="<!--delete-contexts-for-kind-and-openshift-clusters-start-->"
    end="<!--delete-contexts-for-kind-and-openshift-clusters-end-->"
%}



<!--tear-down-kind-end-->
