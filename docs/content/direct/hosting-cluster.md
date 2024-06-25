# The KubeFlex hosting cluster

This document tells you what makes a Kubernetes cluster suitable to serve as the KubeFlex hosting cluster and how to initialize that cluster to play that role.

## Requirements on the KubeFlex hosting cluster

The KubeFlex hosting cluster needs to run an Ingress controller with
SSL passthrough enabled.

## Connectivity from clients

The clients in KubeStellar need to be able to open a TCP connection to where the Ingress controller is listening for HTTPS connections.

The clients in KubeStellar comprise the following.

- The OCM Agent and the OCM Status Add-On Agent in each WEC.
- The KubeStellar controller-manager and the transport controller for each WDS, running in the KubeFlex hosting cluster.

TODO: finish writing this subsection for real. Following are some clues.

When everything runs on one machine, the defaults just work. When core and some WECs are on different machines, it gets more challenging. When the KubeFlex hosting cluster is an OpenShift cluster with a public domain name, the defaults just work.

After the quickstart setup, I looked at an OCM Agent (klusterlet-agent, to be specifc) and did not find a clear passing of kubeconfig. I found adjacent Secrets holding kubeconfigs in which `cluster[0].cluster.server` was `https://kubeflex-control-plane:31048`. Note that `kubeflex-control-plane` is the name of the Docker container running `kind` cluster serving as KubeFlex hosting cluster. I could not find an explanation for the port number 31048; that Docker container maps port 443 inside to 9443 on the outside.

`kflex init` takes a command line flag `--domain string` described as `domain for FQDN (default "localtest.me")`.

## Creating a hosting cluster

### Create and init a kind cluster as hosting cluster with kflex

The following command will use `kind` to create a cluster with an Ingress controller with SSL passthrough and then proceed to install the KubeFlex implementation in it and set your current kubeconfig context to access that cluster as admin.

```shell
kflex init --create-kind
```

### Create and init a kind cluster as hosting cluster with curl-to-bash script

There is a bash script at [`https://raw.githubusercontent.com/kubestellar/kubestellar/v{{ config.ks_latest_regular_release }}/scripts/create-kind-cluster-with-SSL-passthrough.sh`](https://raw.githubusercontent.com/kubestellar/kubestellar/v{{ config.ks_latest_regular_release }}/scripts/create-kind-cluster-with-SSL-passthrough.sh) that can be fed directly into `bash` and will create a `kind` cluster and initialize it as the KubeFlex hosting cluster. This script accepts the following command line flags.

- `--name name`: set a specific name of the kind cluster (default: kubestellar).
- `--port port`: map the specified host port to the kind cluster port 443 (default: 9443).
- `--nowait`: when given, the script proceeds without waiting for the nginx ingress patching to complete.
- `--nosetcontext`: when given, the script does not change the current kubectl context to the newly created cluster.
- `-X` enable verbose execution of the script for debugging.

### Create a k3d cluster

1. Create a K3D hosting cluster with nginx ingress controller:
    ```shell
    k3d cluster create -p "9443:443@loadbalancer" --k3s-arg "--disable=traefik@server:*" kubeflex
    helm install ingress-nginx ingress-nginx --repo https://kubernetes.github.io/ingress-nginx --version 4.6.1 --namespace ingress-nginx --create-namespace
    ```

1. When we use kind, the name of the container is kubeflex-control-plane and that is what we use 
   in the internal URL for `--force-internal-endpoint-lookup`.
   Here the name of the container created by K3D is `k3d-kubeflex-server-0` so we rename it:
    ```shell
    docker stop k3d-kubeflex-server-0
    docker rename k3d-kubeflex-server-0 kubeflex-control-plane
    docker start kubeflex-control-plane
    ```
    Wait 1-2 minutes for all pods to be restarted.
    Use the following command to confirm all are fully running:
    ```shell
    kubectl --context k3d-kubeflex get po -A
    ```

1. Enable SSL passthrough:
   We are using nginx ingress with tls passthru.
   The current install for kubeflex installs also nginx ingress but specifically for kind.
   To specify passthru for K3D, edit the ingress placement controller with the following command and add `--enable-ssl-passthrough` to the list of args for the container
    ```shell
    kubectl edit deployment ingress-nginx-controller -n ingress-nginx  
    ```

## Initialize the KubeFlex hosting cluster

The KubeFlex implementation has to be installed in the cluster chosen to play the role of KubeFlex hosting cluster. This can be done in any of the following ways.

### kflex init

The following command will install the KubeFlex implementation in the cluster that `kubectl` is configured to access, if you have sufficient privileges.

```shell
kflex init
```

#### Using an existing OpenShift cluster as the hosting cluster

When the hosting cluster is an OpenShift cluster, the recipe for registering a WEC with the ITS ([to be written](wec.md)) needs to be modified. In the `clusteradm` command, omit the `--force-internal-endpoint-lookup` flag. If following [the quickstart commands](get-started.md#create-and-register-two-workload-execution-clusters) literally, this means to define `flags=""` rather than `flags="--force-internal-endpoint-lookup"`.

### kflex init --create-kind

See [above](#create-and-init-a-kind-cluster-as-hosting-cluster-with-kflex).

### Curl-to-bash

See [above](#create-and-init-a-kind-cluster-as-hosting-cluster-with-curl-to-bash-script).

### KubeStellar core Helm chart

The [KubeStellar core Helm chart](core-chart.md) will install the KubeFlex implementation in the cluster that `kubectl` is configured to access, as well as create ITSes and WDSes.

