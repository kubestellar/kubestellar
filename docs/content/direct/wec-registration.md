# Registering a Workload Execution Cluster

This document explains how to register a Workload Execution Cluster (WEC) with an Inventory and Transport Space (ITS) in KubeStellar.

## Overview

Registering a WEC with an ITS follows the same process as registering a managed cluster with an OCM hub. KubeStellar uses Open Cluster Management (OCM) for cluster registration and management.

The instructions below provide a comprehensive registration process focused on KubeStellar-specific considerations. For additional OCM registration details, refer to the [official Open Cluster Management documentation](https://open-cluster-management.io/docs/getting-started/installation/register-a-cluster/).

### Terminology Mapping

| OCM Term             | KubeStellar Equivalent                        |
|----------------------|-----------------------------------------------|
| **OCM Hub**          | **ITS** (Inventory and Transport Space) |
| **OCM Managed Cluster** | **WEC** (Workload Execution Cluster) |
| **Klusterlet**       | **Klusterlet** |

## Prerequisites

Before registering a WEC, ensure you have:

1. **An existing ITS** with OCM cluster manager running
2. **A running Kubernetes cluster** that will serve as the WEC
3. **Network connectivity** from the WEC to the ITS (HTTPS connections must be possible)
4. **`kubectl` access** to both the WEC and ITS
5. **`clusteradm` CLI tool** installed on the machine where you will run the registration commands ([installation guide](https://open-cluster-management.io/docs/getting-started/installation/start-the-control-plane/))
6. **Sufficient permissions** to create resources in both clusters (typically cluster-admin or equivalent)

For a complete understanding of how WEC registration fits into the overall KubeStellar architecture, see [The Full Story](user-guide-intro.md#the-full-story).

To verify your ITS is ready for WEC registration:

```shell
# Check if the ITS is accessible (this returns a short result)
kubectl --context <its-context> get ServiceAccount default

# Verify OCM cluster manager is running (look for absence of Pending or failing Pods)
kubectl --context <its-context> get pods -n open-cluster-management-hub

# Check for the customization-properties namespace
kubectl --context <its-context> get ns customization-properties
```

## Registration Process

### Step 1: Get the Registration Token

Obtain a registration token from the ITS:

```shell
# Get the registration token from the ITS
clusteradm --context <its-context> get token
```

This command outputs a `clusteradm join` command that you will use in the next step. The token and apiserver URL from this command will be used in the join command.

### Step 2: Join the WEC to the ITS

**Important:** Before executing the join command, determine if you need additional flags based on your environment. The basic command below may not work for all cluster types.

Execute the join command on the WEC to initiate the registration process. The exact command depends on your environment:

```shell
# Basic join command (additional flags may be required based on your environment)
clusteradm join --hub-token <token> --hub-apiserver <api-server-url> --cluster-name <your-wec-name> --context <wec-context>
```

**Important flags:**
- `--cluster-name`: Choose a unique name for your WEC
- `--context`: Specify the kubectl context for your WEC
- `--singleton`: Use this flag if the WEC is a single-node cluster
- `--force-internal-endpoint-lookup`: Required for Kind clusters and other clusters with internal-only API server endpoints

**Example for Kind clusters:**
```shell
clusteradm join --hub-token <token> --hub-apiserver <api-server-url> --cluster-name cluster1 --context cluster1 --singleton --force-internal-endpoint-lookup
```

**Example for cloud provider clusters:**
```shell
clusteradm join --hub-token <token> --hub-apiserver <api-server-url> --cluster-name prod-cluster-east --context prod-cluster-east
```

### Step 3: Wait for Certificate Signing Request

After the join command completes, wait for a Certificate Signing Request (CSR) to appear on the ITS:

```shell
# Check for pending certificate signing requests
kubectl --context <its-context> get csr
```

The first few attempts might not show the CSR with your WEC name and status `Pending`. Continue checking until the CSR appears. This CSR is created by the OCM registration agent running on your WEC.

#### Loop that waits for Certificate Signing Request

The following bash loop automates the waiting described just above:

```shell
# Wait for CSR to be created (this may take a few moments)
while [ -z "$(kubectl --context <its-context> get csr | grep <your-wec-name>)" ]; do
    echo "Waiting for CSR to appear..."
    sleep 5
done
```

### Step 4: Accept the Registration

Approve the registration request from the ITS side:

```shell
# Accept the WEC registration
clusteradm --context <its-context> accept --clusters <your-wec-name>
```

You can accept multiple WECs at once:
```shell
clusteradm --context <its-context> accept --clusters cluster1,cluster2,cluster3
```

### Step 5: Verify Registration

Confirm that the WEC has been successfully registered:

```shell
# List all managed clusters
kubectl --context <its-context> get managedclusters

# Check the status of your specific WEC
kubectl --context <its-context> get managedcluster <your-wec-name> -o yaml

# Verify the klusterlet is running on the WEC (look for absence of Pending or failing Pods)
kubectl --context <wec-context> get pods -n open-cluster-management-agent
```

A successfully registered WEC will show:
- Status: `Available` and `Joined`
- Klusterlet pods running in the `open-cluster-management-agent` namespace
- A corresponding mailbox namespace in the ITS

## Post-Registration Configuration

### Labeling WECs

After registration, you should label your WECs to make them selectable by BindingPolicies. The following are examples of useful labels - you should choose labels that make sense for your environment:

```shell
# Basic labeling
kubectl --context <its-context> label managedcluster <wec-name> location-group=edge name=<wec-name>

# Geographic labels
kubectl --context <its-context> label managedcluster <wec-name> region=us-east zone=us-east-1

# Capability labels
kubectl --context <its-context> label managedcluster <wec-name> gpu=true cpu-architecture=amd64

# Environment labels
kubectl --context <its-context> label managedcluster <wec-name> environment=production tier=critical

# Custom organizational labels
kubectl --context <its-context> label managedcluster <wec-name> team=platform business-unit=engineering
```

### Customization Properties

You can define additional properties for a WEC using ConfigMaps in the `customization-properties` namespace:

```shell
# Create customization properties
kubectl --context <its-context> create configmap <wec-name> \
  -n customization-properties \
  --from-literal clusterURL=https://my.clusters/1001-abcd \
  --from-literal datacenter=us-east-1a \
  --from-literal maxPods=100
```

These properties can be used for rule-based transformations when workloads are distributed to the WEC.

## Different Deployment Scenarios

### Local Development Clusters (Kind/K3d)

For instructions on creating and registering local development clusters (Kind or K3d), refer to the  
[Create and Register Two Workload Execution Clusters guide](https://docs.kubestellar.io/unreleased-development/direct/get-started/#create-and-register-two-workload-execution-clusters).

### OpenShift Clusters

For OpenShift clusters, omit the `--force-internal-endpoint-lookup` flag:

```shell
# Register OpenShift cluster
clusteradm --context its1 get token | grep '^clusteradm join' | \
  sed "s/<cluster_name>/openshift-cluster/" | \
  awk '{print $0 " --context openshift-cluster"}' | sh

# Accept registration
clusteradm --context its1 accept --clusters openshift-cluster
```

### Cloud Provider Clusters

For clusters from cloud providers (EKS, GKE, AKS, etc.):

```shell
# Ensure you have the correct context
kubectl config use-context <cloud-cluster-context>

# Register the cluster
clusteradm join --hub-token <token> --hub-apiserver <its-api-server> \
  --cluster-name <cloud-cluster-name> \
  --context <cloud-cluster-context>

# Accept from ITS
clusteradm --context <its-context> accept --clusters <cloud-cluster-name>
```

## Managing Registered WECs

### Viewing WEC Status

```shell
# List all WECs
kubectl --context <its-context> get managedclusters

# Get detailed status
kubectl --context <its-context> get managedcluster <wec-name> -o yaml

# Check WEC labels
kubectl --context <its-context> get managedcluster <wec-name> --show-labels
```

### Updating WEC Labels

```shell
# Add new labels
kubectl --context <its-context> label managedcluster <wec-name> new-label=value

# Remove labels
kubectl --context <its-context> label managedcluster <wec-name> old-label-

# Update existing labels
kubectl --context <its-context> label managedcluster <wec-name> existing-label=new-value --overwrite
```

### Deregistering a WEC

To remove a WEC from KubeStellar:

```shell
# Delete the ManagedCluster resource
kubectl --context <its-context> delete managedcluster <wec-name>

# Clean up the WEC (optional)
kubectl --context <wec-context> delete namespace open-cluster-management-agent
kubectl --context <wec-context> delete namespace open-cluster-management-agent-addon
```

## Troubleshooting

For troubleshooting WEC registration issues, see the [general troubleshooting guide](troubleshooting.md). Common issues include:

- Certificate Signing Requests not appearing
- Network connectivity problems
- Klusterlet agent failures
- Registration acceptance issues

## Next Steps

After successfully registering your WEC, you can:

1. **Create BindingPolicies** to distribute workloads to your WEC
2. **Configure customization properties** for workload transformation
3. **Set up monitoring** for your WEC
4. **Register additional WECs** to scale your deployment

For more information on using WECs with KubeStellar, see the [example scenarios](example-scenarios.md) and [BindingPolicy documentation](binding.md/#bindingpolicy).

For the complete picture of how WEC registration fits into the overall KubeStellar architecture, see [The Full Story](user-guide-intro.md#the-full-story).
