# Getting Started with KubeStellar

## Set Up A Demo System

This page shows two ways to create one particular simple configuration that is suitable for kicking the tires (not production usage). This configuration has one `kind` cluster serving as your KubeFlex hosting cluster and two more serving as WECs. This page covers steps 2--7 from [the full installation and usage outline](user-guide-intro.md#the-full-story). This page concludes with forwarding you to some example scenarios that illustrate the remaining steps.

The two ways to create this simple configuration are as follows.

1. A [quick automated setup](#quick-start-using-the-automated-script) using our demo setup script, which creates a basic working environment for those who want to start experimenting right away.

2. A [Step by step walkthrough](#step-by-step-setup) that demonstrates the core concepts and components, showing how to manually set up a simple single-host system.

### Note for Windows users

For some users on WSL, use of the setup procedure on this page and/or the demo environment creation script may require running as the user `root` in Linux. There is a [known issue about this](knownissue-wsl-ghcr-helm.md).

## Quick Start Using the Automated Script

If you want to quickly setup a basic environment, you can use our automated installation script.

### Install software prerequisites

Be sure to [install the software prerequisites](pre-reqs.md) _before_ running the script!

The script will check for the pre-reqs and exit if they are not present.

### Run the script!
The script can install KubeStellar's demonstration environment on top of kind or k3d

For use with kind
```shell
bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/refs/tags/v{{ config.ks_latest_release }}/scripts/create-kubestellar-demo-env.sh) --platform kind
```

For use with k3d
```shell
bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/refs/tags/v{{ config.ks_latest_release }}/scripts/create-kubestellar-demo-env.sh) --platform k3d
```

If successful, the script will output the variable definitions that you would use when proceeding to the example scenarios. After successfully running the script, proceed to the [Exercise KubeStellar](#exercise-kubestellar) section below.

_Note: the script does the same things as described in the [Step by Step Setup](#step-by-step-setup) but with maximum concurrency, so it can complete faster. This makes the script actually more complicated than the step-by-step process below. While this is great for getting started quickly with a demo system, you may want to follow the manual setup below to better understand the components and how to create a [configuration that meets your needs](#next-steps)._

## Step by Step Setup

This walks you through the steps to produce the same configuration as does the script above, suitable for study but not production usage. For general setup information, see [the full story](user-guide-intro.md#the-full-story).

### Install software prerequisites

The following command will check for the prerequisites that you will need for the later steps. See [the prerequisites doc](pre-reqs.md) for more details.

```shell
bash <(curl https://raw.githubusercontent.com/kubestellar/kubestellar/v{{ config.ks_latest_release }}/scripts/check_pre_req.sh) kflex ocm helm kubectl docker kind
```

If that script complains then take it seriously! For example, the following indicates that you have a version of clusteradm that KubeStellar cannot use.

```console
$ bash <(curl https://raw.githubusercontent.com/kubestellar/kubestellar/v0.27.1/scripts/check_pre_req.sh) kflex ocm helm kubectl docker kind
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100  9278  100  9278    0     0   135k      0 --:--:-- --:--:-- --:--:--  137k
✔ KubeFlex (Kubeflex version: v0.8.2.5fd5f9c 2025-03-10T14:58:02Z)
✔ OCM CLI (:v0.11.0-0-g73281f6)
  structured version ':v0.11.0-0-g73281f6' is less than required minimum ':v0.7' or ':v0.10' but less than ':v0.11'
```

This setup recipe uses [kind](https://kind.sigs.k8s.io/) to create three Kubernetes clusters on your machine.
Note that `kind` does not support three or more concurrent clusters unless you raise some limits as described in this `kind` "known issue": [Pod errors due to "too many open files"](https://kind.sigs.k8s.io/docs/user/known-issues/#pod-errors-due-to-too-many-open-files).

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
    --version "$kubestellar_version" \
    --set-json ITSes='[{"name":"its1"}]' \
    --set-json WDSes='[{"name":"wds1"},{"name":"wds2","type":"host"}]' \
    --set verbosity.default=5  # so we can debug your problem reports
```

That command will print some notes about how to get kubeconfig "contexts" named "its1", "wds1", and "wds2" defined. Do that, because those contexts are used in the steps that follow.

```shell
kubectl config use-context kind-kubeflex # this is here only to remind you, it will already be the current context if you are following this recipe exactly
kflex ctx --set-current-for-hosting # make sure the KubeFlex CLI's hidden state is right for what the Helm chart just did
kflex ctx --overwrite-existing-context wds1
kflex ctx --overwrite-existing-context wds2
kflex ctx --overwrite-existing-context its1
```

#### Wait for ITS to be fully initialized

The Helm chart above has a Job that initializes the ITS as an OCM "hub" cluster. Helm does not have a way to wait for that initialization to finish. So you have to do the wait yourself. The following commands will do that.

```shell
kubectl --context kind-kubeflex wait controlplane.tenancy.kflex.kubestellar.org/its1 --for 'jsonpath={.status.postCreateHooks.its-with-clusteradm}=true' --timeout 90s
kubectl --context kind-kubeflex wait -n its1-system job.batch/its-with-clusteradm --for condition=Complete --timeout 150s
```

*To learn more about the Core Helm Chart, refer to the [Core Helm Chart documentation](./core-chart.md)*

### Create and register two workload execution clusters

The following steps show how to create two new `kind` clusters and
register them with the hub as described in the
[official open cluster management docs](https://open-cluster-management.io/docs/getting-started/installation/start-the-control-plane/).

Note that `kind` does not support three or more concurrent clusters unless you raise some limits as described in this `kind` "known issue": [Pod errors due to "too many open files"](https://kind.sigs.k8s.io/docs/user/known-issues/#pod-errors-due-to-too-many-open-files).

1. Execute the following commands to create two kind clusters, named `cluster1` and `cluster2`, and register them with the OCM hub. These clusters will serve as workload clusters. If you have previously executed these commands, you might already have contexts named `cluster1` and `cluster2`. If so, you can remove these contexts using the commands `kubectl config delete-context cluster1` and `kubectl config delete-context cluster2`.

    ```shell
    : set flags to "" if you have installed KubeStellar on an OpenShift cluster
    flags="--force-internal-endpoint-lookup"
    clusters=(cluster1 cluster2);
    for cluster in "${clusters[@]}"; do
       kind create cluster --name ${cluster}
       kubectl config rename-context kind-${cluster} ${cluster}
       clusteradm --context its1 get token | grep '^clusteradm join' | sed "s/<cluster_name>/${cluster}/" | awk '{print $0 " --context '${cluster}' --singleton '${flags}'"}' | sh
    done
    ```

    The `clusteradm` command grabs a token from the hub (`its1` context), and constructs the command to apply the new cluster
    to be registered as a managed cluster on the OCM hub.

2. Repeatedly issue the command:

    ```shell
    kubectl --context its1 get csr
    ```

    until you see that the certificate signing requests (CSR) for both cluster1 and cluster2 exist.
    Note that the CSRs condition is supposed to be `Pending` until you approve them in step 4.

3. Once the CSRs are created, approve the CSRs complete the cluster registration with the command:

    ```shell
    clusteradm --context its1 accept --clusters cluster1
    clusteradm --context its1 accept --clusters cluster2
    ```

4. Check the new clusters are in the OCM inventory and label them:

    ```shell
    kubectl --context its1 get managedclusters
    kubectl --context its1 label managedcluster cluster1 location-group=edge name=cluster1
    kubectl --context its1 label managedcluster cluster2 location-group=edge name=cluster2
    ```

### Variables for running the example scenarios.

Before moving on to try exercising KubeStellar, you will need the following shell variable settings to inform the scenario commands about the configuration.

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

## Exercise KubeStellar

Now that your system is running, you can try some example scenarios

1. Define the needed shell variables, using either the settings output as the script completes or the settings shown just above from the step-by-step instructions. Their meanings are defined [at the start of the example scenarios document](example-scenarios.md#assumptions-and-variables).

2. Proceed to [Scenario 1 (multi-cluster workload deployment with kubectl) in the example scenarios](example-scenarios.md#scenario-1-multi-cluster-workload-deployment-with-kubectl) and/or other examples on the same page, after defining the shell variables that characterize the configuration created above.

## Next Steps

The configuration created here was a basic one suitable for learning. The [full Installation and Usage outline](user-guide-intro.md#the-full-story) shows that KubeStellar has a lot of flexibility.

- Create Kubernetes clusters any way you want
- Multiple Inventory and Transport Spaces (ITS)
- Multiple Workload Definition Spaces (WDS)
- Dynamic addition and removal of ITSes
- Dynamic addition and removal of WDSes
- Use the KubeFlex hosting cluster or a KubeFlex Control Plane as ITS
- Use the KubeFlex hosting cluster or a KubeFlex Control Plane as WDS
- Dynamic addition and removal of Workload Execution Clusters (WECs)

For general setup information, see [the full story](user-guide-intro.md#the-full-story).

## Troubleshooting

In the event something goes wrong, check out the [troubleshooting page](troubleshooting.md) to see if someone else has experienced the same thing
