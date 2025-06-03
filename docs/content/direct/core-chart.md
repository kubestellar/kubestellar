# KubeStellar Core chart usage

## Table of Contents
- [Overview](#overview)
- [Pre-requisites](#pre-requisites)
- [KubeStellar Core Chart values](#kubestellar-core-chart-values)
- [KubeStellar Core Chart usage](#kubestellar-core-chart-usage)
- [Step-by-Step Installation Example](#step-by-step-installation-example)
- [Kubeconfig files and contexts for Control Planes](#kubeconfig-files-and-contexts-for-control-planes)
- [Argo CD integration](#argo-cd-integration)
- [Uninstalling the KubeStellar Core chart](#uninstalling-the-kubestellar-core-chart)

## Overview

This documents explains how to use KubeStellar Core chart to do three
of the 11 installation and usage steps; please see [the
full outline](user-guide-intro.md#the-full-story) for generalities and [Getting Started](get-started.md) for an example of usage.

This Helm chart can do any subset of the following things.

- Initialize a pre-existing cluster to serve as the KubeFlex hosting cluster.
- Create some ITSes.
- Create some WDSes.

The information provided is specific for the following release:

```shell
export KUBESTELLAR_VERSION={{ config.ks_latest_release }}
```

## Pre-requisites

To install the Helm chart the only requirement is [Helm](https://helm.sh/).
However, additional executables may be required to create/manage the cluster(s) (_e.g._, Kind and kubectl),
to join Workload Execution Clusters (WECs) (_e.g._, clusteradm),
and to interact with Control Planes (_e.g._, kubectl), _etc_.
For such purpose, a full list of executable that may be required can be found [here](./pre-reqs.md).

The setup of KubeStellar via the Core chart requires the existence of a KubeFlex hosting cluster.

While not a complete list of supported hosting clusters, here we discuss how to use KubeStellar in:

1. A local **Kind** or **k3s** cluster with an ingress with SSL passthrough and a mapping to host port 9443

    This option is particularly useful for first time users or users that would like to have a local deployment.

    It is important to note that, when the hosting cluster was created by **kind** or **k3s** and its Ingress domain name is left to default to localtest.me, then the name of the container running hosting cluster must be also be referenced during the Helm chart installation by setting `--set "kubeflex-operator.hostContainer=<control-plane-container-name>"`.
    The `<control-plane-container-name>` is the name of the container in which kind or k3d is running the relevant control plane. One may use `docker ps` to find the `<control-plane-container-name>`.

    If a host port number different from the expected 9443 is used for the Kind cluster, then the same port number must be specified during the chart installation by adding the following argument `--set "kubeflex-operator.externalPort=<port>"`.

    By default the KubeStellar Core chart uses a test domain `localtest.me`, which is OK for testing on a single host machine. However, for scenarios that span more than one machine, it is necessary to set `--set "kubeflex-operator.domain=<domain>"` to a more appropriate `<domain>` that can be reached from Workload Execution Clusters (WECs).

    For convenience, a new local **Kind** cluster that satisfies the requirements for KubeStellar setup
    and that can be used to exercises the [examples](./examples.md) can be created with the following command:

    ```shell
    bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v$KUBESTELLAR_VERSION/scripts/create-kind-cluster-with-SSL-passthrough.sh) --name kubeflex --port 9443
    ```

    Alternatively, a new local **k3s** cluster that satisfies the requirements for KubeStellar setup
    and that can be used to exercises the [examples](./examples.md) can be created with the following command:

    ```shell
    bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v$KUBESTELLAR_VERSION/scripts/create-k3s-cluster-with-SSL-passthrough.sh) --port 9443
    ```

2. An **OpenShift** cluster

    When using this option, one is required to explicitly set the `isOpenShift` variable to `true` by including `--set "kubeflex-operator.isOpenShift=true"` in the Helm chart installation command.

## KubeStellar Core Chart values

The KubeStellar chart makes available to the user several values that may be used to customize its installation into an existing cluster:

```yaml
# Control controller log verbosity
# The "default" verbosity value will be used for all controllers unless a specific controller verbosity override is specified
verbosity:
  default: 2
  # Specific controller verbosity overrides:
  # kubestellar: 6 (controller-manager)
  # clusteradm: 6
  # transport: 6

# KubeFlex override values
kubeflex-operator:
  install: true # enable/disable the installation of KubeFlex by the chart (default: true)
  installPostgreSQL: true # enable/disable the installation of the appropriate version of PostgreSQL required by KubeFlex (default: true)
  isOpenShift: false # set this variable to true when installing the chart in an OpenShift cluster (default: false)
  # Kind cluster specific settings:
  domain: localtest.me # used to define the DNS domain name used from outside the KubeFlex hosting cluster to reach that cluster's Ingress endpoint (default: localtest.me)
  externalPort: 9443 # used to set the port to access the Control Planes API (default: 9443)
  hostContainer: kubeflex-control-plane # used to set the name of the container that runs the KubeFlex hosting cluster (default: kubeflex-control-plane, which corresponds to a Kind cluster with name kubeflex)

# Determine if the Post Create Hooks should be installed by the chart
InstallPCHs: true

# List the Inventory and Transport Spaces (ITSes) to be created by the chart
# Each ITS consists of:
# - a mandatory unique name
# - an optional type, which could be host, vcluster, or external (default to vcluster, if not specified)
# - an optional install_clusteradm flag, which could be true  or false  (default to true) to enable/disable the installation of OCM in the control plane
# - an optional bootstrapSecret secion to be used for Control Plabes of type external (more details below)
ITSes: # ==> installs ocm (optional) + ocm-status-addon

# List the Workload Description Spaces (WDSes) to be created by the chart
# Each WDS consists of a mandatory unique name and several optional parameters:
# - type: host or k8s (default to k8s, if not specified)
# - APIGroups: a comma separated list of APIGroups
# - ITSName: the name of the ITS control plane to be used by the WDS. Note that the ITSName MUST be specified if more than one ITS exists.
WDSes: # ==> installs kubestellar + ocm-transport-plugin
```

The first section of the `values.yaml` file refers to parameters that are specific to the KubeFlex installation, see [here](https://github.com/kubestellar/kubeflex/blob/main/docs/users.md) for more information.

In particular:
- `kubeflex-operator.install` accepts a boolean value to enable/disable the installation of KubeFlex into the cluster by the chart
- `kubeflex-operator.isOpenShift` must be set to true by the user when installing the chart into a OpenShift cluster

By default, the chart will install the KubeFlex and its PostgreSQL dependency.

The second section allows a user of the chart to determine if Post Create Hooks (PCHes) needed for creating ITSes and WDSes control planes should be installed by the chart. By default `InstallPCHs` is set to `true` to enable the installation of the PCHes, however one may want to set this value to `false` when installing multiple copies of the chart to avoid conflicts. A single copy of the PCHes is required and allowed per cluster.

The third section of the `values.yaml` file allows one to create a list of Inventory and Transport Spaces (ITSes). By default, this list is empty and no ITS will be created by the chart. A list of ITSes can be specified using the following format:

```yaml
ITSes: # all the CPs in this list will execute the its.yaml PCH
  - name: <its1>          # mandatory name of the control plane
    type: <vcluster|host|external> # optional type of control plane: host, vcluster, or external (default to vcluster, if not specified)
    install_clusteradm: true|false  # optional flag to enable/disable the installation of OCM in the control plane (default to true, if not specified)
    bootstrapSecret: # this section is ignored unless type is "external"
      name: <secret-name> # default: "<control-plane-name>-bootstrap"
      namespace: <secret-namespace> # default: Helm chart installation namespace
      key: <key-name> # default: "kubeconfig-incluster"
  - name: <its2>          # mandatory name of the control plane
    type: <vcluster|host|external> # optional type of control plane: host, vcluster, or external (default to vcluster, if not specified)
    install_clusteradm: true|false  # optional flag to enable/disable the installation of OCM in the control plane (default to true, if not specified)
    bootstrapSecret: # this section is ignored unless type is "external"
      name: <secret-name> # default: "<control-plane-name>-bootstrap"
      namespace: <secret-namespace> # default: Helm chart installation namespace
      key: <key-name> # default: "kubeconfig-incluster"
  ...
```

where `name` must specify a name unique among all the control planes in that KubeFlex deployment, the optional `type` can be vcluster (default), host, or external, see [here](https://github.com/kubestellar/kubeflex/blob/main/docs/users.md) for more information, and the optional `install_clusteradm`can be either true (default) or false to enable or disable the installation of OCM in the control plane.

When the ITS `type` is `external`, the `bootstrapSecret` sub-section can be used to indicate the bootstrap secret used by KubeFlex to connect to the external cluster. Specifically, it can be used to specify any combination of (a) the name of the secret, (b) the namespace containing the secret, and (c) the name of the key containg the kubeconfig of the external cluster if they need to be different from their default value.

If the secret was created using the [create-external-bootstrap-secret.sh](https://github.com/kubestellar/kubestellar/tree/v{{ config.ks_latest_release }}/scripts/create-external-bootstrap-secret.sh) script and the value passed to the argument `--controlplane` matches the name of the Control Plane specified by the Helm chart, then the sub-section `bootstrapSecret` is not required because all default values will identify the bootstrap secret created by the script. More specifically, if an external kind cluster was created with the command `kind create cluster --name its1` and the `create-external-bootstrap-secret.sh --controlplane its1 --verbose` command was used to create the bootstrap secret, then it would be enough to inform the Helm chart with `--set-json='ITSes=[{"name":"its1","type":"external"}]'`.

The fourth section of the `values.yaml` file allows one to create a list of Workload Description Spaces (WDSes). By default, this list is empty and no WDS will be created by the chart. A list of WDSes can be specified using the following format:

```yaml
WDSes: # all the CPs in this list will execute the wds.yaml PCH
  - name: <wds1>     # mandatory name of the control plane
    type: <host|k8s> # optional type of control plane host or k8s (default to k8s, if not specified)
    APIGroups: ""    # optional string holding a comma-separated list of APIGroups
    ITSName: <its1>  # optional name of the ITS control plane, this MUST be specified if more than one ITS exists at the moment the WDS PCH starts
  - name: <wds2>     # mandatory name of the control plane
    type: <host|k8s> # optional type of control plane host or k8s (default to k8s, if not specified)
    APIGroups: ""    # optional string holding a comma-separated list of APIGroups
    ITSName: <its2>  # optional name of the ITS control plane, this MUST be specified if more than one ITS exists at the moment the WDS PCH starts
  ...
```

where `name` must specify a name unique among all the control planes in that KubeFlex deployment (note that this must be unique among both ITSes and WDSes), the optional `type` can be either k8s (default) or host, see [here](https://github.com/kubestellar/kubeflex/blob/main/docs/users.md) for more information, the optional `APIGroups` provides a list of APIGroups, see [here](https://docs.kubestellar.io/release-{{ config.ks_latest_release }}/direct/examples/#scenario-2-using-the-hosting-cluster-as-wds-to-deploy-a-custom-resource) for more information, and `ITSName` specify the ITS connected to the new WDS being created (this parameter MUST be specified if more that one ITS exists in the cluster, if no value is specified and only one ITS exists in the cluster, then it will be automatically selected).

## KubeStellar Core Chart usage

The local copy of the core chart can be installed in an existing cluster using the commands:

```shell
helm dependency update core-chart
helm upgrade --install ks-core core-chart
```

Alternatively, a specific version of the KubeStellar core chart can be simply installed in an existing cluster using the following command:

```shell
helm upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart --version $KUBESTELLAR_VERSION
```

Either if of the previous way of installing KubeStellar chart will install KubeFlex and the Post Create Hooks, but no Control Planes.
Please remember to add `--set "kubeflex-operator.isOpenShift=true"`, when installing into an OpenShift cluster.

User defined control planes can be added using additional value files of `--set` arguments, _e.g._:

- add a single ITS named its1 of default vcluster type: `--set-json='ITSes=[{"name":"its1"}]'`
- add two ITSes named its1 and its2 of of type vcluster and host, respectively: `--set-json='ITSes=[{"name":"its1"},{"name":"its2","type":"host"}]'`
- add a single WDS named wds1 of default k8s type connected to the one and only ITS: `--set-json='WDSes=[{"name":"wds1"}]'`

A KubeStellar Core installation that is consistent with [Getting Started](get-started.md) and and supports [the example scenarios](./example-scenarios.md) could be achieved with the following command:

```shell
helm upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart --version "$KUBESTELLAR_VERSION" \
  --set-json ITSes='[{"name":"its1"}]' \
  --set-json WDSes='[{"name":"wds1"}]'
```

## Step-by-Step Installation Example
For a detailed step-by-step installation guide with expected outputs, see [Step-by-Step Installation Guide](core-chart-step-by-step-installation.md).

The core chart also supports the use of a pre-existing cluster (or any space, really) as an ITS. A specific application is to connect to existing OCM clusters. As an example, create a first local kind cluster with OCM installed in it:

```shell
kind create cluster --name ext1

clusteradm init
```

Then, create a second kind cluster suitable for KubeStellar installation and create a bootstrap secret in the new cluster with the kubeconfig information of the `ext1` cluster:

```shell
bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v$KUBESTELLAR_VERSION/scripts/create-kind-cluster-with-SSL-passthrough.sh) --name kubeflex --port 9443

bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v$KUBESTELLAR_VERSION/scripts/create-external-bootstrap-secret.sh) --controlplane its1 --source-context kind-ext1 --address https://ext1-control-plane:6443 --verbose
```

Note that the last command above creates a secret named `its1-bootstrap` in the Helm chart installation namespace of the `kind-kubeflex` cluster.

The `--address` URL needs to be one that the KubeFlex controller can use to open a connection to the external cluster's Kubernetes apiserver(s). In this example, the external cluster is a kind cluster with one kube-apiserver and it listens on port 6443 in its node's network namespace. This example relies on the DNS resolver in Docker networking to map the domain name `ext1-control-plane` to the Docker network address of the container of that same name.

Finally, install the core chart using the `ext1` cluster as ITS:

```shell
helm upgrade --install core-chart oci://ghcr.io/kubestellar/kubestellar/core-chart --version $KUBESTELLAR_VERSION \
  --set-json='ITSes=[{"name":"its1","type":"external","install_clusteradm":false}]' \
  --set-json='WDSes=[{"name":"wds1"}]'
```

Note that by default, the `its1` Control Plane of type `external` will look for a secret named `its1-bootstrap` in the Helm chart installation namespace. Additionally the `"install_clusteradm":false` value is specified to avoid reinstalling OCM in the `ext1` cluster.

After the initial installation is completed, there are two main ways to install additional control planes (_e.g._, create a second `wds2` WDS):

1. Upgrade the initial chart. This choice requires to relist the existing control planes, which would otherwise be deleted:

    ```shell
    helm upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart --version $KUBESTELLAR_VERSION \
      --set-json='ITSes=[{"name":"its1"}]' \
      --set-json='WDSes=[{"name":"wds1"},{"name":"wds2"}]'
    ```

2. Install a new chart with a different name. This choice does not requires to relist the existing control planes, but requires to disable the reinstallation of KubeFlex and PCHes:

    ```shell
    helm upgrade --install add-wds2 oci://ghcr.io/kubestellar/kubestellar/core-chart --version $KUBESTELLAR_VERSION \
      --set='kubeflex-operator.install=false,InstallPCHs=false' \
      --set-json='WDSes=[{"name":"wds2"}]'
    ```

## Kubeconfig files and contexts for Control Planes

It is convenient to use one kubeconfig file that has a context for
each of your control planes. That can be done in two ways, one using
the `kflex` CLI and one not.

1. Using `kflex` CLI

    The following commands will add a context, named after the given
    control plane, to your current kubeconfig file and make that the
    current context. The deletion is to remove an older vintage if it
    is present.

    ```shell
    kubectl config delete-context $cpname
    kflex ctx $cpname
    ```

    The `kflex ctx` command is unable to create a new context if the
    current context does not access the KubeFlex hosting cluster AND
    the KubeFlex kubeconfig extension remembering that context's name
    is not set; see the KubeFlex user guide for your release of
    KubeFlex for more information.

    To automatically add all Control Planes as contexts of the current kubeconfig, one can use the convenience script below:

    ```shell
    echo "Getting the kubeconfig of all Control Planes..."
    for cpname in `kubectl get controlplane -o name`; do
      cpname=${cpname##*/}
      echo "Getting the kubeconfig of Control Planes \"$cpname\"..."
      kflex ctx $cpname
    done
    ```

    After doing the above context switching you may wish to use `kflex ctx` to switch back to the hosting cluster context.

    Afterwards the content of a Control Plane `$cpname` can be accessed by specifying its context:

    ```shell
    kubectl --context "$cpname" ...
    ```

2. Using plain `kubectl` commands

    The following commands can be used to create a fresh kubeconfig file for each of the KubeFlex Control Planes in the hosting cluster:

    ```shell
    echo "Creating a kubeconfig for each KubeFlex Control Plane:"
    for cpname in `kubectl get controlplane -o name`; do
      cpname=${cpname##*/}
      echo "Getting the kubeconfig of \"$cpname\" ==> \"kubeconfig-$cpname\"..."
      if [[ "$(kubectl get controlplane $cpname -o=jsonpath='{.spec.type}')" == "host" ]] ; then
        kubectl config view --minify --flatten > "kubeconfig-$cpname"
      else
        kubectl get secret $(kubectl get controlplane $cpname -o=jsonpath='{.status.secretRef.name}') \
          -n $(kubectl get controlplane $cpname -o=jsonpath='{.status.secretRef.namespace}') \
          -o=jsonpath="{.data.$(kubectl get controlplane $cpname -o=jsonpath='{.status.secretRef.key}')}" \
          | base64 -d > "kubeconfig-$cpname"
      fi
      curname=$(kubectl --kubeconfig "kubeconfig-$cpname" config current-context)
      if [ "$curname" != "$cpname" ]
      then kubectl --kubeconfig "kubeconfig-$cpname" config rename-context "$curname" $cpname
      fi
    done
    ```

    The code above puts the kubeconfig for a control plane `$cpname` into a file name `kubeconfig-$cpname` in the local folder.
    The current context will be renamed to `$cpname`, if it does not already have that name (which it will for control planes of type "k8s", for example).

    With the above kubeconfig files in place, the control plane named `$cpname` can be accessed as follows.

    ```shell
    kubectl --kubeconfig "kubeconfig-$cpname" ...
    ```

    The individual kubeconfigs can also be merged as contexts of the current `~/.kube/config` with the following commands:

    ```shell
    echo "Merging the Control Planes kubeconfigs into ~/.kube/config ..."
    cp ~/.kube/config ~/.kube/config.bak
    KUBECONFIG=~/.kube/config:$(find . -maxdepth 1 -type f -name 'kubeconfig-*' | tr '\n' ':') kubectl config view --flatten > ~/.kube/kubeconfig-merged
    mv ~/.kube/kubeconfig-merged ~/.kube/config
    ```

    Afterwards the content of a Control Plane `$cpname` can be accessed by specifying its context:

    ```shell
    kubectl --context "$cpname" ...
    ```

3. Using `import-cp-contexts.sh` script

    The following convenience command can also be used to import all the KubeFlex Control Planes in the current hosting cluster as contexts of the current kubeconfig. The script involved requires that you have [`yq`](https://github.com/mikefarah/yq) (also available from [Homebrew](https://formulae.brew.sh/formula/yq)) installed.

    ```shell
    bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v$KUBESTELLAR_VERSION/scripts/import-cp-contexts.sh) --merge
    ```

    The script above only requires `kubectl` and `yq`.

    The script accepts the following arguments:

    - `--kubeconfig <filename>` specify the kubeconfig of the hosting cluster where the KubeFlex Control Planes are located. Note that this argument will override the content of the `KUBECONFIG` environment variable
    - `--context <name>` specify a context of the current kubeconfig where to look for KubeFlex Control Planes. If this argument is not specified, then all contexts will be searched.
    - `--names|-n <name1>,<name2>,..` comma separated list of KubeFlex Control Planes names to import. If this argument is not specified then *all* available KubeFlex Control Planes will be imported.
    - `--replace-localhost|-r <host>` replaces server addresses "127.0.0.1" with a desired `<host>`. This parameter is useful for making KubeFlex Control Planes of type `host` accessible from outside the machine hosting the cluster.
    - `--merge|-m` merge the kubeconfig with the contexts of the control planes with the existing cluster kubeconfig. If this flag is not specified, then only the kubeconfig with the contexts of the KubeFlex Control Planes will be produced.
    - `--output|-o <filename>|-` specify a kubeconfig file to save the kubeconfig to. Use `-` for stdout. If this argument is not provided, then the kubeconfig will be saved to the input specified kubeconfig, if provided, or to `~/.kube/config`.
    - `--silent|-s` quiet mode, do not print information. This may be useful when using `-o -`.
    - `-X` enable verbose execution of the script for debugging

## Argo CD integration

KubeStellar Core Helm chart allows to deploy ArgoCD in the KubeFlex hosting cluster, register every WDS as a target cluster in Argo CD, and create Argo CD applications as specified by chart values, as explained [here](core-chart-argocd.md).

## Uninstalling the KubeStellar Core chart

The chart can be uninstalled using the command:

```shell
helm uninstall ks-core
```

This will remove KubeFlex, PostgreSQL, Post Create Hooks (PCHes), and all KubeFlex Control Planes (_i.e._, ITSes and WDSes) that were created by the chart.

Additionally, if a **Kind** cluster was created with the provide script, it can be deleted with the command:

```shell
kind delete cluster --name kubeflex
```

Alternatively, if a **k3s** cluster was created with the provide script, it can be deleted with the command:

```shell
/usr/local/bin/k3s-uninstall.sh
```