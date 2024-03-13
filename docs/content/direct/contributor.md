# KubeStellar Contributors

**NOTE**: work in progress... write more

## Check pre-requisites for KubeStellar

The [check_pre_req](../../../hack/check_pre_req.sh) script offers a convenient way to check for pre-requisites, needed for [KubeStellar](./pre-reqs.md) deployment and [use case scenarios](./examples.md).

The [check_pre_req](../../../hack/check_pre_req.sh) script check for a pre-requisite presence in the path, by using the `which` command, and it can optionally provide version and path information for pre-requisites that are present, or installation information for missing pre-requisites.

We envision that the [check_pre_req](../../../hack/check_pre_req.sh) script could be useful for user-side debugging as well as for asserting the presence of pre-requisites in higher-level automation scripts.

The [check_pre_req](../../../hack/check_pre_req.sh) script accepts a list of optional flags and arguments.

**Supported flags:**

- `-A|--assert`: exits with error code 2 upon finding the fist missing pre-requisite
- `-L|--list`: prints a list of supported pre-requisites
- `-V|--verbose`: displays version and path information for installed pre-requisites or installation information for missing pre-requisites
- `-X`: enable `set -x` for debugging the script

**Supported arguments:**

The [check_pre_req](../../../hack/check_pre_req.sh) script accepts a list of specific pre-requisites to check, among the list of available ones:

```shell
$ check_pre_req.sh --list
argo brew docker go helm jq kflex kind ko kubectl make ocm yq
```

For example, list of pre-requisites required by Kubestellar can be checked with the command below (add the `-V` flag to get the version of each program na dusggestions on how to install missing pre-requisites):

```shell
$ hack/check_pre_req.sh
Checking pre-requisites for using KubeStellar:
✔ Docker
✔ kubectl
✔ KubeFlex
✔ OCM CLI
✔ Helm
Checking additional pre-requisites for running the examples:
✔ Kind
X ArgoCD CLI
Checking pre-requisites for building KubeStellar:
✔ GNU Make
✔ Go
✔ KO
```

In another example, a specific list of pre-requisites could be asserted by an higher-level script, while providing some installation information, with the command below (note that the script will terminate upon finding a missing pre-requisite):

```shell
$ check_pre_req.sh --assert --verbose helm argo docker kind
Checking KubeStellar pre-requisites:
✔ Helm
  version: version.BuildInfo{Version:"v3.14.0", GitCommit:"3fc9f4b2638e76f26739cd77c7017139be81d0ea", GitTreeState:"clean", GoVersion:"go1.21.5"}
     path: /usr/sbin/helm
X ArgoCD CLI
  how to install: https://argo-cd.readthedocs.io/en/stable/cli_installation/
```

## Unit testing

The Makefile has a target for running all the unit tests.

```shell
make test
```

## Integration testing

There are currently two integration tests. Contributors can run them. There is also [a GitHub workflow](../../../.github/workflows/pr-test-integration.yml) that runs these tests.

These tests require you to already have `etcd` on your `$PATH`.
See https://github.com/kubernetes/kubernetes/blob/v1.28.2/hack/install-etcd.sh for an example of how to do that.

To run the tests sequentially, issue a command like the following.

```shell
CONTROLLER_TEST_NUM_OBJECTS=24 go test -v ./test/integration/controller-manager &> /tmp/test.log
```

If `CONTROLLER_TEST_NUM_OBJECTS` is not set then the number of objects
will be 18. This parameterization by an environment variable is only a
point-in-time hack, it is expected to go away once we have a test that
runs reliably on a large number of objects.

To run one of the individual tests, issue a command like the following example.

```shell
go test -v -timeout 60s -run ^TestCRDHandling$ ./test/integration/controller-manager
```

## Making releases

See [the release process document](release.md).
