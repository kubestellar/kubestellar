
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
   - 1 KubeStellar core cluster (kind-core)
   - 2 KubeStellar edge clusters (kind-edge-cluster1, kind-edge-cluster2)
   
## 1. Install KubeStellar

```
kubectl cluster-info --context kind-core
helm install kubestellar --set EXTERNAL_HOSTNAME="localhost" --set EXTERNAL_PORT=1024
```
   (NOTE: you can use an FQDN for your EXTERNAL_HOSTNAME if you are using DNS in your environment)

-or-

```
helm install kubestellar --set clusterType=OpenShift
```

## 2. Install KubeStellar's Kubectl plugins

```
brew tap kubestellar/kubestellar {{config.repo_url}}/brew
brew install kubestellar-kubectl {{config.ks_tag}}
```

## 3. Access KubeStellar

```
kubectl get secrets kubestellar -n kubestellar -o jsonpath='{.data.external\.kubeconfig}' | base64 -d > > admin-core.kubeconfig
```

## 4. Install KubeStellar Syncers on your Edge Clusters

```
KUBECONFIG=admin-core.config kubectl kubestellar prep-for-cluster --imw root:imw1 edge-cluster1 env=edge-cluster1
KUBECONFIG=admin-core.config kubectl kubestellar prep-for-cluster --imw root:imw1 edge-cluster2 env=edge-cluster2
```

change your kubeconfig to point at edge-cluster1 and edge-cluster2 and apply the files that prep-for-cluster prepared for you

```
kubectl cluster-info --context kind-edge-cluster1
kubectl apply -f edge-cluster1-syncer.yaml
kubectl cluster-info --context kind-edge-cluster2
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
kubectl cluster-info --context kind-edge-cluster1
kubectl get ...
```

### How to use an existing KubeStellar environment

## 1. Install KubeStellar's Kubectl plugins

```
brew install kubestellar-kubectl v0.X.0
```

## 2. Access KubeStellar

```
kubectl get secrets kubestellar -n kubestellar -o jsonpath='{.data.external\.kubeconfig}' | base64 -d > > admin-core.kubeconfig
```

## 3. Create and deploy an Apache Server to edge-cluster1 and edge-cluster2

{%
   include-markdown "quickstart-subs/quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters.md"
   start="<!--quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters-start-->"
   end="<!--quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters-end-->"
%}

## 4. Check the status of your Apache Server on edge-cluster1 and edge-cluster2

```
kubectl cluster-info --context kind-edge-cluster1
kubectl get ...
```