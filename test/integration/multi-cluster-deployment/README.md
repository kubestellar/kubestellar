# Kubestellar multi-cluster workload deployment with kubectl

## Prerequisites
See the list [here](https://github.com/kubestellar/kubestellar/blob/ks-0.20/test/pre-reqs.md).

## Running the test using a script
- Clone the repo and checkout the ks-0.20 branch
```
git clone git@github.com:kubestellar/kubestellar.git
cd kubestellar
git checkout -b ks-0.20 origin/ks-0.20
```
- Run the test
```
cd test/integration/multi-cluster-deployment
./run-test.sh
```

## Running the test steps manually
### Setup Kubestellar
- Clone the repo and checkout the ks-0.20 branch
```
git clone git@github.com:kubestellar/kubestellar.git
git branch -b ks-0.20 origin/ks-0.20
cd kubestellar
```

Create a Kind hosting cluster with nginx ingress controller and KubeFlex operator.
```
kflex init --create-kind
```

Create an inventory & mailbox space of type vcluster running OCM (Open Cluster Management) directly in KubeFlex. Note that -p ocm runs a post-create hook on the vcluster control plane which installs OCM on it.
```
kflex create imbs1 --type vcluster -p ocm
```

Create a Workload Description Space wds1 directly in KubeFlex.
```
kflex create wds1
kubectl config use-context kind-kubeflex
kubectl label cp wds1 kflex.kubestellar.io/cptype=wds

cd ../../../
make ko-build
make install-local-chart
cd -
```

Create clusters and register with OCM.
```
kind create cluster --name cluster1
kubectl config rename-context kind-cluster1 cluster1
clusteradm --context imbs1 get token | grep '^clusteradm join' | sed "s/<cluster_name>/cluster1/" | awk '{print $0 " --context 'cluster1' --force-internal-endpoint-lookup"}' | sh

kind create cluster --name cluster2
kubectl config rename-context kind-cluster2 cluster2
clusteradm --context imbs1 get token | grep '^clusteradm join' | sed "s/<cluster_name>/cluster2/" | awk '{print $0 " --context 'cluster2' --force-internal-endpoint-lookup"}' | sh

kubectl config use-context imbs1
```

Wait for csr on imbs1. You might need to wait a few seconds before the resource is created.
```
kubectl --context imbs1 get csr 
```

Once you see the csr object accept the clusters.
```
clusteradm --context imbs1 accept --clusters cluster1
clusteradm --context imbs1 accept --clusters cluster2
```

Label the clusters as edge.
```
kubectl --context imbs1 get managedclusters
kubectl --context imbs1 label managedcluster cluster1 location-group=edge
kubectl --context imbs1 label managedcluster cluster2 location-group=edge
```

List all deployments and statefulsets running in the hosting cluster.
```
kubectl --context kind-kubeflex get deployments,statefulsets --all-namespaces
```

List available clusters with label location-group=edge and check there are two of them.
```
kubectl --context imbs1 get managedclusters -l location-group=edge | tee out
if (("$(wc -l < out)" != "3")); then
  echo "Failed to see two clusters."
  exit 1
fi
```

### Run the workload 
Create a placement to deliver an app to all clusters in wds1.
This placement configuration determines where to deploy the workload by using the label selector expressions found in clusterSelectors. It also specifies what to deploy through the downsync.objectSelectors expressions. When there are multiple matchLabels expressions, they are combined using a logical AND operation. Conversely, when there are multiple objectSelectors, they are combined using a logical OR operation.
```
kubectl --context wds1 apply -f - <<EOF
apiVersion: edge.kubestellar.io/v1alpha1
kind: Placement
metadata:
  name: nginx-placement
spec:
  clusterSelectors:
  - matchLabels: {"location-group":"edge"}
  downsync:
  - objectSelectors:
    - matchLabels: {"app.kubernetes.io/name":"nginx"}
EOF
```

Deploy the app.
```
kubectl --context wds1 apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  labels:
    app.kubernetes.io/name: nginx
  name: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: nginx
  labels:
    app.kubernetes.io/name: nginx
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: public.ecr.aws/nginx/nginx:latest
        ports:
        - containerPort: 80
EOF
```

Verify that manifestworks wrapping the objects have been created in the mailbox namespaces.
```
kubectl --context imbs1 get manifestworks -n cluster1 | tee out
kubectl --context imbs1 get manifestworks -n cluster2 | tee -a out
if (("$(wc -l < out)" != "6")); then
  echo "Failed to see expected manifestworks."
  exit 1
fi
```

Verify that the deployment has been created in both clusters. It can take a few seconds for both deployments to be propegated so you might need to issue these commands several times.
```
kubectl --context cluster1 get deployments nginx-deployment -n nginx 
kubectl --context cluster2 get deployments nginx-deployment -n nginx 
```

### Cleanup
```
kind delete cluster --name cluster1
kind delete cluster --name cluster2
kind delete cluster --name kubeflex
kubectl config delete-context cluster1
kubectl config delete-context cluster2
rm out
```


