<!--quickstart-2-apache-example-deployment-a-kind-clusters-start-->
### a. Stand up two kind clusters: florin and guilder

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
<!--quickstart-2-apache-example-deployment-a-kind-clusters-end-->