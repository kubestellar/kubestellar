# KubeStellar Quickstart Setup

This Quick Start outlines step 1 and shows a concrete example of steps 2--7 in the [Installation and Usage outline](usage-basics.md). In this example you will create three new `kind` clusters to serve as your KubeFlex hosting cluster and two WECs.

  1. Before you begin, prepare your system (get the software prerequisites)
  2. Create the KubeFlex hosting cluster and Kubestellar core components
  3. Create and register two WECs.

---
## Before You Begin


{%
    include-markdown "pre-reqs.md"
    rewrite-relative-urls=true
    start="<!-- begin software prerequisites -->"
    end="Additional Software For Running"
    heading-offset=1
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

### Delete debris from preivous trials

If you have run this quickstart or any related recipe previously then
you will first want to remove any related debris. The following
commands tear down the state established by this quickstart.

```shell
kind delete cluster --name kubeflex
kind delete cluster --name cluster1
kind delete cluster --name cluster2
kubectl config delete-context kind-kubeflex
kubectl config delete-context cluster1
kubectl config delete-context cluster2
```

## Set the Version appropriately as an environment variable

```shell
export KUBESTELLAR_VERSION=0.23.0
```

## Create a kind cluster to host KubeFlex

For convenience, a new local **Kind** cluster that satisfies the requirements for KubeStellar setup and that can be used to commission the quickstart workload can be created with the following command:

```shell
bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v0.23.0/scripts/create-kind-cluster-with-SSL-passthrough.sh) --name kubeflex --port 9443
```

## Use Core Helm chart to initialize KubeFlex and create ITS and WDS

```shell
helm upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart \
    --version $KUBESTELLAR_VERSION \
    --set-json='ITSes=[{"name":"its1"}]' \
    --set-json='WDSes=[{"name":"wds1"}]'
```

## Create and register two workload execution cluster(s)

 {%
    include-markdown "example-wecs.md"
    heading-offset=2
 %}

## Exercise KubeStellar

Proceed to exercise any of the [example scenarios](example-scenarios.md) after defining the shell variables that charaterize the setup done above. Following are setting for those variables.

```shell
host_context=kind-kubeflex
its_cp=its1
its_context=its1
wds_cp=wds1
wds_context=wds1
wec1_name=cluster1
wec2_name=cluster2
wec1_context=$wec1_name
wec2_context=$wec2_name
label_query_both=location-group=edge
label_query_one=name=cluster1
```
