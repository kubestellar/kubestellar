# Inventory and Transport Spaces

- [What is an ITS?](#what-is-an-its)
- [Creating an ITS](#creating-an-its)
  - [Using the KubeStellar Core Helm Chart](#using-the-kubestellar-core-helm-chart)
  - [Using the KubeFlex CLI](#using-the-kubeflex-cli)
- [KubeFlex Hosting Cluster as ITS](#kubeflex-hosting-cluster-as-its)
- [Important Note on ITS Registration](#important-note-on-its-registration)
- [Architecture and Components](#architecture-and-components)

---

## Introduction

An Inventory and Transport Space (ITS) is a core component of the KubeStellar architecture. It combines:

- **Inventory Space** – part of the interface to KubeStellar that maintains a registry of all Workload Execution Clusters (WECs)
- **Transport Space** – part of the implementation that delivers workloads from Workload Description Spaces (WDSes) to the appropriate WECs

An ITS serves as an OCM (Open Cluster Management) "hub" and enables two primary functions:

- **Inventory Management**  
  Maintains a registry of all WECs available in the system.

- **Transport Facilitation**  
  Handles the movement of workloads from WDSes to their targeted WECs.

---

## What is an ITS?

An ITS is a Kubernetes-like API server (with storage) that:

- Holds inventory information about all registered WECs using [`ManagedCluster.v1.cluster.open-cluster-management.io`](https://github.com/open-cluster-management-io/api/blob/v0.12.0/cluster/v1/types.go#L33) objects
- Contains a `customization-properties` namespace with ConfigMaps carrying additional properties for each WEC
- Manages mailbox namespaces that correspond 1:1 with each WEC, holding `ManifestWork` objects
- Runs the OCM Cluster Manager to synchronize objects with WECs

---

## Creating an ITS

> **Note:** The only supported method for creating an ITS is using the KubeStellar Core Helm Chart or KubeFlex CLI.

### Using the KubeStellar Core Helm Chart

To create an ITS using the KubeStellar Core Chart:

```bash
helm upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart \
  --set-json='ITSes=[{"name":"its1", "type":"vcluster"}]'
