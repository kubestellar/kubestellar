# Kubestellar end to end testing

This e2e tests can test either (a) the  local copy of the repo or (b) the release identified in the kubestellar PostCreateHook (which will be the last release created, regardless of quality, except for that brief time when it identifies the release about to be made). Testing the local copy is the default behavior; to test the release identified in the PostCreateHook, pass `--released` on the command line of `run-test.sh`.

The kubestellar controller-manager will be invoked with `-v=2` unless otherwise specified on the command line with `--kubestellar-controller-manager-verbosity $number`. This verbosity can not be set to a value other than 2 when using `--released`.

The transport controller will be invoked with `-v=4` unless othewise specified on the command line with `--transport-controller-verbosity $number`.

## Running the test on three new Kind clusters

**PRE-REQ**: All of these tests use three `kind` clusters, so you need to [raise the relevant OS limits](https://kind.sigs.k8s.io/docs/user/known-issues/#pod-errors-due-to-too-many-open-files).

Starting from a local directory containing the git repo, do the following.

a) Bash tests:
```bash
cd test/e2e/multi-cluster-deployment
./run-test.sh
```

b) Ginkgo tests:
```bash
cd test/e2e/multi-cluster-deployment
./run-test.sh --test-type ginkgo
```

## Running the test in three existing OCP clusters

**NOTE**: at present this _only_ works with `--released`.

1. Create a kubeconfig with contexts named `kscore` (for the kubeflex hosting cluster), `cluster1` and `cluster2`. The following shows what the result should look like.

```bash
$ kubectl config get-contexts
CURRENT   NAME          CLUSTER                   AUTHINFO               NAMESPACE
          kscore       <url>:port               <defaul-value>            default
          cluster1     <url>:port               <defaul-value>            default
*         cluster2     <url>:port               <defaul-value>            default
```

Use the following command to rename the default context name for host cluster and workload execution clusters:

```bash 
$ kubectl config rename-context <default-wec1-context-name> cluster1
```

2. Run e2e test in your ocp clusters:

a) Bash tests:
```bash
cd test/e2e/multi-cluster-deployment
./run-test.sh --env ocp --released
```

b) Ginkgo tests:
```bash
cd test/e2e/multi-cluster-deployment
./run-test.sh --env ocp --released --test-type ginkgo
```
