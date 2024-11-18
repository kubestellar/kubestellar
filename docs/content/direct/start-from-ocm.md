# Adding KubeStellar to OCM

In general, the key idea is to use the OCM hub cluster as the
KubeStellar "Inventory and Transport Space" (ITS) and the KubeFlex
hosting cluster.

This page shows one concrete example of adding KubeStellar to an
existing OCM system. In particular, a hub plus two managed clusters
almost exactly as created by [the OCM Quick Start
instructions](https://open-cluster-management.io/docs/getting-started/quick-start/).
In terms of the [full Installation and Usage outline of
KubeStellar](user-guide-intro.md#the-full-story), the modified OCM
Quick Start has already: established some, but not all, of the
software prerequisites; acquired the ability to use a Kube cluster as
KubeFlex hosting cluster; created an Inventory and Transport Space;
created two Workload Execution Clusters (WECs) and registered
them. These are the boxes outlined in red in the following flowchart.

![this copy of the general installation and usage flowchart](images/ocm-usage-outline.svg).

  1. [Setup](#setup)
    1. Install remaining software prerequisites
    1. Cleanup from previous runs
    1. OCM Quick Start with Ingress
    1. Label WECs for selection by examples
    1. Install Kubestellar core components
  2. [Exercise KubeStellar](#exercise-kubestellar)
  3. [Troubleshooting](#troubleshooting)

## Setup

Continuing with the spirit of the OCM Quick Start, this is one way to
produce a very simple system --- suitable for study but not production
usage. For general setup information, see [the full
story](user-guide-intro.md#the-full-story).

### Install software prerequisites

The following command will check for the prerequisites that KubeStellar will need for the later steps. See [the prerequisites doc](pre-reqs.md) for more details.

```shell
bash <(curl https://raw.githubusercontent.com/kubestellar/kubestellar/v{{ config.ks_latest_release }}/hack/check_pre_req.sh) kflex ocm helm kubectl docker kind
```

### Cleanup from previous runs

If you have run this recipe or any related recipe previously then
you will first want to remove any related debris. The following
commands tear down the state established by this recipe.

```shell
kind delete cluster --name hub
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

### OCM Quick Start with Ingress

This recipe uses a modified version of [the OCM Quick Start script](https://raw.githubusercontent.com/open-cluster-management-io/OCM/v0.15.0/solutions/setup-dev-environment/local-up.sh). The modification is necessary because KubeStellar requires the hosting cluster to have an Ingress controller with SSL passthrough enabled. The modified Quick Start script has the following modifications compared to the baseline.

1. The `kind` cluster created for the hub has an additional port mapping, where the Ingress controller listens.
1. The script installs [the NGINX Ingress Controller](https://docs.nginx.com/nginx-ingress-controller/) into the hub cluster, then patches the controller to enable SSL passthrough, and later waits for it to be in service.

You can invoke the modified OCM Quick Start as follows.

```shell
curl -L https://raw.githubusercontent.com/kubestellar/kubestellar/refs/tags/v{{ config.ks_latest_release }}/scripts/ocm-local-up-for-ingress.sh | bash
```

Like the baseline, this script creates a `kind` cluster named "hub" to
serve as hub cluster (known in KubeStellar as an Inventory and
Transport Space, ITS) and two `kind` clusters named "cluster1" and
"cluster2" to serve as managed clusters (known in KubeStellar as
Workload Execution Clusters, WECs), and registers them in the hub.

### Label WECs for selection by examples

The examples will use label selectors to direct workload to WECs (in
terms of their ManagedCluster representations). The following commands
apply labels that will be used in the examples.

```shell
kubectl --context kind-hub label managedcluster cluster1 location-group=edge name=cluster1
kubectl --context kind-hub label managedcluster cluster2 location-group=edge name=cluster2
```

### Use Core Helm chart to initialize KubeFlex, recognize ITS, and create WDS

This chart instance will do the following.

- Install KubeFlex in the hosting cluster.
- Assign the hosting cluster the role of an ITS, named "its1".
- Create a KubeFlex ControlPlane named "wds1" to play the role of a WDS.
- Install the KubeStellar core stuff in the hosting cluster.

```shell
helm --kube-context kind-hub upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart \
    --version $kubestellar_version \
    --set-json='ITSes=[{"name":"its1", "type":"host"}]' \
    --set-json='WDSes=[{"name":"wds1"}]' \
    --set-json='verbosity.default=5' # so we can debug your problem reports
```

That command will print some notes about how to get a kubeconfig
"context" named "wds1" defined. Do that, because this context is used
in the steps that follow. The notes assume that your current
kubeconfig context is the one where the Helm chart was installed,
which is not necessarily true --- so take care for that too.

```shell
kubectl config use-context kind-hub
kflex ctx --set-current-for-hosting # make sure the KubeFlex CLI's hidden state is right for what the Helm chart just did
kflex ctx --overwrite-existing-context wds1
```

For more information about this Helm chart, see [its documentation](core-chart.md).

## Exercise KubeStellar

Use the following commands to wait for the KubeStellar core Helm chart to finish setting up the WDS, because the examples assume that this has completed.

```shell
while [ -z "$(kubectl --context wds1 get crd bindingpolicies.control.kubestellar.io --no-headers -o name 2> /dev/null)" ] ;  do
    sleep 5
done
kubectl --context wds1 wait --for condition=Established crd bindingpolicies.control.kubestellar.io
```

Proceed to [Scenario 1 (multi-cluster workload deployment with kubectl) in the example scenarios](example-scenarios.md#scenario-1-multi-cluster-workload-deployment-with-kubectl) _after_ defining the shell variables that characterize the setup done above. Following are the settings for those variables, whose meanings are defined [at the start of the example scenarios document](example-scenarios.md#assumptions-and-variables).

```shell
host_context=kind-hub
its_cp=its1
its_context=${host_context}
wds_cp=wds1
wds_context=wds1
wec1_name=cluster1
wec2_name=cluster2
wec1_context=kind-$wec1_name
wec2_context=kind-$wec2_name
label_query_both=location-group=edge
label_query_one=name=cluster1
```
## Troubleshooting

In the event something goes wrong, check out the [troubleshooting page](troubleshooting.md) to see if someone else has experienced the same thing
