
[![QuickStart test]({{config.repo_url}}/actions/workflows/docs-ecutable-qs.yml/badge.svg?branch={{config.ks_branch}})]({{config.repo_url}}/actions/workflows/docs-ecutable-qs.yml)&nbsp;&nbsp;&nbsp;

<!-- <img width="500px" src="../../KubeStellar-with-Logo.png" title="KubeStellar"> -->
{%
   include-markdown "quickstart-subs/quickstart-0-demo.md"
   start="<!--quickstart-0-demo-start-->"
   end="<!--quickstart-0-demo-end-->"
%}

!!! tip "Estimated time to complete this example:" 
    ~4 minutes (after installing prerequisites)


## Setup Instructions

Table of contents:

1. [Check Required Packages](#1-check-required-packages)
2. [Install and run kcp and **KubeStellar**](#2-install-and-run-kcp-and-kubestellar)
3. [Example deployment of Apache HTTP Server workload into two local kind clusters](#3-example-deployment-of-apache-http-server-workload-into-two-local-kind-clusters)
      1. [Stand up two kind clusters: florin and guilder](#a-stand-up-two-kind-clusters-florin-and-guilder)
      2. [Onboarding the clusters](#b-onboarding-the-clusters)
      3. [Create and deploy the Apache Server workload into florin and guilder clusters](#c-create-and-deploy-the-apache-server-workload-into-florin-and-guilder-clusters)
4. [Teardown the environment](#4-teardown-the-environment)
5. [Next Steps](#5-next-steps)


This guide is intended to show how to (1) quickly bring up a **KubeStellar** environment with its dependencies from a binary release and then (2) run through a simple example usage.

## 1. Check Required Packages
   
{%
   include-markdown "../common-subs/required-packages.md"
   start="<!--required-packages-start-->"
   end="<!--required-packages-end-->"
%}

## 2. Install and run kcp and **KubeStellar**

{%
   include-markdown "quickstart-subs/quickstart-1-install-and-run-kubestellar.md"
   start="<!--quickstart-1-install-and-run-kubestellar-start-->"
   end="<!--quickstart-1-install-and-run-kubestellar-end-->"
%}

## 3. Example deployment of Apache HTTP Server workload into two local kind clusters

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

### b. Onboarding the clusters

{%
   include-markdown "quickstart-subs/quickstart-2-apache-example-deployment-c-onboarding-clusters.md"
   start="<!--quickstart-2-apache-example-deployment-c-onboarding-clusters-start-->"
   end="<!--quickstart-2-apache-example-deployment-c-onboarding-clusters-end-->"
%}

### c. Create and deploy the Apache Server workload into florin and guilder clusters

{%
   include-markdown "quickstart-subs/quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters.md"
   start="<!--quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters-start-->"
   end="<!--quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters-end-->"
%}

## 4. Teardown the environment

{%
   include-markdown "../common-subs/teardown-the-environment.md"
   start="<!--teardown-the-environment-start-->"
   end="<!--teardown-the-environment-end-->"
%}

## 5. Next Steps

What you just did is a shortened version of the 
[more detailed example](../../Coding%20Milestones/PoC2023q1/example1/) on the website,
but with the some steps reorganized and combined and the special
workload and summarization aspiration removed.  You can continue
from here, learning more details about what you did in the QuickStart,
and adding on some more steps for the special workload.

