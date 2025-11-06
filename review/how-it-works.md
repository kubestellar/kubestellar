# How KubeStellar Works

KubeStellar orchestrates your multi-cluster environment through a well-defined architecture and workflow:

## 1. Set Up Your Environment

### Prerequisites and Core Infrastructure

```shell
# Install required tools
- kubectl, helm, docker, kind/k3d
- KubeFlex
- Open Cluster Management (OCM) CLI
```

- Create a KubeFlex hosting cluster
- Initialize core components:
  - Inventory and Transport Space (ITS)
  - Workload Definition Space (WDS)
- Set up Workload Execution Clusters (WECs)

## 2. Register and Label Clusters

```yaml
# Example cluster labeling
kubectl label managedcluster cluster1 \
location-group=edge \
name=cluster1
```

- Register WECs with the ITS using OCM
- Apply labels to clusters for targeting
- Establish secure connections between control plane and WECs
- Verify cluster registration status

## 3. Define Workload Placement

```yaml
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: example-policy
spec:
  clusterSelectors:
    - matchLabels:
        location-group: edge
  downsync:
    - objectSelectors:
        - matchLabels:
          app.kubernetes.io/name: myapp
```

- Create BindingPolicy objects to specify:
  - Which clusters receive workloads (clusterSelectors)
  - Which workloads to distribute (downsync rules)
  - Optional transformations and status collection

## 4. Deploy Your Workloads

```yaml
# Deploy workloads in native Kubernetes format
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-app
  labels:
    app.kubernetes.io/name: myapp
spec:
  replicas: 3
  # ... rest of deployment spec
```

Supported deployment methods:

- Direct kubectl apply
- Helm charts
- ArgoCD integration
- Custom Resources (CRDs)

## 5. Monitor and Manage

KubeStellar provides comprehensive monitoring capabilities:

- View deployment status across clusters
- Monitor workload health
- Collect and aggregate status information
- Manage policy compliance
- Handle cluster additions/removals dynamically

### Status Collection Options

```yaml
# Enable detailed status collection
spec:
  downsync:
    - wantSingletonReportedState: true
      objectSelectors:
        - matchLabels:
          app.kubernetes.io/name: myapp
```

## Real-World Usage Patterns

KubeStellar supports various deployment scenarios:

1. **Standard Kubernetes Resources**
   - Deployments, Services, ConfigMaps
   - Native format, no wrapping required

2. **Helm Chart Distribution**
   - Deploy charts across clusters
   - Maintain release information
   - Automated chart synchronization

3. **Custom Resource Management**
   - Distribute CRDs and custom resources
   - Maintain consistency across clusters
   - Handle specialized workloads

4. **Policy-Based Deployments**
   - Define cluster selection rules
   - Set resource constraints
   - Implement governance policies

Each step is backed by KubeStellar's resilient architecture, ensuring reliable workload distribution and management across your entire Kubernetes estate.

[Get started with KubeStellar](get-started.md)
