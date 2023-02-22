# 2023q1 PoC

## Status of this memo

This summarizes the current state of design work that is still in
progress.

## Introduction

This is a quick demo of a fragment of what we think is needed for edge
multi-cluster.  It is intended to demonstrate the following points.

- Separation of inventory and workload management.
- The focus here is on workload management, and that strictly reads
  inventory.
- What passes from inventory to workload management is kcp TMC
  Location and SyncTarget objects.
- Use of a kcp workspace as the container for the central spec of a workload.
- Propagation of desired state from center to edge, as directed by
  EdgePlacement objects and the Location and SyncTarget objects they reference.
- Interfaces designed for a large number of edge clusters.
- Interfaces designed with the intention that edge clusters operate
  independently of each other and the center (e.g., can tolerate only
  occasional connectivity) and thus any "service providers" (in the
  technical sense from kcp) in the center or elsewhere.
- Rule-based customization of desired state.
- Propagation of reported state from edge to center.
- Summarization of reported state in the center.
- The edge opens connections to the center, not vice-versa.
- An edge computing platform "product" that can be deployed (as
  opposed to a service that is used).

Some important things that are not attempted in this PoC include the following.

- An implementation that supports a large number of edge clusters or
  any other thing that requires sharding for scale.
- More than one SyncTarget per Location.
- Return or summarization of reported state from associated objects
  (e.g., ReplicaSet or Pod objects associated with a given Deployment
  object).
- A hierarchy with more than two levels.
- User control over ordering of propagation from center to edge,
  either among destinations or kinds of objects.
- More than baseline security (baseline being, e.g., HTTPS, Secret
  objects, non-rotating bearer token based service authentication).
- A good design for bootstrapping the workload management in the edge
  clusters.
- Support for workload object types that are not either built into kcp
  or imported via a kcp APIBinding.
- Very strong isolation between tenants in the edge computing
  platform.

It is TBD whether the implementation will support intermittent
connectivity.  This depends on whether we can quickly and easily get a
syncer that creates the appropriately independent objects in the edge
cluster and itself tolerates intermittent connectivity.

As further matters of roadmapping development of this PoC:
customization may be omitted at first, and summarization may start
with only a limited subset of the implicit functionality.

This PoC builds on TMC and makes some compromises to accommodate that.
The implementation involves workload components (syncers) writing
status information to inventory objects (SyncTargets).

## Roles and Responsibilities

### Developers/deployers/admins/users of the inventory management layer

### Developers of the workload management layer

### Deployers/admins of the workload management layer

### Users of the workload management layer

## Design overview

In very brief: the design is to reduce each edge placement problem to
many instances of kcp's TMC problem.

See [the overview picture](Edge-PoC-2023q1.svg) for an overview
picture.

## Inventory Management workspaces

This design takes as a given that something maintains some kcp
workspaces that contain dynamic collections of Location and SyncTarget
objects as defined in [kcp
TMC](https://github.com/kcp-dev/kcp/tree/main/pkg/apis), and that one
view can be used to read those Location objects and one view can be
used to read those SyncTarget objects.

To complete the plumbing of the syncers, each inventory workspace that
contains a SyncTarget needs to also contain the following associated
objects.  FYI, these are the things that `kubectl kcp workload sync`
directly creates besides the SyncTarget.  Ensuring their presence is
part of the problem of bootstrapping the workload management layer and
is not among the things that this PoC takes a position on.

1. A ServiceAccount that the syncer will authenticate as.
2. A ClusterRole manipulating that SyncTarget and the
   APIResourceImports (what are these?).
3. A ClusterRoleBinding that links that ServiceAccount with that
   ClusterRole.

## Edge Service Provider workspace

The edge multi-cluster service is provided by one workspace that
includes the following things.

- An APIExport of the edge API group.
- The edge controllers: scheduler, placement translator, mailbox
  controller, and status sumarizer.

## Workload Management workspaces

The users of edge multi-cluster primarily maintain these.  Each one of
these has both control (API objects that direct the behavior of the
edge computing platform) and data (API objects that hold workload
desired and reported state).

### Data objects

The workload desired state is represented by kube-style API objects,
in the way that is usual in the Kubernetes milieu.  For edge computing
we need to support both cluster-scoped (AKA non-namespaced) kinds as
well as namespaced kinds of objects.

The use of a workspace as a mere container presents a challenge,
because some kinds of kubernetes API objects at not merely data but
also modify the behavior of the apiserver holding them.  To resolve
this dilemma, the edge users of such a workspace will use a special
view of the workspace that holds only data objects.  The ones that
modify apiserver behavior will be translated by the view into
"denatured" versions of those objects in the actual workspace so that
they have no effect on it.  And for these objects, the transport from
center-to-edge will do the inverse: translate the denatured versions
into the regular ("natured"?) versions for appearance in the edge
cluster.  Furthermore, for some kinds of objects that modify apiserver
behavior we want them "natured" at both center and edge.  There are
thus a few categories of kinds of objects.  Following is a listing,
with with the particular kinds that appear in kcp or plain kubernetes.

#### Needs to be denatured in center, natured in edge

For these kinds of objects, clients of the real workspace can
manipulate such objects and they will modify the behavior of the
workspace, while clients of the edge computing view will manipulate
distinct objects that have no effect on the behavior of the workspace.

| APIVERSION | KIND | NAMESPACED |
| ---------- | ---- | ---------- |
| admissionregistration.k8s.io/v1 | MutatingWebhookConfiguration | false |
| admissionregistration.k8s.io/v1 | ValidatingWebhookConfiguration | false |
| flowcontrol.apiserver.k8s.io/v1beta2 | FlowSchema | false |
| flowcontrol.apiserver.k8s.io/v1beta2 | PriorityLevelConfiguration | false |
| rbac.authorization.k8s.io/v1 | ClusterRole | false |
| rbac.authorization.k8s.io/v1 | ClusterRoleBinding | false |
| rbac.authorization.k8s.io/v1 | Role | true |
| rbac.authorization.k8s.io/v1 | RoleBinding | true |
| v1 | LimitRange | true |
| v1 | ResourceQuota | true |
| v1 | ServiceAccount | true |

#### Needs to be natured in center and edge

These should have their usual effect in both center and edge; they
need no distinct treatment.

Note, however, that they _do_ have some sequencing implications.  They
have to be created before any dependent objects, deleted after all
dependent objects.

| APIVERSION | KIND | NAMESPACED |
| ---------- | ---- | ---------- |
| apiextensions.k8s.io/v1 | CustomResourceDefinition | false |
| v1 | Namespace | false |

### Needs to be natured in center, not destined for edge

| APIVERSION | KIND | NAMESPACED |
| ---------- | ---- | ---------- |
| apis.kcp.io/v1alpha1 | APIBinding | false |

A workload management workspace generally has APIBindings to workload
APIs.  These bindings cause corresponding CRDs to be created in the
same workspace.  The CRDs propagate to the edge, the APIBindings do
not.

#### For features not supported

These are part of k8s or kcp APIs that are not supported by the edge
computing platform.

| APIVERSION | KIND | NAMESPACED |
| ---------- | ---- | ---------- |
| apiregistration.k8s.io/v1 | APIService | false |
| apiresource.kcp.io/v1alpha1 | APIResourceImport | false |
| apiresource.kcp.io/v1alpha1 | NegotiatedAPIResource | false |
| apis.kcp.io/v1alpha1 | APIConversion | false |

The APIService objects are of two sorts: (a) those that are built-in
and describe object types built into the apiserver and (b) those that
are added by admins to add API groups served by custom external
servers.  Sort (b) is not supported because this PoC does not support
custom external servers in the edge clusters.  Sort (a) is not
programmable in this PoC, but it might be inspectable.

#### Not destined for edge

These kinds of objects are concerned with either (a) TMC control or
(b) workload data that should only exist in the edge clusters.  These
will not be available in the view used by edge clients to maintain
their workload desired and reported state.

| APIVERSION | KIND | NAMESPACED |
| ---------- | ---- | ---------- |
| apis.kcp.io/v1alpha1 | APIExport | false |
| apis.kcp.io/v1alpha1 | APIExportEndpointSlice | false |
| apis.kcp.io/v1alpha1 | APIResourceSchema | false |
| apps/v1 | ControllerRevision | true |
| authentication.k8s.io/v1 | TokenReview | false |
| authorization.k8s.io/v1 | LocalSubjectAccessReview | true |
| authorization.k8s.io/v1 | SelfSubjectAccessReview | false |
| authorization.k8s.io/v1 | SelfSubjectRulesReview | false |
| authorization.k8s.io/v1 | SubjectAccessReview | false |
| certificates.k8s.io/v1 | CertificateSigningRequest | false |
| coordination.k8s.io/v1 | Lease | true |
| core.kcp.io/v1alpha1 | LogicalCluster | false |
| core.kcp.io/v1alpha1 | Shard | false |
| events.k8s.io/v1 | Event | true |
| scheduling.kcp.io/v1alpha1 | Location | false |
| scheduling.kcp.io/v1alpha1 | Placement | false |
| tenancy.kcp.io/v1alpha1 | ClusterWorkspace | false |
| tenancy.kcp.io/v1alpha1 | Workspace | false |
| tenancy.kcp.io/v1alpha1 | WorkspaceType | false |
| topology.kcp.io/v1alpha1 | Partition | false |
| topology.kcp.io/v1alpha1 | PartitionSet | false |
| v1 | Binding | true |
| v1 | ComponentStatus | false |
| v1 | Event | true |
| v1 | Node | false |
| workload.kcp.io/v1alpha1 | SyncTarget | false |

#### Already denatured in center, want natured in edge

These are kinds of objects that kcp already gives no interpretation
to.

This is the default category of kind of object --- any kind of data
object not specifically listed in another category is implicitly in
this category.  Following are the kinds from k8s and kcp that fall in
this category.

| APIVERSION | KIND | NAMESPACED |
| ---------- | ---- | ---------- |
| apps/v1 | DaemonSet | true |
| apps/v1 | Deployment | true |
| apps/v1 | ReplicaSet | true |
| apps/v1 | StatefulSet | true |
| autoscaling/v2 | HorizontalPodAutoscaler | true |
| batch/v1 | CronJob | true |
| batch/v1 | Job | true |
| networking.k8s.io/v1 | Ingress | true |
| networking.k8s.io/v1 | IngressClass | false |
| networking.k8s.io/v1 | NetworkPolicy | true |
| node.k8s.io/v1 | RuntimeClass | false |
| policy/v1 | PodDisruptionBudget | true |
| scheduling.k8s.io/v1 | PriorityClass | false |
| storage.k8s.io/v1 | CSIDriver | false |
| storage.k8s.io/v1 | CSINode | false |
| storage.k8s.io/v1 | CSIStorageCapacity | true |
| storage.k8s.io/v1 | StorageClass | false |
| storage.k8s.io/v1 | VolumeAttachment | false |
| v1 | ConfigMap | true |
| v1 | Endpoints | true |
| v1 | PersistentVolume | false |
| v1 | PersistentVolumeClaim | true |
| v1 | Pod | true |
| v1 | PodTemplate | true |
| v1 | ReplicationController | true |
| v1 | Secret | true |
| v1 | Service | true |

### Control objects

These are the EdgePlacement objects, their associated
SinglePlacementSlice objects, and the objects that direct
customization and summarization.

## Mailbox workspaces

The mailbox controller maintains one mailbox workspace for each
SyncTarget.  A mailbox workspace acts as a workload source for TMC,
prescribing the workload to go to the corresponding edge pcluster and
holding the corresponding TMC Placement object.

A mailbox workspace contains the following items.

1. APIBindings (maintained by the mailbox controller) to APIExports of
   workload object types.
2. Workload namespaces holding workload objects, post customization.
3. A TMC Placement object.

## Edge cluster

Also called edge pcluster.

One of these contains the following items.  FYI, these are the things
in the YAML output by `kubectl kcp workload sync`.  The responsibility
for creating and maintaining these objects is part of the problem of
bootstrapping the workload management layer and is not among the
things that this PoC takes a position on.

- A namespace that holds the syncer and associated objects.
- A ServiceAccount that the syncer authenticates as when accessing the
  views of the center and when accessing the edge cluster.
- A Secret holding that ServiceAccount's authorization token.
- A ClusterRole listing the non-namespaced privileges that the
  syncer will use in the edge cluster.
- A ClusterRoleBinding linking the syncer's ServiceAccount and ClusterRole.
- A Role listing the namespaced privileges that the syncer will use in
  the edge cluster.
- A RoleBinding linking the syncer's ServiceAccount and Role.
- A Secret holding the kubeconfig that the syncer will use to access
  the edge cluster.
- A Deployment of the syncer.

## Mailbox Controller

This controller maintains one mailbox workspace per SyncTarget.  Each
of these mailbox workspaces is used for a distinct TMC problem (e.g.,
Placement object).  These workspaces are all children of the edge
service provider workspace.

## Edge Scheduler

This controller monitors the EdgePlacement, Location, and SyncTarget
objects and maintains the results of matching.  For each EdgePlacement
object this controller maintains an associated collection of
SinglePlacementSlice objects holding the matches for that
EdgePlacement.  These SinglePlacementSlice objects appear in the same
workspace as the corresponding EdgePlacement; the remainder of how
they are linked is TBD.

## Placement Translator

This controller translates each EdgePlacement object into a collection
of TMC Placement objects and corresponding related objects.  For each
matching SinglePlacement, the placement translator maintains a TMC
Placement object and copy of the workload --- customized as directed.
Note that this involves sequencing constraints: CRDs and namespaces
have to be created before anything that uses them, and deleted after
everything that uses them.  Note also that everything that has to be
denatured in the workload management workspace also has to be
denatured in the mailbox workspace.

The job of the placement translator can be broken down into the
following four parts.

- Resolve each EdgePlacement's "what" part to a list of particular
  workspace items (namespaces and non-namespaced objects).
- Maintain the association between the resolved "where" from the edge
  scheduler and the resolved what.
- Maintain the copies, with customization, of the workload objects
  from source workspace to mailbox workspaces.
- Maintain the TMC Placement objects that derive from the
  EdgePlacement objects.

## Syncers

This design nominally uses TMC and its syncers, but that can not be
exactly true because these syncers need to translate between denatured
objects in the mailbox workspace and natured objects in the edge
cluster.  Or perhaps not, if there is an additional controller in the
edge cluster that handles the denatured-natured relation.

## Status Summarizer

For each EdgePlacement object and related objects this controller
maintains the directed status summary objects.

## Usage Scenario

The usage scenario breaks, at the highest level, into two parts:
inventory and workload.

### Inventory Usage

A user with infrastructure authority creates one or more inventory
management workspaces.  Each such workspace needs to have the
following items, which that user will create if they are not
pre-populated by the workspace type.

- An APIBinding to the `workload.kcp.io` APIExport to get
  `SyncTarget`.
- An APIBinding to the `scheduling.kcp.io` APIExport to get
  `Location`.
- A ServiceAccount (with associated token-bearing Secret) (details
  TBD) that the mailbox controller authenticates as.
- A ClusterRole and ClusterRoleBinding that authorize said
  ServiceAccount to do what the mailbox controller needs to do.

This user also creates one or more edge clusters.

For each of those edge clusters, this user creates the following.

- a corresponding SyncTarget, with an annotation referring to the
  following Secret object, in one of those inventory management
  workspaces;
- a Secret, in the same workspace, holding a kubeconfig that the
  central automation will use to install the syncer in the edge
  cluster;
- a Location, in the same workspace, that matches only that
  SyncTarget.

### Workload usage

A user with workload authority starts by creating one or more workload
management workspaces.  Each needs to have the following, which that
user creates if the workload type did not already provide.

- An APIBinding to the APIExport of `edge.kcp.io` from the edge
  service provider workspace.
- For each of the Edge Scheduler, the Placement Translator, and the
  Status Summarizer:
  - A ServiceAccount for that controller to authenticate as;
  - A ClusterRole granting the privileges needed by that controller;
  - A ClusterRoleBinding that binds those two.

This user also uses the edge-workspace-as-container view of each such
workspace to describe the workload desired state.

This user creates one or more EdgePlacement objects to say which
workload goes where.  These may be accompanied by API objects that
specify rule-baesd customization, specify how status is to be
summarized.

The edge-mc implementation propagates the desired state from center to
edge and collects the specified information from edge to center.

The edge user monitors status summary objects in their workload
management workspaces.

The status summaries may include limited-length lists of broken
objects.

Full status from the edge is available in the mailbox workspaces.
