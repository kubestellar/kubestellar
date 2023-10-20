
[![QuickStart test]({{config.repo_url}}/actions/workflows/docs-ecutable-qs.yml/badge.svg?branch={{config.ks_branch}})]({{config.repo_url}}/actions/workflows/docs-ecutable-qs.yml)&nbsp;&nbsp;&nbsp;


!!! tip "Estimated time to complete this example:" 
    ~4 minutes (after installing prerequisites)

## How to deploy and use KubeStellar

This guide is intended to show how to (1) quickly bring up a **KubeStellar** environment using helm, (2) install plugins with brew, (3) retrieve the KubeStellar kubeconfig, (4) install KubeStellar Syncers on 2 edge clusters, (5) and deploy an example workload to both edge clusters.

For this quickstart you will need to know how to use kubernetes kubeconfig context to access multiple clusters.  You can learn about it [here](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/)

## 0. Pre-reqs

- helm
- brew
- kubectl
- Kind clusters (3) 
   - 1 KubeStellar core cluster (kind-ks-host) [instructions](https://docs.kubestellar.io/main/Coding%20Milestones/PoC2023q1/environments/dev-env/#hosting-kubestellar-in-a-kind-cluster)
   - 2 KubeStellar edge clusters (kind-edge-cluster1, kind-edge-cluster2)

create kind cluster 'ks-host'
```
kind create cluster --name ks-host --config=- <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 443
    hostPort: 1024
    protocol: TCP
EOF
```

Apply an ingress control with SSL passthrough to 'ks-host'. This is a special requirement for Kind that allows access to the KubeStellar core running on 'ks-host'.
```
kubectl apply -f https://raw.githubusercontent.com/kubestellar/kubestellar/main/example/kind-nginx-ingress-with-SSL-passthrough.yaml
```

create kind cluster 'edge-cluster1'
```
kind create cluster --name edge-cluster1 --config=- <<EOF
   kind: Cluster
   apiVersion: kind.x-k8s.io/v1alpha4
   nodes:
   - role: control-plane
     extraPortMappings:
     - containerPort: 8081
       hostPort: 8094
EOF
```

create kind cluster 'edge-cluster2'

```
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
   
## 1. Install KubeStellar

```
KUBECONFIG=~/.kube/config kubectl config use-context kind-ks-host
helm repo add kubestellar https://helm.kubestellar.io
kubectl create namespace kubestellar
helm install kubestellar/kubestellar-core --set EXTERNAL_HOSTNAME="$(hostname -f | tr '[:upper:]' '[:lower:]')" --set EXTERNAL_PORT=1024 --namespace kubestellar --generate-name
```
<!-- 
-or-

```
oc login
oc new-project kubestellar
helm install kubestellar --set clusterType=OpenShift
``` -->

## 2. Install KubeStellar's Kubectl plugins

```
brew tap kubestellar/kubestellar
brew install kcp_cli
brew install kubestellar_cli@{{config.ks_tag}}
```

## 3. View KubeStellar Space environment

```
kubectl get secrets kubestellar -n kubestellar -o jsonpath='{.data.external\.kubeconfig}' | base64 -d > ks-host.kubeconfig
KUBECONFIG=ks-host.kubeconfig kubectl config use-context root
kubectl ws tree
```

## 4. Install KubeStellar Syncers on your Edge Clusters
change your kubeconfig context to point at edge-cluster1 and edge-cluster2 and apply the files that prep-for-cluster prepared for you

```
KUBECONFIG=~/.kube/config kubectl config use-context kind-ks-host
kubectl kubestellar prep-for-cluster --imw root:imw1 edge-cluster1 env=edge-cluster1
kubectl kubestellar prep-for-cluster --imw root:imw1 edge-cluster2 env=edge-cluster2
```

```
KUBECONFIG=~/.kube/config kubectl config use-context kind-edge-cluster1
kubectl apply -f edge-cluster1-syncer.yaml
kubectl apply -f edge-cluster2-syncer.yaml
```

## 5. Create and deploy an Apache Server to edge-cluster1 and edge-cluster2

{%
   include-markdown "quickstart-subs/quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters.md"
   start="<!--quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters-start-->"
   end="<!--quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters-end-->"
%}

## 6. Check the status of your Apache Server on edge-cluster1 and edge-cluster2

```
KUBECONFIG=~/.kube/config kubectl --context kind-edge-cluster1 get deploy,rs -A | egrep 'NAME|stuff'
```

### How to use an existing KubeStellar environment

## 1. Install KubeStellar's Kubectl plugins

```
brew tap kubestellar/kubestellar
brew install kcp_cli
brew install kubestellar_cli@{{config.ks_tag}}
```

## 2. View KubeStellar Space environment

```
kubectl get secrets kubestellar -n kubestellar -o jsonpath='{.data.external\.kubeconfig}' | base64 -d > ks-host.kubeconfig
KUBECONFIG=ks-host.kubeconfig kubectl config use-context root
kubectl ws tree
```

## 3. Create and deploy an Apache Server to edge-cluster1 and edge-cluster2

{%
   include-markdown "quickstart-subs/quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters.md"
   start="<!--quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters-start-->"
   end="<!--quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters-end-->"
%}

## 4. Check the status of your Apache Server on edge-cluster1 and edge-cluster2

```
KUBECONFIG=~/.kube/config kubectl --context kind-edge-cluster1 get deploy,rs -A | egrep 'NAME|stuff'
```