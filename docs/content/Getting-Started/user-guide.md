# User Guide

## KubeStellar Primer

This is a description of the concepts behind KubeStellar.

## Glossary

**Downsynced Object** - One of two categories of workload object, complementary to "upsynced object".  In KubeStellar, a downsynced object first appears in a Workload Description Space and the object's desired state propagates from there through Mailbox Spaces to Workload Execution Clusters and that object's reported state originates in the Workload Execution Clusters and propagates back to the Mailbox Spaces and in the future will be summarized into the Workload Description Space.

**EdgePlacement** - A kind of Kubernetes API object, in a Workload Description Space. One of objects these binds some workload ("what") with a set of workload execution clusters ("where") it should run. The workload is identified by a predicate over namespaced objects and a predicate over cluster-scoped objects. The where is identified by a predicate over workload execution clusters as represented by `Location` objects.

**Inventory Space (IS)** - Holds the `SyncTarget` and `Location` objects describing the Workload Execution Clusters. 

**KubeStellar Core Space (KCS)** - Exports the Kubestellar API.

**KubeStellar Syncer** - The KubeStellar agent in a Workload Execution Cluster; syncs workload objects between the Workload Execution Cluster and the corresponding Mailbox Space.

**Location** - A kind of Kubernetes API object, in an Inventory Space. Paired one-to-one with a `SyncTarget` object in the same space. Together these describe a workload execution cluster. The Location's labels are tested by the "where predicate" in an `EdgePlacement` object, and this object's labels and annotations provide values used in customization of workload objects going to the workload execution cluster.

**Mailbox Controller** - One of the central KubeStellar controllers; maintains a Mailbox Space for each `SyncTarget` object. This includes putting an APIBinding to the KubeStellar API into those mailbox spaces.

**Mailbox Space** - There is one Mailbox Space for each workload execution cluster. It stores the `SyncerConfig` object and copies of the workload(s).

**PlacementTranslator** - One of the central KubeStellar controllers; maintains the `SyncerConfig` objects in the Mailbox Spaces and syncs workload objects between the Workload Description Spaces and the Mailbox Spaces.

**SinglePlacementSlice** - A kind of Kubernetes API object, in a Workload Description Space.  Such an object holds a list of references to `Location` & `SyncTarget` objects that match the "where predicate" of an `EdgePlacement`.  Currently there is exactly one `SinglePlacementSlice` for each `EdgePlacement` but in the future the matches for one `EdgePlacement` could be spread among several `SinglePlacementSlice` objects (analogously to `EndpointSlice` vs `Service` in Kubernetes).

**Space** - A thing that behaves like a Kubernetes kube-apiserver (and the persistent storage behind it) and the subset of controllers in the kube-controller-manager that are concerned with API machinery generalities (not management of containerized workloads). A kcp logical cluster is an example. A regular Kubernetes cluster includes a space.

**Status Summarizer** - A planned central KubeStellar controller that will maintain the status summary objects in the Workload Description Spaces as a function of the `EdgePlacement` objects and the workload objects in the Mailbox Spaces.

**SyncerConfig** - A kind of Kubernetes API object, in a Mailbox Space. Such an object holds the dynamic configuration for the syncer in the corresponding workload execution cluster.

**SyncTarget** - A kind of Kubernetes API object, in an Inventory Space. Paired one-to-one with a Location in the same space, jointly representing a Workload Execution Cluster.

**Upsynced Object** - One of two categories of workload object, complementary to "downsynced object".  Upsynced objects originate in Workload Execution Clusters and propagate inward to Mailbox Spaces and in the future will be summarized into Workload Description Spaces.

**Where Resolver** - One of the central KubeStellar controllers; tests the `Location` objects against the "where predicates" in the `EdgePlacement` objects to maintain the corresponding `SinglePlacementSlice` objects.

**Workload Description Space (WDS)** - Holds workload objects and the adjacent KubeStellar control objects, which are the `EdgePlacement`, `SinglePlacementSlice`, and `Customizer` objects and, eventually, the ones developed to prescribe summarization.

**Workload Execution Cluster** - A Kubernetes cluster which can execute a workload. In the examples on this website, we use [Kind](https://kind.sigs.k8s.io/) clusters.

### Older Terminology

There have been some terminology shifts since the start of the project.  The project started with a focus on edge computing scenarios; later we realized that the technical problems addressed are not limited to those scenarios.  The project started in the context of [kcp](https://github.com/kcp-dev/kcp) and unabashedly used concepts from kcp; later we began working on generalizing KubeStellar so that it can run in the context of kcp and also can run in other contexts that do not have kcp.

- The term "space" is intended to be a generalization covering both a kcp "logical cluster" or "workspace" and other things that have the same essential behavior
- The term "workload execution cluster" was formerly "edge cluster"
- The term "workload description space" was formerly "workload management workspace"
- The term "inventory space" was formerly "inventory management workspace"
- The term "kubestellar core space" was formerly "edge service provider workspace"
- The term "mailbox space" was formerly "mailbox workspace"
