## Workload Benchmark for KubeStellar

To generate the sample workload for KubeStellar performance experiments proceed as following:

### Generate the workload traffic

1. Clone the following clusterloader2 repo: 

   ```console 
   git clone -b release-1.31 https://github.com/kubernetes/perf-tests.git
   ```

   Set the variable `CL2_DIR` to the path of the subdirectory `clusterloader2/` of the cloned repo. For example: 

   ```console 
   export CL2_DIR=$HOME/perf-tests/clusterloader2
   ```

2. Configure clusterloader2 to generate KS performance workloads:

   Starting from a local directory containing this KubeStellar git repo, run the following script based on your cluster environment:

   a) cd into the `test/performance/common` directory from your local copy of the KubeStellar repo, for example:

   ```console 
   cd .$HOME/kubestellar/test/performance/common
   ```

   b) set-up your environment:

   ```console 
   ./setup-clusterloader2.sh --exp l
   ```

3. Configure the parameters of your workload:  

   a) cd into the load configuration directory

   ```console 
   cd  $CL2_DIR/testing/load/
   ```
  
   b) configure the parameters to create the workload traffic (e.g., RandomizedLoad, SteppedLoad, Sequence, etc.)
   
   ```console 
   vi long-duration-exp-config.yaml
   ``` 

   More specifically, configure the following parameter: 

   - numberOfWorkloadObjects: total number of objects to be created in the experiment `perf-test` namespace (*default value: "10"*)
  
   To learn more about clusterloader2 tuningSet, see the following: [tuningSets](https://github.com/kubernetes/perf-tests/blob/master/clusterloader2/docs/design.md#tuning-set) and [configurations](https://github.com/kubernetes/perf-tests/blob/fac2a5eec96fab76a4bc4858795df4544b729b0b/clusterloader2/api/types.go#L249).


4. Deploy your first workload:

   First, use a kubeconfig that includes contexts for WDS spaces and set its current context to a target WDS space (e.g., `wds1`):

   ```console 
   kubectl config use-context wds1
   ```

   Output:
   ```console 
   Switched to context "wds1".
   ```
   
   Optionally, check the kubeconfig contexts: 
   ```console 
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

   ```console 
   cd $CL2_DIR/testing/load/
   kubectl apply -f long-duration-exp-setup.yaml
   ```

   Lastly, in the following set ``--kubeconfig`` flag to the path of the kubeconfig file of the cluster and use the newly added Kubestellar provider (`--provider=ks`) and run the following command to create the workload:

   ```console 
   cd $CL2_DIR
   go run cmd/clusterloader.go --testconfig=./testing/load/long-duration-exp-config.yaml --kubeconfig=<path>/wds-kubeconfig --provider=ks --v=2
   ```
    
   The generated workload is a pod that sleeps for 20 seconds([see here](workloads/long-duration-exp-workload.yaml)) and created using the clusterloader `RSteppedLoad` tuningSet (burstSize=1 and stepDelay=60 sec) - tuningSet is a type of workload generator function used by clusterloader2 (e.g., RandomizedLoad, QpsLoad, SteppedLoad, etc). 


5. Run the metrics collection script: collect the creation and status_update timestamps for the benchmark workload objects proceed as following

   a) cd into the `test/performance/common` directory from your local copy of the KubeStellar repo, for example:

   ```console 
   cd .$HOME/kubestellar/test/performance/common
   ```

   b) Run the metrics collection script:

   ```console 
   python3 metrics_collector.py <kubeconfig> <wds-context-name> <its-context-name> <wec-context-name> <number-of-namespaces> <output-directory> <exp-type> <number-of-pods> <watch-interval>
   ```

   For example:
   
   ```console 
   python3 metrics_collector.py $HOME/.kube/config wds1 its1 cluster1 1 $HOME/data long_duration 10 30
   ```

   Observations: 
      - The execution of the above python script requires the installation of the dependencies listed in `requirements.txt`.
      - The script creates a custom controller that deletes a pod after reaching the completed state 
      - The workload generation tool (i.e., clusterloader2) launched at `step-4` will create another pod after 1 minute with a different name.

   Below is a detailed explanation of the input parameters:
   - `kubeconfig`: path to the kubeconfig file, e.g., `$HOME/.kube/config`
   - `wds-context-name`: name of the context for the target WDS, e.g., `wds1`
   - `its-context-name`: name of the context for the target ITS, e.g., `its1`
   - `wec-context-name`: name of the context for the target WEC, e.g., `cluster1`
   - `number-of-namespaces`: number of namespaces created in your experiment, e.g., `1`
   - `output-directory`: path to the directory for the output data files, e.g., `$HOME/data`
   - `exp-type`: experiment type (e.g., `short_duration` or `long_duration`)
   - `number-of-pods`: total number of pods to be created in the experiment
   - `watch-interval`: time duration in seconds of your experiment (e.g., `30`)

   The creation and status_update timestamps will be collected for the workload objects created in `wds1`, manifestwork & workstatus objects created in `its1`, and workload & appliedmanifestworks objects created in `wec1`, for example:

   ```console  
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

   The generated output files have the following structure:

   ```console 
   <obj-name> <obj-creation-time> <obj-status-update-time> <obj-status-condition> <obj-controller-manager>
   ```
   
   Where: `<obj-controller-manager>` - it is the FieldManager value used by the controller that updated the status of the workload object.

   For example: 

   ```console 
   stress-pod-2	2024-03-14 19:51:19+00:00	2024-03-14 19:51:44+00:00	Succeeded	controller-manager
   ```

6. Clean up the generated workload Kubernetes API objects from your cluster:

   ```console 
   cd $CL2_DIR
   ./cleanup.sh
   ```