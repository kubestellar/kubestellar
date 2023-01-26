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

- An implementation that supports intermittent connectivity.
- An implementation that supports a large number of edge clusters.
- More than one SyncTarget per Location.
- A hierarchy with more than two levels.

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

## Edge Service Provider workspace

The edge multi-cluster service is provided by one workspace that
includes the following things.

- An APIExport of the edge API group.
- The edge controllers: scheduler, placement translator, mailbox
  controller, and status sumarizer.

## Workload Management workspaces

The users of edge multi-cluster primarily maintain these.  Each one
contains the following things.

- A workload spec.  That is, some namespaces containing namespaced
  workload objects --- and possibly later on some non-namespaced
  workload objects.  Initially the workload objects must be of types
  drawn from a list of those supported by the EMC implementation
  (because extension by user CRDs is not supported yet).
- EMC interface objects: EdgePlacement, SinglePlacementSlice, those
  for customization and summarization.

## Mailbox Controller

This controller maintains one mailbox workspace per SyncTarget.  Each
of these mailbox workspaces is used for a distinct TMC problem (e.g.,
Placement object).  These workspaces are all children of the edge
service provider workspace.

## Edge Scheduler

This controller monitors the EdgePlacement, Location, and SyncTarget
objects and maintains the results of matching.  For each EdgePlacement
object this controller maintains an adjacent collection of
SinglePlacementSlice objects holding the matches for that
EdgePlacement.

## Placement Translator

This controller translates each EdgePlacement object into a collection
of TMC Placement objects and corresponding related objects.  For each
matching SinglePlacement, the placement translator maintains a TMC
Placement object and copy of the workload namespaces and their
contents --- customized as directed.

## Syncers

Open question: what maintains them?

## Status Summarizer

For each EdgePlacement object and related objects this controller
maintains the directed status summary objects.
