### End-to-end testing


KubeStellar end-to-end testing covers the following test matrix.


- Run either of two scenarios.
- Either create three new `kind` clusters or use three pre-existing OCP clusters.
- Test either the local copy of the repo or the latest release before the local copy's version.


However there is a restriction: when using OCP, only a release can be tested.


This directory has a script that will run a given one of the six allowed cells in that matrix. The script is [run-test.sh](https://github.com/kubestellar/kubestellar/blob/main/test/e2e/run-test.sh). The command line flags say which cell to run. The default is the bash scenario using three new `kind` clusters and the local copy of the repo.


## Version


This script will test the relevant release if `--released` appears on the command line, otherwise will test the local copy of the repo.


## Scenario


Select the scenario by putting `--test-type $scenario` on the command line, where `$scenario` is either `bash` (for the scenario in the [bash subdirectory](bash)) or `ginkgo` (for the scenario in the [ginkgo subdirectory](ginkgo)). In order to run the ginkgo scenario you **need to** have [ginkgo](https://onsi.github.io/ginkgo/) installed; see [ginkgo Getting Started](https://onsi.github.io/ginkgo/#getting-started).


## Infrastructure


Select the infrastructure by putting `--env $env` on the command line, where `$env` is either `kind` (for three new `kind` clusters) or `ocp` (for three pre-existing OCP clusters).


When using `kind`, this test involves three `kind` clusters, so you **need to** [raise the relevant OS limits](https://kind.sigs.k8s.io/docs/user/known-issues/#pod-errors-due-to-too-many-open-files).


When using three pre-existing OCP clusters, your kubeconfig must include contexts named `kscore` (for the kubeflex hosting cluster), `cluster1` and `cluster2`. The following shows an example listing of adequate contexts.


$ kubectl config get-contexts
CURRENT   NAME          CLUSTER                   AUTHINFO               NAMESPACE
          kscore       <url>:port               <defaul-value>            default
          cluster1     <url>:port               <defaul-value>            default

        cluster2     <url>:port               <defaul-value>            default




FYI, if you need to rename a kubeconfig context in order to reach the above configuration then you can use the `kubectl config rename-context` command. For example:


$ kubectl config rename-context <default-wec1-context-name> cluster1




## Fail fast or run every test case


For the ginkgo-based test, normally every test case is run. However, the script accepts a `--fail-fast` flag --- which will get passed on to `ginkgo`, making it stop after the first failed test case.


## Verbosity


The kubestellar controller-manager will be invoked with `-v=2` unless otherwise specified on the command line with `--kubestellar-controller-manager-verbosity $number`. This verbosity can not be set to a value other than 2 when using `--released`.


The transport controller will be invoked with `-v=4` unless otherwise specified on the command line with `--transport-controller-verbosity $number`.