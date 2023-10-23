<!--create-ks-edge-cluster1-kind-cluster-start-->
create the **ks-edge-cluster1** kind cluster
```shell
KUBECONFIG=~/.kube/config kind create cluster --name ks-edge-cluster1 --config - <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 8081
    hostPort: 8094
EOF
```
<!--create-ks-edge-cluster1-kind-cluster-end-->
