[![Go Report Card](https://goreportcard.com/badge/github.com/kubestellar/kubeflex)](https://goreportcard.com/report/github.com/kubestellar/kubeflex)
[![GitHub release](https://img.shields.io/github/release/kubestellar/kubeflex/all.svg?style=flat-square)](https://github.com/kubestellar/kubeflex/releases)
[![CI](https://github.com/kubestellar/kubeflex/actions/workflows/ci.yaml/badge.svg)](https://github.com/kubestellar/kubeflex/actions/workflows/ci.yaml)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=kubestellar_kubeflex&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=kubestellar_kubeflex)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=kubestellar_kubeflex&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=kubestellar_kubeflex)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/kubestellar/kubeflex)

# KubeFlex

A flexible and scalable platform for running Kubernetes control plane APIs with multi-tenancy support.

## Overview

KubeFlex is a CNCF sandbox project under the KubeStellar umbrella that enables "control-plane-as-a-service" multi-tenancy for Kubernetes. It provides a new approach to multi-tenancy by offering each tenant their own dedicated Kubernetes control plane and data-plane nodes in a cost-effective manner.

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

- **k8s**: Lightweight Kubernetes API server (~350MB) with essential controllers, using shared Postgres via Kine
- **vcluster**: Full virtual clusters based on the vCluster project, sharing host cluster worker nodes
- **host**: The hosting cluster itself exposed as a control plane for management scenarios
- **ocm**: Open Cluster Management control plane for multi-cluster federation scenarios
- **external**: Import existing external clusters under KubeFlex management (roadmap)

For detailed architecture information, see the [Architecture Guide](./architecture.md).

## Multi-Tenancy Approach

KubeFlex addresses the fundamental challenge of Kubernetes multi-tenancy by providing each tenant with a dedicated control plane while maintaining cost efficiency through shared infrastructure. This approach delivers strong isolation at both control and data plane levels.

For a comprehensive analysis of multi-tenancy approaches and KubeFlex's solution, see the [Multi-Tenancy Guide](./multi-tenancy.md).

## Installation

[kind](https://kind.sigs.k8s.io) and [kubectl](https://kubernetes.io/docs/tasks/tools/) are
required. A kind hosting cluster is created automatically by the kubeflex CLI. You may
also install KubeFlex on other Kube distros, as long as they support an nginx ingress
with SSL passthru, or on OpenShift. See the [User's Guide](./users.md) for more details.

Download the latest kubeflex CLI binary release for your OS/Architecture from the
[release page](https://github.com/kubestellar/kubeflex/releases) and copy it
to `/usr/local/bin` using the following command:

```shell
sudo su <<EOF
bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubeflex/main/scripts/install-kubeflex.sh) --ensure-folder /usr/local/bin --strip-bin
EOF
```

If you have [Homebrew](https://brew.sh), use the following commands to install kubeflex:

```shell
brew tap kubestellar/kubeflex https://github.com/kubestellar/kubeflex
brew install kflex
```

To upgrade the kubeflex CLI to the latest release, you may run:

```shell
brew upgrade kflex
```

## Quick Start

Get started with KubeFlex quickly by following our [Quick Start Guide](./quickstart.md). The guide includes:

- Basic multi-tenant setup with step-by-step commands
- Advanced development team scenarios with complete isolation
- Context switching and control plane management
- Cleanup and best practices

## Goals and Features

### Core Capabilities
- **Lightweight API Servers**: Provide dedicated Kubernetes API servers with minimal resource footprint
- **Flexible Storage Architecture**: Support shared databases, dedicated storage, or external systems
- **Custom API Server Builds**: Use upstream Kubernetes or specialized builds like multicluster-controlplane
- **Unified Management**: Single CLI for all control plane lifecycle operations

### Architecture Flexibility
- **Storage Options**: Shared Postgres, dedicated etcd, or Kine+Postgres configurations
- **API Server Variants**: Standard kubernetes API servers or trimmed-down specialized builds
- **Integration Ready**: Designed to work with existing Kubernetes ecosystem tools

### Operational Excellence
- **Zero-Touch Provisioning**: Automated control plane creation and configuration
- **Context Management**: Seamless switching between tenant environments
- **Lifecycle Management**: Complete control plane creation, update, and deletion workflows

## Documentation

- [Quick Start Guide](./quickstart.md): Get up and running quickly with KubeFlex
- [User Guide](./users.md): Detailed usage instructions and advanced scenarios
- [Architecture Guide](./architecture.md): Deep-dive into technical architecture
- [Multi-Tenancy Guide](./multi-tenancy.md): Comprehensive multi-tenancy analysis and use cases
- [Contributing Guide](./CONTRIBUTING.md): How to contribute to KubeFlex development

## Community and Support

- **Issues and Features**: [GitHub Issues](https://github.com/kubestellar/kubeflex/issues)
- **Community Discussion**: [KubeStellar Slack](https://kubestellar.io/slack)
- **Documentation**: [KubeStellar Website](https://docs.kubestellar.io/release-0.28.0/direct/kubeflex-intro/)

## License

KubeFlex is licensed under the Apache 2.0 License. See [LICENSE](./LICENSE) for the full license text.

---

*KubeFlex is part of the [KubeStellar](https://kubestellar.io) project, a CNCF sandbox initiative focused on multi-cluster configuration management for edge, multi-cloud, and hybrid cloud environments.*
