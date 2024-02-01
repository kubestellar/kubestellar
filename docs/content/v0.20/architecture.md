# KubeStellar Architecture

KubeStellar provides multi-cluster deployment of Kubernetes objects, controlled by a simple 
placement policy, where Kubernetes objects are expressed in their native format with no wrapping
or bundling. The high-level architecture for KubeStellar is illustrated in Figure 1.

<figure>
  <img src="./high-level-architecture.png"  alt="High Level Architecture">
  <figcaption align="center">Figure 1 - High Level Architecture </figcaption>
</figure>

KubeStellar relies on the concept of *spaces*.  
A Space is an abstraction to represent an API service that 
behaves like a Kubernetes kube-apiserver (including the persistent storage behind it) 
and the subset of controllers in the kube-controller-manager that are concerned with 
API machinery generalities (not management of containerized workloads). 
A KubeFlex `ControlPlane` is an example. A regular Kubernetes cluster is another example.
Users can use spaces to perform these tasks:

1. Create *Workload Definition Spaces* (WDSs) to store the definitions of their workloads.
A Kubernetes workload is an application that runs on Kubernetes. A workload can be made by a 
single Kubernetes object or several objects that work together.
2. Create *Inventory and Transport Spaces* (ITSs) to manage the inventory of clusters and 
the transport of workloads.
3. Register and label Workload Execution Clusters (WECs) with the Inventory and 
Transport Space, to keep track of the available clusters and their characteristics.
4. Define Placement Policies to specify *what* objects and *where* should be 
deployed on the WECs.
5. Submit objects in the native Kubernetes format to the WDSs, 
and let the Placement Policies govern which WECs should receive them.
6. Check the status of submitted objects from the WDS.

In KubeStellar, users can assume a variety of roles and responsibilities. 
These roles could range from system administrators and application owners 
to CISOs and DevOps Engineers. However, for the purpose of this document, 
we will not differentiate between these roles. Instead we will use the term 
'user' broadly, without attempting to make distinctions among roles.

Examples of users interaction with KubeStellar are illustrated in the
[KubeStellar Usage Examples](./examples.md) section.

The KubeStellar architecture has these main modules:

- *KubeStellar Controller Manager*: this module is responsible of delivering workload
objects from the WDS to the ITS according to placement policies, and 
updating the status of objects in the WDS.

- *Space Manager*: This module manages the lifecycle of spaces.

- *OCM Cluster Manager*: This module syncs objects from the ITS to the Workload Execution 
Clusters (WECs). In the ITS, each mailbox namespace is associated with one WEC. Objects 
that are put in a mailbox namespace are delivered to the matching WEC.

- *Status Add-On Controller*: This module installs the OCM status add-on agent 
on all WECs and sets the RBAC permissions for it using the OCM add-on framework.

- *OCM Agent*: This module registers the WEC to the OCM Hub, watches for 
ManifestWorks and unwraps and syncs the objects into the WEC.

- *OCM Status Add-On Agent*: This module watches *AppliedManifestWorks* 
to find objects that are synced by the OCM agent, gets their status 
and updates *WorkStatus* objects in the ITS namespace associated with the WEC.

<figure>
  <img src="./main-modules.png"  alt="Main Modules">
  <figcaption align="center">Figure 2 - Main Modules </figcaption>
</figure>

## KubeStellar Controller Manager

This module manages the placement and status controllers. The placement controller watches 
placement and workload objects on the Workload Definition Space 
(WDS) and wraps workload objects into a manifest for delivery through the 
Inventory and Transport Space (ITS). The status controller watches for *WorkStatus* objects on the 
ITS and updates the status of objects in the WDS when singleton status is requested 
in the placement for those objects. There is one instance of a KubeStellar Controller 
Manager for each WDS. Currently this controller-manager runs in the KubeFlex hosting cluster and 
is responsible of installing the required CRDs in the associated WDS.
More details on the internals of this module are provided in [KubeStellar Controllers Architecture](#kubestellar-controllers-architecture)

## Space Manager

The Space Manager handles the lifecycle of spaces. 
KubeStellar uses the [KubeFlex project](https://github.com/kubestellar/kubeflex)
for space management. In KubeFlex, a space is named a `ControlPlane`, and we will use 
both terms in this document. KubeSteller currently prereqs KubeFlex to 
provide one or more spaces. We plan to make this optional in the near future.

KubeFlex is a flexible framework that supports various kinds of control planes, such
as *k8s*, a basic Kubernetes API Server with a subset of kube controllers, and 
*vcluster*: a virtual cluster that runs on the hosting cluster based on the
[vCluster Project](https://www.vcluster.com). More detailed information
on the different types of control planes and architecture are described
in the [KubeFlex Architecture](https://github.com/kubestellar/kubeflex/blob/main/docs/architecture.md).

There are currently two roles for spaces managed by KubeFlex: Inventory and Transport Space 
(ITS) and Workload Description Space (WDS). The former runs the [OCM Cluster Manager](#ocm-cluster-manager)
on a vcluster-type control plane, and the latter runs on a k8s-type control plane.

An ITS holds the OCM inventory (`ManagedCluster`) objects and mailbox namespaces. The mailbox 
namespaces and their contents are implementation details that users do not deal with. Each 
mailbox workspace corresponds 1:1 with a WEC and holds `ManifestWork` objects managed by the 
central KubeStellar controllers.

A WDS holds user workload objects and the user's objects that form the interface to KubeStellar 
control. Currently the only control objects are `Placement` objects. We plan to add `Customization`
objects soon, and later objects to specify summarization.

KubeFlex provides the ability to start controllers connected to a
Control Plane API Server or to deploy Helm Charts into a Control Plane
API server with [<u>post-create
hooks</u>](https://github.com/kubestellar/kubeflex/blob/main/docs/users.md#post-create-hooks).
This feature is currently adopted for KubeStellar modules startup, as it allows to
create a Workload Description Space (WDS) and start the KubeStellar Controller Manager, and create an Inventory and Transport Space (ITS) in a
`vcluster` and install the [<u>Open Cluster Management
Hub</u>](https://open-cluster-management.io/) there.

## OCM Cluster Manager

This module is based on the [Open Cluster Management Project](https://open-cluster-management.io),
a community-driven project that focuses on multicluster and multicloud scenarios for Kubernetes apps. 
It provides APIs for cluster registration, work distribution and much more. 
The project is based on a hub-spoke architecture, where a single hub cluster 
handles the distribution of workloads through manifests, and one or more spoke clusters 
receive and apply the workload objects from the manifests. In Open Cluster Management, spoke clusters 
are called *managed clusters*, and the component running on the hub cluster is the *cluster manager*.
Manifests provide a summary for the status of each object, however in some use 
cases this might not be sufficient as the full status for objects may be required. 
OCM provides an add-on framework that allows to automatically install additional 
agents on the managed clusters to provide specific features. This framework is used to
install the status add-on on all managed clusters.
KubeStellar currently exposes users directly to OCM inventory management and WEC registration.
We plan to make the transport features provided by the OCM project pluggable in the near future.

## Status Add-On Controller

This module automates the installation of the status add-on agent 
on all managed clusters. It is based on the 
[OCM Add-on Framework](https://open-cluster-management.io/concepts/addon/), 
which is a framework that helps developers to develop extensions 
for working with multiple clusters in custom cases. A module based on 
the add-on framework has two components: a controller and the 
add-on agent. The controller interacts with the add-on manager to register 
the add-on, manage the distribution of the add-on to all clusters, and set 
up the RBAC permissions required by the add-on agent to interact with the mailbox 
namespace associated with the managed cluster. More specifically, the status 
add-on controller sets up RBAC permissions to allow the add-on agent to 
list and get *ManifestWork* objects and create and update *WorkStatus* objects.

## OCM Agent

The OCM Agent Module (a.k.a klusterlet) has two main controllers: the *registration agent*
and the *work agent*. 

The **registration agent** is responsible for registering 
a new cluster into OCM. The agent creates an unaccepted *ManagedCluster* into 
the hub cluster along with a temporary *CertificateSigningRequest* (CSR) resource. 
The cluster will be accepted by the hub control plane if the CSR is approved and 
signed by any certificate provider setting filling `.status.certificate` with legit 
X.509 certificates, and the ManagedCluster resource is approved by setting 
`.spec.hubAcceptsClient` to true in the spec. Upon approval, the registration 
agent observes the signed certificate and persists them as a local secret 
named `hub-kubeconfig-secret` (by default in the `open-cluster-management-agent` namespace) 
which will be mounted to the other fundamental components of klusterlet such as 
the work agent. The registration process in OCM is called *double opt-in* mechanism, 
which means that a successful cluster registration requires both sides of approval 
and commitment from the hub cluster and the managed cluster.

The **work agent** monitors the `ManifestWork` resource in the cluster namespace 
on the hub cluster. The work agent tracks all the resources defined in ManifestWork 
and updates its status. There are two types of status in ManifestWork: the *resourceStatus* 
tracks the status of each manifest in the ManifestWork, and *conditions* reflects the overall 
status of the ManifestWork. The work agent checks whether a resource is *Available*, 
meaning the resource exists on the managed cluster, and *Applied*, meaning the resource 
defined in ManifestWork has been applied to the managed cluster. To ensure the resources 
applied by ManifestWork are reliably recorded, the work agent creates an *AppliedManifestWork* 
on the managed cluster for each ManifestWork as an anchor for resources relating to ManifestWork. 
When ManifestWork is deleted, the work agent runs a *Foreground* deletion, and that ManifestWork 
will stay in deleting state until all its related resources have been fully cleaned in the managed 
cluster.

## OCM Status Add-On Agent

The OCM Status Add-On Agent is a controller that runs alongside the OCM Agent 
in the managed cluster. Its primary function is to track objects delivered 
by the work agent and report the full status of those objects back to the ITS. 
The KubeStellar controller can then use different user-defined summarization 
policies to report status, such as the `wantSingletonReportedState` policy that reports 
full status for each deployed object when the workload is delivered only to one 
cluster. The controller watches *AppliedManifestWork* objects to determine which 
objects have been delivered through the work agent. It then starts dynamic informers 
to watch those objects, collect their individual statuses, and report back the status 
updating *WorkStatus* objects in the namespace associated with the WEC in the ITS.
Installing the status add-on cause status to be returned to `WorkStatus` 
objects for all downsynced objects.

## KubeStellar Controllers Architecture

The KubeStellar controllers architecture is based on common patterns and best 
practices for Kubernetes controllers, such as the 
[<u>Kubernetes Sample Controller</u>](https://github.com/kubernetes/sample-controller). 
A Kubernetes controller uses informers to watch for changes in Kubernetes
objects, caches to store the objects, event handlers to react to
events, work queues for parallel processing of tasks, and a reconciler
to ensure the actual state matches the desired state. However, that
pattern has been extended to provide the following features:

- Using dynamic informers
- Starting informers on all API Resources (except some that do not need
  watching)
- Informers and Listers references are maintained in a hash map and
  indexed by GVK (Group, Version, Kind) of the watched objects.
- Using a common work queue and set of workers, where the key is defined as follows:
  - Key is a struct instead than a string, and contains the following:
    - GVK of the informer and lister for the object that generated the
      event
    - Structured Name of the object 
    - For delete event: Shallow copy of the object being deleted. This
      is required for objects that need to be deleted
      from the managed clusters (WECs)
- Starting & stopping informers dynamically based on creation or
  deletion of CRDs (which add/remove APIs on the WDS).
- One client connected to the WDS space and one (or more in the future)
  to connect to one or more OCM shards.
  - The WDS-connected client is used to start the dynamic
    informers/listers for most API resources in the WDS
  - The OCM-connected client is used to start informers/listers for OCM
    ManagedClusters and to copy/update/remove the wrapped objects
    into/from the OCM mailbox namespaces.

Currently there are two controllers in the KubeStellar controller
manager: the placement controller and the status controllor.

### Placement Controller

The Placement controller is responsible for watching workload and 
placement objects, and wrapping and delivering objects to the ITS
based on placement policies.

The architecture and the event flow for create/update object events is
illustrated in Figure 3 (some details might be omitted to make the flow easier
to understand).

<figure>
  <img src="./placement-controller.png"  alt="Placement Controller">
  <figcaption align="center">Figure 3 - Placement Controller</figcaption>
</figure>

At startup, the controller code sets up the dynamic informers, the event
handler and the work queue as follows:

- lists all API resources
- Filters out some resources
- For each resource:
  - Creates GVK key
  - Registers Event Handler
  - Starts Informer
  - Indexes informer and lister in a map by GVK key
- Waits for all caches to sync
- Starts N workers to process work queue

The reflector is started as part of the informer and watches specific
resources on the WDS API Server; on create/update/delete object events it
puts a copy of the object into the local cache. The informer invokes the
event handler. The handler implements the event handling functions
(AddFunc, UpdateFunc, DeleteFunc)

A typical event flow for a create/update object event will run as
follows:

1.  Informer invokes the event handler AddFunc or UpdateFunc

2.  The event handler does some filtering (for example, to ignore update
    events where the object content is not modified) and then creates a
    key to store a reference to the object in the work queue. The key
    contains the GVK key used to retrieve a reference to the informer
    and lister for that kind of object, and a namespace + name key to
    retrieve the actual object. Storing the key in the work queue is a
    common pattern in client-go as the object may have changed in the
    cache (which is kept updated by the reflector) by the time a worker
    gets a copy from the work queue. Workers should always receive the
    key and use it to retrieve the object from the cache.

3.  A worker pulls a key from the work queue, and then does the
    following processing:

    -  Uses the GVK key to get the reference to the lister for the
        object
    -  Gets the object from the lister cache using the NamespacedName of 
       the object.
    -  If the object was not found (because it was deleted) worker returns. 
       A delete event for that object consumed by the event handler enqueues 
       a key for [Object Deleted](#object-deleted).
    -  Gets the lister for Placement objects and list all placements
    -  Iterates on all placements, and for each placement:
        - evaluates whether the object matches the downsync selection 
          criteria in the Placement.
        - If there is a match, list ManagedClusters and
          find the matching clusters using the label selector expression
          for clusters.
        - If there are matching clusters, add the names of the cluster
          to a hashmap setting the name of the cluster as a key. Cluster
          groups from different placements are merged together.
        - If any of the matched Placements has ` WantSingletonReportedState`
          set to true, clusters are sorted in alphanumerical order and
          only the first cluster is selected for delivery. Note that setting 
          `WantSingletonReportedState` in one of the placements that 
          matches the object affects the behavior for all matching placements, 
          those with matching clusters and those w/o matching clusters. 
    -  If there are no matching clusters, the worker returns without
        actions and is ready to process other events from the queue.
    -  If there are matching clusters:
        - Wraps the object into a ManifestWork
        - Adds a label for each matched Placement to the ManifestWork that 
          is used to track the placement that caused the object to be 
          delivered to one or more clusters. The label contains both the 
          placement name (note that placement is cluster-scoped) and the WDS name. 
          This way, when a placement is deleted or updated it is possible to 
          locate the associated ManifestWorks for deletion or label removal 
          (if more than one label is present, as other placements may “own” the object).
        - For each cluster:
          -  Sets the manifest namespace == name of the cluster
          -  Uses the client connected to the ITS to do a server-side
             apply patch of the manifest into the namespace.
          - At this time there is only one field manager name for the server-side
            apply (`kubestellar`) but the name should include also the WDS
            name to allow detecting conflicts when two different field managers
            try to patch a manifest with the same name.   
        - Worker returns and is ready to pick other keys from the queue.

There are other event flows, based on the object GVK and type of event. 
Error conditions may cause the re-enqueing of keys, resulting in retries.
The following sections broadly describe these flows.

#### Object Deleted 

When an object is deleted from the WDS, the handler’s *DeleteFunc* is
invoked. A shallow copy of the object is added to a field in the key 
before pushing it to the work queue. Then:

- Worker pulls the key from the work queue

- Flow continues the same way as in the create/update scenario, however
  the deletedObject field in the key indicates that the object has been
  deleted and that it needs to be removed from all clusters.

- Running the selectors matching logic as in the create/delete scenario
  once again produces a list of clusters.

- Worker iterates over the list of clusters and deletes the manifest
  work for the object from each namespace associated with the cluster.

- Worker returns and is ready to pick other keys from the queue.

#### Placement Created or Updated 

Worker pulls key from queue; if it is a placement and it has not been
deleted (deletion timestamp not set) it follows the following flow:

- Re-enqueues all objects to force re-evaluation: this is done by 
  iterating all GVK-indexed listers, listing objects for each lister
  and re-enqueuing the key for each object.
- Remove ManifestWorks from ITS for objects no longer matching: generate
  the ManagedByPlacementLabelKey for the current (Placement, WDS) and use
  that to retrieve all the manifestworks associated with the (Placement, WDS).
  Then, for each manifestwork, extract the wrapped object, and re-evaluate the
  object vs. the current placement. If no longer a match (either for the
  "what" or the "where" part) check if each manifestwork has other (Placement, WDS)
  labels. If yes, remove the label, if not, delete the manifestwork.

Re-enqueuing all object keys forces the re-evaluation of all objects vs.
all placements. This is a shortcut as it would be more efficient to
re-evaluate all objects vs. the changed placement only, but it saves
some additional complexity in the code.

#### Placement Deleted

When the placement controller first processes a new Placement, the placement 
controller sets a finalizer on it. The Worker pulls a key from queue; if it is 
a placement and it has been deleted (deletion timestamp is set) it follows the flow below:

- Lists all ManifestWorks on all namespaces in ITS by
  placement label.
- Iterates on all matching ManifestWorks, for each one:
  - Deletes the ManifestWork (if no other placement labels) or the label
    (if other placement labels are present).
- Deletes placement finalizer.

#### New CRD Added

When a new CRD is added, the placement controller needs to start a new informer to watch instances of the new CRD on the WDS.

The worker pulls a key from queue and creates a GVK Key; if it is a CRD and
not deleted:

- Checks if an informer for that GVK was already started, return if that
  is the case.
- If not, creates a new informer
- Registers the event handler for the informer (same one used for all
  other api resources)
- Starts the new informer with a stopper channel, so that it can be
  stopped later on by closing the channel.
- adds informer, lister and stopper channel references to the hashmap
  indexed by the GVK key.

#### CRD Deleted

When a CRD is deleted, the controller needs to stop the informer that
was used to watch instances of that CRD on the WDS. This is because
informers on CRs will keep on throwing exceptions for missing CRDs.

The worker pulls a key from queue and creates a GVK Key; if it is a CRD and it
has been deleted:

- Uses the GVK key to retrieve the stopper channel for the informer.
- Closes the stopper channel
- Removes informer, lister and stopper channel references from the
  hashmap indexed by the GVK key.

### Status Controller

The status controller watches for *WorkStatus* objects on the ITS, and
for WDS objects propagated by a placement with  the flag
`wantSingletonReportedState` set to true, updates the status of those
objects with the corresponding status found in the workstatus object.

The *WorkStatus* objects are created and updated on the ITS by the status add-on.
The high-level flow for the singleton status update is described in Figure 4.

<figure>
  <img src="./status-controller.png"  alt="Status Controller">
  <figcaption align="center">Figure 4 - Status Controller</figcaption>
</figure>

The status add-on tracks objects applied by the work agent by watching 
*AppliedManifestWork* objects. These objects list the GVR, name
and namespace (the latter for namespaced objects) of each object applied
by the related *ManifestWork*. The status add-on then uses this information 
to ensure that a singleton informer is started for each GVR, 
and to track status updates of each tracked object. The status add-on
then creates/updates *WorkStatus* objects in the ITS with the status
of tracked objects in the namespace associated with the WEC cluster.
A `WorkStatus` object contains status for exactly one object, so that 
status updates for one object do not require updates of a whole bundle. 

