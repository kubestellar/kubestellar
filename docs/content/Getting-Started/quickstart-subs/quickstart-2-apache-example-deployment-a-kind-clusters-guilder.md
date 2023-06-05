<!--quickstart-2-apache-example-deployment-a-kind-clusters-guilder-start-->
Create the second edge cluster:

```shell
kind create cluster --name guilder --config - <<EOF
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

Note: if you already have a cluster named 'guilder' from a previous exercise of KubeStellar, please delete the guilder cluster ('kind delete cluster --name guilder') and create it using the instruction above.
<!--quickstart-2-apache-example-deployment-a-kind-clusters-guilder-end-->