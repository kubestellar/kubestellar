# KubeStellar prerequisites

The following prerequisites are required.

Use the [check_pre_req](contributor.md#check-key-pre-requisites-for-kubestellar) script to quickly check which pre-requisites are already installed:

```shell
$ hack/check_pre_req.sh
Checking pre-requisites for using KubeStellar:
✔ Docker
✔ kubectl
✔ KubeFlex
✔ OCM CLI
✔ Helm
Checking additional pre-requisites for running the examples:
✔ Kind
X ArgoCD CLI
Checking pre-requisites for building KubeStellar:
✔ GNU Make
✔ Go
✔ KO
```

## For Using KubeStellar

- kubeflex version 0.4.2 or higher
    To install kubeflex go to [https://github.com/kubestellar/kubeflex/blob/main/docs/users.md#installation](https://github.com/kubestellar/kubeflex/blob/main/docs/users.md#installation). To upgrade from an existing installation,
follow [these instructions](https://github.com/kubestellar/kubeflex/blob/main/docs/users.md#upgrading-kubeflex). At the end of the install make sure that the kubeflex CLI, kflex, is in your path.

- OCM CLI (clusteradm)
    To install OCM CLI use:

    ```shell
    curl -L https://raw.githubusercontent.com/open-cluster-management-io/clusteradm/main/install.sh | bash
    ```

    Note that the default installation of clusteradm will install in /usr/local/bin which will require root access. If you prefer to avoid root, you can specify an alternative installation path using the INSTALL_DIR environment variable, as follows:

    ```shell
    mkdir -p ocm
    export INSTALL_DIR="$PWD/ocm"
    curl -L https://raw.githubusercontent.com/open-cluster-management-io/clusteradm/main/install.sh | bash
    export PATH=$PWD/ocm:$PATH
    ```

    At the end of the install make sure that the OCM CLI, clusteradm, is in your path.

- helm - to deploy the kubestellar and kubeflex charts
- kubectl - to access the kubernetes clusters
- docker (or compatible docker engine that works with kind)

## For running the examples

- kind - to create a few small kubernetes clusters
- argocd - for the examples that use it

## For Building KubeStellar

- go version 1.20 or higher - to build kubestellar
- make - to build kubestellar and create the kubestellar image
- ko - to create the kubestellar image
