<!--create-ks-edge-cluster2-kind-cluster-start-->
create the **edge-cluster2** kind cluster
```shell
export KUBECONFIG=~/.kube/config
kind create cluster --name edge-cluster2 --config=- <<EOF
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
<!--create-ks-edge-cluster2-kind-cluster-end-->
