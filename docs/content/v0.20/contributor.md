# KubeStellar Contributors

**NOTE**: work in progress... write more

## Check pre-requisites for KubeStellar

The [check_pre_req](../../../hack/check_pre_req.sh) script offers a convenient way to check for pre-requisites, including the ones needed for [KubeStellar](./pre-reqs.md) deployment and [use case scenarios](./examples.md).

The [check_pre_req](../../../hack/check_pre_req.sh) script check for a pre-requisite presence in the path, by using the `which` command, and it can optionally provide version and path information for pre-requisites that are present, or installation information for missing pre-requisites.

We envision that the [check_pre_req](../../../hack/check_pre_req.sh) script could be useful for user-side debugging as well as for asserting the presence of pre-requisites in higher-level automation scripts.

The [check_pre_req](../../../hack/check_pre_req.sh) script accepts a list of optional flags and arguments.

**Supported flags:**

- `-A|--assert`: exits with error code 2 upon finding the fist missing pre-requisite
- `-L|--list`: prints a list of supported pre-requisites
- `-V|--verbose`: displays version and path information for installed pre-requisites or installation information for missing pre-requisites
- `-X`: enable `set -x` for debugging the script

**Supported arguments:**

The [check_pre_req](../../../hack/check_pre_req.sh) script accepts a list of specific pre-requisites to check, among the list of available ones:

```shell
$ check_pre_req.sh --list
argo brew docker go helm jq kcp kflex kind ko kubectl make ocm yq
```

For example, the complete list of pre-requisites could be checked with the command below:

```shell
$ check_pre_req.sh
X ArgoCD CLI
✔ Home Brew
✔ Docker
✔ Go
✔ Helm
✔ jq
X kcp
✔ KubeFlex
✔ Kind
✔ KO
✔ kubectl
✔ GNU Make
✔ OCM CLI
✔ yq
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
