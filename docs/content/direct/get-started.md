# Getting Started with KubeStellar

This page shows one concrete example of steps 2--7 from the [full Installation and Usage outline](user-guide-intro.md#the-full-story). This example produces a simple single-host system that is suitable for kicking the tires, using [kind](https://kind.sigs.k8s.io/) to create three new clusters to serve as your KubeFlex hosting cluster and two WECs. This page concludes with forwarding you to one example of the remaining steps. For general setup information, see [the full story](user-guide-intro.md#the-full-story).

This shows just one particular choice for these initial steps, intended only to produce a system suitable for initial study. This is not what you would do for a production system. Remember that KubeStellar supports more sophisticated configurations including the following.

- Multiple Inventory and Transport Spaces (ITS)
- Multiple Workload Definition Spaces (WDS)
- Dynamic addition and removal of Workload Execution Clusters (WECs)
- Various deployment patterns to suit different organizational needs

This page is primarily organized as a sequence of steps. There is also a script that covers a subset of the steps.

## Quick Start Using the Automated Script

When reading through the steps in [Setup](#setup), after you have [established the software prerequisites](#install-software-prerequisites) (including making `kind` able to create _three_ clusters) you can then use our automated installation script to do the remaining [Setup](#setup) steps, from _checking_ the software prerequisites through creating the example WECs.

```shell
bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/refs/heads/main/scripts/create-kubestellar-demo-env.sh)
```

If successful, the script will output the variable definitions that you would use when proceeding to the example scenarios. After successfully running the script, proceed to the [Exercise KubeStellar](#exercise-kubestellar) section below.

This script does the same things as described in [Setup](#setup) but with maximum concurrency. The concurrency is there to enable the script to complete faster. This makes the script more complicated than what is presented below.

## Detailed Step walkthrough

The following steps show one concrete example of steps 2--7 from the [full Installation and Usage outline](user-guide-intro.md#the-full-story).

1. [Setup](#setup)
    1. Install software prerequisites
    1. Cleanup from previous runs
    1. Create the KubeFlex hosting cluster and Kubestellar core components
    1. Create and register two WECs.
2. [Exercise KubeStellar](#exercise-kubestellar)
3. [Troubleshooting](#troubleshooting)

## Setup

This is one way to produce a very simple system, suitable for study but not production usage. For general setup information, see [the full story](user-guide-intro.md#the-full-story).

### Note for Windows users

For some users on WSL, use of the setup procedure on this page and/or the demo environment creation script may require running as the user `root` in Linux. There is a [known issue about this](knownissue-wsl-ghcr-helm.md).

### Install software prerequisites

The following command will check for the prerequisites that you will need for the later steps. See [the prerequisites doc](pre-reqs.md) for more details.

```shell
bash <(curl https://raw.githubusercontent.com/kubestellar/kubestellar/v{{ config.ks_latest_release }}/hack/check_pre_req.sh) kflex ocm helm kubectl docker kind
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
kubectl config delete-context cluster1
kubectl config delete-context cluster2
```

After that cleanup, you may want to `set -e` so that failures do not
go unnoticed (the various cleanup commands may legitimately "fail" if
there is nothing to clean up).

### Set the Version appropriately as an environment variable

```shell
kubestellar_version={{ config.ks_latest_release }}
```

### Create a kind cluster to host KubeFlex

For convenience, a new local **Kind** cluster that satisfies the requirements for playing the role of KubeFlex hosting cluster can be created with the following command:

```shell
bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v{{ config.ks_latest_release }}/scripts/create-kind-cluster-with-SSL-passthrough.sh) --name kubeflex --port 9443
```

### Use Core Helm chart to initialize KubeFlex and create ITS and WDS

```shell
helm upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart \
    --version $kubestellar_version \
    --set-json='ITSes=[{"name":"its1"}]' \
    --set-json='WDSes=[{"name":"wds1"}]' \
    --set-json='verbosity.default=5' # so we can debug your problem reports
```

That command will print some notes about how to get kubeconfig "contexts" named "its1" and "wds1" defined. Do that, because those contexts are used in the steps that follow.

```shell
kubectl config use-context kind-kubeflex # this is here only to remind you, it will already be the current context if you are following this recipe exactly
kflex ctx --set-current-for-hosting # make sure the KubeFlex CLI's hidden state is right for what the Helm chart just did
kflex ctx --overwrite-existing-context wds1
kflex ctx --overwrite-existing-context its1
```

#### Wait for ITS to be fully initialized

The Helm chart above has a Job that initializes the ITS as an OCM "hub" cluster. Helm does not have a way to wait for that initialization to finish. So you have to do the wait yourself. The following commands will do that.

```shell
kubectl --context kind-kubeflex wait controlplane.tenancy.kflex.kubestellar.org/its1 --for 'jsonpath={.status.postCreateHooks.its}=true' --timeout 90s
kubectl --context kind-kubeflex wait -n its1-system job.batch/its --for condition=Complete --timeout 150s
```

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
## Troubleshooting

In the event something goes wrong, check out the [troubleshooting page](troubleshooting.md) to see if someone else has experienced the same thing
