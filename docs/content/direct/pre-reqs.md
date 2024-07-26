# KubeStellar Prerequisites

The following prerequisites are required.
You can use the [check-pre-req](#automated-check-of-pre-requisites-for-kubestellar) script, to validate if all needed prerequisites are installed.


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

- **kubeflex** version 0.6.1 or higher
    To install kubeflex go to [https://github.com/kubestellar/kubeflex/blob/main/docs/users.md#installation](https://github.com/kubestellar/kubeflex/blob/main/docs/users.md#installation). To upgrade from an existing installation,
follow [these instructions](https://github.com/kubestellar/kubeflex/blob/main/docs/users.md#upgrading-kubeflex). At the end of the install make sure that the kubeflex CLI, kflex, is in your `$PATH`.

- **OCM CLI (clusteradm)**
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

- **helm** - to deploy the Kubestellar and kubeflex charts
- [**kubectl**](https://kubernetes.io/docs/tasks/tools/) - to access the kubernetes clusters

## Additional Software for the Getting Started setup

- [**kind**](https://kind.sigs.k8s.io/)
- **docker** (or compatible docker engine that works with kind)

## Additional Software For Running the Examples

- [**argocd**](https://argo-cd.readthedocs.io/en/stable/getting_started/) - for the examples that use it

## Additional Software For Building KubeStellar from Source

- [**go**](https://go.dev/doc/install) version 1.21 or higher - to build Kubestellar
- [**make**](https://www.gnu.org/software/make/) - to build Kubestellar and create the Kubestellar container images
- [**ko**](https://ko.build/install/) - to create some of the Kubestellar container images
- **docker** (or equivalent that implements `docker buildx`) - to create other KubeStellar container images


To build and _**test**_ KubeStellar properly, you will also need

- [**kind**](https://kind.sigs.k8s.io/)
- [**OCP**](https://docs.openshift.com/container-platform/4.13/installing/index.html)

<!-- start tag for check script  include -->

## Automated Check of Prerequisites for KubeStellar
The [check_pre_req](https://github.com/kubestellar/kubestellar/blob/main/hack/check_pre_req.sh) script offers a convenient way to check for the prerequisites needed for [KubeStellar](./pre-reqs.md) deployment and [use](./example-scenarios.md).

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

<!-- end tag for check-prereq script -->

In another example, a specific list of prerequisites could be asserted by a higher-level script, while providing some installation information, with the command below (note that the script will terminate upon finding a missing prerequisite):

```shell
$ check_pre_req.sh --assert --verbose helm argo docker kind
Checking KubeStellar pre-requisites:
✔ Helm
  version: version.BuildInfo{Version:"v3.14.0", GitCommit:"3fc9f4b2638e76f26739cd77c7017139be81d0ea", GitTreeState:"clean", GoVersion:"go1.21.5"}
     path: /usr/sbin/helm
X ArgoCD CLI
  how to install: https://argo-cd.readthedocs.io/en/stable/cli_installation/
```
