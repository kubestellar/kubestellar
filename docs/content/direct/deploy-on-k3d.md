# Deploy KubeStellar on K3D

This document shows how to deploy kubestellar on K3D hub and wec clusters.

This is a point-in-time statement that worked with KubeStellar release v0.22.0.

## Prereqs

In addition to [pre-reqs](pre-reqs.md), install k3d v5.6.0 (only k3d version tested so far)

## Common Setup for standard examples

1. You may want to `set -e` in your shell so that any failures in the setup or usage scenarios are not lost.

1. If you previously installed KS on K3D:
    ```shell
    k3d cluster delete kubeflex
    k3d cluster delete cluster1
    k3d cluster delete cluster2
    kubectl config delete-context kubeflex || true
    kubectl config delete-context cluster1 || true
    kubectl config delete-context cluster2 || true
    ```
   If previously running KS on Kind, clean that up with the Kind cleanup script (in `test/e2e/common/cleanup.sh`).

1. Set environment variables to hold KubeStellar and OCM-status-addon desired versions:
    ```shell
    export KUBESTELLAR_VERSION=0.22.0
    export OCM_STATUS_ADDON_VERSION=0.2.0-rc8
    export OCM_TRANSPORT_PLUGIN=0.1.7
    ```

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

1. Install kubestellar controller and OCM space:
   We are using nginx ingress with tls passthru.
   The current install for kubeflex installs also nginx ingress but specifically for kind.
   To specify passthru for K3D, edit the ingress placement controller with the following command and add `--enable-ssl-passthrough` to the list of args for the container
    ```shell
    kubectl edit deployment ingress-nginx-controller -n ingress-nginx  
    ```
   Then initialize kubeflex and create the its1 space with OCM running in it:
    ```shell
    kflex init
    kubectl apply -f https://raw.githubusercontent.com/kubestellar/kubestellar/v${KUBESTELLAR_VERSION}/config/postcreate-hooks/kubestellar.yaml
    kubectl apply -f https://raw.githubusercontent.com/kubestellar/kubestellar/v${KUBESTELLAR_VERSION}/config/postcreate-hooks/ocm.yaml
    kubectl config rename-context k3d-kubeflex kind-kubeflex
    kflex create its1 --type vcluster -p ocm
    ```

1. Install OCM status addon
   First wait until managedclusteraddons resource shows up on its1 using:
    ```shell
   kubectl --context its1 api-resources | grep managedclusteraddons
    ```
   then install status addon:
    ```shell
    helm --kube-context its1 upgrade --install status-addon -n open-cluster-management oci://ghcr.io/kubestellar/ocm-status-addon-chart --version v${OCM_STATUS_ADDON_VERSION}
    ```

1. Create a Workload Description Space `wds1` in KubeFlex.
    ```shell
    kflex create wds1 -p kubestellar
    ```

1. Deploy the OCM based transport controller.

    ```shell
    helm --kube-context kind-kubeflex upgrade --install ocm-transport-plugin oci://ghcr.io/kubestellar/ocm-transport-plugin/chart/ocm-transport-plugin --version ${OCM_TRANSPORT_PLUGIN} \
     --set transport_cp_name=its1 \
     --set wds_cp_name=wds1 \
     -n wds1-system
    ```

1. Create the Workload Execution Cluster `cluster1` and register it
   Make sure `cluster1` shares the same docker network as the `kubeflex` hosting cluster.
    ```shell
    k3d cluster create -p "31080:80@loadbalancer"  --network k3d-kubeflex cluster1
    kubectl config rename-context k3d-cluster1 cluster1
    ```
   Register `cluster1`:
    ```shell
    flags="--force-internal-endpoint-lookup"
    clusteradm --context its1 get token | grep '^clusteradm join' | sed "s/<cluster_name>/cluster1/" | awk '{print $0 " --context 'cluster1' '${flags}'"}' | sh
    ```
   Wait for csr to be created:
    ```shell
    kubectl --context its1 get csr --watch || true
    ```
    and then accept pending cluster1 cluster
    ```shell
    clusteradm --context its1 accept --clusters cluster1
    ```
    Confirm cluster1 is accepted and label it for the BindingPolicy:
    ```shell
    kubectl --context its1 get managedclusters
    kubectl --context its1 label managedcluster cluster1 location-group=edge name=cluster1
    ```

1. In the same way, make another WEC named "cluster2".
    ```shell
    k3d cluster create -p "31180:80@loadbalancer"  --network k3d-kubeflex cluster2
    kubectl config rename-context k3d-cluster2 cluster2
    ```
   Register `cluster2`:
    ```shell
    flags="--force-internal-endpoint-lookup"
    clusteradm --context its1 get token | grep '^clusteradm join' | sed "s/<cluster_name>/cluster2/" | awk '{print $0 " --context 'cluster2' '${flags}'"}' | sh
    ```
   Wait for csr to be created:
    ```shell
    kubectl --context its1 get csr --watch || true
    ```
    and then accept pending cluster2 cluster
    ```shell
    clusteradm --context its1 accept --clusters cluster2
    ```
    Confirm cluster2 is accepted and label it for the BindingPolicy:
    ```shell
    kubectl --context its1 get managedclusters
    kubectl --context its1 label managedcluster cluster2 location-group=edge name=cluster2
    ```

1. (optional) Check relevant deployments and statefulsets running in the hosting cluster. Expect to
see the `kubestellar-controller-manager` in the `wds1-system` namespace and the 
statefulset `vcluster` in the `its1-system` namespace, both fully ready.

    ```shell
    kubectl --context kind-kubeflex get deployments,statefulsets --all-namespaces
    ```
   The output should look something like the following:
    ```
    NAMESPACE         NAME                                             READY   UP-TO-DATE   AVAILABLE   AGE
    kube-system       deployment.apps/coredns                          1/1     1            1           10m
    kube-system       deployment.apps/local-path-provisioner           1/1     1            1           10m
    kube-system       deployment.apps/metrics-server                   1/1     1            1           10m
    ingress-nginx     deployment.apps/ingress-nginx-controller         1/1     1            1           9m50s
    kubeflex-system   deployment.apps/kubeflex-controller-manager      1/1     1            1           5m45s
    wds1-system       deployment.apps/kube-apiserver                   1/1     1            1           3m54s
    wds1-system       deployment.apps/kube-controller-manager          1/1     1            1           3m54s
    wds1-system       deployment.apps/kubestellar-controller-manager   1/1     1            1           3m29s
    wds1-system       deployment.apps/transport-controller             1/1     1            1           2m52s

    NAMESPACE         NAME                                   READY   AGE
    kubeflex-system   statefulset.apps/postgres-postgresql   1/1     6m12s
    its1-system       statefulset.apps/vcluster              1/1     5m17s
    ```

You may now proceed with Sceario 1 and others in [the examples](examples.md) (do not do the "Common Setup", that is written for `kind` rather than `k3d`).
