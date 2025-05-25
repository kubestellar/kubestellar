# Kubestellar performance regression testing
This is a simple performance test that aims to identify major regressions. 

The test adds a deployment to the WDS and measures the cumulative time it takes to create the WDS workload object, the Binding object, the ManifestWork object, and the deployment object on the WEC. The starting time is just prior to adding the workload to the WDS and the end time is just after the object is observed on the WEC. The timings are obtained from the objects creation timestamp which have seconds granularity. We expect this test to run on VMs and and anything below seconds is likely just noise. 

This is what the results look like:

```shell
  Run 0: wds deployment=0, binding=1, manifestwork=1, wec deployment=1, total=2
  Run 1: wds deployment=0, binding=0, manifestwork=0, wec deployment=5, total=6
  Run 2: wds deployment=0, binding=1, manifestwork=1, wec deployment=7, total=7
  Run 3: wds deployment=0, binding=0, manifestwork=0, wec deployment=5, total=5
  Run 4: wds deployment=0, binding=0, manifestwork=0, wec deployment=4, total=5
  Run 5: wds deployment=0, binding=0, manifestwork=0, wec deployment=4, total=5
  Run 6: wds deployment=0, binding=0, manifestwork=0, wec deployment=4, total=5
  Run 7: wds deployment=0, binding=1, manifestwork=1, wec deployment=6, total=7
  Run 8: wds deployment=0, binding=0, manifestwork=0, wec deployment=4, total=5
  Run 9: wds deployment=0, binding=0, manifestwork=0, wec deployment=5, total=5
  ----------------------------------------------------------------------------------
  Avg:   wds deployment=0, binding=0, manifestwork=0, wec deployment=4, total=5
```

## Running the tests
To install Ginkgo, follow the instructions in [Ginkgo's getting started](https://onsi.github.io/ginkgo/#getting-started).

To execute these tests, issue the following command. 

```shell
ginkgo -v 
```

To test the latest release up through the commit of test/e2e being used, pass `-released` to the test. For example:

```shell
ginkgo -v -- -released
```

To output just the summary, pass `-just-summary` to the test. For example:

```shell
ginkgo -v -- -just-summary
```

To skip the cleanup and setup phase, pass `-skip-setup` to the test. For example:

```shell
ginkgo -v -- -skip-setup
```



