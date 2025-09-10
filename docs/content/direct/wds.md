# Workload Description Spaces

- [What is a WDS?](#what-is-a-wds)
- [Creating a WDS](#creating-a-wds)
  - [Using the KubeFlex CLI](#using-the-kubeflex-cli)
- [KubeFlex Hosting Cluster as WDS](#kubeflex-hosting-cluster-as-wds)
- [WDS vs. ControlPlane Registration](#wds-vs-controlplane-registration)
- [Controllers Running in a WDS](#controllers-running-in-a-wds)
- [Working with a WDS](#working-with-a-wds)
  - [Accessing the WDS](#accessing-the-wds)

A Workload Description Space (WDS) is a core component of the KubeStellar architecture that serves as the primary interface for users to define and manage workloads for multi-cluster deployment.

## What is a WDS?

A WDS is a space (a Kubernetes-like API server with storage) that:

- Stores the definitions of workloads in their native Kubernetes format
- Hosts the control objects (`BindingPolicy` and `Binding`) that define how workloads are distributed
- Maintains status information about deployed workloads
- Acts as the main user interface to the KubeStellar system

## Creating a WDS

A WDS can be created in several ways:

### Using the KubeStellar Core Helm Chart

The recommended approach is to use the KubeStellar Core Chart:

```shell
helm upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart \
  --set-json='WDSes=[{"name":"wds1", "type":"k8s"}]'
```

You can customize your WDS by specifying:

- `name`: A unique name for the WDS
- `type`:
  - `k8s` (default): Creates a basic Kubernetes API Server with a subset of kube controllers
  - `host`: Uses the KubeFlex hosting cluster itself
- `APIGroups`: A comma-separated list of API Groups to include
- `ITSName`: The name of the ITS to be used by this WDS (required if multiple ITSes exist)

### Using the KubeFlex CLI

You can also create a WDS using the KubeFlex CLI:

```shell
kflex create wds1 -p kubestellar
```

This command creates a WDS and runs a post-create hook that deploys the KubeStellar controller manager and transport controller.

## KubeFlex Hosting Cluster as WDS

The KubeFlex hosting cluster can be configured to act as a WDS by specifying `type: host` when creating the WDS:

```shell
helm upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart \
  --set-json='WDSes=[{"name":"wds1", "type":"host"}]'
```

This approach:

- Avoids creating a separate control plane
- Simplifies the architecture by reusing the hosting cluster
- Makes the WDS directly accessible through the hosting cluster's API server

## WDS vs. ControlPlane Registration

It's important to distinguish between:

1. **Creating a space that can serve as a WDS**: This involves setting up a Kubernetes-like API server.
2. **Registering it with KubeFlex as a ControlPlane and deploying KubeStellar components**: This is the step that makes the space function as a WDS in the KubeStellar ecosystem.

When using the Core Helm Chart or KubeFlex CLI with appropriate parameters, both steps happen automatically.

## Controllers Running in a WDS

When a space is configured as a WDS, the following controllers are deployed:

1. **KubeStellar Controller Manager**: Watches `BindingPolicy` objects and creates corresponding `Binding` objects that contain references to concrete workload objects and destination clusters.

2. **Transport Controller**: Projects KubeStellar workload and control objects from the WDS into the Inventory and Transport Space (ITS).

These controllers are managed as Deployment objects in the KubeFlex hosting cluster.

## Working with a WDS

Once your WDS is created, you can:

1. **Create workload objects** in their native Kubernetes format
2. **Define BindingPolicy objects** to specify which workloads should be deployed to which WECs
3. **Monitor the status** of your deployed workloads

### Accessing the WDS

You can access your WDS using the kubeconfig context provided by KubeFlex:

```shell
# Set up the WDS context
kflex ctx --overwrite-existing-context wds1

# Switch to the WDS context
kubectl config use-context wds1
```
