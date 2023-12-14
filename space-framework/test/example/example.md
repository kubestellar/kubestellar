Install the backend - space manager, kcp, and kubeflex and create a kind cluster which will host the space manager and kubeflex and also act as a space provider:

[clone_and_build_space_manager.sh](clone_and_build_space_manager.sh)<br/>
[create_kind_cluster.sh](create_kind_cluster.sh)<br/>
[run_space_manager.sh](run_space_manager.sh)<br/>
[install_and_run_kcp.sh](install_and_run_kcp.sh)<br/>
[install_and_run_kubeflex.sh](install_and_run_kubeflex.sh)<br/>

Set the KUBECONFIG to point to the space manager kubeconfig file:
```shell
export KUBECONFIG=$PWD/sm.kubeconfig
```

The space manager access providers using kubernetes secrets.  We now create these secrets for kcp, kubeflex and kind. Since kubeflex is actually running on our kind cluster we could use the same secret for both.  
```shell
kubectl create secret generic kcpsec --from-file=kubeconfig=kcp/.kcp/admin.kubeconfig --from-file=incluster=kcp/.kcp/admin.kubeconfig
kubectl create secret generic kfsec --from-file=kubeconfig=$HOME/.kube/config 
kubectl create secret generic kindsec --from-file=kubeconfig=$HOME/.kube/config 
```

Now we add the kcp, kubeflex, and kind as space providers by creating spaceproviderdesc objects for them.
```shell
kubectl apply -f kcp-provider.yaml
kubectl apply -f kf-provider.yaml
kubectl apply -f kind-provider.yaml
```

Create three spaces: one in kind, one in kubeflex, and one in kcp.  
```shell
kubectl apply -f kcp-space.yaml
kubectl apply -f kf-space.yaml
kubectl apply -f kind-space.yaml
```

To access a space, you extract its KUBECONFIG from a secret stored within the space object. You then access a space by referencing the relevant KUBECONFIG file either by setting the KUBECONFIG environment variable or by using the --kubeconfig parameter on the kubectl command.
```shell
get_kubeconfig_for_space kindspace kind > kindspace.kubeconfig
kubectl --kubeconfig kindspace.kubeconfig get pods -A

get_kubeconfig_for_space kcpspace kcp > kcpspace.kubeconfig
kubectl --kubeconfig kcpspace.kubeconfig get pods -A

get_kubeconfig_for_space kfspace kf > kfspace.kubeconfig
kubectl --kubeconfig kfspace.kubeconfig get pods -A
```

