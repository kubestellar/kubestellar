#  Kubestellar end to end testing

**PRE-REQ**: All of these tests use three `kind` clusters, so you need to [raise the relevant OS limits](https://kind.sigs.k8s.io/docs/user/known-issues/#pod-errors-due-to-too-many-open-files).

This test is an executable variant of the "multi-cluster workload deployment with kubectl" scenario in 1 in [the examples doc](../../../docs/content/direct/examples.md). In this scenario, there are one hosting cluster and two workload execution clusters (WECs). Using a single binding policy, a nginx deployment is synced from the hosting cluster to both WECs. For more details refer to [scenario 1](https://github.com/dumb0002/kubestellar/blob/e2e-test-reorg/docs/content/direct/examples.md#scenario-1---multi-cluster-workload-deployment-with-kubectl). 

## Running the test using a script

Starting from a local directory containing the git repo, do the following.

```
cd test/e2e/bash
../run-test.sh
```