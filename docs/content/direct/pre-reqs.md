# KubeStellar Prerequisites

The following prerequisites are required.
You can use the [check-pre-req](#automated-check-of-prerequisites-for-kubestellar) script, to validate if all needed prerequisites are installed.


## Infrastructure (clusters)

Because of its multicluster architecture, KubeStellar requires that you have the necessary privileges and infrastructure access to create and/or configure the necessary Kubernetes clusters. These are the following; see [the architecture document](architecture.md) for more details.

- One cluster to serve as the [KubeFlex](https://github.com/kubestellar/kubeflex/) hosting cluster.
- Any additional Kubernetes clusters that are not created by KubeFlex but you will use as a WDS or ITS.
- Your WECs.

Our documentation has remarks about using the following sorts of clusters:

- **kind**
- **k3s**
- **openshift** 

<!-- begin software prerequisites -->
## Software Prerequisites: for Using KubeStellar

- **kubeflex** version 0.7.1 or higher.
    To install kubeflex go to [https://github.com/kubestellar/kubeflex/blob/main/docs/users.md#installation](https://github.com/kubestellar/kubeflex/blob/main/docs/users.md#installation). To upgrade from an existing installation,
follow [these instructions](https://github.com/kubestellar/kubeflex/blob/main/docs/users.md#upgrading-kubeflex). At the end of the install make sure that the kubeflex CLI, kflex, is in your `$PATH`.

- **OCM CLI (clusteradm)** version >= 0.7.
    To install OCM CLI use:

    ```shell
    curl -L https://raw.githubusercontent.com/open-cluster-management-io/clusteradm/main/install.sh | bash
    ```

    Note that the default installation of clusteradm will install in /usr/local/bin which will require root access. If you prefer to avoid root, you can specify an alternative installation location using the INSTALL_DIR environment variable, as follows:

    ```shell
    mkdir -p ocm
    export INSTALL_DIR="$PWD/ocm"
    curl -L https://raw.githubusercontent.com/open-cluster-management-io/clusteradm/main/install.sh | bash
    export PATH=$PWD/ocm:$PATH
    ```

    At the end of the install make sure that the OCM CLI, clusteradm, is in your `$PATH`.

- **helm** version >= 3 - to deploy the Kubestellar and kubeflex charts
- [**kubectl**](https://kubernetes.io/docs/tasks/tools/) version >= 1.27 - to access the kubernetes clusters

## Additional Software for the Getting Started setup

- [**kind**](https://kind.sigs.k8s.io/) version >= 0.20
- **docker** (or compatible docker engine that works with kind) (client version >= 20)

## Additional Software for monitoring

The setup in `montoring/` additional uses the following.

- [`yq`](https://github.com/mikefarah/yq) (also available from [Homebrew](https://formulae.brew.sh/formula/yq)) version >= 1.5

## Additional Software For Running the Examples

- [**argocd**](https://argo-cd.readthedocs.io/en/stable/getting_started/) version >= 2 - for the examples that use it

## Additional Software For Building KubeStellar from Source and Testing

- [**go**](https://go.dev/doc/install) version 1.21 or higher - to build Kubestellar
- [**GNU make**](https://www.gnu.org/software/make/) version >= 3.5 - to build Kubestellar and create the Kubestellar container images
- [**ko**](https://ko.build/install/) version >= 0.15 - to create some of the Kubestellar container images
- **docker** (or equivalent that implements `docker buildx`) (client version >= 20) - to create other KubeStellar container images


To build and _**test**_ KubeStellar properly, you will also need

- [**kind**](https://kind.sigs.k8s.io/) version >= 0.20
- [**OCP**](https://docs.openshift.com/container-platform/4.13/installing/index.html), if you are testing a scenario involving OCP
- [**ginkgo**](https://onsi.github.io/ginkgo/), if you will run the ginkgo-based end-to-end test
- [`yq`](https://github.com/mikefarah/yq) (also available from [Homebrew](https://formulae.brew.sh/formula/yq)) version >= 4 - for running tests

<!-- start tag for check script  include -->

## Automated Check of Prerequisites for KubeStellar
The [check_pre_req](https://github.com/kubestellar/kubestellar/blob/main/hack/check_pre_req.sh) script offers a convenient way to check for the prerequisites needed for [KubeStellar](./pre-reqs.md) deployment and [use](./example-scenarios.md).

This script is self-contained, so it is suitable for "curl-to-bash" style usage. The latest development version is at [https://raw.githubusercontent.com/kubestellar/kubestellar/refs/heads/main/hack/check_pre_req.sh](https://raw.githubusercontent.com/kubestellar/kubestellar/refs/heads/main/hack/check_pre_req.sh). To check the prerequisites for using a particular release of KubeStellar, you will want to use the script from that release.

The script checks for a prerequisite presence in the `$PATH`, by using the `which` command, and it can optionally provide version and path information for prerequisites that are present, or installation information for missing prerequisites.

We envision that this script could be useful for user-side debugging as well as for asserting the presence of prerequisites in higher-level automation scripts.

The script accepts a list of optional flags and arguments.

### **Supported flags:**

- `-A|--assert`: exits with error code 2 upon finding the first missing prerequisite
- `-L|--list`: prints a list of supported prerequisites
- `-V|--verbose`: displays version and path information for installed prerequisites or installation information for missing prerequisites
- `-X`: enable `set -x` for debugging the script

### **Supported arguments:**

The script accepts a list of specific prerequisites to check, among the list of available ones:

```shell
$ check_pre_req.sh --list
argo brew docker go helm jq kflex kind ko kubectl make ocm yq
```

### Examples
For example, list of prerequisites required by KubeStellar can be checked with the command below (add the `-V` flag to get the version of each program and a suggestions on how to install missing prerequisites):

```shell
$ hack/check_pre_req.sh
Checking pre-requisites for using KubeStellar:
✔ Docker (Docker version 27.2.1-rd, build cc0ee3e)
✔ kubectl (v1.29.2)
✔ KubeFlex (Kubeflex version: v0.6.3.672cc8a 2024-09-23T16:15:47Z)
✔ OCM CLI (:v0.9.0-0-g56e1fc8)
✔ Helm (v3.16.1)
Checking additional pre-requisites for running the examples:
✔ Kind (kind v0.22.0 go1.22.0 darwin/arm64)
✔ ArgoCD CLI (v2.10.1+a79e0ea)
Checking pre-requisites for building KubeStellar:
✔ GNU Make (GNU Make 3.81)
✔ Go (go version go1.23.2 darwin/arm64)
✔ KO (0.16.0)
```

<!-- end tag for check-prereq script -->

In another example, a specific list of prerequisites could be asserted by a higher-level script, while providing some installation information, with the command below (note that the script will terminate upon finding a missing prerequisite):

```shell
$ check_pre_req.sh --assert --verbose helm argo docker kind
Checking KubeStellar pre-requisites:
✔ Helm
  version (unstructured): version.BuildInfo{Version:"v3.14.0", GitCommit:"3fc9f4b2638e76f26739cd77c7017139be81d0ea", GitTreeState:"clean", GoVersion:"go1.21.5"}
     path: /usr/sbin/helm
X ArgoCD CLI
  how to install: https://argo-cd.readthedocs.io/en/stable/cli_installation/; get at least version v2
```
