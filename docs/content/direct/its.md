# Inventory and Transport Spaces

An Inventory and Transport Space (ITS) is a core component of the KubeStellar architecture that serves two primary functions:

1. **Inventory Management**: It maintains a registry of all Workload Execution Clusters (WECs) available in the system.
2. **Transport Facilitation**: It handles the movement of workloads from Workload Description Spaces (WDSes) to the appropriate WECs.

## What is an ITS?

An ITS is a space (a Kubernetes-like API server with storage) that:

- Holds inventory information about all registered WECs using [ManagedCluster.v1.cluster.open-cluster-management.io](https://github.com/open-cluster-management-io/api/blob/v0.12.0/cluster/v1/types.go#L33) objects
- Contains a "customization-properties" namespace with ConfigMaps carrying additional properties for each WEC
- Manages mailbox namespaces that correspond 1:1 with each WEC, holding ManifestWork objects
- Runs the OCM (Open Cluster Management) Cluster Manager to synchronize objects with the WECs

## Creating an ITS

An ITS can be created in several ways:

### Using the KubeStellar Core Helm Chart

The recommended approach is to use the KubeStellar Core Chart:

```shell
helm upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart \
  --set-json='ITSes=[{"name":"its1", "type":"vcluster"}]'
```

You can customize your ITS by specifying:
- `name`: A unique name for the ITS
- `type`: 
  - `vcluster` (default): Creates a virtual cluster
  - `host`: Uses the KubeFlex hosting cluster itself
  - `external`: Uses an external cluster
- `install_clusteradm`: `true` (default) or `false` to control OCM installation

### Using the KubeFlex CLI

You can also create an ITS using the KubeFlex CLI:

```shell
kflex create its1 --type vcluster -p ocm
```

## KubeFlex Hosting Cluster as ITS

The KubeFlex hosting cluster can be configured to act as an ITS by specifying `type: host` when creating the ITS:

```shell
helm upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart \
  --set-json='ITSes=[{"name":"its1", "type":"host"}]'
```

This approach:
- Avoids creating a separate virtual cluster
- Simplifies the architecture by reusing the hosting cluster
- Makes the ITS directly accessible through the hosting cluster's API server

## Important Note on ITS Registration

Creating an ITS includes installing the relevant OCM (Open Cluster Management) machinery in it. However, registering the ITS as a KubeFlex control plane is a separate step that happens automatically when using the Core Helm Chart or KubeFlex CLI with the appropriate parameters.

## Architecture and Components

The ITS runs the OCM Cluster Manager, which:
- Accepts registrations from WECs through the OCM registration agent
- Manages the distribution of workloads to WECs
- Maintains status information from the WECs
- Creates and manages mailbox namespaces for each registered WEC

## Transport Implementation

The current implementation uses [Open Cluster Management](https://open-cluster-management.io), but the architecture is designed to be pluggable, allowing for different transport mechanisms in the future.

The OCM-based Transport Controller in the ITS:
- Maintains a set of `ManifestWork` objects that represent the KubeStellar workload
- Bundles workload objects from the WDS into `ManifestWork` objects
- Delivers these objects to the appropriate WECs
- Collects status information from the WECs

## Next Steps

After creating an ITS, you'll need to:
1. [Create a Workload Description Space (WDS)](wds.md)
2. [Create and register Workload Execution Clusters (WECs)](wec.md)
3. [Create BindingPolicies](binding.md) to associate workloads with WECs
