# Kubestellar multi-cluster workload deployment with kubectl

This test is an executable variant of the "multi-cluster workload deployment with kubectl" scenario in [the examples doc](../../../docs/content/v0.20/examples.md), testing the local copy of the repo rather than things built earlier, and has the same prerequisites.

In addition, the following test cases are executed on top of the setup created by the "multi-cluster workload deployment with kubectl" scenario:
1. Update of the workload object on WDS should update the object on the WECs. Increase the number of replicas from 1 to 2, verify they are updated on the WECs.
2. Changing the placement objectSelector to no longer match should delete the object from the WECs.
3. Changing the placement objectSelector to match should create the object on the WECs.
4. Delete of the placement object should delete the object on the WECs.
5. Delete of the overlapping placement object should not delete the object on the WECs.
6. Delete of the workload object on WDS should delete the object on the WECs.
7. Re-create of the workload object on WDS should re-create the object on the WECs.

## Running the test using a script

Starting from a local directory containing the git repo, do the following.

```
cd test/e2e/multi-cluster-deployment
./run-test.sh
```
