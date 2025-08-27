# ginkgo end-to-end testing

This end to end testing includes:
1. Deployments are downsync propagated to the WECs
1. Update of the workload object on WDS should update the object on the WECs. Increase the number of replicas from 1 to 2, verify they are updated on the WECs.
1. Supports to the 'create-only' mode
1. Changing the BindingPolicy objectSelector to no longer match should delete the object from the WECs
1. Changing the BindingPolicy objectSelector to match should create the object on the WECs
1. Delete of an overlapping BindingPolicy object should not delete objects on the WECs
1. Delete of the workload object on WDS deletes the relevant objects on the WECs
1. Delete of a BindingPolicy deletes the relevant objects on the WECs
1. Downsync objects that fully match on object and cluster selector
1. Handles OR of cluster and object selectors
1. Downsync based on object labels and object name
1. Singleton status update
1. Object cleaning for services
1. Resiliency testing - killing kubestellar manager on the WDS
1. Resiliency testing - killing kubeflex
1. Resiliency testing - killing both kubestellar and kubeflex

## Running the tests

In order to run this test suite you must have ginkgo installed; see [Ginkgo's getting started](https://onsi.github.io/ginkgo/#getting-started).

You can run this test suite using the general runner, as described in [the parent README](../README.md). Alternatively, you can run this suite directly using ginkgo commands shown below.

To execute these tests, issue the following command. This will make a local image and run the end to end tests. Omit the `--no-color` if you want pretty terminal output; omit the `KFLEX_DISABLE_CHATTY=true` to get progress logging from kubeflex.

```shell
KFLEX_DISABLE_CHATTY=true ginkgo --vv --trace --no-color
```

To pass additional command line flags/args to the embedded usage of the [setup-kubestellar.sh script](../common/setup-kubestellar.sh), pass one additional flag to the test suite, where this outer flag's name is `kubestellar-setup-flags` and the outer flag's value is the space-separated concatenation of all the additional setup-kubestellar flags/args. Remember that test suite args/flags must be preceded on the `ginkgo` command line by `--`. Following is an example.

```shell
KFLEX_DISABLE_CHATTY=true ginkgo --vv --trace --no-color -- -kubestellar-setup-flags="--kubestellar-controller-manager-verbosity 5" 
```

To test the latest release image, either (a) pass the `--released` flag to setup-kubestellar using the technique above or (b) pass `--released` to the test suite. For example:

```shell
KFLEX_DISABLE_CHATTY=true ginkgo --vv --trace --no-color -- -released
```

To test a specific test use ginkgo's `--focus` parameter.  For example:

```shell
KFLEX_DISABLE_CHATTY=true ginkgo --vv --trace --no-color --focus "survives ITS vcluster coming down"
```

To skip the cleanup and setup phase, pass `-skip-setup` to the test. For example:

```shell
KFLEX_DISABLE_CHATTY=true ginkgo --vv --trace --no-color -- -skip-setup
```



