# Kubestellar end to end testing
The end to end testing includes:
1. deployments are downsync propegated to the WECs
1. Update of the workload object on WDS should update the object on the WECs. Increase the number of replicas from 1 to 2, verify they are updated on the WECs.
1. Changing the bindingpolicy objectSelector to no longer match should delete the object from the WECs
1. Changing the bindingpolicy objectSelector to match should create the object on the WECs
1. Delete of an overlapping bindingpolicy object should not delete objects on the WECs
1. Delete of the workload object on WDS deletes the relevant objects on the WECs
1. Delete of a bindingpolicy deletes the relevant objects on the WECs
1. Downsync objects that fully match on object and cluster selector
1. Handles OR of cluster and object selectors
1. Downsync based on object labels and object name
1. Singleton status update
1. Object cleaning for services
1. Resiliency testing - killing kubestellar manager on the WDS
1. Resiliency testing - killing kubeflex
1. Resiliency testing - killing both kubestellar and kubeflex

## Running the tests
To install Ginkgo, follow the instructions in [Ginkgo's getting started](https://onsi.github.io/ginkgo/#getting-started).

To execute these tests, issue the following command. This will make a local image and run the end to end tests. Omit the `--no-color` if you want pretty terminal output; omit the `KFLEX_DISABLE_CHATTY=true` to get progress logging from kubeflex.

```shell
KFLEX_DISABLE_CHATTY=true ginkgo --vv --trace --no-color
```

To test the latest release image, pass `--released` to the test. For example:

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



