<!--user-example1-pre-start-->

### Create two kind clusters

This example uses two [kind](https://kind.sigs.k8s.io/) clusters as
workload execution clusters (WEC).  We will call them "ren" and "stimpy". 

- WEC Ren will host a single workload (a kubernetes namespace, configset, 
a customizer object, and a replicaset with nginx inside), and 
- WEC Stimpy will
host two workloads similar in composition to Ren's.

```shell
kind create cluster --name ren --config - <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 8081
    hostPort: 8094
EOF

kind create cluster --name stimpy --config - <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 8081
    hostPort: 8096
  - containerPort: 8082
    hostPort: 8097
EOF
```
<!--user-example1-pre-end-->