# Testing a new KubeStellar Release

The following testing process should be applied to every KubeStellar new release in order to validate the release. All the tests should be done while the KubeStellar code is still under code-freeze and new code shouldn't be merged into the main branch until all tests are passed and the release is declared as stable.  
In case the release tests fail (even one of them), the release should be marked as instable and a fix should be worked on ASAP. The KubeStellar code-freeze should be lifted only after all tests are passed and a stable releases was released.  
To reduce the exposure of instable releases the update of the KubeStellar site [kubestellar.io](https://docs.kubestellar.io) should be done only once all release tests passed sussceffuly. 

## Release tests
The following section describe the tests that must be executed for each release.

Our release tests consists of:
   * Automatic tests running on Ubuntu X86 (see below)
   * Manually initiated tests running on OCP (TODO: add specific version and machine details)

Note:  We plan to automate all release tests in the future

### Automatic (github based) release tests
KubeStellar CICD automatically runs couple of e2e tests on each new release. Currently these tests include 2 main scenarios - [multi-cluster workload deployment with kubectl](examples.md#scenario-1---multi-cluster-workload-deployment-with-kubectl), and [Singleton status](examples.md#scenario-4---singleton-status), however, this may be changed in the future. We will refer to those tests as the **e2e release tests**. 
The automatic tests are running on github hosted runners of type **Ubuntu latest (currently 22.04) X86 64 bit** 
Note: When a new release is created please verify that the automatic tests indeed executed and passed. 

### e2e release tests on OCP
As many of the KubeStellar customers are using OCP, the release tests should be executed on an OCP cluster as well.  
Currently this test should be initiated manually on a dedicated OCP cluster that is reserved for the release testing process. 

TODO: The details on how to setup and run the test
![](./images/construction.png){: style="height:100px;width:100px"}

## Other platforms
KubeStellar is also used on other platforms such as ARM64, MacOS, etc.. Currently these platforms are not part of the routine release testing, however the KubeStellar team will try its best to help and solve issues detected on other platforms as well. Users should go through the regular procedure of opening issues against the KubeStellar git.
