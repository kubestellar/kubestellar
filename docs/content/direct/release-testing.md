# Testing a new KubeStellar Release

The following testing process should be applied to every new KubeStellar release in order to validate it, this include both regular releases and release candidates. All the tests should be done while the KubeStellar code is still under code-freeze and new code shouldn't be merged into the main branch until all tests are passed and the release is officially declared as ready.  
In case the release tests fail (even one of them), the release should be declared as unstable and a fix through a new release candidate should be worked on ASAP. The KubeStellar code-freeze should be lifted only after all tests are passed and a the release was completed.  
To reduce the exposure of unstable releases the update of the KubeStellar site [kubestellar.io](https://docs.kubestellar.io) should be done only once all release tests passed successfully. 

## Release tests
The following section describe the tests that must be executed for each release.

Our release tests consists of:
   * Automatic tests running on Ubuntu X86 (see below)
   * Manually initiated tests running on OCP (TODO: add specific version and machine details)

Due to the lack of OCP based automatic testing, these tests will be performed only once a release candidate passed all other tests and is a candidate to become a regular release. 

Note:  We plan to automate all release tests in the future

### Automatic (github based) release tests
KubeStellar CICD automatically runs a set of e2e tests on each new release. Currently these tests include 2 main test types bash based e2e tests and ginkgo based e2e tests. The bash test basically tests the scenario of  [multi-cluster workload deployment with kubectl](example-scenarios.md#scenario-1-multi-cluster-workload-deployment-with-kubectl). The ginkgo test cover the [Singleton status test](example-scenarios.md#scenario-4-singleton-status), and several other tests that are listed in the test [README](https://github.com/kubestellar/kubestellar/blob/main/test/e2e/ginkgo/README.md). Note, however, that the content of the releases tests may be changed in the future. We will refer to those tests as the **e2e release tests**. 
The automatic tests are running on github hosted runners of type **Ubuntu latest (currently 22.04) X86 64 bit** 
Note: When a new release is created please verify that the automatic tests indeed executed and passed. 

### e2e release tests on OCP
As many of the KubeStellar customers are using OCP, the release tests should be executed on an OCP cluster as well.  
Currently these tests should be initiated manually on a dedicated OCP cluster that is reserved for the release testing process. 

TODO: The details on how to setup and run the test
![](./images/construction.png){: style="height:100px;width:100px"}

## Other platforms
KubeStellar is also used on other platforms such as ARM64, MacOS, etc.. Currently these platforms are not part of the routine release testing, however the KubeStellar team will try its best to help and solve issues detected on other platforms as well. Users should go through the regular procedure of opening issues against the KubeStellar [project](https://github.com/kubestellar/kubestellar/) .
