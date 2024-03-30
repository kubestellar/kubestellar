# KubeStellar Contributors

![](./images/construction.png){: style="height:100px;width:100px"}


Make sure all pre-requisites are installed as described in [pre-reqs](pre-reqs.md)

## Unit testing

The Makefile has a target for running all the unit tests.

```shell
make test
```

## Integration testing

There are currently two integration tests. Contributors can run them. There is also a GitHub Actions workflow (in `.github/workflows/pr-test-integration.yml`) that runs these tests.

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
