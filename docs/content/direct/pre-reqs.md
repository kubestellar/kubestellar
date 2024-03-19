# KubeStellar prerequisites

The following prerequisites are required.
You can use the [check-pre-req](#check-pre-requisites-for-kubestellar) script, to validate if all needed pre-requisites are installed. 

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

## Check pre-requisites for KubeStellar
The [check_pre_req](https://github.com/kubestellar/kubestellar/blob/main/hack/check_pre_req.sh) script offers a convenient way to check for the pre-requisites eeded for [KubeStellar](./pre-reqs.md) deployment and [use case scenarios](./examples.md).

The script checks for a pre-requisite presence in the path, by using the `which` command, and it can optionally provide version and path information for pre-requisites that are present, or installation information for missing pre-requisites.

We envision that this script could be useful for user-side debugging as well as for asserting the presence of pre-requisites in higher-level automation scripts.

The script accepts a list of optional flags and arguments.

**Supported flags:**

- `-A|--assert`: exits with error code 2 upon finding the fist missing pre-requisite
- `-L|--list`: prints a list of supported pre-requisites
- `-V|--verbose`: displays version and path information for installed pre-requisites or installation information for missing pre-requisites
- `-X`: enable `set -x` for debugging the script

**Supported arguments:**

The script accepts a list of specific pre-requisites to check, among the list of available ones:

```shell
$ check_pre_req.sh --list
argo brew docker go helm jq kflex kind ko kubectl make ocm yq
```

For example, list of pre-requisites required by KubeStellar can be checked with the command below (add the `-V` flag to get the version of each program and a suggestions on how to install missing pre-requisites):

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

In another example, a specific list of pre-requisites could be asserted by an higher-level script, while providing some installation information, with the command below (note that the script will terminate upon finding a missing pre-requisite):

```shell
$ check_pre_req.sh --assert --verbose helm argo docker kind
Checking KubeStellar pre-requisites:
✔ Helm
  version: version.BuildInfo{Version:"v3.14.0", GitCommit:"3fc9f4b2638e76f26739cd77c7017139be81d0ea", GitTreeState:"clean", GoVersion:"go1.21.5"}
     path: /usr/sbin/helm
X ArgoCD CLI
  how to install: https://argo-cd.readthedocs.io/en/stable/cli_installation/
```
