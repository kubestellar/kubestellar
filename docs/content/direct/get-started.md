# KubeStellar Quickstart Setup

This Quick Start is based on Scenario 1 of our [examples](examples.md).
In a nutshell, you will:

  1. Before you begin, prepare your system (get the prerequisites)
  2. Create the Kubestellar core components on a cluster
  3. Commission a workload to a WEC

---
## Before You Begin


{%
    include-markdown "pre-reqs.md"
    rewrite-relative-urls=true
    end="Additional Software For Running"
    heading-offset=2
%}
---
{%
    include-markdown "pre-reqs.md"
    rewrite-relative-urls=true
    heading-offset=2
    start="<!-- start tag for check script  include -->"
    end="<!-- end tag for check-prereq script -->"
%}

---

## Create the KubeStellar Core components

Use our helm chart to set up the main core and establish its initial state using our helm chart:

### Set the Version appropriately as an environment variable

```shell
export KUBESTELLAR_VERSION=0.23.0
export OCM_TRANSPORT_PLUGIN=0.1.11
```
### Use the Helm chart  to deploy the KubeStellar Core to a Kind, K3s, or OpenShift cluster:
!!! tip "Pick the cluster configuration which applies to your system:"
    === "Using Kind"

        For convenience, a new local **Kind** cluster that satisfies the requirements for KubeStellar setup and that can be used to commission the quickstart workload can be created with the following command:

        ```shell
        bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v${KUBESTELLAR_VERSION}/scripts/create-kind-cluster-with-SSL-passthrough.sh) --name   kubeflex --port 9443
        ```
        After the cluster is created, deploy the Kubestellar Core installation on it with the helm chart command

        ```shell
        helm upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart --version $KUBESTELLAR_VERSION \
        --set-json='ITSes=[{"name":"its1"}]' \
        --set-json='WDSes=[{"name":"wds1"}]'
        ```

    === "Using K3S"

        A new local **k3s** cluster that satisfies the requirements for KubeStellar setup and that can be used to commission the quickstart workload can be created with the following command:

        ```shell
        bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v0.23.0-alpha.4/scripts/create-k3s-cluster-with-SSL-passthrough.sh) --port 9443
        ```
        After the cluster is created, deploy the Kubestellar Core installation on it with the helm chart command

        ```shell
        helm upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart --version $KUBESTELLAR_VERSION \
        --set-json='ITSes=[{"name":"its1"}]' \
        --set-json='WDSes=[{"name":"wds1"}]'
        ```

    === "Using OpenShift"

        When using this option, one is required to explicitly set the `isOpenShift` variable to `true` by including `--set "kubeflex-operator.isOpenShift=true"` in the Helm chart installation command.

        After the cluster is created, deploy the Kubestellar Core installation on it with the helm chart command

        ```shell
        helm upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart --version $KUBESTELLAR_VERSION \
          --set "kubeflex-operator.isOpenShift=true" \ 
          --set-json='ITSes=[{"name":"its1"}]' \
          --set-json='WDSes=[{"name":"wds1"}]'
        ```

Once you have done this, you should have the KubeStellar core components plus the required workload definition space and inventory and transport space control planes running on your cluster.

---

## Define, Bind and Commission a workload on a WEC

### Set up and define the workload execution cluster(s)

 {%
    include-markdown "example-wecs.md"
    heading-offset=2
 %}
  
### Bind and Commission the workload

 {% 
    include-markdown "examples.md"
    heading-offset=2
    start="<!-- Start for app commissioning in quickstart -->"
    end="<!-- End for app commissioning in quickstart -->"
 %}

