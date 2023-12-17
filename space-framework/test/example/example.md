# Space Framework quick start 

The Space Framework (SF) allows you to use different space providers to obtain spaces. 

You can find a more detailed document describe the Space Framework and the Space Manager in:   
[Space Framework high level architecture](https://github.com/kubestellar/kubestellar/blob/main/space-framework/docs/space-framework.md)

The following guide describes how to setup and use the SF.  
As part of this guide we also describe how to install and setup the space providers. Please note that, this is not part of the SF configuration and and is described here only as a helpful reference.
The following sections describe how to setup and use the Space Manager (SM) which is the main management component of the Space Framework.

The setup includes the following sections:
   * Get (build if needed) space providers
   * Get and run the SM
   * Use the SM
        * Create space-provider-description 
        * Create spaces
        * Access spaces and perform operations on spaces


**Note**: All helper scripts and predefined object's yaml files that are referenced or used in this guide are under `space-framework/test/example`

## Setup the environment (clusters & space providers)
### Create a kind cluster fot the SM
We need to create an k8s api-server to host the SM CRDs. We will be using a simple KIND cluster. 
The following script can be used to create the needed cluster:  
[create_kind_cluster.sh](create_kind_cluster.sh)<br/>

Please note the special configuration (e.g., ingress) in the above script is not needed for the SM. The SM itself needs only a **simple k8s api-service** to hosts its CRDs (the SM doesn't need to be run as a k8s workload). The configuration is needed for KubeFlex that will be used as a space provider in this guide.   

Note: We recommend to set the KUBECONFIG to the location of the kind config assuming the kind cluster is using `$HOME/.kube/config`
```sell
export KUBECONFIG=$HOME/.kube/config
```

### Build the SM
Building the SM is simple:  
```shell
git clone git@github.com:kubestellar/kubestellar.git
cd kubestellar/space-framework
make 
```
You can also use the following script:  
[clone_and_build_space_manager.sh](clone_and_build_space_manager.sh)<br/>

### Run the SM
The `space-manager` executable should be under `space-framework/bin` 

```shell
# Assuming you are under the space-framework directory
kubectl apply -f config/crds

bin/space-manager --context kind-sm-mgt &> /tmp/space-manager.log &
```
The above can be done using:    
[run_space_manager.sh](run_space_manager.sh)<br/>

## Install provider backends
### KubeFlex
Follow the installation section on the Kubeflex instructions:  
[https://github.com/kubestellar/kubeflex/tree/main](https://github.com/kubestellar/kubeflex/tree/main)<br/>

Alternatively you can use:
[install_and_run_kubeflex.sh](install_and_run_kubeflex.sh)<br/>

### KCP
Clone and build KCP:
```shell
git clone https://github.com/kcp-dev/kcp
cd kcp
make
bin/kcp start &> /tmp/kcp.log &
```
Alternatively yu can use:  
[install_and_run_kcp.sh](install_and_run_kcp.sh)<br/>

## Using the SM
Before we can use a space provider through the SM we need to register it with the SM. This process has two steps:
1. Create a secret holding the access config to the space provider
2. Create a SpaceProviderDesc object (used by the SM)

### creating secrets 
```shell
# Create a secret for KCP
kubectl create secret generic kcpsec --from-file=kubeconfig=kcp/.kcp/admin.kubeconfig --from-file=incluster=kcp/.kcp/admin.kubeconfig

#Create a secret for KubeFlex
kubectl create secret generic kfsec --from-file=kubeconfig=$HOME/.kube/config 
```
### Create SpaceProviderDesc objects 
The space provider object is a very simple object, for this guide we will use pre-defined objects
```shell
kubectl apply -f kcp-provider.yaml
kubectl apply -f kf-provider.yaml
```

## Using Spaces
**Create spaces**  
We can now manage spaces on the providers we registered. 
Create 2 spaces: one in kubeflex, and one in kcp.  
```shell
kubectl apply -f kcp-space.yaml
kubectl apply -f kf-space.yaml
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

**Accessing a space**
Working with spaces is easy, you first get a config file for that space and then access it in a regular kubectl commands. You can use the `get_kubeconfig_for_space` function to get the config file.

```shell
#Get the config for the kcpspace we create
get_kubeconfig_for_space external-kcpspace kcp > kcpspace.kubeconfig

#Accessing the space
kubectl --kubeconfig kcpspace.kubeconfig get ns

#Get the config for the kubeflex we create
get_kubeconfig_for_space external-kfspace kf > kfspace.kubeconfig

#Accessing the space
kubectl --kubeconfig kfspace.kubeconfig get configmaps -A
```

**Delete a space**
To delete a space you simply need to delete the space object, the SM will delete the actual space on the space provider. 
```shell
#delete the kfspace
kubectl delete space kfspace -n spaceprovider-kf
```

You can verify that the space was deleted from the space provider as well

## Teardown
```shell
kubectl delete space kcpspace -n spaceprovider-kcp
pkill -f space-manager
pkill -f kcp
kind delete cluster --name sm-mgt
```
