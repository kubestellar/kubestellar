# User Guide

## KubeStellar Primer

This is a description of the concepts behind KubeStellar.

## Glossary

There have been some nomenclature changes: 

- The term `edge cluster` has been replaced by `workload execution cluster`
- The term `workload management workspace` has been replaced by `workload description space`
- The term `inventory management workspace` has been by `inventory space`
- The term `edge service provider workspace` has been replaced by `kubestellar core space`
- The term `mailbox workspace` has been replaced by `mailbox space`

**Edge Cluster** - A Kubernetes cluster which can execute a workload. In the examples on this website, we use Kind clusters. 

**EdgePlacement** - Contains the 'what' and 'where' of the workload in the `spec.namespaceSelector` and `spec.locationSelectors` fields respectively. It describes which cluster(s) to send the workload to.

**Edge Service Provider Workspace (ESPW)** - Exports the Kubestellar API.

**Inventory Management Workspace (IMW)** - Stores the SyncTarget and Location objects for the edge clusters. 

**Kubestellar Syncer** - Syncs Kubernetes resource objects from the cluster to the mailbox workspace, and vice versa. 

**Location** - An object that tells Kubestellar where the cluster is located. It is a vendor-facing API. 

**Mailbox Controller** - Creates a Mailbox Workspace for each SyncTarget object and puts an APIBinding in each Mailbox Workspace. 

**Mailbox Workspace** - There is one Mailbox Workspace for each cluster. It stores the SyncerConfig object and copies of the workload(s).

**PlacementTranslator** - 

**SinglePlacementSlice** - 

**Status Summarizer** - Creates a StatusSummary object with a summary of the Deployment objects in a workload's workspace.

**SyncerConfig** - 

**Sync Target** - A vendor-facing API that describes the cluster. This is referenced by the Location object. 

**Where Resolver** - Creates a SinglePlacementSlice object for each Edge Placement object. It runs from the ESPW. 

**Workload Management Workspace (WMW)** - This can store one workload description.

