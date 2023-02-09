# 2023q1 PoC

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
- Return or summarization of status from associated objects (e.g.,
  ReplicaSet or Pod objects associated with a given Deployment
  object).
- A hierarchy with more than two levels.
- More than baseline security (baseline being, e.g., HTTPS, Secret
  objects, non-rotating bearer token based service authentication).
- A good design for bootstrapping the workload management in the edge
  clusters.
- Very strong isolation between tenants in the edge computing
  platform.

It is TBD whether the implementation will support intermittent
connectivity.  This depends on whether we can quickly and easily get a
syncer that creates the appropriately independent objects in the edge
cluster and itself tolerates intermittent connectivity.

Also: initially this PoC will not transport non-namespaced objects.
CustomResourceDefinitions are thus among the objects that will not be
transported.  Transport of non-namespaced objects may be added after
the other goals are achieved.

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

To complete the plumbing of the syncers, for each SyncTarget the
mailbox controller creates the following needed accompanying items in
that SyncTarget's workspace (which must authorize the mailbox
controller to do this).

1. A ServiceAccount that the syncer will authenticate as.
2. A ClusterRole manipulating that SyncTarget and the
   APIResourceImports (what are these?).
3. A ClusterRoleBinding that links that ServiceAccount with that
   ClusterRole.

FYI, those are the things that `kubectl kcp workload sync` directly
creates besides the SyncTarget.

## Edge Service Provider workspace

The edge multi-cluster service is provided by one workspace that
includes the following things.

- An APIExport of the edge API group.
- The edge controllers: scheduler, placement translator, mailbox
  controller, and status sumarizer.

## Workload Management workspaces

The users of edge multi-cluster primarily maintain these.  Each one
contains the following things.

- APIBindings to the APIExports of workload object types.
- A workload spec.  That is, some namespaces containing namespaced
  workload objects --- and possibly later on some non-namespaced
  workload objects.  Initially the workload objects must be of types
  drawn from a list of those supported by the EMC implementation
  (because extension by user CRDs is not supported yet).
- EMC interface objects: EdgePlacement, SinglePlacementSlice, those
  for customization and summarization.

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
for creating and maintaining these objects is called the problem of
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
Placement object and copy of the workload namespaces and their
contents --- customized as directed.

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

This user also creates one or more namespaces containing specs of
desired state of workload.

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
