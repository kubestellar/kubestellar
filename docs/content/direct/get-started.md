# KubeStellar Quickstart Setup

This Quick Start outlines step 1, shows a concrete example of steps 2--7 in the [Installation and Usage outline](user-guide-intro.md), and forwards you to one example of the remaining steps. In this example you will create three new `kind` clusters to serve as your KubeFlex hosting cluster and two WECs.

  1. Install software prerequisites
  1. Cleanup from previous runs
  1. Create the KubeFlex hosting cluster and Kubestellar core components
  1. Create and register two WECs.
  1. Use KubeStellar to distribute a Deployment object to the two WECs.

---
## Install software prerequisites

The following command will check for the prerequisites that you will need for this quickstart. See [the prerequisites doc](pre-reqs.md) for more details.

```shell
bash <(curl https://raw.githubusercontent.com/kubestellar/kubestellar/v{{ config.ks_latest_regular_release }}/hack/check_pre_req.sh) kflex ocm helm kubectl docker kind
```

This quickstart uses [kind](https://kind.sigs.k8s.io/) to create three Kubernetes cluster on your machine.
Note that `kind` does not support three or more concurrent clusters unless you raise some limits as described in this `kind` "known issue": [Pod errors due to “too many open files”](https://kind.sigs.k8s.io/docs/user/known-issues/#pod-errors-due-to-too-many-open-files).

### Cleanup from previous runs

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

Proceed to Scenario 1 (multi-cluster workload deployment with kubectl) in [the example scenarios](example-scenarios.md) after defining the shell variables that characterize the setup done above. Following are setting for those variables, whose meanings are defined at the start of the example scenarios document.

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
