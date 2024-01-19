# KubeStellar 0.20

KubeStellar 0.20  supports multi-cluster deployment of Kubernetes objects, controlled by a 
simple placement policy and deploying Kubernetes objects in their native format.  It uses OCM as 
transport, with standard OCM agents (Klusterlet). We show examples of deploying workloads to 
multi-cluster with kubectl, helm and ArgoCD using a simple label-selectors-based placement policy.


## Supported Features:

1. *Multi-cluster Deployment:* Kubernetes objects are deployed across multiple clusters, controlled by a 
straightforward placement policy.
2. *Pure-Kube User Experience:* Deployment of non-wrapped objects is handled in a pure Kubernetes manner.
3. *Object Management via WDS:* Creation, update, and deletion of objects in managed clusters are performed from WDS.
4. *OCM as Transport:* The Open Cluster Management (OCM) is used as transport, with standard OCM agents (Klusterlet).
5. *Multi-WDS and single OCM Shard:* Multiple WDSs and a single OCM shard are supported.
6. *Resiliency:* All components are running in Kubernetes, ensuring continued operation even after restarts of any component.
7. *Re-evaluation of Objects:* Existing objects are re-evaluated when a new placement is added or updated.
8. *Object Removal:* Objects are removed from clusters when the placement that led to their deployment on
 those clusters is deleted or updated and the what or where no longer match.
9. *Dynamic Handling of APIs:* Dynamically start/stop informers when adding/removing CRDs.
10. *Simplified setup:* Just 3 commands to get a fully functional setup (`kflex init`, `kflex create imbs`, `kflex create wds`)
11. *OpenShift Support:* Same commands to set it up. All components have been tested in OpenShift, 
including OCM Klusterlet for the WECs.
12. *Singleton Status* Addressed by the status controller in KubeStellar 0.20 and the [Status Add-On for OCM](link to be added)

## To be supported

1. Status summarization
2. Customization
3. OCM sharding
4. Upsync
5. "Pluggable Transport" 

## Architecture

See [Architecture](architecture.md).

## Packaging and Delivery

See [Packaging and Delivery](packaging.md)

## Usage examples

See [Examples](examples.md).
