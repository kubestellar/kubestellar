# KubeFlex

A flexible and scalable platform for running Kubernetes control plane APIs with multi-tenancy support.

<div className="flex flex-wrap gap-2 my-4">
  <a href="https://goreportcard.com/report/github.com/kubestellar/kubeflex" target="_blank" rel="noopener noreferrer">
    <img src="https://goreportcard.com/badge/github.com/kubestellar/kubeflex" alt="Go Report Card" />
  </a>
  <a href="https://github.com/kubestellar/kubeflex/releases" target="_blank" rel="noopener noreferrer">
    <img src="https://img.shields.io/github/release/kubestellar/kubeflex/all.svg?style=flat-square" alt="GitHub release" />
  </a>
</div>

> **KubeFlex** is a CNCF sandbox project under the KubeStellar umbrella that enables "control-plane-as-a-service" multi-tenancy for Kubernetes.

## Overview

KubeFlex provides a new approach to multi-tenancy by offering each tenant their own dedicated Kubernetes control plane and data-plane nodes in a cost-effective manner. It addresses the fundamental challenge of Kubernetes multi-tenancy by providing strong isolation at both control and data plane levels while maintaining cost efficiency through shared infrastructure.

## Key Features

| Feature | Description |
|---------|-------------|
| **Lightweight API Servers** | Dedicated Kubernetes API servers with minimal resource footprint (~350MB) |
| **Flexible Storage** | Choose from shared Postgres, dedicated etcd, or Kine+Postgres configurations |
| **Multiple Control Plane Types** | Support for k8s, vcluster, host, ocm, and external cluster types |
| **Unified CLI (kflex)** | Single binary for initializing, managing, and switching between control planes |
| **Zero-Touch Provisioning** | Automated control plane creation and configuration |

## Architecture

KubeFlex implements a sophisticated multi-tenant architecture that separates control plane management from workload execution:

![KubeFlex Architecture](./images/kubeflex-architecture.png)

### Core Components

1. **KubeFlex Controller**: Orchestrates the lifecycle of tenant control planes through the ControlPlane CRD
2. **Tenant Control Planes**: Isolated API server and controller manager instances per tenant
3. **Flexible Data Plane**: Choose between shared host nodes, vCluster virtual nodes, or dedicated KubeVirt VMs
4. **Unified CLI (kflex)**: Single binary for initializing, managing, and switching between control planes
5. **Storage Abstraction**: Configurable backends from shared Postgres to dedicated etcd

### Supported Control Plane Types

| Type | Description |
|------|-------------|
| **k8s** | Lightweight Kubernetes API server (~350MB) with essential controllers, using shared Postgres via Kine |
| **vcluster** | Full virtual clusters based on the vCluster project, sharing host cluster worker nodes |
| **host** | The hosting cluster itself exposed as a control plane for management scenarios |
| **ocm** | Open Cluster Management control plane for multi-cluster federation scenarios |
| **external** | Import existing external clusters under KubeFlex management (roadmap) |

For detailed architecture information, see the [Architecture Guide](./architecture.md).

## Installation

[kind](https://kind.sigs.k8s.io) and [kubectl](https://kubernetes.io/docs/tasks/tools/) are required. A kind hosting cluster is created automatically by the kubeflex CLI. You may also install KubeFlex on other Kube distros, as long as they support an nginx ingress with SSL passthru, or on OpenShift. See the [User's Guide](./users.md) for more details.

### Using the install script

```shell
sudo su <<EOF
bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubeflex/main/scripts/install-kubeflex.sh) --ensure-folder /usr/local/bin --strip-bin
EOF
```

### Using Homebrew

```shell
brew tap kubestellar/kubeflex https://github.com/kubestellar/kubeflex
brew install kflex
```

To upgrade:

```shell
brew upgrade kflex
```

## Quick Start

Get started with KubeFlex quickly by following our [Quick Start Guide](./quickstart.md). The guide includes:

- Basic multi-tenant setup with step-by-step commands
- Advanced development team scenarios with complete isolation
- Context switching and control plane management
- Cleanup and best practices

## Documentation

- [Quick Start Guide](./quickstart.md): Get up and running quickly with KubeFlex
- [User Guide](./users.md): Detailed usage instructions and advanced scenarios
- [Architecture Guide](./architecture.md): Deep-dive into technical architecture
- [Multi-Tenancy Guide](./multi-tenancy.md): Comprehensive multi-tenancy analysis and use cases

## Community and Support

- **Issues and Features**: [GitHub Issues](https://github.com/kubestellar/kubeflex/issues)
- **Community Discussion**: [KubeStellar Slack](https://kubestellar.io/slack)
- **Source Code**: [GitHub Repository](https://github.com/kubestellar/kubeflex)

## License

KubeFlex is licensed under the Apache 2.0 License.

---

*KubeFlex is part of the [KubeStellar](https://kubestellar.io) project, a CNCF sandbox initiative focused on multi-cluster configuration management for edge, multi-cloud, and hybrid cloud environments.*
