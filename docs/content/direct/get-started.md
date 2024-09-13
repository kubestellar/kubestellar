# Getting Started with KubeStellar

This page shows one concrete example of steps 2--7 from the [full Installation and Usage outline](user-guide-intro.md#the-full-story). This example produces a simple single-host system suitable for kicking the tires, using [kind](https://kind.sigs.k8s.io/) to create three new clusters to serve as your KubeFlex hosting cluster and two WECs. This page concludes with forwarding you to one example of the remaining steps.

  1. Setup
    1. Install software prerequisites
    1. Cleanup from previous runs
    1. Create the KubeFlex hosting cluster and Kubestellar core components
    1. Create and register two WECs.
  1. Exercise KubeStellar

## Setup

This is one way to produce a very simple system, suitable for study but not production usage. For general setup information, see [the full story](user-guide-intro.md#the-full-story).

### Install software prerequisites

The following command will check for the prerequisites that you will need for the later steps. See [the prerequisites doc](pre-reqs.md) for more details.

```shell
bash <(curl https://raw.githubusercontent.com/kubestellar/kubestellar/v{{ config.ks_latest_regular_release }}/hack/check_pre_req.sh) kflex ocm helm kubectl docker kind
```

This setup recipe uses [kind](https://kind.sigs.k8s.io/) to create three Kubernetes clusters on your machine.
Note that `kind` does not support three or more concurrent clusters unless you raise some limits as described in this `kind` "known issue": [Pod errors due to “too many open files”](https://kind.sigs.k8s.io/docs/user/known-issues/#pod-errors-due-to-too-many-open-files).

### Cleanup from previous runs

If you have run this recipe or any related recipe previously then
you will first want to remove any related debris. The following
commands tear down the state established by this recipe.

```shell
kind delete cluster --name kubeflex
kind delete cluster --name cluster1
kind delete cluster --name cluster2
kubectl config delete-context kind-kubeflex
kubectl config delete-context cluster1
kubectl config delete-context cluster2
```

### Set the Version appropriately as an environment variable

```shell
export KUBESTELLAR_VERSION={{ config.ks_latest_release }}
```

### Create a kind cluster to host KubeFlex

For convenience, a new local **Kind** cluster that satisfies the requirements for playing the role of KubeFlex hosting cluster can be created with the following command:

```shell
bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v{{ config.ks_latest_release }}/scripts/create-kind-cluster-with-SSL-passthrough.sh) --name kubeflex --port 9443
```

### Use Core Helm chart to initialize KubeFlex and create ITS and WDS

```shell
helm upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart \
    --version $KUBESTELLAR_VERSION \
    --set-json='ITSes=[{"name":"its1"}]' \
    --set-json='WDSes=[{"name":"wds1"}]'
```

That command will print some notes about how to get kubeconfig "contexts" named "its1" and "wds1" defined. Do that, because those contexts are used in the following.

### Create and register two workload execution cluster(s)

 {%
    include-markdown "example-wecs.md"
    heading-offset=2
 %}

## Exercise KubeStellar

Proceed to [Scenario 1 (multi-cluster workload deployment with kubectl) in the example scenarios](example-scenarios.md#scenario-1-multi-cluster-workload-deployment-with-kubectl) after defining the shell variables that characterize the setup done above. Following are setting for those variables, whose meanings are defined [at the start of the example scenarios document](example-scenarios.md#assumptions-and-variables).

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
