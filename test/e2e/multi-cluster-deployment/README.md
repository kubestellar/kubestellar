# Kubestellar multi-cluster workload deployment with kubectl

This test is an executable variant of the "multi-cluster workload deployment with kubectl" scenario in [the examples doc](../../../docs/content/direct/examples.md). This test has the same prerequisites as the cited one. This test can test either (a) the  local copy of the repo or (b) the release identified in the kubestellar PostCreateHook (which will be the last release created, regardless of quality, except for that brief time when it identifies the release about to be made). Testing the local copy is the default behavior; to test the release identified in the PostCreateHook, pass `--released` on the command line of `run-test.sh`.

The kubestellar controller-manager will be invoked with `-v=2` unless otherwise specified on the command line with `--kubestellar-controller-manager-verbosity $number`. This verbosity can not be set to a value other than 2 when using `--released`.

The transport controller will be invoked with `-v=4` unless othewise specified on the command line with `--transport-controller-verbosity $number`.

## Running the test using a script

Starting from a local directory containing the git repo, do the following.

```
cd test/e2e/multi-cluster-deployment
./run-test.sh
```
