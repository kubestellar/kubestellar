<!--quickstart-2-apache-example-deployment-a-kind-clusters-florin-start-->
Create the first edge cluster:

```shell
kind create cluster --name florin --config - <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 8081
    hostPort: 8094
EOF
```  

Note: if you already have a cluster named 'florin' from a previous exercise of KubeStellar, please delete the florin cluster ('kind delete cluster --name florin') and create it using the instruction above.
<!--quickstart-2-apache-example-deployment-a-kind-clusters-florin-end-->