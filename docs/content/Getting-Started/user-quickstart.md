
[![QuickStart test]({{config.repo_url}}/actions/workflows/docs-ecutable-qs.yml/badge.svg?branch={{config.ks_branch}})]({{config.repo_url}}/actions/workflows/docs-ecutable-qs.yml)&nbsp;&nbsp;&nbsp;

<!-- 
!!! tip "Estimated time to complete this example:" 
    ~4 minutes (after installing prerequisites) -->

## How to deploy and use KubeStellar

This guide will show how to:

1. quickly deploy the KubeStellar Core component on a kind cluster using helm, 
2. install the KubeStellar user commands and kubectl plugins on your computer with brew,
3. retrieve the KubeStellar Core component kubeconfig, 
4. install the KubeStellar Syncer component on two edge clusters, 
5. deploy an example kubernetes workload to both edge clusters from KubeStellar Core,
6. view the status of your deployment across both edge clusters from KubeStellar Core

**important:** For this quickstart you will need to know how to use kubernetes' kubeconfig *context* to access multiple clusters.  You can learn more about kubeconfig context [here](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/)

!!! tip ""
    === "Pre-reqs"
        + [__kubectl__](https://kubernetes.io/docs/tasks/tools/) (version range expected: 1.24-1.26)

        + [__helm__](https://helm.sh/docs/intro/install/) - to deploy the kubestellar-core helm chart
        
        + [__brew__](https://helm.sh/docs/intro/install/) - to install the kubestellar user commands and kubectl plugins
        
        + [__kind__](https://kind.sigs.k8s.io) - to create a few small kubernetes clusters

        + 3 kind clusters
        
        {%
          include-markdown "../common-subs/create-ks-core-kind-cluster.md"
          start="<!--create-ks-core-kind-cluster-start-->"
          end="<!--create-ks-core-kind-cluster-end-->"
        %}

        {%
          include-markdown "../common-subs/create-ks-edge-cluster1-kind-cluster.md"
          start="<!--create-ks-edge-cluster1-kind-cluster-start-->"
          end="<!--create-ks-edge-cluster1-kind-cluster-end-->"
        %}

        {%
          include-markdown "../common-subs/create-ks-edge-cluster2-kind-cluster.md"
          start="<!--create-ks-edge-cluster2-kind-cluster-start-->"
          end="<!--create-ks-edge-cluster2-kind-cluster-end-->"
        %}
    === "uh oh, error?"
        if you apply the ingress and then receive an error while waiting:
          `error: no matching resources found`

        this might mean that you did not wait long enough before issuing the check command. Simply try the check command again.
   
#### 1. Deploy your KubeStellar Core component
deploy the KubeStellar Core components on the **ks-core** kind cluster you created in the pre-req section above  
{%
  include-markdown "../common-subs/deploy-your-kubestellar-core-component.md"
  start="<!--deploy-your-kubestellar-core-component-start-->"
  end="<!--deploy-your-kubestellar-core-component-end-->"
%}

#### 2. Install KubeStellar's user commands and kubectl plugins

{%
   include-markdown "../common-subs/install-brew.md"
   start="<!--install-brew-start-->"
   end="<!--install-brew-end-->"
%}

#### 3. View your KubeStellar Core Space environment
!!! tip ""
    === "show all available KubeStellar Core Spaces"
         ```
         KUBECONFIG=~/.kube/config kubectl --context kind-ks-core get secrets kubestellar \
           -o jsonpath='{.data.external\.kubeconfig}' \
           -n kubestellar | base64 -d > ks-core.kubeconfig

         KUBECONFIG=ks-core.kubeconfig kubectl ws --context root tree
         ```
    === "uh oh, error?"
         Did you received the following error:
         ```Error: Get "https://some_hostname.some_domain_name:{{config.ks_kind_port_num}}/clusters/root/apis/tenancy.kcp.io/v1alpha1/workspaces": dial tcp: lookup some_hostname.some_domain_name on x.x.x.x: no such host``

         A common error occurs if you set your port number to a pre-occupied port number and/or you set your EXTERNAL_HOSTNAME to something other than "localhost" so that you can reach your KubeStellar Core from another host, check the following:
         
         Check if the port specified in the **ks-core** kind cluster configuration and the EXTERNAL_PORT helm value are occupied by another application:

         &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;1. is the port specified in this example occupied by another process?  If so, delete the **ks-core** kind cluster and create it again using an available port for your 'hostPort' value

         &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;2. if you change the port for your **ks-core** 'hostPort', remember to also use that port as the helm 'EXTERNAL_PORT' value

         Check that your EXTERNAL_HOSTNAME helm value is reachable via DNS:

         &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;1. use 'nslookup <value of EXTERNAL_HOSTNAME>' to make sure there is a valid IP address associated with the hostname you have specified

         &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;2. make sure your EXTERNAL_HOSTNAME and associated ip address are listed in your /etc/hosts file.

         &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;3. make sure the IP address is associated with the system where you have deployed the **ks-core** kind cluster

         if there is nothing obvious, [open a bug report and we can help you out](https://github.com/kubestellar/kubestellar/issues/new?assignees=&labels=kind%2Fbug&projects=&template=bug_report.yaml&title=bug%3A+)

    === "open a bug report"
        Stuck? [open a bug report and we can help you out](https://github.com/kubestellar/kubestellar/issues/new?assignees=&labels=kind%2Fbug&projects=&template=bug_report.yaml&title=bug%3A+)


#### 4. Install KubeStellar Syncers on your Edge Clusters
prepare KubeStellar Syncers, with `kubestellar prep-for-cluster`, for **edge-cluster1** and **edge-cluster2** and then apply the files that `kubestellar prep-for-cluster`` prepared for you

**important:** make sure you created kind clusters for **edge-cluster1** and **edge-cluster2** from the pre-req step above before proceeding [how-to-deploy-and-use-kubestellar](#how-to-deploy-and-use-kubestellar)

!!! tip ""
    === "Prep and apply"
        ``` hl_lines="3 7"
        KUBECONFIG=ks-core.kubeconfig kubectl kubestellar prep-for-cluster --imw root:imw1 edge-cluster1 \
          env=edge-cluster1 \
          location-group=edge     #add edge-cluster1 and edge-cluster2 to the same group

        KUBECONFIG=ks-core.kubeconfig kubectl kubestellar prep-for-cluster --imw root:imw1 edge-cluster2 \
          env=edge-cluster2 \
          location-group=edge     #add edge-cluster1 and edge-cluster2 to the same group
        ```

        ```
        export KUBECONFIG=~/.kube/config

        #apply edge-cluster1 syncer
        kubectl --context kind-edge-cluster1 apply -f edge-cluster1-syncer.yaml
        sleep 3
        kubectl --context kind-edge-cluster1 get pods -A | grep kubestellar  #check if syncer deployed to edge-cluster1 correctly

        #apply edge-cluster2 syncer
        kubectl --context kind-edge-cluster2 apply -f edge-cluster2-syncer.yaml
        sleep 3
        kubectl --context kind-edge-cluster2 get pods -A | grep kubestellar  #check if syncer deployed to edge-cluster2 correctly
        ```

#### 5. Create and deploy an Apache Server to edge-cluster1 and edge-cluster2

{%
   include-markdown "quickstart-subs/quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters-user-qs.md"
   start="<!--quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters-start-->"
   end="<!--quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters-end-->"
%}

#### 6. Check the status of your Apache Server on edge-cluster1 and edge-cluster2

```
TODO
```

what's next...  
how to upsync a resource  
how to create, but not overrite/update a synchronized resource  

<br>
---

---
<br>

## How to use an existing KubeStellar environment

## 1. Install KubeStellar's user commands and kubectl plugins

{%
   include-markdown "../common-subs/install-brew.md"
   start="<!--install-brew-start-->"
   end="<!--install-brew-end-->"
%}


## 2. View your KubeStellar Core Space environment

!!! tip ""
    === "show all available KubeStellar Core Spaces"
         ```
         KUBECONFIG=~/.kube/config kubectl --context kind-ks-core get secrets kubestellar \
           -o jsonpath='{.data.external\.kubeconfig}' \
           -n kubestellar | base64 -d > ks-core.kubeconfig

         KUBECONFIG=ks-core.kubeconfig kubectl ws --context root tree
         ```
    === "uh oh, error?"
         Did you received the following error:
         ```Error: Get "https://some_hostname.some_domain_name:{{config.ks_kind_port_num}}/clusters/root/apis/tenancy.kcp.io/v1alpha1/workspaces": dial tcp: lookup some_hostname.some_domain_name on x.x.x.x: no such host``

         A common error occurs if you set your port number to a pre-occupied port number and/or you set your EXTERNAL_HOSTNAME to something other than "localhost" so that you can reach your KubeStellar Core from another host, check the following:
         
         Check if the port specified in the **ks-core** kind cluster configuration and the EXTERNAL_PORT helm value are occupied by another application:

         &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;1. is the port specified in this example occupied by another process?  If so, delete the **ks-core** kind cluster and create it again using an available port for your 'hostPort' value

         &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;2. if you change the port for your **ks-core** 'hostPort', remember to also use that port as the helm 'EXTERNAL_PORT' value

         Check that your EXTERNAL_HOSTNAME helm value is reachable via DNS:

         &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;1. use 'nslookup <value of EXTERNAL_HOSTNAME>' to make sure there is a valid IP address associated with the hostname you have specified

         &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;2. make sure your EXTERNAL_HOSTNAME and associated ip address are listed in your /etc/hosts file.

         &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;3. make sure the IP address is associated with the system where you have deployed the **ks-core** kind cluster

         if there is nothing obvious, [open a bug report and we can help you out](https://github.com/kubestellar/kubestellar/issues/new?assignees=&labels=kind%2Fbug&projects=&template=bug_report.yaml&title=bug%3A+)

    === "open a bug report"
        Stuck? [open a bug report and we can help you out](https://github.com/kubestellar/kubestellar/issues/new?assignees=&labels=kind%2Fbug&projects=&template=bug_report.yaml&title=bug%3A+)

## 3. Create and deploy an Apache Server to edge-cluster1 and edge-cluster2

{%
   include-markdown "quickstart-subs/quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters-user-qs.md"
   start="<!--quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters-start-->"
   end="<!--quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters-end-->"
%}

## 4. Check the status of your Apache Server on edge-cluster1 and edge-cluster2

```
TODO
```