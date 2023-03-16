---
categories: ["Coding", "Sprints", "Milesones", "PoC"]
tags: ["code","milestone","poc2023q1"] 
title: "Possible Roadmaps for Particular Use Cases"
linkTitle: "Details"
weight: 2
---

{{% pageinfo %}}
This document outlines thoughts about how soon some particular use cases can work.
{{% /pageinfo %}}

## Background

The outline [mentions features that need not be implemente at
first](outline/#development-roadmap).  In the following sections we
consider some particular use cases.

## MVI

MVI needs customization.  We can demo an MVI sceario without:
self-sufficient edge clusters, summarization, upsync (Return and/or
summarization of reported state from associated objects),
sophisticated handling of workload conflicts.

What about denaturing?

## Compliance-to-Policy

I am not sure what is workable here.  Following are some
possibilities.  They vary in two dimensions.

### How C2P controller consumes reports

#### Through an APIExport view

As outlined in [PR 241](https://github.com/kcp-dev/edge-mc/pull/241):
- the C2P team maintains CRDs, APIResourceSchemas, and an APIExport for
  the policy and report resourcess;
- the workload management workspace has an APIBinding to that APIExport;
- the EdgePlacement selects that APIBinding for downsync;
- the APIBinding goes to the mailbox workspace but not the edge cluster;
- those CRDs are pre-installed on the edge clusters;
- the APIExport's view shows the report objects in the mailbox
  workspaces (as well as anywhere else they exist).

#### Through an API for consuming from mailboxes

The C2P Controller uses the API proposed in [PR
240](https://github.com/kcp-dev/edge-mc/pull/240) to read the report
objects from the mailbox workspaces.  This has the downside of
exposing the mailbox workspaces as part of the edge-mc interface ---
which they were NOT originally intended to be.

#### C2P Controller consumes report summaries prepared by edge-mc

In this scenario:
- we have defined and implemented summarization in edge-mc;
- that summarization is adequate for the needs of the C2P Controller;
- that controller consumes summaries rather than the reports themselves.

### Mailbox vs. Edge

#### Using the current TMC syncer

In this scenario the edge clusters are not self-sufficient; the
workload containers in an edge cluster use kube api services from the
corresponding mailbox workspace.  The key insight here is that from an
outside perspective, a pair of (edge cluster, corresponding mailbox
workspace) operates as a unit and the rest of the world does not care
about internal details of that unit.  But that is only true if you do
not require too much from the networking.  In this scenario, a
workload container runs in the edge cluster and a workload Service
object is about proxying/load-balancing in the edge cluster.  An
admission control webhook normally directs the apiserver to call out
to a virtual IP address associated with a Service; that is a problem
in this scenario because the apiserver in question is the one holding
the mailbox workspace but the Service that gets connections to the
workload containers is in the edge cluster.  This scenario will work
if the C2P workload does not include admission control webhooks.  Note
that Kubernetes release 1.26 introduces CEL-based validating admission
control _policies_, so using them would not involve webhooks.

#### Expand TMC to support webhooks

The problem with webhooks would go away if TMC were expanded to
support them, perhaps through some sort of tunneling so that a client
in the center can open connections to a Service at the edge.

#### Pre-deploy controllers and resources on edge clusters

In this scenario, the PVP/PEP is predeployed on the edge clusters, and
the policy and report resources (which are cluster-scoped) are
predefined there too.  This scenario would continue to use the TMC
syncer, but only need it to downsync the policies and upsync the
reports.

#### Use full EMC

No shortcuts here, no limitations.

