# Kubestellar multi-cluster workload deployment with kubectl

This test is an executable variant of the "multi-cluster workload deployment with kubectl" scenario in [the examples doc](../../../docs/content/direct/examples.md). This test has the same prerequisites as the cited one. This test can test either (a) the  local copy of the repo or (b) the release identified in the kubestellar PostCreateHook (which will be the last release created, regardless of quality, except for that brief time when it identifies the release about to be made). Testing the local copy is the default behavior; to test the release identified in the PostCreateHook, pass `--released` on the command line of `run-test.sh`.

The kubestellar controller-manager will be invoked with `-v=2` unless otherwise specified on the command line with `--kubestellar-controller-manager-verbosity $number`. This verbosity can not be set to a value other than 2 when using `--released`.

The transport controller will be invoked with `-v=4` unless othewise specified on the command line with `--transport-controller-verbosity $number`.

## Running the test on three new Kind clusters

Starting from a local directory containing the git repo, do the following.

```
cd test/e2e/multi-cluster-deployment
./run-test.sh
```

## Running the test in three existing OCP clusters


1. Create a kubeconfig with contexts for the `kscore`, `cluster1` and `cluster2`. The following shows what the result should look like.

```bash
$ kubectl config get-contexts
CURRENT   NAME          CLUSTER                   AUTHINFO               NAMESPACE
          kscore       <url>:port               <defaul-value>            default
          cluster1     <url>:port               <defaul-value>            default
*         cluster2     <url>:port               <defaul-value>            default
```

Use the following command to rename the default context name for `cluster1` workload execution cluster:

```bash 
$ kubectl config rename-context <default-wec1-context-name> kscore
```

2. Run e2e test in your ocp cluster:

```
 export KUBESTELLAR_VERSION=0.23.0-alpha.1
 export OCM_STATUS_ADDON_VERSION=0.2.0-rc8
 export OCM_TRANSPORT_PLUGIN=0.1.7
 bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/release-$KUBESTELLAR_VERSION/test/e2e/multi-cluster-deployment/run-test.sh) --env ocp
```
