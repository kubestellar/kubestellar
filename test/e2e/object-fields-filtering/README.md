# Kubestellar object fields filtering

In some objects, k8s api-server adds server managed fields to the object that have to be cleaned from the object before it's distributed to WECs.
Not cleaning those fields may result in a failure on the WEC when the agent tries to apply the object.
One example for such an object type is "Service". When applying a service to WDS, the api server adds the clusterIP and additional fields that are specific to the WDS network.
Trying to distribute it to the WEC may result in a failure since the WEC has a different IP range.
This is only one example for such a use case, there are more object types and the ones that are handled by KubeStellar should be added to this e2e test.

This e2e test is validating that such objects with unique behavior are handled correctly in KubeStellar. 
This test has the same prerequisites as [mutli-cluster-deployment](../multi-cluster-deployment/) and [singleton-status](../singleton-status/) e2e tests.

This test can test either (a) the  local copy of the repo or (b) the release identified in the kubestellar PostCreateHook (which will be the last release created, regardless of quality, except for that brief time when it identifies the release about to be made). Testing the local copy is the default behavior; to test the release identified in the PostCreateHook, pass `--released` on the command line of `run-test.sh`.

## Running the test using a script

Starting from a local directory containing the git repo, do the following.

```
cd test/e2e/object-fields-filtering
./run-test.sh
```
