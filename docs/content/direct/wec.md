# Workload Execution Clusters (WEC)

A Workload Execution Cluster (WEC) is a Kubernetes cluster that receives and runs workloads distributed by KubeStellar. WECs are registered with an Inventory and Transport Space (ITS), which manages their inventory and delivers workload bundles to them.

## What Makes a Cluster Suitable as a WEC?

A cluster can serve as a WEC if it meets these requirements:

- **Network Connectivity**: The cluster must be reachable from the ITS (network connectivity and credentials required)
- **OCM Registration**: The cluster must be registered with the ITS using Open Cluster Management (OCM) tooling
- **Agent Installation**: The OCM agent (klusterlet) and KubeStellar's status add-on agent must be installed on the WEC
- **Kubernetes Compatibility**: The cluster must be running a compatible version of Kubernetes

## WEC Registration Process

WEC registration follows the standard Open Cluster Management process. The registration involves:

1. **Obtaining a registration token** from the ITS
2. **Installing the OCM agent** on the WEC using `clusteradm join`
3. **Approving the registration** from the ITS side
4. **Verifying the connection** and agent status

For detailed step-by-step instructions, see [Registering a WEC](wec-registration.md).

## WEC Management

### Labeling WECs

After registration, WECs should be labeled to make them selectable by BindingPolicies:

```shell
# Basic labeling
kubectl --context <its-context> label managedcluster <wec-name> location-group=edge name=<wec-name>

# Geographic labels
kubectl --context <its-context> label managedcluster <wec-name> region=us-east zone=us-east-1

# Capability labels
kubectl --context <its-context> label managedcluster <wec-name> gpu=true cpu-architecture=amd64

# Environment labels
kubectl --context <its-context> label managedcluster <wec-name> environment=production tier=critical
```

### Customization Properties

You can define additional properties for each WEC using ConfigMaps in the `customization-properties` namespace:

```shell
# Create customization properties
kubectl --context <its-context> create configmap <wec-name> \
  -n customization-properties \
  --from-literal clusterURL=https://my.clusters/1001-abcd \
  --from-literal datacenter=us-east-1a \
  --from-literal maxPods=100
```

These properties can be used for rule-based transformations when workloads are distributed to the WEC.

## Supported Cluster Types

WECs can be created using any Kubernetes distribution:

- **Local Development**: kind, k3d, minikube
- **Cloud Providers**: EKS, GKE, AKS
- **On-Premises**: OpenShift, Rancher, vanilla Kubernetes
- **Edge/IoT**: k3s, microk8s

## WEC Lifecycle

### Registration
- WEC registers with ITS using OCM
- ITS creates a mailbox namespace for the WEC
- OCM agents establish secure communication

### Workload Distribution
- WDS creates workload objects and BindingPolicies
- KubeStellar controller manager creates Binding objects
- Transport controller bundles workloads and sends to ITS
- ITS delivers workload bundles to WEC mailbox

### Monitoring
- WEC agents report status back to ITS
- Status is aggregated and available in WDS
- CombinedStatus objects provide unified view

### Deregistration
- Remove WEC from KubeStellar management
- Clean up OCM agents and resources
- Delete mailbox namespace in ITS

## Notes
- WECs are managed through OCM's ManagedCluster resources in the ITS
- Each WEC has its own mailbox namespace in the ITS for workload delivery
- The OCM agents handle secure communication between WEC and ITS
- For more details, see the [architecture](architecture.md) and [example scenarios](example-scenarios.md)
