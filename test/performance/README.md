## Stress Test Experiments


## Generate the workload traffic

1. Clone the following fork of the clusterloader2 repo: 

```bash
$ git clone -b ks-provider https://github.com/dumb0002/perf-tests.git
```

2. Configure the parameters of your experiment:  

   a) cd into the load configuration directory

```bash
$ cd  perf-tests/clusterloader2/testing/load
```
   b) create the `stress-test` namespace - you workload needs to be deployed into this exact namespace (temporary solution)

```bash
$ kubectl apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: stress-test
  labels:
    app.kubernetes.io/cluster: wec1
EOF
```

   c) configure your sample workload: 
   
```bash
$ vi test-workload.yaml
```
  
   d) configure the parameters to create the workload traffic (e.g., RandomizedLoad, RSteppedLoad, Sequence, etc.)
   
```bash
$ vi performance-test-config.yaml
``` 

To learn more about the clusterloader2 load experiments configuration paramaters, see the following: [clusterloader tunningSets](https://github.com/kubernetes/perf-tests/blob/master/clusterloader2/docs/design.md#tuning-set)


3. Run your first workload:


Set ``--kubeconfig`` flag to the path of the kubeconfig file of your cluster (e.g., WDS kubeconfig) and use the newly added Kubestellar provider

```bash
$ cd perf-tests/clusterloader2/
$ go run cmd/clusterloader.go --testconfig=./testing/load/performance-test-config.yaml --kubeconfig=<path>/wds-kubeconfig --provider=ks --v=2
```


## Collect the measurements 

To collect the creation and status update for a k8s object proceed as following:  

1. Create a kubeconfig with contexts for the `wds1`, `imbs1` and `wec1`:

For example: 

```bash
$ kubectl config get-contexts
CURRENT   NAME     CLUSTER                                          AUTHINFO                                                                 NAMESPACE
          imbs1    imbs1-cluster                                    imbs1-admin                                                               default
          wec1     <url>:port#                                      <defaul-value>                                                            default
*         wds1     wds1-cluster                                     wds1-admin                                                                default
```

Use the following command to rename the default context name for `wec1` workload execution cluster:

```bash 
$ kubectl config rename-context <default-wec1-context-name> wec1
```

2. Configure the following parameters in the `main` function in the `metrics_collectory.py` script

```bash
    kubeconfig = "kscore-config"  # path to the kubeconfig file
    ns = "stress-test"  # name of the namespace where the workload are generated 
    mypath="data/"  # path to the directory for the output files
    run_ID="1" # run ID for your experiment 
    freq="1"   # workload frequency parameter set in your experiment 
    numPods=20 # total number of workload objects set in your experiment
    method = "exp-RandomizedLoad" # clusterloader method used to generate your traffic
    watch_interval = 3600  # watch for events that to occur within time interval threshold
```

3. Run the metrics collection script:

a) Clone the KubeStellar repo:

```bash
$ git clone https://github.com/kubestellar/kubestellar.git
$ cd kubestellar/test/performance
```

b) Start collecting metrics:

```bash 
$ python3 metrics_collector.py
```

Metrics will be collected for pods in `wds1`, manifestworks & workstatuses in `imbs1` and pods & appliedmanifestworks in `wec1` - In total 5 files will be output, for example:

```bash 
      20 exp-RandomizedLoad-freq-20-appliedmanifestwork-wec1-run1.txt
      20 exp-RandomizedLoad-freq-20-manifestwork-imbs1-run1.txt
      20 exp-RandomizedLoad-freq-20-pod-wds-kscore-stage-run1.txt
      20 exp-RandomizedLoad-freq-20-pod-wec1-run1.txt
      20 exp-RandomizedLoad-freq-20-workstatus-imbs1-run1.txt
```

The output is a file with the following structure:

```bash
<obj_name> <obj_creation_time> <obj_status_update_time> <obj_status_condition>  <obj_status_condition>  <obj_controller_manager>
```

For example: 

```bash
stress-pod-2	2024-03-14 19:51:19+00:00	2024-03-14 19:51:44+00:00	Succeeded	controller-manager
```