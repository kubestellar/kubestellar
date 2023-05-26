## Background

The outline [mentions features that need not be implement at
first](../outline.md/#development-roadmap).  In the following sections we
consider some particular use cases.

## MVI

MVI needs customization.  We can demo an MVI scenario without:
self-sufficient edge clusters, summarization, upsync (Return and/or
summarization of reported state from associated objects),
sophisticated handling of workload conflicts.

What about denaturing?

## Compliance-to-Policy

I am not sure what is workable here.  Following are some
possibilities.  They vary in two dimensions.

### How C2P controller consumes reports

#### Through view of workload APIExport and workload APIBindings

As outlined in [PR 241](https://github.com/kcp-dev/edge-mc/pull/241):
- the C2P team maintains CRDs, APIResourceSchemas, and an APIExport for
  the policy and report resources;
- the C2P team puts those APIResourceSchemas and that APIExport in a
  kcp workspace of their choice;
- the workload management workspace has an APIBinding to that APIExport;
- the EdgePlacement selects that APIBinding for downsync;
- the APIBinding goes to the mailbox workspace but not the edge cluster;
- those CRDs are pre-installed on the edge clusters;
- the APIExport's view shows the report objects in the mailbox
  workspaces (as well as anywhere else they exist).

This is not a great choice because of "those CRDs are pre-installed on
the edge clusters".

It is also not a great choice because it requires the C2P team to
maintain two copies of the Kyverno resource definitions.

This is a bad choice because it is not consistent with the preferred
way to demonstrate installation of Kyverno, which is to have Helm
install Kyverno into the workload managemet workspace.

#### Through a new kind of view

We could define a new kind of view that does what we want.

#### Through view of EMC APIExport and mailbox APIBindings

This approach also uses APIExport and APIBinding objects but in a
different way than above.  In this approach the placement translator
maintains one APIExport in the edge service provider workspace and a
corresponding APIBinding object in each mailbox workspace, and they
work together as follows.

The APIExport has an empty
[LatestResourceSchemas](https://github.com/kcp-dev/kcp/blob/v0.11.0/pkg/apis/apis/v1alpha1/types_apiexport.go#L108)
but a large dynamic
[PermissionClaims](https://github.com/kcp-dev/kcp/blob/v0.11.0/pkg/apis/apis/v1alpha1/types_apiexport.go#L165)
slice.  In particular, there is a PermissionClaim for every resource
involved in downsync or upsync in any EdgePlacement object.  Some day
we might try something more granular, but today is not that day.

In each mailbox workspace, the corresponding APIBinding's [list of
accepted
PermissionClaims](https://github.com/kcp-dev/kcp/blob/v0.11.0/pkg/apis/apis/v1alpha1/types_apibinding.go#L81)
has an etry for every resource downsynced or upsynced to that
workspace.

As a consequence, the APIExport's view holds all the objects whose
kind/resource is defined by those APIBindings.

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

The latest plan is to [use full EMC](#use-full-emc).

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

In this scenario, the PVP/PEP is pre-deployed on the edge clusters, and
the policy and report resources (which are cluster-scoped) are
predefined there too.  This scenario would continue to use the TMC
syncer, but only need it to downsync the policies and upsync the
reports.

#### Use full EMC

No shortcuts here, no limitations.

