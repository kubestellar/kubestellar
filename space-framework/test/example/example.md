# Space Framework quick start 

The Space Framework (SF) is a generic management framework for space providers and pSpaces. The framework defines an abstraction layer for space providers and pSpace management that allows clients (both script and kubectl based and client-go based clients) to use spaces while maintaining the clients decoupled from the specific pSpace and space provider that is being used. For a more detailed document describing the space framework and the space manager go to: [Space Framework high level architecture](https://github.com/kubestellar/kubestellar/blob/main/space-framework/docs/space-framework.md)

The following guide describes how to install, setup, and use the SF. Please note that installing the space provider servers is technically not part of the SF and is described here as a helpful reference.

**Note**: All helper scripts and predefined object's yaml files are located in `space-framework/test/example`

## Setup the environment (clusters & space providers)
### Create a kind cluster for the Space Manager
We need to create a k8s api-server to host the Space Manager (SM) CRDs. We will be using a simple Kind cluster which we create using the following script: [create_kind_cluster](create_kind_cluster.sh)<br/>

Please note the special ingress configuration in the above script is not needed for the SM. The SM itself needs only a **simple k8s api-service** to hosts its CRDs. The SM doesn't need to be run as a k8s workload. The configuration is needed for KubeFlex which is one of the space provider used in this guide.   

Note: We recommend you set the KUBECONFIG to the location of the kind config used by the space manager which by default is `$HOME/.kube/config`:
```shell
export KUBECONFIG=$HOME/.kube/config
```

### Build the SM
Building the SM is simple:  
```shell
git clone git@github.com:kubestellar/kubestellar.git
cd kubestellar/space-framework
make 
```

### Run the SM
The `space-manager` executable should be under `space-framework/bin` 

```shell
# Assuming you are already in the space-framework directory
kubectl apply -f config/crds
bin/space-manager --context kind-sm-mgt &> /tmp/space-manager.log &
```

## Install provider backends
### KubeFlex
Follow the installation section on the Kubeflex instructions:  
[https://github.com/kubestellar/kubeflex/tree/main](https://github.com/kubestellar/kubeflex/tree/main)<br/>

Alternatively you can issue the following commands:
```shell
git clone git@github.com:kubestellar/kubeflex.git
cd kubeflex
make
bin/kflex init
```

### KCP
Clone and build KCP:
```shell
git clone https://github.com/kcp-dev/kcp
cd kcp
make
bin/kcp start &> /tmp/kcp.log &
```

## Using the SM
Before we can use a space provider through the SM we need to register it with the SM. This process has two steps:
1. Create a secret holding the access config to the space provider
2. Create a SpaceProviderDesc object (used by the SM)

**Note:** the scripts below assume you are in the kubestellar/space-framework directory.
### Creating secrets 
```shell
# Create a secret for KCP. The --from-file assumes your directory is above the kcp directory.
kubectl create secret generic kcpsec --from-file=kubeconfig=kcp/.kcp/admin.kubeconfig --from-file=incluster=kcp/.kcp/admin.kubeconfig
# Create a secret for KubeFlex
kubectl create secret generic kfsec --from-file=kubeconfig=$HOME/.kube/config 
```
### Create SpaceProviderDesc objects 
The space provider object is a very simple object, for this guide we will use pre-defined objects
```shell
kubectl apply -f test/example/kcp-provider.yaml
kubectl apply -f test/example/kf-provider.yaml
```
## Using Spaces
**Create spaces**  
We can now manage spaces on the providers we registered. 
Create 2 spaces: one in kubeflex, and one in kcp.  
```shell
kubectl apply -f test/example/kcp-space.yaml
kubectl apply -f test/example/kf-space.yaml
```
**List spaces**  
You can list the existing spaces and see their status. Those are regular Kube objects so you can use regular kubectl commands
```shell
# List all spaces
kubectl get spaces -A

# Get detailed info on the kcpspace space
kubectl get spaces kcpspace -n spaceprovider-kcp -o yaml

# Get detailed info on the kfspace space
kubectl get space kfspace -n spaceprovider-kf -o yaml
```

Wait for `status.Phase: Ready` for both spaces
```shell
kubectl wait --for=jsonpath='{.status.Phase}'=Ready space/kcpspace -n spaceprovider-kcp --timeout=90s
kubectl wait --for=jsonpath='{.status.Phase}'=Ready space/kfspace -n spaceprovider-kf --timeout=90s
```

**Accessing a space**
Working with spaces is easy, you first get a config file for that space and then access it in a regular kubectl commands. You can use the `get_kubeconfig_for_space` function to get the config file.
```shell
# Get the config for the kcpspace we create
test/example/get_kubeconfig_for_space external-kcpspace kcp > kcpspace.kubeconfig

# Accessing the space
kubectl --kubeconfig kcpspace.kubeconfig get ns

# Get the config for the kubeflex we create
test/example/get_kubeconfig_for_space external-kfspace kf > kfspace.kubeconfig

# Accessing the space
kubectl --kubeconfig kfspace.kubeconfig get configmaps -A
```

**Delete a space**
To delete a space you simply need to delete the space object, the SM will delete the actual space on the space provider. 
```shell
# delete the kfspace
kubectl delete space kfspace -n spaceprovider-kf
```

You can verify that the physical space was deleted from the space provider as well.
## Teardown
To remove the SM you should kill its process (pkill). If you started a kcp server you would need to kill its process as well. Instructions on cleaning up KubeFlex are located [here](https://github.com/kubestellar/kubeflex/blob/main/docs/users.md#uninstalling-kubeflex).
```shell
kubectl delete space kcpspace -n spaceprovider-kcp
pkill -f space-manager
pkill -f kcp
kind delete cluster --name sm-mgt
```
