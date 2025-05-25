## Workload Benchmark for KubeStellar

To generate the sample workload for KubeStellar performance experiments for long running tests proceed as following:

1. Clone the following clusterloader2 repo. It's the tool used to deploy the benchmark workload in a KubeStellar environment:

   ```bash 
   git clone -b release-1.31 https://github.com/kubernetes/perf-tests.git
   ```

2. Configure clusterloader2 to generate KS performance workloads:

   Starting from a local directory containing this KubeStellar git repo, run the following script to set-up your environment:

   a) cd into the `test/performance/common` directory from your local copy of the KubeStellar repo, for example:

   ```bash 
   cd $HOME/kubestellar/test/performance/common
   ```

   b) set-up your environment:

   First, set the variable `CL2_DIR` to the path of the subdirectory `clusterloader2/` of the cloned repo in step 1. For example: 

   ```bash 
   export CL2_DIR=$HOME/perf-tests/clusterloader2
   ```

   Then, run the set-up script as following:
   ```bash 
   ./setup-clusterloader2.sh --exp l
   ```

3. Configure the parameters of your workload:  

   a) cd into the load configuration directory.

   ```bash 
   cd  $CL2_DIR/testing/load/
   ```
  
   b) configure the parameters to create the workload traffic.
   
   ```bash 
   vi long-duration-exp-config.yaml
   ``` 

   More specifically, configure the following parameter: 

   - numberOfWorkloadObjects: total number of objects to be created in the experiment `perf-test` namespace (default value: `10`)
   - burstSize: number of workload objects created at a burst event (default value: `1`) 
   - stepDelay: specifies the time interval between burst events (default value: `60s`)

   The workload objects are created using the clusterloader `SteppedLoad` tuningSet. It defines a load that generates a burst of a given size every X seconds, see more [here](https://github.com/kubernetes/perf-tests/blob/fac2a5eec96fab76a4bc4858795df4544b729b0b/clusterloader2/api/types.go#L240). To learn more about clusterloader2 tuningSet, see the following: [tuningSets configurations](https://github.com/kubernetes/perf-tests/blob/master/clusterloader2/docs/design.md#tuning-set).
   
   Use the following equation to understand how long an experiment should take in seconds:

   $$t_{exp} = [stepDelay \times \frac{numberOfWorkloadObjects}{burstSize}] + \alpha$$

   Where $\alpha <= 60$ seconds and it corresponds to the amount of time that clusterloader2 takes to cleanup resources after successful execution of a test. 
   
   For example, using the default parameters (i.e., burstSize=1, stepDelay=60 seconds, 10 workload objects) the experiment will take 660 seconds at most. Below are some sample parameters:

   | burstSize (#)   | stepDelay (sec) |  numberOfWorkloadObjects  | duration |
   | --------------- | --------------- | ------------------------- |--------- |
   |       1         |      60         |           1,440           |  1 day   |
   |       1         |      5          |          120,960          |  7 days  |


4. Deploy your first workload:

   First, use a kubeconfig that includes contexts for WDS spaces and set its current context to a target WDS space (e.g., `wds1`):

   ```bash 
   kubectl config use-context wds1
   ```

   Output:
   ```console 
   Switched to context "wds1".
   ```
   
   Optionally, check the kubeconfig contexts: 
   ```bash 
   kubectl config get-contexts
   ```
   
   Output:
   ```console 
   CURRENT   NAME            CLUSTER         AUTHINFO        NAMESPACE
             cluster1        kind-cluster1   kind-cluster1   
             cluster2        kind-cluster2   kind-cluster2   
             its1            its1-cluster    its1-admin      
             kind-kubeflex   kind-kubeflex   kind-kubeflex   
   *         wds1            wds1-cluster    wds1-admin      default
             wds2            wds2-cluster    wds2-admin      default
   ```

   Second, create the binding Policy and namespace for your experiment using the following commands:

   ```bash 
   cd $CL2_DIR/testing/load/
   kubectl apply -f long-duration-exp-setup.yaml
   ```

   Use a command of the following form to create the workload. The value given to the `--kubeconfig` flag should be a pathname of the kubeconfig file used above. This command uses the newly added KubeStellar provider (`--provider=ks`):

   ```bash
   cd $CL2_DIR
   go run cmd/clusterloader.go --testconfig=./testing/load/long-duration-exp-config.yaml --kubeconfig=${KUBECONFIG:-$HOME/.kube/config} --provider=ks --v=2
   ```

   The generated workload is a pod that sleeps for 20 seconds ([see here](workloads/long-duration-exp-workload.yaml)). Proceed to the next step below and we recommend to execute it in parallel to step 4, if you want to maintain the minimal number of created workload objects in your environment at all times. Beware that delaying the execution of step 5 leads to an accumulation of created workload objects, and potentially overloading the system if the delay continues for an extended period.

5. Open a new terminal and run the metrics collection script to collect the creation and status_update timestamps for the benchmark workload objects:

   a) cd into the `test/performance/common` directory from your local copy of the KubeStellar repo, for example:
   
   ```bash 
   cd $HOME/kubestellar/test/performance/common
   ```

   b) Run the metrics collection script:

   ```bash
   python3 metrics_collector.py <kubeconfig> <wds-context-name> <its-context-name> <wec-context-name> <number-of-namespaces> <output-directory> <exp-type> <number-of-pods> <watch-interval>
   ```

   For example:
   
   ```bash 
   python3 metrics_collector.py $HOME/.kube/config wds1 its1 cluster1 1 $HOME/data l 10 660
   ```

   Observations: 
      - The script creates a custom controller that deletes a pod after reaching the completed state.
      - The workload generation tool (i.e., clusterloader2) launched at step 4 will create another pod after `stepDelay` (default value: 1 minute) with a different name. 
      - The above procedure guarantees that the running system is not overloaded. 

   Below is a detailed explanation of the input parameters:
   - `kubeconfig`: path to the kubeconfig file, e.g., `$HOME/.kube/config`
   - `wds-context-name`: name of the context for the target WDS, e.g., `wds1`
   - `its-context-name`: name of the context for the target ITS, e.g., `its1`
   - `wec-context-name`: name of the context for the target WEC, e.g., `cluster1`
   - `number-of-namespaces`: it must be set to `1` always. There is only one namespace created: `perf-test`. 
   - `output-directory`: path to the directory for the output data files, e.g., `$HOME/data`
   - `exp-type`: experiment type (e.g., `l`)
   - `number-of-pods`: total number of pods to be created in the experiment. The same value configured at step 3-b. This script runs until it has collected data for this number of Pods.
   - `watch-interval`: time duration in seconds of your experiment (e.g., `660`). The same value configured at step 3-b.

   The creation and status_update timestamps will be collected for the workload objects created in `wds1`, manifestwork & workstatus objects created in `its1`, and workload & appliedmanifestworks objects created in `wec1`, for example:

   ```bash  
   cd $HOME/data
   tree
   ```

   Output:
   ```console  
      .
      ├── 10-appliedmanifestwork.csv
      ├── 10-manifestworks.csv
      ├── 10-pod-wds.csv
      ├── 10-pod-wec.csv
      └── 10-workstatus.csv
   ```

   The generated output from the script are tab-delimited files with the following structure:

   ```console 
   <obj-name> <obj-creation-time> <obj-status-update-time> <obj-status-condition> <obj-controller-manager>
   ```
   
   Where: `<obj-controller-manager>` - it is the FieldManager value used by the controller that updated the status of the workload object.

   For example: 

   ```console 
   stress-pod-2	2024-03-14 19:51:19+00:00	2024-03-14 19:51:44+00:00	Succeeded	controller-manager
   ```

6. Clean up the generated workload Kubernetes API objects from your cluster:

   ```bash 
   cd $CL2_DIR
   ./cleanup.sh
   ```