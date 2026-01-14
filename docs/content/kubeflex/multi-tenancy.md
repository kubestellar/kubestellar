# Multi-Tenancy with KubeFlex

## The Multi-Tenancy Challenge

As organizations scale their Kubernetes adoption, they face a fundamental question: how to efficiently share cluster resources across teams and applications while maintaining proper isolation, security, and cost efficiency? Traditional Kubernetes multi-tenancy approaches present significant trade-offs:

| Approach | Control Plane Isolation | Data Plane Isolation | Operational Cost | Tenant Flexibility |
|----------|------------------------|---------------------|------------------|-------------------|
| **Cluster-as-a-Service** | Full | Full | Very High | Full |
| **Namespace-as-a-Service** | None | Partial | Low | Limited |
| **Control-Plane-as-a-Service** | Full | Shared | Medium | High |
| **KubeFlex (Enhanced CaaS)** | Full | Full | Medium | High |

**The Problem**: Organizations need the isolation benefits of dedicated clusters without the operational overhead and cost. Namespace-based sharing is cost-effective but creates security and noisy-neighbor risks. Full cluster-per-tenant approaches provide excellent isolation but lead to cluster sprawl and wasted resources.

**KubeFlex's Solution**: Provides each tenant with a dedicated Kubernetes control plane (API server + controllers) while offering optional dedicated data-plane nodes through integration with KubeVirt. This approach delivers strong isolation at both control and data plane levels while maintaining cost efficiency through shared infrastructure.

*Learn more about multi-tenancy isolation approaches in this [comprehensive analysis](https://medium.com/@brauliodumba/cloud-computing-multi-tenancy-isolation-a-new-approach-815ff3e6dfd1).*

## KubeFlex Scope and Third-Party Integration Boundaries

**What KubeFlex Provides:**
- Control plane provisioning and lifecycle management
- Multi-tenant API server isolation
- Flexible storage backend abstraction
- CLI tooling for tenant management
- Integration hooks for post-creation workflows

**What KubeFlex Integrates With:**
- **KubeVirt**: For VM-based worker nodes providing complete tenant isolation
- **vCluster**: As a control plane type for lightweight virtual clusters
- **Open Cluster Management**: For multi-cluster scenarios and edge deployments
- **Standard Kubernetes Storage**: CSI drivers, persistent volumes, and storage classes

**Integration Boundaries:**
KubeFlex focuses on control plane management and provides integration points rather than reimplementing existing solutions. For example, when using KubeVirt for data plane isolation, KubeFlex creates the control plane while KubeVirt handles VM provisioning and management.

## Use Cases and Benefits

### 1. Multi-Tenant SaaS Platforms
- **Challenge**: Provide isolated environments for hundreds of customers
- **Solution**: Create lightweight control planes per customer using the `k8s` type
- **Benefit**: Strong isolation without the cost of dedicated clusters

### 2. Enterprise Development Teams
- **Challenge**: Multiple teams need Kubernetes access without cluster sprawl
- **Solution**: Dedicated control planes with shared infrastructure
- **Benefit**: Teams get cluster-admin privileges in their own control plane

### 3. CI/CD and Testing
- **Challenge**: Isolated environments for parallel testing
- **Solution**: Ephemeral control planes created and destroyed per test run
- **Benefit**: True isolation between test runs with quick provisioning

### 4. Edge and Multi-Cluster Management
- **Challenge**: Manage multiple edge locations with varying connectivity
- **Solution**: Use `ocm` type control planes for edge cluster federation
- **Benefit**: Centralized management with distributed execution

## Advanced Configuration

### Custom Control Plane Components

```yaml
apiVersion: tenancy.kflex.kubestellar.org/v1alpha1
kind: ControlPlane
metadata:
  name: custom-tenant
spec:
  type: k8s
  backend: dedicated  # Use dedicated etcd instead of shared Postgres
  tokenExpirationSeconds: 7200  # 2-hour token expiration
  postCreateHooks:
    - hookName: "setup-monitoring"
      vars:
        prometheus_namespace: "monitoring"
    - hookName: "configure-networking"
      vars:
        network_policy: "strict"
```

### Storage Backend Options

1. **Shared Postgres (Default)**:
   - Multiple tenants share a Postgres instance
   - Uses Kine for etcd-compatible API
   - Most cost-effective for large numbers of tenants

2. **Dedicated etcd**:
   - Each tenant gets their own etcd instance
   - Best performance and isolation
   - Higher resource usage

3. **External Database**:
   - Connect to existing database infrastructure
   - Useful for compliance or existing investments

### Integration with KubeVirt for Data Plane Isolation

For scenarios requiring complete workload isolation:

```yaml
apiVersion: tenancy.kflex.kubestellar.org/v1alpha1
kind: ControlPlane
metadata:
  name: secure-tenant
spec:
  type: k8s
  postCreateHooks:
    - hookName: "kubevirt-nodes"
      vars:
        node_count: "3"
        vm_memory: "4Gi"
        vm_cpu: "2"
```

This creates a control plane where workloads run in dedicated KubeVirt VMs, providing:
- Complete isolation from other tenants
- Protection against container breakout attacks
- Dedicated compute resources per tenant

## Next Steps

- Start with the [Quick Start Guide](quickstart.md) to get hands-on experience
- Read the [User's Guide](users.md) for detailed usage instructions
- Explore [Architecture](architecture.md) to understand the technical implementation
