## Workload Benchmark for KubeStellar

*Pre-requisite*: in order to follow the instructions below, you must have [python3](https://www.python.org/downloads/) and a environment with KubeStellar installed; see [KubeStellar getting started](https://docs.kubestellar.io/release-0.23.1/direct/get-started/). Alternatively, you can also use KubeStellar e2e script [run-test.sh](https://github.com/kubestellar/kubestellar/blob/main/test/e2e/run-test.sh) to setup an environment.

To generate the sample workload for KubeStellar performance experiments proceed as following:

### Generate the workload traffic

1. Clone the following clusterloader2 repo: 

   ```bash
   $ git clone -b release-1.31 https://github.com/kubernetes/perf-tests.git
   ```

   Set the variable `CL2_DIR` to the path of the subdirectory `clusterloader2/` of the cloned repo. For example: 

   ```bash
   $ export CL2_DIR=$HOME/perf-tests/clusterloader2
   ```

2. Configure clusterloader2 to generate KS performance workloads:

   Starting from a local directory containing this KubeStellar git repo, run the following script based on your cluster environment:

   a) cd into the `test/performance` directory from your local copy of the KubeStellar repo, for example:

   ```bash
   $ cd .$HOME/kubestellar/test/performance
   ```

   b) If using plain K8s clusters:

   ```bash
   $ ./setup-clusterloader2.sh
   ```

   c) If using OpenShift clusters: 

   ```bash
   $ ./setup-clusterloader2.sh --env ocp
   ```

3. Configure the parameters of your workload:  

   a) cd into the load configuration directory

   ```bash
   $ cd  $CL2_DIR/testing/load/
   ```
  
   b) configure the parameters to create the workload traffic (e.g., RandomizedLoad, SteppedLoad, Sequence, etc.)
   
   ```bash
   $ vi performance-test-config.yaml
   ``` 

   More specifically, configure the following parameters: 

   - namespaces: number of namespaces to be created in your cluster to deploy the workload defined in step-2 above (*default value: 1*). For example: if `namespaces=2`, then the following namespaces will be created: `perf-exp-0` and `perf-exp-1`
   - K8S_CLUSTER: set to `true` for plain Kubernetes clusters (*default value: "true"*)
   - OPENSHIFT_CLUSTER: set to `true` for Kubernetes clusters (*default value: "false"*)
   - tuningSet: workload generation function (*default value: "RandomizedLoad"*)

   To learn more about clusterloader2 tuningSet, see the following: [tuningSets](https://github.com/kubernetes/perf-tests/blob/master/clusterloader2/docs/design.md#tuning-set) and [configurations](https://github.com/kubernetes/perf-tests/blob/fac2a5eec96fab76a4bc4858795df4544b729b0b/clusterloader2/api/types.go#L249).




4. Deploy your first workload:

   First, use a kubeconfig that includes contexts for WDS spaces and set its current context to a target WDS space (e.g., `wds1`):

   ```bash
   $ kubectl config use-context wds1
   Switched to context "wds1".

   $ kubectl config get-contexts
   CURRENT   NAME            CLUSTER         AUTHINFO        NAMESPACE
             cluster1        kind-cluster1   kind-cluster1   
             cluster2        kind-cluster2   kind-cluster2   
             its1            its1-cluster    its1-admin      
             kind-kubeflex   kind-kubeflex   kind-kubeflex   
   *         wds1            wds1-cluster    wds1-admin      default
             wds2            wds2-cluster    wds2-admin      default
   ```


   In the following set ``--kubeconfig`` flag to the path of the kubeconfig file of your cluster and use the newly added Kubestellar provider (`--provider=ks`) and run the following command to create your workload:

   ```bash
   $ cd $CL2_DIR
   $ go run cmd/clusterloader.go --testconfig=./testing/load/performance-test-config.yaml --kubeconfig=<path>/wds-kubeconfig --provider=ks --v=2
   ```

    In plain Kubernetes environments, a modified version of the kube-burner [cluster-density](https://github.com/kube-burner/kube-burner/tree/main/examples/workloads/cluster-density) workload is generated per namespace. This workload consists of the following objects:

   - 1 deployments, with two pod replicas (pause), mounting 2 secrets, 2 config maps
   - 3 services, the first service points to the TCP/8080 port of the deployments, respectively.
   - 10 secrets containing a 2048-character random string.
   - 10 config maps

   In Openshift environments, a modified version of the kube-burner [cluster-density-ms](https://github.com/kube-burner/kube-burner-ocp/tree/478bb42e1842a94ca3210d26a08633b70a443005/cmd/config/cluster-density-ms) workload is generated per namespace. This workload consists of the following objects:

    - 1 image stream.
    - 4 deployments, each with two pod replicas (pause), mounting 4 secrets, 4 config maps, and
      1 downward API volume each.
    - 2 services, each pointing to the TCP/8080 and TCP/8443 ports of the first and second
      deployments, respectively.
    - 1 edge route pointing to the first service.
    - 20 secrets containing a 2048-character random string.
    - 10 config maps containing a 2048-character random string



### Collect the creation & update timestamps for workload benchmark (OPTIONAL)


To collect the creation and status_update timestamps for the benchmark workload objects proceed as following:  

1. Run the metrics collection script:

   a) cd into the `test/performance` directory from your local copy of the KubeStellar repo, for example:

   ```bash
   $ cd .$HOME/kubestellar/test/performance
   ```

   b) Run the metrics collection script:

   NB: the execution of the following python script requires the installation of the dependencies listed in `requirements.txt`.

   ```bash 
   $ python3 metrics_collector.py <kubeconfig> <wds-context-name> <its-context-name> <wec-context-name> <number-of-namespaces> <output-directory> 
   ```

   For example:
   
   ```bash 
   $ python3 metrics_collector.py $HOME/.kube/config wds1 its1 cluster1 2 $HOME/data
   ```

   Below is a detailed explanation of the input parameters:
   - `kubeconfig`: path to the kubeconfig file, e.g., `$HOME/.kube/config`
   - `wds-context-name`: name of the context for the target WDS, e.g., `wds1`
   - `its-context-name`: name of the context for the target ITS, e.g., `its1`
   - `wec-context-name`: name of the context for the target WEC, e.g., `cluster1`
   - `number-of-namespaces`: number of namespaces created in your experiment, e.g., `2`
   - `output-directory`: path to the directory for the output data files, e.g., `$HOME/data`

   For each namespace, the creation and status_update timestamps will be collected for the workload objects created in `wds1`, manifestwork & workstatus objects created in `its1`, and workload & appliedmanifestworks objects created in `wec1`, for example: let's assume that `num_ns=2`, then the following directories and files will be created: 

   ```bash 
      $ cd $HOME/data
      $ tree
      .
      ├── perf-exp-0
      │   ├── appliedmanifestworks
      │   ├── configmaps-wds1
      │   ├── configmaps-wec
      │   ├── deployments-wds1
      │   ├── deployments-wec
      │   ├── manifestworks
      │   ├── secrets-wds1
      │   ├── secrets-wec
      │   ├── services-wds1
      │   ├── services-wec
      │   └── workstatuses
      └── perf-exp-1
         ├── appliedmanifestworks
         ├── configmaps-wds1
         ├── configmaps-wec
         ├── deployments-wds1
         ├── deployments-wec
         ├── manifestworks
         ├── secrets-wds1
         ├── secrets-wec
         ├── services-wds1
         ├── services-wec
         └── workstatuses
   ```

   The generated output files have the following structure:

   ```bash
   <obj-name> <obj-creation-time> <obj-status-update-time> <obj-status-condition> <obj-controller-manager>
   ```
   
   Where: `<obj-controller-manager>` - it is the FieldManager value used by the controller that updated the status of the workload object.

   For example: 

   ```bash
   stress-pod-2	2024-03-14 19:51:19+00:00	2024-03-14 19:51:44+00:00	Succeeded	controller-manager
   ```

2. Clean up the generated workload Kubernetes API objects from your cluster:

   ```bash
   $ cd $CL2_DIR
   $ ./cleanup.sh
   ```