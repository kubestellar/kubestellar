# Workload Description Spaces

- [What is a WDS?](#what-is-a-wds)
- [Creating a WDS](#creating-a-wds)

  - [Using the KubeFlex CLI](#using-the-kubeflex-cli)
  - [Accessing the WDS](#accessing-the-wds)

- [Working with a WDS](#working-with-a-wds)
- [WDS vs. ControlPlane Registration](#wds-vs-controlplane-registration)
- [Controllers Running in a WDS](#controllers-running-in-a-wds)



## What is a WDS?

A Workload Description Space (WDS) is a space in the [KubeStellar architecture](user-guide-intro.md) that serves as the primary interface for users to define and manage workloads for multi-cluster deployment. The WDS constitue of a Kubernetes API server with storage that:

- Stores the definitions of workloads in their native Kubernetes format
- Hosts the control objects (`BindingPolicy`, `Binding`, `Status Collector`, `CombinedStatus` and `CustomTransform`) that define how workloads are distributed
- Maintains status information about deployed workloads
- Acts as the main user interface to the KubeStellar system


## Creating a WDS

Currently, we support the use of our Kubestellar core Helm chart as the only way to  create a WDS. This is because the core Helm chart also automatically creates a kubestellar-controller-manager and transport-controller, which are contorllers that your WDS requires to function properly.

### Use the KubeStellar Core Helm Chart to create your WDS

The recommended approach is to use the KubeStellar Core Chart:

```shell
helm upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart \
  --set-json='WDSes=[{"name":"<wds-name>", "type":"<space-type>"}]'
```

You can customize your WDS by specifying:  

- `<wds-name>`: A unique name for the WDS
- `<space-type>`:  

    - `k8s` (default): Creates a basic Kubernetes API Server with a subset of kube controllers
    - `host`: Uses the KubeFlex hosting cluster itself. 
    
- `APIGroups`: A comma-separated list of API Groups to include
- `ITSName`: The name of the ITS to be used by this WDS (required if multiple ITSes exist)

The type of space you choose determines the type of controllers that are included in your WDS. When creating a WDS, the recommended space types are "K8s" (the default space) and "host". If you the decide to use your hosting cluster's control plane (which is typically a KubeFlex cluster for Kubstellar) by specifying `type:host`, it presents you some benefits such as:

- Saves the cost of creating an additional space.
- Simplifies the architecture by reusing the hosting cluster
- Makes the WDS directly accessible through the hosting cluster's API server

> You can create multiple WDSes of the same space types and even a mix of different space types, where you can have the default and host WDSes in your cluster. A multiple WDS architecure can be valueable in cases where you want to manage workload runs across different sets of users or groups. Thus each user or group can use the same ITS for WEC inventory but with their different WDS

### Accessing the WDS

After creating your WDS, you will need to access it to submit workload objects. To do this you have to create a context for your WDS in kubeconfig context and set it context as your current cluster context:

- **Setup WDS context with KubeFlex CLI**: KubeFlex CLI enables you to create the context for your WDS in the kubeconfig context. Then automatically switch to it by overwiting the current kubeconfig context as your WDS context.

```shell
# create a the WDS context and make it the current context
kflex ctx --overwrite-existing-context <wds-name>
```

- **Access WDS context with Kubernetes API**: In a case where the context for your WDS already exists in the kubeconfig context, then, you can just switch to it as your current context with the Kubernetes API. 

```shell
# Switch to the WDS context
kubectl config use-context <wds-name>
```

Both methods provide flexibility depending on your preferred way of working with the system.

## Working with a WDS

Once your WDS is created, you can:

1. **Create workload objects** in their native Kubernetes format
2. **Define BindingPolicy objects** to specify which workloads should be deployed to which WECs
3. **Monitor StatusCollector and CombinedStatus objects** for a status update on your deployed workloads

## WDS vs. ControlPlane Registration

It's important to distinguish between:

1. **Creating a space that can serve as a WDS**: This involves setting up a Kubernetes-like API server.
2. **Registering it with KubeFlex as a ControlPlane and deploying KubeStellar components**: This is the step that makes the space function as a WDS in the KubeStellar ecosystem.

When using the Core Helm Chart or KubeFlex CLI with appropriate parameters, both steps happen automatically.

## Controllers that work with WDS

For the WDS to execute its tasks after it is configured, it will need to interact with the following controllers that live in the Hosting cluster:

1. **KubeStellar Controller Manager**: Watches the WDS `BindingPolicy` objects and creates corresponding `Binding` objects that contain references to concrete workload objects and destination clusters.

2. **Pluggable Transport Controller**: This projects KubeStellar's workload and control objects from the WDS into the Inventory and Transport Space (ITS).

These controllers are managed as Deployment objects in the KubeFlex hosting cluster.


