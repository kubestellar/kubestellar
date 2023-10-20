
[![QuickStart test]({{config.repo_url}}/actions/workflows/docs-ecutable-qs.yml/badge.svg?branch={{config.ks_branch}})]({{config.repo_url}}/actions/workflows/docs-ecutable-qs.yml)&nbsp;&nbsp;&nbsp;

<!-- 
!!! tip "Estimated time to complete this example:" 
    ~4 minutes (after installing prerequisites) -->

## How to deploy and use KubeStellar

This guide will show how to:

1. quickly deploy the KubeStellar Core component on a kind cluster using helm, 
2. install the KubeStellar user commands and kubectl plugins on your computer with brew,
3. retrieve the KubeStellar Core component kubeconfig, 
4. install the KubeStellar Syncer component on 2 edge clusters, 
5. deploy an example kubernetes workload to both edge clusters from KubeStellar Core,
6. view the status of your deployment across both edge clusters from KubeStellar Core

NOTE: For this quickstart you will need to know how to use kubernetes' kubeconfig *context* to access multiple clusters.  You can learn more about kubeconfig context [here](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/)

!!! tip "Pre-reqs"
    === "General"
        + [__kubectl__](https://kubernetes.io/docs/tasks/tools/) (version range expected: 1.24-1.26)

        + [__helm__](https://helm.sh/docs/intro/install/) - to deploy the kubestellar-core helm chart
        
        + [__brew__](https://helm.sh/docs/intro/install/) - to install the kubestellar user commands and kubectl plugins
        
        + [__kind__](https://kind.sigs.k8s.io) - to create a few small kubernetes clusters

        + 3 kind clusters (see tabs for 'ks-core', 'edge-cluster1', and 'edge-cluster2' above)
        
    === "ks-core cluster"
        <!-- [instructions](https://docs.kubestellar.io/main/Coding%20Milestones/PoC2023q1/environments/dev-env/#hosting-kubestellar-in-a-kind-cluster) -->
        ```
        kind create cluster --name ks-core --config=- <<EOF
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

        Be sure to apply an ingress control with SSL passthrough to 'ks-core'. This is a special requirement for Kind that allows access to the KubeStellar core running on 'ks-core'.
        ```
        kubectl apply -f https://raw.githubusercontent.com/kubestellar/kubestellar/main/example/kind-nginx-ingress-with-SSL-passthrough.yaml
        ```
        **Wait about 10 seconds** and then check for ingress to be ready:
        ```
        kubectl wait --namespace ingress-nginx \
          --for=condition=ready pod \
          --selector=app.kubernetes.io/component=controller \
          --timeout=90s
        ```

    === "edge-cluster1"
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

    === "edge-cluster2"
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
   
#### 1. Deploy your KubeStellar Core component

!!! tip ""
    === "deploy"
         ```
         # deploy KubeStellar core components on the 'ks-core' kind cluster you created in the pre-req section above
         KUBECONFIG=~/.kube/config kubectl config use-context kind-ks-core
         kubectl create namespace kubestellar
         helm repo add kubestellar https://helm.kubestellar.io
         helm install kubestellar/kubestellar-core --set EXTERNAL_HOSTNAME="$(hostname -f | tr '[:upper:]' '[:lower:]')" --set EXTERNAL_PORT=1024 --namespace kubestellar --generate-name
         ```
    === "wait"
         run the following to wait for KubeStellar to be ready to take requests:
         ```
         echo -n 'Waiting for KubeStellar to be ready'
         while ! kubectl exec $(kubectl get pod --selector=app=kubestellar -o jsonpath='{.items[0].metadata.name}' -n kubestellar) -c init -- ls /home/kubestellar/ready &> /dev/null; do
            sleep 10
            echo -n "."
         done
         echo "KubeStellar is now ready to take requests"
         ```
    === "debug"
         you can also check logs:

         Checking the initialization log to see if there are errors:
         ```
         kubectl logs $(kubectl get pod --selector=app=kubestellar -o jsonpath='{.items[0].metadata.name}' -n kubestellar) -n kubestellar -c init
         ```

#### 2. Install KubeStellar's user commands and kubectl plugins

!!! tip ""
    === "install"
         ```
         brew tap kubestellar/kubestellar
         brew install kcp_cli
         brew install kubestellar_cli@{{config.ks_tag}}
         ```
    === "remove"
         ```
         brew remove kubestellar_cli@{{config.ks_tag}}
         brew remove kcp_cli
         brew untap kubestellar/kubestellar
         ```


#### 3. View your KubeStellar Core Space environment

```
kubectl get secrets kubestellar -n kubestellar -o jsonpath='{.data.external\.kubeconfig}' | base64 -d > ks-core.kubeconfig
KUBECONFIG=ks-core.kubeconfig kubectl ws --context root tree
```

#### 4. Install KubeStellar Syncers on your Edge Clusters
change your kubeconfig context to point at edge-cluster1 and edge-cluster2 and apply the files that prep-for-cluster prepared for you

```
KUBECONFIG=~/.kube/config kubectl config use-context kind-ks-core
kubectl kubestellar prep-for-cluster --imw root:imw1 edge-cluster1 env=edge-cluster1
kubectl kubestellar prep-for-cluster --imw root:imw1 edge-cluster2 env=edge-cluster2
```

```
KUBECONFIG=~/.kube/config kubectl config use-context kind-edge-cluster1
kubectl apply -f edge-cluster1-syncer.yaml

KUBECONFIG=~/.kube/config kubectl config use-context kind-edge-cluster2
kubectl apply -f edge-cluster2-syncer.yaml
```

#### 5. Create and deploy an Apache Server to edge-cluster1 and edge-cluster2

{%
   include-markdown "quickstart-subs/quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters.md"
   start="<!--quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters-start-->"
   end="<!--quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters-end-->"
%}

#### 6. Check the status of your Apache Server on edge-cluster1 and edge-cluster2

```
KUBECONFIG=~/.kube/config kubectl --context kind-edge-cluster1 get deploy,rs -A | egrep 'NAME|stuff'
```

## How to use an existing KubeStellar environment

## 1. Install KubeStellar's user commands and kubectl plugins

```
brew tap kubestellar/kubestellar
brew install kcp_cli
brew install kubestellar_cli@{{config.ks_tag}}
```

## 2. View your KubeStellar Core Space environment

```
kubectl get secrets kubestellar -n kubestellar -o jsonpath='{.data.external\.kubeconfig}' | base64 -d > ks-core.kubeconfig
KUBECONFIG=ks-core.kubeconfig kubectl config use-context root
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