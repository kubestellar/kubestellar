# KubeStellar

KubeStellar supports multi-cluster deployment of Kubernetes objects, controlled by a 
simple binding policy and deploying Kubernetes objects in their native format.  It uses OCM as 
transport, with standard OCM agents (Klusterlet). We show examples of deploying workloads to 
multi-cluster with kubectl, helm and ArgoCD using a simple label-selectors-based binding policy.


## Supported Features:

1. *Multi-cluster Deployment:* Kubernetes objects are deployed across multiple clusters, controlled by a 
straightforward binding policy.
2. *Pure-Kube User Experience:* Deployment of non-wrapped objects is handled in a pure Kubernetes manner.
3. *Object Management via WDS:* Creation, update, and deletion of objects in managed clusters are performed from WDS.
4. *OCM as Transport:* The Open Cluster Management (OCM) is used as transport, with standard OCM agents (Klusterlet).
5. *Multi-WDS and single OCM Shard:* Multiple WDSs and a single OCM shard are supported.
6. *Resiliency:* All components are running in Kubernetes, ensuring continued operation even after restarts of any component.
7. *Re-evaluation of Objects:* Existing objects are re-evaluated when a new binding policy is added or updated.
8. *Object Removal:* Objects are removed from clusters when the binding policy that led to their deployment on
 those clusters is deleted or updated and the what or where no longer match.
9. *Dynamic Handling of APIs:* Dynamically start/stop informers when adding/removing CRDs.
10. *Simplified setup:* Just 3 commands to get a fully functional setup (`kflex init`, `kflex create imbs`, `kflex create wds`)
11. *OpenShift Support:* Same commands to set it up. All components have been tested in OpenShift, 
including OCM Klusterlet for the WECs.
12. *Singleton Status* Addressed by the status controller in KubeStellar and the [Status Add-On for OCM](https://github.com/kubestellar/ocm-status-addon).

## To be supported

1. Status summarization
2. Customization
3. OCM sharding
4. Upsync

## Latest stable release

We do not have one that is proven very good yet.
The first release using the new architecture is 0.20.0; it is feture-incomplete.
Release 0.21.0 works but is feature-incomplete and is not well documented on the website; [view the missing information directly at GitHub](https://github.com/kubestellar/kubestellar/tree/release-0.21.0/docs/content/direct).
The latest release is 0.21.2-rc1; it is also feature-incomplete.
See also [the release notes](release-notes.md).

## Architecture

See [Architecture](architecture.md).

## Packaging and Delivery

See [Packaging and Delivery](packaging.md)

## Usage examples, and testing

[Examples](examples.md) shows a few examples of how to deploy and use a release of KubeStellar.

The `test/e2e` directory holds end-to-end tests.

The contributor guide has a section on [integration testing](contributor.md#integration-testing).

The contributor guide has a section on [unit testing](contributor.md#unit-testing).

## Using an existing hosting cluster

See [the instructions for using an existing hosting cluster](hosting-cluster.md) for notes on that and coping with multi-machine scenarios.

## Contributor Guide

See [the contributor guide](contributor.md), which is also just beginning to be written.
