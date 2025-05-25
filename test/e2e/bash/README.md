#  Kubestellar end to end test written in bash

This test is an executable variant of the "multi-cluster workload deployment with kubectl" scenario in 1 in [the examples doc](../../../docs/content/direct/examples.md). In this scenario, there are one hosting cluster and two workload execution clusters (WECs). Using a single binding policy, a nginx deployment is synced from the hosting cluster to both WECs. For more details refer to [scenario 1](https://github.com/dumb0002/kubestellar/blob/e2e-test-reorg/docs/content/direct/examples.md#scenario-1---multi-cluster-workload-deployment-with-kubectl). 

The bash scripting in this scenario and the common setup/cleanup illustrates how a contributor might do their own testing of their modified version of the repo.
