# Registering a Workload Execution Cluster

This document explains how to register a Workload Execution Cluster (WEC) with an Inventory and Transport Space (ITS) in KubeStellar.

## Overview

Registering a WEC in an ITS is the same process as registering a managed cluster with an OCM hub cluster. KubeStellar uses Open Cluster Management (OCM) for cluster registration and management.

**Terminology Mapping:**
- **OCM Hub** → **KubeStellar ITS** (Inventory and Transport Space)
- **OCM Managed Cluster** → **KubeStellar WEC** (Workload Execution Cluster)
- **OCM Agent** → **OCM Agent** (same component, installed on WEC)

For the complete OCM registration process, refer to the [official Open Cluster Management documentation](https://open-cluster-management.io/docs/getting-started/installation/register-a-cluster/).

## Prerequisites

Before registering a WEC, ensure you have:

1. **A running Kubernetes cluster** that will serve as the WEC
2. **Network connectivity** between the WEC and the ITS
3. **`kubectl` access** to both the WEC and ITS
4. **`clusteradm` CLI tool** installed ([installation guide](https://open-cluster-management.io/docs/getting-started/installation/start-the-control-plane/))
5. **An existing ITS** with OCM cluster manager running
6. **Sufficient permissions** to create resources in both clusters

To verify your ITS is ready for WEC registration:

```shell
# Check if the ITS is accessible (this returns a short result)
kubectl --context <its-context> get ServiceAccount default

# Verify OCM cluster manager is running (look for pods in Running state)
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

This command outputs a `clusteradm join` command that you'll use in the next step.

### Step 2: Join the WEC to the ITS

Execute the join command on the WEC to initiate the registration process:

```shell
# Basic join command
clusteradm join --hub-token <token> --hub-apiserver <api-server-url> --cluster-name <your-wec-name> --context <wec-context>
```

**Important flags:**
- `--cluster-name`: Choose a unique name for your WEC
- `--context`: Specify the kubectl context for your WEC
- `--singleton`: Use this flag if the WEC is a single-node cluster
- `--force-internal-endpoint-lookup`: Required for Kind clusters and other local development setups

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

You should see a CSR with your WEC name and status `Pending`. This CSR is created by the OCM registration agent running on your WEC.

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

# Verify the OCM agent is running on the WEC (look for pods in Running state)
kubectl --context <wec-context> get pods -n open-cluster-management-agent
```

A successfully registered WEC will show:
- Status: `Available` and `Joined`
- OCM agent pods running in the `open-cluster-management-agent` namespace
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

## Different Deployment Scenarios

### Local Development Clusters (Kind/K3d)

For local development with Kind or K3d clusters:

```shell
# Create Kind cluster (context renaming is optional but convenient)
kind create cluster --name cluster1
kubectl config rename-context kind-cluster1 cluster1

# Register with ITS (note the --force-internal-endpoint-lookup flag)
clusteradm --context its1 get token | grep '^clusteradm join' | \
  sed "s/<cluster_name>/cluster1/" | \
  awk '{print $0 " --context cluster1 --singleton --force-internal-endpoint-lookup"}' | sh

# Wait for CSR to appear, then accept registration
clusteradm --context its1 accept --clusters cluster1
```

For K3d clusters, the process is identical except for the cluster creation:
```shell
# Create K3d cluster with port mapping
k3d cluster create -p "9443:443@loadbalancer" cluster1
kubectl config rename-context k3d-cluster1 cluster1

# Then follow the same registration steps as above
```

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

## Troubleshooting

### Common Issues and Solutions

**1. CSR not appearing**
```shell
# Check if the join command completed successfully (look for Running pods)
kubectl --context <wec-context> get pods -n open-cluster-management-agent

# Verify network connectivity
kubectl --context <wec-context> get events --all-namespaces | grep -i error
```

**2. CSR stuck in Pending state**
```shell
# Check CSR details
kubectl --context <its-context> get csr <csr-name> -o yaml

# Manually approve if needed
kubectl --context <its-context> certificate approve <csr-name>
```

**3. ManagedCluster shows as Unknown**
```shell
# Check OCM agent logs
kubectl --context <wec-context> logs -n open-cluster-management-agent deployment/klusterlet-registration-agent

# Verify ITS connectivity
kubectl --context <wec-context> get events -n open-cluster-management-agent
```

**4. Network connectivity issues**
```shell
# Test connectivity from WEC to ITS
kubectl --context <wec-context> run test-connectivity --image=curlimages/curl --rm -it -- curl -k <its-api-server>

# Check firewall rules and security groups
```

### Diagnostic Commands

```shell
# Check OCM components status
kubectl --context <its-context> get pods -n open-cluster-management-hub
kubectl --context <wec-context> get pods -n open-cluster-management-agent

# Verify ManagedCluster resource
kubectl --context <its-context> describe managedcluster <wec-name>

# Check agent logs
kubectl --context <wec-context> logs -n open-cluster-management-agent deployment/klusterlet-registration-agent
kubectl --context <wec-context> logs -n open-cluster-management-agent deployment/klusterlet-work-agent
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

## Next Steps

After successfully registering your WEC, you can:

1. **Create BindingPolicies** to distribute workloads to your WEC
2. **Configure customization properties** for workload transformation
3. **Set up monitoring** for your WEC
4. **Register additional WECs** to scale your deployment

For more information on using WECs with KubeStellar, see the [example scenarios](example-scenarios.md) and [BindingPolicy documentation](binding-policies.md).
