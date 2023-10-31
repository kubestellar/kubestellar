---
short_name: user-quickstart-openshift
manifest_name: 'docs/content/Getting-Started/user-quickstart-openshift.md'
syncer_image_sets: ''
---
<!-- [![User QuickStart OpenShift test]({{config.repo_url}}/actions/workflows/docs-ecutable-user-quickstart-openshift.yml/badge.svg?branch={{config.ks_branch}})]({{config.repo_url}}/actions/workflows/docs-ecutable-user-quickstart-openshift.yml)&nbsp;&nbsp;&nbsp; -->

<!-- 
!!! tip "Estimated time to complete this example:" 
    ~20 minutes (after installing prerequisites) -->

## How to deploy and use <span class="Space-Bd-BT">KUBESTELLAR</span> on Red Hat OpenShift Kubernetes Clusters

!!! tip ""
    === "Goals"
        This guide will show how to:

        1. quickly deploy the KubeStellar Core component on an OpenShift cluster using helm (ks-core), 
        2. install the KubeStellar user commands and kubectl plugins on your computer with brew,
        3. retrieve the KubeStellar Core component kubeconfig, 
        4. install the KubeStellar Syncer component on two edge OpenShift clusters (ks-edge-cluster1 and ks-edge-cluster2), 
        5. deploy an example kubernetes workload to both edge OpenShift clusters from KubeStellar Core (ks-core),
        6. view the example kubernetes workload running on two edge OpenShift clusters (ks-edge-cluster1 and ks-edge-cluster2)
        7. view the status of your deployment across both edge OpenShift clusters from KubeStellar Core (ks-core)

        **important:** For this quickstart you will need to know how to use kubernetes' kubeconfig *context* to access multiple clusters.  You can learn more about kubeconfig context [here](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/)

!!! tip ""
    === "Pre-reqs"
        + [__kubectl__](https://kubernetes.io/docs/tasks/tools/) (version range expected: 1.24-1.26)

        + [__helm__](https://helm.sh/docs/intro/install/) - to deploy the KubeStellar-core helm chart
        
        + [__brew__](https://brew.sh) - to install the KubeStellar user commands and kubectl plugins
        
        + 3 Red Hat OpenShift clusters - we will refer to them as **ks-core**, **ks-edge-cluster1**, and **ks-edge-cluster2** in this document

        {%
          include-markdown "../common-subs/delete-contexts-for-kind-and-openshift-clusters.md"
          start="<!--delete-contexts-for-kind-and-openshift-clusters-start-->"
          end="<!--delete-contexts-for-kind-and-openshift-clusters-end-->"
        %}
        {%
          include-markdown "../common-subs/create-files-and-contexts-for-openshift-clusters.md"
          start="<!--create-files-and-contexts-for-openshift-clusters-start-->"
          end="<!--create-files-and-contexts-for-openshift-clusters-end-->"
        %}
   
#### 1. Deploy the <span class="Space-Bd-BT">KUBESTELLAR</span> Core component
{%
  include-markdown "../common-subs/deploy-your-kubestellar-core-component-openshift.md"
  start="<!--deploy-your-kubestellar-core-component-openshift-start-->"
  end="<!--deploy-your-kubestellar-core-component-openshift-end-->"
%}

#### 2. Install KubeStellar's user commands and kubectl plugins
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
    === "uh oh, no brew?"
         {%
           include-markdown "../common-subs/brew-no.md"
           start="<!--brew-no-start-->"
           end="<!--brew-no-end-->"
         %}

#### 3. View your <span class="Space-Bd-BT">KUBESTELLAR</span> Core Space environment
!!! tip ""
    === "show all available KubeStellar Core Spaces"
         Let's store the KubeStellar kubeconfig to a file we can reference later and then check out the Spaces KubeStellar created during installation
         {%
           include-markdown "../common-subs/kubestellar-show-available-spaces.md"
           start="<!--kubestellar-show-available-spaces-start-->"
           end="<!--kubestellar-show-available-spaces-end-->"
         %}
    
    === "uh oh, error?"

    === "open a bug report"
        Stuck? [open a bug report and we can help you out](https://github.com/kubestellar/kubestellar/issues/new?assignees=&labels=kind%2Fbug&projects=&template=bug_report.yaml&title=bug%3A+)


#### 4. Install <span class="Space-Bd-BT">KUBESTELLAR</span> Syncers on your Edge Clusters
!!! tip ""
    === "Prep and apply"
        prepare KubeStellar Syncers, with `kubestellar prep-for-cluster`, for **ks-edge-cluster1** and **ks-edge-cluster2** and then apply the files that `kubestellar prep-for-cluster` prepared for you

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
           include-markdown "../common-subs/kubestellar-apply-apache-openshift.md"
           start="<!--kubestellar-apply-apache-openshift-start-->"
           end="<!--kubestellar-apply-apache-openshift-end-->"
         %}

#### 6. View the Apache Web Server running on ks-edge-cluster1 and ks-edge-cluster2
!!! tip ""
    === "view"
         {%
           include-markdown "../common-subs/kubestellar-test-apache-openshift.md"
           start="<!--kubestellar-test-apache-openshift-start-->"
           end="<!--kubestellar-test-apache-openshift-end-->"
         %}
    === "uh oh, error?"
         {%
           include-markdown "../common-subs/kubestellar-check-syncers.md"
           start="<!--kubestellar-check-syncers-start-->"
           end="<!--kubestellar-check-syncers-end-->"
         %}

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

#### 1. Install KubeStellar's user commands and kubectl plugins
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
    === "uh oh, no brew?"
         {%
           include-markdown "../common-subs/brew-no.md"
           start="<!--brew-no-start-->"
           end="<!--brew-no-end-->"
         %}

#### 2. View your <span class="Space-Bd-BT">KUBESTELLAR</span> Core Space environment

!!! tip ""
    === "show all available KubeStellar Core Spaces"
         Let's store the KubeStellar kubeconfig to a file we can reference later and then check out the Spaces KubeStellar created during installation

         ```
         KUBECONFIG=~/.kube/config kubectl --context ks-core get secrets kubestellar \
           -o jsonpath='{.data.external\.kubeconfig}' \
           -n kubestellar | base64 -d > ks-core.kubeconfig

         KUBECONFIG=ks-core.kubeconfig kubectl ws tree
         ```

    === "open a bug report"
        Stuck? [open a bug report and we can help you out](https://github.com/kubestellar/kubestellar/issues/new?assignees=&labels=kind%2Fbug&projects=&template=bug_report.yaml&title=bug%3A+)

#### 3. Deploy an Apache Web Server to ks-edge-cluster1 and ks-edge-cluster2
!!! tip ""
    === "deploy"
         {%
           include-markdown "../common-subs/kubestellar-apply-apache-openshift.md"
           start="<!--kubestellar-apply-apache-openshift-start-->"
           end="<!--kubestellar-apply-apache-openshift-end-->"
         %}

#### 4. View the Apache Web Server running on ks-edge-cluster1 and ks-edge-cluster2
!!! tip ""
    === "view"
         {%
           include-markdown "../common-subs/kubestellar-test-apache-openshift.md"
           start="<!--kubestellar-test-apache-openshift-start-->"
           end="<!--kubestellar-test-apache-openshift-end-->"
         %}
    === "uh oh, error?"
         {%
           include-markdown "../common-subs/kubestellar-check-syncers.md"
           start="<!--kubestellar-check-syncers-start-->"
           end="<!--kubestellar-check-syncers-end-->"
         %}


## 5. Check the status of your Apache Server on ks-edge-cluster1 and ks-edge-cluster2

{%
           include-markdown "../common-subs/kubestellar-list-syncing.md"
           start="<!--kubestellar-list-syncing-start-->"
           end="<!--kubestellar-list-syncing-end-->"
%}
