# What is KubeStellar?

KubeStellar is a multi-cluster Kubernetes orchestration platform that simplifies how organizations manage distributed workloads across multiple Kubernetes clusters. At its core, KubeStellar provides a sophisticated yet accessible approach to multi-cluster management through three fundamental capabilities:

## Single Control Plane

KubeStellar introduces a revolutionary approach to multi-cluster management through its unified control plane architecture:

- **Native Kubernetes Objects**: Work with Kubernetes objects in their native format without wrapping or bundling
- **Workload Definition Spaces (WDSes)**: Store and manage workload definitions centrally
- **Inventory and Transport Spaces (ITSes)**: Efficiently manage cluster inventory and workload transport
- **Simplified Operations**: Eliminate context switching between clusters
- **Centralized Management**: Control multiple clusters from a single point

[Learn more about KubeStellar Control Plane](architecture.md)

## Intelligent Workload Placement

KubeStellar's sophisticated workload placement system ensures optimal distribution of applications:

- **Label-Based Selection**: Use cluster labels for precise workload targeting
- **Binding Policies**: Declaratively specify what workloads run where
- **Flexible Distribution**: Support for:
  - Standard Kubernetes resources
  - Custom Resources (CRDs)
  - Helm charts
  - Out-of-tree workload types
- **Automatic Synchronization**: Keep workloads and configurations in sync across clusters

[Learn more about Workload Binding](binding.md)

## Policy-Driven Management

Implement comprehensive governance through centralized policy enforcement:

- **BindingPolicy System**: Define rules for workload placement and configuration
- **Status Management**:
  - Singleton status reporting for individual cluster monitoring
  - Combined status aggregation across clusters
- **Custom Transforms**: Apply transformations to workloads during distribution
- **Resilient Architecture**:
  - Automatic recovery after disruptions
  - State reconciliation across clusters
  - Robust error handling

[Learn more about Policy Control](control.md)

## Real-World Use Cases

KubeStellar excels in several key scenarios:

1. **Declarative Multi-Cluster Management**: Deploy and manage workloads across clusters using native Kubernetes objects
2. **Custom Resource Distribution**: Manage CRDs and custom resources across multiple clusters
3. **Advanced Status Management**: Monitor and aggregate workload status across clusters
4. **Helm Chart Distribution**: Deploy and manage Helm charts consistently across clusters
5. **Template-Based Customization**: Customize workload configurations for different clusters while maintaining a single source of truth

## Why Choose KubeStellar?

- **Native Experience**: Work with standard Kubernetes objects and practices
- **Scalable Architecture**: Efficiently manage workloads across any number of clusters
- **Operational Simplicity**: Reduce complexity in multi-cluster management
- **Enterprise Ready**: Built for production use cases with resilience and scalability
- **Flexible Control**: Fine-grained control over workload placement and configuration

Whether you're managing edge deployments, implementing multi-cloud strategies, or scaling your Kubernetes infrastructure, KubeStellar provides the tools and capabilities needed for effective multi-cluster orchestration.

[Get Started with KubeStellar](get-started.md)
