# Common Setup - step-by-step

The following steps create a deployment of KubeStellar Core

1. You may want to `set -e` in your shell so that any failures in the setup or usage scenarios are not lost.

1. If you ran through these scenarios previously then you will need to do a bit of cleanup first. See how it is done in the cleanup script for our E2E tests (in `test/e2e/common/cleanup.sh`).

1. Set environment variables to hold KubeStellar and OCM-transport-plugin desired versions:

    ```shell
    export KUBESTELLAR_VERSION=0.23.0
    export OCM_TRANSPORT_PLUGIN=0.1.11
    ```

1. Create a Kind hosting cluster with nginx ingress controller and KubeFlex controller-manager installed:

    ```shell
    kflex init --create-kind
    ```
   If you are installing KubeStellar on an existing Kubernetes or OpenShift cluster, just use the command `kflex init`.

1. Update the post-create-hooks in KubeFlex to install kubestellar with the desired images:

    ```shell
    kubectl apply -f https://raw.githubusercontent.com/kubestellar/kubestellar/v${KUBESTELLAR_VERSION}/config/postcreate-hooks/kubestellar.yaml
    kubectl apply -f https://raw.githubusercontent.com/kubestellar/kubestellar/v${KUBESTELLAR_VERSION}/config/postcreate-hooks/ocm.yaml
    ```

1. Create an inventory & mailbox space of type `vcluster` running *OCM* (Open Cluster Management)
in KubeFlex. Note that `-p ocm` runs a post-create hook on the *vcluster* control plane
which installs OCM on it.
This step includes the installation of the status add-on controller.
See [here](./architecture.md#ocm-status-add-on-agent) for more details on the add-on.

    ```shell
    kflex create its1 --type vcluster -p ocm
    ```

1. Create a Workload Description Space `wds1` in KubeFlex. Similarly to before, `-p kubestellar`
runs a post-create hook on the *k8s* control plane that starts an instance of a KubeStellar controller
manager which connects to the `wds1` front-end and the `its1` OCM control plane back-end.
Additionally the OCM based transport controller will be deployed.

    ```shell
    kflex create wds1 -p kubestellar
    ```

1. (optional) Check relevant deployments and statefulsets running in the hosting cluster. Expect to
see the `kubestellar-controller-manager` in the `wds1-system` namespace and the
statefulset `vcluster` in the `its1-system` namespace, both fully ready.

    ```shell
    kubectl --context kind-kubeflex get deployments,statefulsets --all-namespaces
    ```
   The output should look something like the following:

    ```shell
    NAMESPACE            NAME                                             READY   UP-TO-DATE   AVAILABLE   AGE
    ingress-nginx        deployment.apps/ingress-nginx-controller         1/1     1            1           22h
    kube-system          deployment.apps/coredns                          2/2     2            2           22h
    kubeflex-system      deployment.apps/kubeflex-controller-manager      1/1     1            1           22h
    local-path-storage   deployment.apps/local-path-provisioner           1/1     1            1           22h
    wds1-system          deployment.apps/kube-apiserver                   1/1     1            1           22m
    wds1-system          deployment.apps/kube-controller-manager          1/1     1            1           22m
    wds1-system          deployment.apps/kubestellar-controller-manager   1/1     1            1           21m
    wds1-system          deployment.apps/transport-controller             1/1     1            1           21m

    NAMESPACE         NAME                                   READY   AGE
    its1-system       statefulset.apps/vcluster              1/1     11h
    kubeflex-system   statefulset.apps/postgres-postgresql   1/1     22h
    ```