---
short_name: user-quickstart-kind
manifest_name: 'docs/content/Getting-Started/user-quickstart-kind.md'
---
[![User QuickStart Kind test]({{config.repo_url}}/actions/workflows/docs-ecutable-user-quickstart-kind.yml/badge.svg?branch={{config.ks_branch}})]({{config.repo_url}}/actions/workflows/docs-ecutable-user-quickstart-kind.yml)&nbsp;&nbsp;&nbsp;

<!-- 
!!! tip "Estimated time to complete this example:" 
    ~4 minutes (after installing prerequisites) -->

## How to deploy and use <span class="Space-Bd-BT">KUBESTELLAR</span> on Kind Kubernetes Clusters
!!! tip ""
    === "Goals"
        This guide will show how to:

        1. quickly deploy the <span class="Space-Bd-BT">KUBESTELLAR</span> Core component on a Kind cluster using helm (ks-core), 
        2. install the <span class="Space-Bd-BT">KUBESTELLAR</span> user commands and kubectl plugins on your computer with brew,
        3. retrieve the <span class="Space-Bd-BT">KUBESTELLAR</span> Core component kubeconfig, 
        4. install the <span class="Space-Bd-BT">KUBESTELLAR</span> Syncer component on two edge Kind clusters (ks-edge-cluster1 and ks-edge-cluster2), 
        5. deploy an example kubernetes workload to both edge Kind clusters from <span class="Space-Bd-BT">KUBESTELLAR</span> Core (ks-core),
        6. view the example kubernetes workload running on two edge Kind clusters (ks-edge-cluster1 and ks-edge-cluster2)
        7. view the status of your deployment across both edge Kind clusters from <span class="Space-Bd-BT">KUBESTELLAR</span> Core (ks-core)

        **important:** For this quickstart you will need to know how to use kubernetes' kubeconfig *context* to access multiple clusters.  You can learn more about kubeconfig context [here](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/)

!!! tip ""
    === "Pre-reqs"
        + [__kubectl__](https://kubernetes.io/docs/tasks/tools/) (version range expected: 1.24-1.26)

        + [__helm__](https://helm.sh/docs/intro/install/) - to deploy the <span class="Space-Bd-BT">KUBESTELLAR</span>-core helm chart
        
        + [__brew__](https://helm.sh/docs/intro/install/) - to install the <span class="Space-Bd-BT">KUBESTELLAR</span> user commands and kubectl plugins
        
        + [__Kind__](https://kind.sigs.k8s.io) - to create a few small kubernetes clusters

        + 3 Kind clusters configured as follows
        
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

        {%
          include-markdown "../common-subs/create-files-and-contexts-for-kind-clusters.md"
          start="<!--create-files-and-contexts-for-kind-clusters-start-->"
          end="<!--create-files-and-contexts-for-kind-clusters-end-->"
        %}
    === "uh oh, error?"
        if you apply the ingress and then receive an error while waiting:
          `error: no matching resources found`

        this might mean that you did not wait long enough before issuing the check command. Simply try the check command again.
   
#### 1. Deploy the <span class="Space-Bd-BT">KUBESTELLAR</span> Core component  
{%
  include-markdown "../common-subs/deploy-your-kubestellar-core-component-kind.md"
  start="<!--deploy-your-kubestellar-core-component-kind-start-->"
  end="<!--deploy-your-kubestellar-core-component-kind-end-->"
%}

#### 2. Install <span class="Space-Bd-BT">KUBESTELLAR</span>'s user commands and kubectl plugins
!!! tip ""
    === "install"
         {%
           include-markdown "../common-subs/brew-install.md"
           start="<!--brew-install-start-->"
           end="<!--brew-install-end-->"
         %}
    === "remove"
         {%
           include-markdown "../common-subs/brew-remove.md"
           start="<!--brew-remove-start-->"
           end="<!--brew-remove-end-->"
         %}

#### 3. View your <span class="Space-Bd-BT">KUBESTELLAR</span> Core Space environment
!!! tip ""
    === "show all available <span class="Space-Bd-BT">KUBESTELLAR</span> Core Spaces"
         {%
           include-markdown "../common-subs/kubestellar-show-available-spaces.md"
           start="<!--kubestellar-show-available-spaces-start-->"
           end="<!--kubestellar-show-available-spaces-end-->"
         %}
    
    === "uh oh, error?"
         Did you received the following error:
         ```Error: Get "https://some_hostname.some_domain_name:{{config.ks_kind_port_num}}/clusters/root/apis/tenancy.kcp.io/v1alpha1/workspaces": dial tcp: lookup some_hostname.some_domain_name on x.x.x.x: no such host``

         A common error occurs if you set your port number to a pre-occupied port number and/or you set your EXTERNAL_HOSTNAME to something other than "localhost" so that you can reach your <span class="Space-Bd-BT">KUBESTELLAR</span> Core from another host, check the following:
         
         Check if the port specified in the **ks-core** Kind cluster configuration and the EXTERNAL_PORT helm value are occupied by another application:

         &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;1. is the port specified in this example occupied by another process?  If so, delete the **ks-core** Kind cluster and create it again using an available port for your 'hostPort' value

         &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;2. if you change the port for your **ks-core** 'hostPort', remember to also use that port as the helm 'EXTERNAL_PORT' value

         Check that your EXTERNAL_HOSTNAME helm value is reachable via DNS:

         &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;1. use 'nslookup <value of EXTERNAL_HOSTNAME>' to make sure there is a valid IP address associated with the hostname you have specified

         &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;2. make sure your EXTERNAL_HOSTNAME and associated ip address are listed in your /etc/hosts file.

         &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;3. make sure the IP address is associated with the system where you have deployed the **ks-core** Kind cluster

         if there is nothing obvious, [open a bug report and we can help you out](https://github.com/kubestellar/kubestellar/issues/new?assignees=&labels=kind%2Fbug&projects=&template=bug_report.yaml&title=bug%3A+)

    === "open a bug report"
        Stuck? [open a bug report and we can help you out](https://github.com/kubestellar/kubestellar/issues/new?assignees=&labels=kind%2Fbug&projects=&template=bug_report.yaml&title=bug%3A+)


#### 4. Install <span class="Space-Bd-BT">KUBESTELLAR</span> Syncers on your Edge Clusters
!!! tip ""
    === "Prep and apply"
        prepare <span class="Space-Bd-BT">KUBESTELLAR</span> Syncers, with `kubestellar prep-for-cluster`, for **ks-edge-cluster1** and **ks-edge-cluster2** and then apply the files that `kubestellar prep-for-cluster` prepared for you

        **important:** make sure you created Kind clusters for **ks-edge-cluster1** and **ks-edge-cluster2** from the pre-req step above before proceeding [how-to-deploy-and-use-kubestellar](#how-to-deploy-and-use-kubestellar)

         {%
           include-markdown "../common-subs/kubestellar-prep-syncer.md"
           start="<!--kubestellar-prep-syncer-start-->"
           end="<!--kubestellar-prep-syncer-end-->"
         %}

         {%
           include-markdown "../common-subs/kubestellar-apply-syncer.md"
           start="<!--kubestellar-apply-syncer-start-->"
           end="<!--kubestellar-apply-syncer-end-->"
         %}

#### 5. Deploy an Apache Web Server to ks-edge-cluster1 and ks-edge-cluster2
!!! tip ""
    === "deploy"
         {%
           include-markdown "../common-subs/kubestellar-apply-apache.md"
           start="<!--kubestellar-apply-apache-start-->"
           end="<!--kubestellar-apply-apache-end-->"
         %}

#### 6. View the Apache Web Server running on ks-edge-cluster1 and ks-edge-cluster2
!!! tip ""
    === "view"
         {%
           include-markdown "../common-subs/kubestellar-test-apache.md"
           start="<!--kubestellar-test-apache-start-->"
           end="<!--kubestellar-test-apache-end-->"
         %}
    === "uh oh, error?"
         {%
           include-markdown "../common-subs/kubestellar-check-syncers.md"
           start="<!--kubestellar-check-syncers-start-->"
           end="<!--kubestellar-check-syncers-end-->"
         %}

          If you see a `connection refused` error in either <span class="Space-Bd-BT">KUBESTELLAR</span> Syncer log(s):

          `E1021 21:22:58.000110       1 reflector.go:138] k8s.io/client-go@v0.0.0-20230210192259-aaa28aa88b2d/tools/cache/reflector.go:215: Failed to watch *v2alpha1.EdgeSyncConfig: failed to list *v2alpha1.EdgeSyncConfig: Get "https://kubestellar.core:1119/apis/edge.kubestellar.io/v2alpha1/edgesyncconfigs?limit=500&resourceVersion=0": dial tcp 127.0.0.1:1119: connect: connection refused`

          it means that your `/etc/hosts` does not have a proper ip address (NOT `127.0.0.1`) listed for the `kubestellar.core` hostname. Once there is a valid address in `/etc/hosts` for `kubestellar.core`, the syncer will begin to work properly and pull the namespace, deployment, and configmap from this instruction set. 

#### 7. Check the status of your Apache Server on ks-edge-cluster1 and ks-edge-cluster2

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

## How to use an existing <span class="Space-Bd-BT">KUBESTELLAR</span> environment

#### 1. Install <span class="Space-Bd-BT">KUBESTELLAR</span>'s user commands and kubectl plugins
!!! tip ""
    === "install"
         {%
           include-markdown "../common-subs/brew-install.md"
           start="<!--brew-install-start-->"
           end="<!--brew-install-end-->"
         %}
    === "remove"
         {%
           include-markdown "../common-subs/brew-remove.md"
           start="<!--brew-remove-start-->"
           end="<!--brew-remove-end-->"
         %}

#### 2. View your <span class="Space-Bd-BT">KUBESTELLAR</span> Core Space environment

!!! tip ""
    === "show all available <span class="Space-Bd-BT">KUBESTELLAR</span> Core Spaces"
         ```
         KUBECONFIG=~/.kube/config kubectl --context ks-core get secrets kubestellar \
           -o jsonpath='{.data.external\.kubeconfig}' \
           -n kubestellar | base64 -d > ks-core.kubeconfig

         KUBECONFIG=ks-core.kubeconfig kubectl ws --context root tree
         ```
    === "uh oh, error?"
         Did you received the following error:
         ```Error: Get "https://some_hostname.some_domain_name:{{config.ks_kind_port_num}}/clusters/root/apis/tenancy.kcp.io/v1alpha1/workspaces": dial tcp: lookup some_hostname.some_domain_name on x.x.x.x: no such host``

         A common error occurs if you set your port number to a pre-occupied port number and/or you set your EXTERNAL_HOSTNAME to something other than "localhost" so that you can reach your <span class="Space-Bd-BT">KUBESTELLAR</span> Core from another host, check the following:
         
         Check if the port specified in the **ks-core** Kind cluster configuration and the EXTERNAL_PORT helm value are occupied by another application:

         &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;1. is the port specified in this example occupied by another process?  If so, delete the **ks-core** Kind cluster and create it again using an available port for your 'hostPort' value

         &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;2. if you change the port for your **ks-core** 'hostPort', remember to also use that port as the helm 'EXTERNAL_PORT' value

         Check that your EXTERNAL_HOSTNAME helm value is reachable via DNS:

         &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;1. use 'nslookup <value of EXTERNAL_HOSTNAME>' to make sure there is a valid IP address associated with the hostname you have specified

         &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;2. make sure your EXTERNAL_HOSTNAME and associated ip address are listed in your /etc/hosts file.

         &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;3. make sure the IP address is associated with the system where you have deployed the **ks-core** Kind cluster

         if there is nothing obvious, [open a bug report and we can help you out](https://github.com/kubestellar/kubestellar/issues/new?assignees=&labels=kind%2Fbug&projects=&template=bug_report.yaml&title=bug%3A+)

    === "open a bug report"
        Stuck? [open a bug report and we can help you out](https://github.com/kubestellar/kubestellar/issues/new?assignees=&labels=kind%2Fbug&projects=&template=bug_report.yaml&title=bug%3A+)

#### 3. Deploy an Apache Web Server to ks-edge-cluster1 and ks-edge-cluster2
!!! tip ""
    === "deploy"
         {%
           include-markdown "../common-subs/kubestellar-apply-apache.md"
           start="<!--kubestellar-apply-apache-start-->"
           end="<!--kubestellar-apply-apache-end-->"
         %}

#### 4. View the Apache Web Server running on ks-edge-cluster1 and ks-edge-cluster2
!!! tip ""
    === "view"
         {%
           include-markdown "../common-subs/kubestellar-test-apache.md"
           start="<!--kubestellar-test-apache-start-->"
           end="<!--kubestellar-test-apache-end-->"
         %}
    === "uh oh, error?"
         {%
           include-markdown "../common-subs/kubestellar-check-syncers.md"
           start="<!--kubestellar-check-syncers-start-->"
           end="<!--kubestellar-check-syncers-end-->"
         %}

          If you see a `connection refused` error in either <span class="Space-Bd-BT">KUBESTELLAR</span> Syncer log(s):

          `E1021 21:22:58.000110       1 reflector.go:138] k8s.io/client-go@v0.0.0-20230210192259-aaa28aa88b2d/tools/cache/reflector.go:215: Failed to watch *v2alpha1.EdgeSyncConfig: failed to list *v2alpha1.EdgeSyncConfig: Get "https://kubestellar.core:1119/apis/edge.kubestellar.io/v2alpha1/edgesyncconfigs?limit=500&resourceVersion=0": dial tcp 127.0.0.1:1119: connect: connection refused`

          it means that your `/etc/hosts` does not have a proper ip address (NOT `127.0.0.1`) listed for the `kubestellar.core` hostname. Once there is a valid address in `/etc/hosts` for `kubestellar.core`, the syncer will begin to work properly and pull the namespace, deployment, and configmap from this instruction set. 

## 5. Check the status of your Apache Server on ks-edge-cluster1 and ks-edge-cluster2

```
TODO
```