# Basics of KubeStellar Usage

KubeStellar operates with a multiple-cluster [architecture](/readme/):

* One for the KubeStellar core components
* One or more Workload Execution Clusters (WECs) to which the core deploys workloads


## General Flow of Initial Setup
The general flow for setting up and using KubeStellar is in 3 steps:

1. Set Up the KubeStellar core components
2. Set Up and Register the Workload Execution Clusters
3. Define workloads and the bindings between workloads and the WECs
 _This will deploy the workloads to the WECs_

## After Deployment of Workloads
Once workload(s) are deployed, KubeStellar lets you

* Confirm/monitor status of workload(s)
* Modify workloads as necessary (update, delete, or redeploy workloads) on WECs by redefining the workloads and/or bindings.

## Getting Started: Try the "Common Setup"
Given the multitude of infrastructure configurations possible, the details of any particular installation can obviously vary.

We have developed a [common setup](common-setup-intro.md) for our examples which you may find a useful starting point. You can use a [helm chart](common-setup-hub-chart.md) to automate the process of setting up the core components, or you may [work through the steps individually](common-setup-step-by-step.md) to more completely customize the installation.

<!--
* Set up infrastructure to host the hub and workload clusters
* Install prerequisite software to do the setup
* Set up the KubeStellar core components (hub) cluster(s)
* Set up Workload Execution Cluster(s)
* Register WECs with the hub
* Define workloads for deployment
* Deploy workloads
* Confirm/monitor status of workload(s)
* Redefine workloads as necessary (Updates/Undeploys/Redeploys workload on WECs)
-->
<!-- ## Prereqs

### Set up your core and workload cluster infrastructure

### install appropriate software there

## Set up the KubeStellar Core components

### - Prepare the environment ###
    
### - Initialize Kubeflex ###

### - Install the Kubestellar core components ###

   
### - Create the Inventory & Transport Space ###

### - Create the Workload Description Spaces ###

## Create Workload Execution Clusters (WECs)

## Register WECs with Kubestellar core

## Create and Define workloads for deployment -->
