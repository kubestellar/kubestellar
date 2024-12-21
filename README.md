[![Go Report Card](https://goreportcard.com/badge/github.com/kubestellar/kubeflex)](https://goreportcard.com/report/github.com/kubestellar/kubeflex)
[![GitHub release](https://img.shields.io/github/release/kubestellar/kubeflex/all.svg?style=flat-square)](https://github.com/kubestellar/kubeflex/releases)
[![CI](https://github.com/kubestellar/kubeflex/actions/workflows/ci.yaml/badge.svg)](https://github.com/kubestellar/kubeflex/actions/workflows/ci.yaml)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=kubestellar_kubeflex&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=kubestellar_kubeflex)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=kubestellar_kubeflex&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=kubestellar_kubeflex)

# <img alt="Logo" width="90px" src="./docs/images/kubeflex-logo.png" style="vertical-align: middle;" />  KubeFlex

A flexible and scalable platform for running Kubernetes control plane APIs.

## Goals

- Provide lightweight Kube API Server instances and selected controllers as a service.
- Provide a flexible architecture for the storage backend, e.g.:
    - shared DB for API servers,
    - dedicated DB for each API server,
    - etcd DB or Kine + Postgres DB
- Flexibility in choice of API Server build:
    - upstream Kube (e.g. `registry.k8s.io/kube-apiserver:v1.27.1`),
    - trimmed down API Server builds (e.g. [multicluster control plane](https://github.com/open-cluster-management-io/multicluster-controlplane))
- Single binary CLI for improved user experience:
    - initialize, install operator, manage lifecycle of control planes and contexts.

## Installation

[kind](https://kind.sigs.k8s.io) and [kubectl](https://kubernetes.io/docs/tasks/tools/) are
required. A kind hosting cluster is created automatically by the kubeflex CLI. You may
also install KubeFlex on other Kube distros, as long as they support an nginx ingress
with SSL passthru, or on OpenShift. See the [User's Guide](docs/users.md) for more details.

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
brew install kubeflex
```

To upgrade the kubeflex CLI to the latest release, you may run:

```shell
brew upgrade kubeflex
```

## Quickstart

Create the hosting kind cluster with ingress controller and install
the kubeflex operator:

```shell
kflex init --create-kind
```

Create a control plane:

```shell
kflex create cp1
```

Interact with the new control plane, for example get namespaces and create a new one:

```shell
kubectl get ns
kubectl create ns myns
```

Create a second control plane and check that the namespace created in the
first control plane is not present:

```shell
kflex create cp2
kubectl get ns
```

To go back to the hosting cluster context, use the `ctx` command:

```shell
kflex ctx
```

To switch back to a control plane context, use the
`ctx <control plane name>` command, e.g:

```shell
kflex ctx cp1
```

To delete a control plane, use the `delete <control plane name>` command, e.g:

```shell
kflex delete cp1
```

## Next Steps

Read the [User's Guide](docs/users.md) to learn more about using KubeFlex for your project
and how to create and interact with different types of control planes, such as
[vcluster](https://www.vcluster.com) and [Open Cluster Management](https://github.com/open-cluster-management-io/multicluster-controlplane).

## Architecture

![image info](./docs/images/kubeflex-high-level-arch.png)
