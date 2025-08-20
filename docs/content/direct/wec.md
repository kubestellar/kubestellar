# Workload Execution Clusters
- [What is a WEC?](#what-is-a-wec)
- [Creating a WEC](#creating-a-wec)
  - [Using Kind (for development/testing)](#using-kind-for-developmenttesting)
  - [Using K3d (for development/testing)](#using-k3d-for-developmenttesting)
  - [Using MicroShift (for edge deployments)](#using-microshift-for-edge-deployments)
  - [Using Production Kubernetes Distributions](#using-production-kubernetes-distributions)
- [Registering a WEC](#registering-a-wec)
- [Labeling WECs](#labeling-wecs)
- [WEC Customization Properties](#wec-customization-properties)
- [WEC Status and Monitoring](#wec-status-and-monitoring)
- [Workload Transformation](#workload-transformation)
Workload Execution Clusters (WECs) are the Kubernetes clusters where KubeStellar deploys and runs the workloads defined in the Workload Description Spaces (WDSes).

## What is a WEC?

A WEC is a standard Kubernetes cluster that:

- Runs the actual workloads distributed by KubeStellar
- Has the OCM Agent (klusterlet) installed for communication with the ITS
- Is registered with an Inventory and Transport Space (ITS)
- May have specific characteristics (location, resources, capabilities) that make it suitable for particular workloads

## Requirements for a WEC

For a Kubernetes cluster to function as a WEC in the KubeStellar ecosystem, it must:

1. **Have network connectivity** to the Inventory and Transport Space (ITS)
2. **Be a valid Kubernetes cluster** with a working control plane
3. **Have the OCM Agent installed** for integration with KubeStellar
4. **Be registered with an ITS** to receive workloads

## Creating a WEC

You can use any existing Kubernetes cluster as a WEC, or create a new one using your preferred method:

### Using Kind (for development/testing)

```shell
kind create cluster --name cluster1
kubectl config rename-context kind-cluster1 cluster1
```

### Using K3d (for development/testing)

```shell
k3d cluster create -p "9443:443@loadbalancer" cluster1
kubectl config rename-context k3d-cluster1 cluster1
```

### Using MicroShift (for edge deployments)

For resource-constrained environments like edge devices, you can use MicroShift:

```shell
# Instructions for setting up MicroShift can be found at
# https://community.ibm.com/community/user/cloud/blogs/alexei-karve/2021/11/28/microshift-4
```

### Using Production Kubernetes Distributions

For production environments, consider using:

- Red Hat OpenShift
- Amazon EKS
- Google GKE
- Microsoft AKS
- Any conformant Kubernetes distribution

## Registering a WEC

After creating your cluster, you need to register it with an ITS. This process installs the OCM Agent and establishes the communication channel.

```shell
# Get the join command from the ITS
clusteradm --context its1 get token

# Execute the join command with your WEC name
clusteradm join --hub-token <token> --hub-apiserver <api-server-url> --cluster-name cluster1 --context cluster1

# Accept the registration on the ITS side
clusteradm --context its1 accept --clusters cluster1
```

For detailed registration instructions, see [WEC Registration](wec-registration.md).

## Labeling WECs

After registration, you should label your WEC to make it selectable by BindingPolicies:

```shell
kubectl --context its1 label managedcluster cluster1 location-group=edge name=cluster1
```

These labels can represent any characteristics relevant to your workload placement decisions:

- Geographic location (`region=us-east`, `location=edge`)
- Hardware capabilities (`gpu=true`, `cpu-architecture=arm64`)
- Environment type (`environment=production`, `environment=development`)
- Compliance requirements (`pci-dss=compliant`, `hipaa=compliant`)
- Custom organizational labels (`team=retail`, `business-unit=finance`)

## WEC Customization Properties

For each WEC, you can define additional properties in a ConfigMap stored in the "customization-properties" namespace of the ITS. These properties can be used for rule-based transformations of workloads.

## WEC Status and Monitoring

You can check the status of your registered WECs:

```shell
kubectl --context its1 get managedclusters
```

## Workload Transformation

KubeStellar performs transformations on workloads before they are deployed to WECs:

1. **Generic transformations** that apply to all workloads
2. **Rule-based customizations** that adapt workloads to specific WEC characteristics

For more information, see [Transforming Desired State](transforming.md).
