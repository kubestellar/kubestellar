# Workload Execution Clusters
- [What is a WEC?](#what-is-a-wec)
- [Creating a WEC](#creating-a-wec)
  - [Using Kind (for development/testing)](#using-kind-for-developmenttesting)
  - [Using K3d (for development/testing)](#using-k3d-for-developmenttesting)
  - [Using MicroShift (for edge deployments)](#using-microshift-for-edge-deployments)
  - [Using Production Kubernetes Distributions](#using-production-kubernetes-distributions)
- [Authorization and Security](#authorization-and-security)
  - [Authorization Model Overview](#authorization-model-overview)
  - [Default Permissions](#default-permissions)
  - [When Additional Authorization is Needed](#when-additional-authorization-is-needed)
  - [Expanding Authorization](#expanding-authorization)
  - [Troubleshooting Authorization Issues](#troubleshooting-authorization-issues)
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

## Authorization and Security

### Authorization Model Overview

KubeStellar implements a **security-first approach** to workload deployment. The system follows the **principle of least privilege** - the OCM Agent (klusterlet) that runs on each WEC operates with carefully limited permissions by default. This design ensures that workloads can only be deployed with explicitly granted permissions, providing better security posture and audit capabilities.

> **⚠️ Important**: KubeStellar does not intend to violate the principle of least privilege any more than necessary. The default authorizations are intentionally minimal and must be expanded when deploying workloads that require additional permissions.

### Default Permissions

The OCM work agent (`klusterlet-work-sa` service account) comes with default permissions for common Kubernetes resources. These default authorizations are inherited from OCM and include:

- **Core resources**: Pods, Services, ConfigMaps, Secrets, Namespaces
- **Apps resources**: Deployments, ReplicaSets, DaemonSets, StatefulSets  
- **Batch resources**: Jobs, CronJobs
- **Limited RBAC**: Basic roles and role bindings

> **📖 Complete Permission List**: For the authoritative and up-to-date list of default permissions, see the [OCM ManifestWork documentation](https://open-cluster-management.io/docs/concepts/work-distribution/manifestwork/#permission-setting-for-work-agent).

### When Additional Authorization is Needed

You'll need to expand OCM agent permissions when downsyncing:

| Workload Type | Default Support | Additional Auth Required |
|---------------|-----------------|-------------------------|
| Standard Kubernetes resources | ✅ Supported | ❌ No |
| Custom Resource Definitions and instances | ❌ Not supported | ✅ Yes |
| Cluster-scoped resources beyond defaults | ❌ Limited support | ✅ Usually |
| Operator-managed workloads | ⚠️ Partial support | ✅ Often required |

**Common scenarios requiring additional authorization:**
- Deploying custom resources (CRDs and their instances)
- Using operators that manage cluster-scoped resources
- Workloads requiring access to specific API groups not covered by defaults

### Expanding Authorization

KubeStellar provides two approaches for expanding OCM agent permissions:

#### Method 1: Downsyncing RBAC Objects (Recommended)

**Advantages**: Maintains KubeStellar's declarative model, provides audit trail, works with GitOps workflows

Include RBAC objects as part of your workload definition in the WDS:

```yaml
# Define additional permissions for custom resources
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: my-custom-resource-access
rules:
- apiGroups: ["my-api-group.io"]
  resources: ["mycustomresources"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: klusterlet-my-custom-resource-access
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: my-custom-resource-access
subjects:
- kind: ServiceAccount
  name: klusterlet-work-sa
  namespace: open-cluster-management-agent
```

This approach ensures that:
- Permissions are version-controlled with your workload
- All WECs receive consistent authorization
- Changes are auditable and reversible
- No direct WEC access is required

#### Method 2: Direct WEC RBAC Manipulation

**Advantages**: Immediate effect, doesn't require modifying workload definitions

Apply RBAC directly to each WEC:

```shell
# Apply to each WEC individually
kubectl --context wec1 apply -f rbac-expansion.yaml
kubectl --context wec2 apply -f rbac-expansion.yaml
```

**Considerations**: 
- Requires manual management across all WECs
- May not align with GitOps practices
- Risk of inconsistent permissions across WECs

### Troubleshooting Authorization Issues

#### Common Error Patterns

```
cannot create resource "mycustomresources" in API group "my-api-group.io"
User "system:serviceaccount:open-cluster-management-agent:klusterlet-work-sa" cannot create resource
```

#### Debugging Steps

1. **Identify the missing permission** from the error message
2. **Check if the resource is custom** - custom resources always require additional RBAC
3. **Verify the API group and resource name** for your RBAC rules
4. **Test permissions** using `kubectl auth can-i`:

```shell
# Test if the work agent has the required permissions
kubectl --context <wec-context> auth can-i create mycustomresources.my-api-group.io \
  --as=system:serviceaccount:open-cluster-management-agent:klusterlet-work-sa
```

#### Complete Example

For a complete working example of authorization expansion, see [Scenario 2 - Out-of-tree workload](example-scenarios.md#scenario-2---out-of-tree-workload), which demonstrates expanding permissions for AppWrapper custom resources.

**Security Best Practices:**
- Grant only the minimum necessary permissions
- Use namespaced permissions (Role/RoleBinding) when possible instead of cluster-scoped
- Document all permission expansions for audit purposes
- Regularly review and clean up unused permissions

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
