# KubeFlex Architecture

KubeFlex implements a sophisticated multi-tenant architecture that separates control plane management from workload execution:

![KubeFlex Architecture](./images/kubeflex-architecture.png)

## Core Components

1. **KubeFlex Controller**: Orchestrates the lifecycle of tenant control planes through the ControlPlane CRD
2. **Tenant Control Planes**: Isolated API server and controller manager instances per tenant
3. **Flexible Data Plane**: Choose between shared host nodes, vCluster virtual nodes, or dedicated KubeVirt VMs
4. **Unified CLI (kflex)**: Single binary for initializing, managing, and switching between control planes
5. **Storage Abstraction**: Configurable backends from shared Postgres to dedicated etcd

## Supported Control Plane Types

KubeFlex is a flexible framework that supports various kinds of control planes, such as:

- *k8s*: a basic Kubernetes API Server with a subset of kube controllers. 
The control plane in this context does not execute workloads, such as pods, 
because the controllers associated with these objects are not activated. 
This environment is referred to as 'denatured' because it lacks the typical 
characteristics and functionalities of a standard Kubernetes cluster
It uses about 350 MB of memory per instance with a shared Postgres Database Backend.

- *vcluster*: a virtual cluster that runs on the hosting cluster, 
based on the  [vCluster Project](https://www.vcluster.com). This type of control 
plane can run pods using worker nodes of the hosting cluster.

- *host*: the KubeFlex hosting cluster, which is exposed as a control plane.

- *external*: an external cluster that is imported as a control plane (this
is in the roadmap but not yet implemented)

- *ocm*: a control plane that uses the 
[multicluster-controlplane project](https://github.com/open-cluster-management-io/multicluster-controlplane) 
for managing multiple clusters.

When using KubeFlex, users interact with the API server 
of the hosting cluster to create or delete control planes.
KubeFlex defines a ControlPlane CRD that represents a Control Plane.

![image info](./images/kubeflex-architecture.png)

When a user initiates the creation of a Control Plane Custom Resource (CR) by
executing the `kflex create <cp>` command or the `kubectl apply -f mycontrolplane.yaml` 
command for a control plane of type k8s, the KubeFlex controller creates a new namespace 
within the hosting cluster, and then deploys the following artifacts in that namespace:

1. **Deploys a Kubernetes API server** instance within a pod, which
    serves the API for the control plane.

    - For control planes designated as type `k8s`, the API server is
      configured to use a **shared Postgres database** as a backend DB.
      This database is located in the `kubeflex-system` namespace. KubeFlex takes advantage of
      [**kine**](https://github.com/k3s-io/kine), a tool that
      emulates the etcd interface for Postgres, allowing the API server
      to interact with the database.
      
      > **Note**: PostgreSQL is installed using PostCreateHook Jobs rather than Helm subchart dependencies. 
      > For detailed information about this architectural decision, see [PostgreSQL Architecture Decision](./postgresql-architecture-decision.md).

    - For control planes designated as type `vcluster`, the API
      server and an embedded etcd database run as a single process in
      the pod, and mount a persistent volume for etcd.

2.  **Deploys a Kubernetes controller manager** within a pod in the new
    namespace. This controller manager operates a select group of
    essential Kubernetes controllers, such as namespace, garbage
    collection (gc), and service account controllers. The controller
    manager is configured to connect to the hosted Kubernetes API
    server.

3.  Creates a **Service for the Kubernetes API server**

4.  Creates either an **Ingress** or a **Route** for providing external
    connectivity to the API server. The latter is adopted if the hosting
    cluster is an OpenShift cluster.

5.  Creates a **Secret** that contains the `kubeconfig` file required by
    clients to access the Control Plane API server. The secret has
    actually two Kubeconfigs, one for off-cluster access and one for
    in-cluster access. The latter is used by clients running in the
    hosting cluster, such as controllers.

This flow might be slightly different for other types of control planes, such as
`vcluster` or `host`, but there are common elements for each type:

- A namespace with the name of the control plane and the `-system` suffix

- A secret with the Kubeconfigs for in-cluster and off-cluster access.
  For clusters of type `host` only the in-cluster Kubeconfig is available,

- A service to connect to the hosted API server

- A route or ingress to connect to the hosted API server 

The last two elements are not present for ControlPlanes of type `host` and `external`.

