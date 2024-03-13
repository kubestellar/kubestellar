# Kubestellar multi-cluster workload deployment with kubectl

This test is an executable and extended variant of the "multi-cluster workload deployment with kubectl" scenario in [the examples doc](../../../docs/content/direct/examples.md). This test has the same prerequisites as the cited one. This test can test either (a) the  local copy of the repo or (b) the release identified in the kubestellar PostCreateHook (which will be the last release created, regardless of quality, except for that brief time when it identifies the release about to be made). Testing the local copy is the default behavior; to test the release identified in the PostCreateHook, pass `--released` on the command line of `run-test.sh`.

This variant extends the cited one by executing the following test cases at the end:
1. Update of the workload object on WDS should update the object on the WECs. Increase the number of replicas from 1 to 2, verify they are updated on the WECs.
2. Changing the bindingpolicy objectSelector to no longer match should delete the object from the WECs.
3. Changing the bindingpolicy objectSelector to match should create the object on the WECs.
4. Delete of the bindingpolicy object should delete the object on the WECs.
5. Delete of the overlapping bindingpolicy object should not delete the object on the WECs.
6. Delete of the workload object on WDS should delete the object on the WECs.
7. Re-create of the workload object on WDS should re-create the object on the WECs.

## Running the test using a script

Starting from a local directory containing the git repo, do the following.

```
cd test/e2e/multi-cluster-deployment
./run-test.sh
```
