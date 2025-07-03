# Inventory and Transport Spaces (ITS)

An Inventory and Transport Space (ITS) is a special type of space in KubeStellar that manages the inventory of clusters (Workload Execution Clusters, or WECs) and handles the transport of workloads to those clusters. The ITS is typically implemented as a KubeFlex ControlPlane of type `vcluster` with Open Cluster Management (OCM) components installed.

## Role of ITS
- Maintains the inventory of WECs using OCM's ManagedCluster objects
- Acts as the hub for distributing workloads to registered WECs
- Hosts mailbox namespaces for each WEC, where workload bundles (ManifestWork objects) are delivered
- Provides the OCM hub functionality for cluster registration and management

## Creating an ITS

Currently, the only supported way to create an ITS is using the KubeStellar core Helm chart. This is because the core Helm chart automatically sets up all the required components and configurations.

### Using the KubeStellar Core Helm Chart

The recommended approach is to use the KubeStellar Core Chart:

```shell
helm upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart \
  --set-json='ITSes=[{"name":"its1", "type":"vcluster"}]'
```

You can customize your ITS by specifying:

- `name`: A unique name for the ITS
- `type`: 
  - `vcluster` (default): Creates a virtual cluster with OCM components
  - `host`: Uses the KubeFlex hosting cluster itself as the ITS
  - `external`: Uses an external cluster as the ITS (requires bootstrap secret)

### What the Core Helm Chart Does

When you create an ITS using the core Helm chart, it:

1. **Creates a KubeFlex ControlPlane** of the specified type
2. **Installs OCM components** via the `its-with-clusteradm` post-create hook:
   - OCM hub (cluster manager)
   - KubeStellar's OCM status add-on controller
   - Required RBAC and service accounts
3. **Sets up the necessary infrastructure** for WEC registration and workload distribution

### KubeFlex Hosting Cluster as ITS

The KubeFlex hosting cluster itself can also serve as an ITS by specifying `type: host`:

```shell
helm upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart \
  --set-json='ITSes=[{"name":"its1", "type":"host"}]'
```

This approach:
- Avoids creating a separate virtual cluster
- Simplifies the architecture by reusing the hosting cluster
- Makes the ITS directly accessible through the hosting cluster's API server

## Notes
- Creating an ITS includes installing OCM hub components and KubeStellar's OCM status add-on
- The ITS is automatically configured to work with the KubeStellar transport system
- For more details, see the [architecture](architecture.md) and [core chart usage](core-chart.md) docs
