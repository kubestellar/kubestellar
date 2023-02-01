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

Some important things that are not attempted in this PoC include the following.

- An implementation that supports a large number of edge clusters.
- More than one SyncTarget per Location.
- A hierarchy with more than two levels.

It is TBD whether the implementation will support intermittent
connectivity.  This depends on whether we can quickly and easily get a
syncer that creates the appropriately independent targets and itself
tolerates intermittent connectivity.

Also: initially this PoC will not transport non-namespaced objects.
CustomResourceDefinitions are thus among the objects that will not be
transported.  Transport of non-namespaced objects may be added after
the other goals are achieved.

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

To enable the workload management layer to manipulate each edge
cluster, that cluster's SyncTarget has an annotation that refers to a
Secret that holds a kubeconfig that can be used to maintain the syncer
and workload in that edge cluster.  These annotations and Secrets are
created by the inventory management layer for use by the workload
management layer.

To complete the plumbing of the syncers, for each SyncTarget the
mailbox controller creates the following needed accompanying items in
that SyncTarget's workspace (which must authorize the mailbox
controller to do this).

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

One of these contains the following items, maintained by the ??? in
the workload management layer.

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
