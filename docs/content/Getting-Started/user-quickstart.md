
[![QuickStart test]({{config.repo_url}}/actions/workflows/docs-ecutable-qs.yml/badge.svg?branch={{config.ks_branch}})]({{config.repo_url}}/actions/workflows/docs-ecutable-qs.yml)&nbsp;&nbsp;&nbsp;


!!! tip "Estimated time to complete this example:" 
    ~4 minutes (after installing prerequisites)

## How to deploy and use KubeStellar

This guide is intended to show how to (1) quickly bring up a **KubeStellar** environment using helm, (2) install plugins with brew, (3) retrieve your kubeconfig, (4) install KubeStellar Syncers on 2 edge clusters, (5) and deploy an example workload to both edge clusters.

## 0. Pre-reqs

- helm
- brew
- kubectl
- kind clusters (3) 
   - 1 KubeStellar core cluster (kind-core) [instructions](https://docs.kubestellar.io/main/Coding%20Milestones/PoC2023q1/environments/dev-env/#hosting-kubestellar-in-a-kind-cluster)
   - 2 KubeStellar edge clusters (kind-edge-cluster1, kind-edge-cluster2)

```

   kind create cluster --name edge-cluster1 --config - <<EOF
      kind: Cluster
      apiVersion: kind.x-k8s.io/v1alpha4
      nodes:
      - role: control-plane
      extraPortMappings:
      - containerPort: 8081
         hostPort: 8094
      EOF

   kind create cluster --name edge-cluster2 --config - <<EOF
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
KUBECONFIG=~/.kube/config kubectl use-context kind-core
kubectl create namespace kubestellar
helm install kubestellar --set EXTERNAL_HOSTNAME="$(hostname -f)" --set EXTERNAL_PORT=1024 --namespace kubestellar
```

-or-

```
oc login
oc new-project kubestellar
helm install kubestellar --set clusterType=OpenShift
```

## 2. Install KubeStellar's Kubectl plugins

```
brew tap kubestellar/kubestellar
brew install kcp_cli
brew install kubestellar_cli {{config.ks_tag}}
```

## 3. Access KubeStellar

```
kubectl get secrets kubestellar -n kubestellar -o jsonpath='{.data.external\.kubeconfig}' | base64 -d > > admin-core.kubeconfig
```

## 4. Install KubeStellar Syncers on your Edge Clusters

```
KUBECONFIG=~/.kube/config kubectl --context kind-core kubestellar prep-for-cluster --imw root:imw1 edge-cluster1 env=edge-cluster1
KUBECONFIG=~/.kube/config kubectl --context kind-core kubestellar prep-for-cluster --imw root:imw1 edge-cluster2 env=edge-cluster2
```

change your kubeconfig to point at edge-cluster1 and edge-cluster2 and apply the files that prep-for-cluster prepared for you

```
KUBECONFIG=~/.kube/config kubectl --context kind-edge-cluster1 apply -f edge-cluster1-syncer.yaml
KUBECONFIG=~/.kube/config kubectl --context kind-edge-cluster2 apply -f edge-cluster2-syncer.yaml
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
brew install kubestellar_cli {{config.ks_tag}}
```

## 2. Access KubeStellar

```
KUBECONFIG=~/.kube/config kubectl --context kind-core get secrets kubestellar -n kubestellar -o jsonpath='{.data.external\.kubeconfig}' | base64 -d > > admin-core.kubeconfig
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