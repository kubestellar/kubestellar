[![QuickStart test]({{config.repo_url}}/actions/workflows/docs-ecutable-qs.yml/badge.svg?branch={{config.ks_branch}})]({{config.repo_url}}/actions/workflows/docs-ecutable-qs.yml)&nbsp;&nbsp;&nbsp;

<!-- <img width="500px" src="../../KubeStellar-with-Logo.png" title="KubeStellar"> -->
{%
   include-markdown "quickstart-subs/quickstart-0-demo.md"
   start="<!--quickstart-0-demo-start-->"
   end="<!--quickstart-0-demo-end-->"
%}

!!! tip "Estimated time to complete this example:" 
    ~4 minutes
   



!!! tip "Required Packages:"
    === "Mac"
        ``` title="jq - https://stedolan.github.io/jq/download/"
        brew install jq
        ```
        ``` title="docker - https://docs.docker.com/engine/install/"
        brew install docker
        open -a Docker
        ```
        ``` title="kind - https://kind.sigs.k8s.io/docs/user/quick-start/"
        brew install kind
        ```
        ``` title="kubectl - https://kubernetes.io/docs/tasks/tools/ (version range expected: 1.23-1.25)"
        brew install kubectl
        ```
        [GO v1.19](https://gist.github.com/jniltinho/8758e15a9ef80a189fce) - You will need GO to compile and run kcp and the KubeStellar scheduler.  Currently kcp requires go version 1.19.
    === "Ubuntu"
        ``` title="jq - https://stedolan.github.io/jq/download/"
        sudo apt-get install jq
        ```
        ``` title="docker - https://docs.docker.com/engine/install/"
        sudo mkdir -p /etc/apt/keyrings
        curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
        echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
        sudo apt update
        sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
        ```
        ``` title="kind - https://kind.sigs.k8s.io/docs/user/quick-start/"
        curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.14.0/kind-linux-$(dpkg --print-architecture) && chmod +x ./kind && sudo mv ./kind /usr/local/bin
        ```
        ``` title="kubectl - https://kubernetes.io/docs/tasks/tools/ (version range expected: 1.23-1.25)"
        curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/$(dpkg --print-architecture)/kubectl && chmod +x kubectl && sudo mv ./kubectl /usr/local/bin/kubectl
        ```
        ``` title="GO - You will need GO to compile and run kcp and the KubeStellar scheduler.  Currently kcp requires go version 1.19"
        curl -L "https://go.dev/dl/go1.19.5.linux-$(dpkg --print-architecture).tar.gz" -o go.tar.gz
        tar -C /usr/local -xzf go.tar.gz
        rm go.tar.gz
        echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
        source /etc/profile
        go version
        ```
    === "RHEL"
        ``` title="jq - https://stedolan.github.io/jq/download/"
        yum -y install jq
        ```
        ``` title="docker - https://docs.docker.com/engine/install/"
        yum -y install epel-release && yum -y install docker && systemctl enable --now docker && systemctl status docker
        ```
        ``` title="kind - https://kind.sigs.k8s.io/docs/user/quick-start/"
        # For AMD64 / x86_64
        [ $(uname -m) = x86_64 ] && curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
        # For ARM64
        [ $(uname -m) = aarch64 ] && curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-arm64
        chmod +x ./kind
        sudo mv ./kind /usr/local/bin/kind
        ```
        ``` title="kubectl - https://kubernetes.io/docs/tasks/tools/ (version range expected: 1.23-1.25)"
        curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl && chmod +x kubectl && mv ./kubectl /usr/local/bin/kubectl
        ```
        [GO v1.19](https://gist.github.com/jniltinho/8758e15a9ef80a189fce) - You will need GO to compile and run kcp and the KubeStellar scheduler.  Currently kcp requires go version 1.19.
    === "WSL"
        ``` title="jq - https://stedolan.github.io/jq/download/"
        choco install jq -y
        choco install curl -y
        ```
        ``` title="docker - https://docs.docker.com/engine/install/"
        choco install docker -y
        ```
        ``` title="kind - https://kind.sigs.k8s.io/docs/user/quick-start/"
        curl.exe -Lo kind-windows-amd64.exe https://kind.sigs.k8s.io/dl/v0.14.0/kind-windows-amd64
        ```
        ``` title="kubectl - https://kubernetes.io/docs/tasks/tools/install-kubectl-windows/ (version range expected: 1.23-1.25)"
        curl.exe -LO "https://dl.k8s.io/release/v1.27.2/bin/windows/amd64/kubectl.exe"
        ```
        [GO v1.19](https://gist.github.com/jniltinho/8758e15a9ef80a189fce) - You will need GO to compile and run kcp and the KubeStellar scheduler.  Currently kcp requires go version 1.19.
<!--required-packages-end-->
<!-- 
## 
  - [docker](https://docs.docker.com/engine/install/)
  - [kind](https://kind.sigs.k8s.io/)
  - [kubectl](https://kubernetes.io/docs/tasks/tools/) (version range expected: 1.23-1.25)
  - [jq](https://stedolan.github.io/jq/download/) -->
  
## Setup Instructions

Table of contents:

1. [Install and run **KubeStellar**](#1-install-and-run-kubestellar)
2. [Example deployment of Apache HTTP Server workload into two local kind clusters](#2-example-deployment-of-apache-http-server-workload-into-two-local-kind-clusters)
      1. [Stand up two kind clusters: florin and guilder](#a-stand-up-two-kind-clusters-florin-and-guilder)
      2. [Create a KubeStellar Inventory Management Workspace (IMW) and Workload Management Workspace (WMW)](#b-create-a-kubestellar-inventory-management-workspace-imw-and-workload-management-workspace-wmw)
      3. [Onboarding the clusters](#c-onboarding-the-clusters)
      4. [Create and deploy the Apache Server workload into florin and guilder clusters](#d-create-and-deploy-the-apache-server-workload-into-florin-and-guilder-clusters)
3. [Teardown the environment](#3-teardown-the-environment)
4. [Next Steps](#4-next-steps)


This guide is intended to show how to (1) quickly bring up a **KubeStellar** environment with its dependencies from a binary release and then (2) run through a simple example usage.

{%
   include-markdown "quickstart-subs/quickstart-1-install-and-run-kubestellar.md"
   start="<!--quickstart-1-install-and-run-kubestellar-start-->"
   end="<!--quickstart-1-install-and-run-kubestellar-end-->"
%}

## 2. Example deployment of Apache HTTP Server workload into two local kind clusters

In this example you will create two edge clusters and define one
workload that will be distributed from the center to those edge
clusters.  This example is similar to the one described more
expansively [on the
website](../../Coding%20Milestones/PoC2023q1/example1/),
but with the some steps reorganized and combined and the special
workload and summarization aspirations removed.

### a. Stand up two kind clusters: florin and guilder

{%
   include-markdown "quickstart-subs/quickstart-2-apache-example-deployment-a-kind-clusters-florin.md"
   start="<!--quickstart-2-apache-example-deployment-a-kind-clusters-florin-start-->"
   end="<!--quickstart-2-apache-example-deployment-a-kind-clusters-florin-end-->"
%}

{%
   include-markdown "quickstart-subs/quickstart-2-apache-example-deployment-a-kind-clusters-guilder.md"
   start="<!--quickstart-2-apache-example-deployment-a-kind-clusters-guilder-start-->"
   end="<!--quickstart-2-apache-example-deployment-a-kind-clusters-guilder-end-->"
%}
### b. Create a KubeStellar Inventory Management Workspace (IMW) and Workload Management Workspace (WMW)

{%
   include-markdown "quickstart-subs/quickstart-2-apache-example-deployment-b-create-imw-and-wmw.md"
   start="<!--quickstart-2-apache-example-deployment-b-create-imw-and-wmw-start-->"
   end="<!--quickstart-2-apache-example-deployment-b-create-imw-and-wmw-end-->"
%}

### c. Onboarding the clusters

{%
   include-markdown "quickstart-subs/quickstart-2-apache-example-deployment-c-onboarding-clusters.md"
   start="<!--quickstart-2-apache-example-deployment-c-onboarding-clusters-start-->"
   end="<!--quickstart-2-apache-example-deployment-c-onboarding-clusters-end-->"
%}

### d. Create and deploy the Apache Server workload into florin and guilder clusters

{%
   include-markdown "quickstart-subs/quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters.md"
   start="<!--quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters-start-->"
   end="<!--quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters-end-->"
%}

## 3. Teardown the environment

{%
   include-markdown "../common-subs/teardown-the-environment.md"
   start="<!--teardown-the-environment-start-->"
   end="<!--teardown-the-environment-end-->"
%}

## 4. Next Steps

What you just did is a shortened version of the 
[more detailed example](../../Coding%20Milestones/PoC2023q1/example1/) on the website,
but with the some steps reorganized and combined and the special
workload and summarization aspiration removed.  You can continue
from here, learning more details about what you did in the QuickStart,
and adding on some more steps for the special workload.
