# edge-mc
Edge Multi-Cluster

Driving Concerns

General Context: hierarchy, infrastructure & platform, roles & responsibilities, integration architecture, security issues
Runtime in[ter]dependence: An edge location may need to operate independently of the center and other edge locations​
Non-namespaced objects: need general support
Cardinality of destinations: A source object may propagate to many thousands of destinations. ​
for more detailed info see comments below.

Consequent Issues

Desired placement expression​: Need a way for one center object to express large number of desired copies​
Scheduling/syncing interface​: Need something that scales to large number of destinations​
Rollout control​: Client needs programmatic control of rollout, possibly including domain-specific logic​
Customization: Need a way for one pattern in the center to express how to customize for all the desired destinations​
Status from many destinations​: Center clients may need a way to access status from individual edge copies.​
Status summarization​: Client needs a way to say how statuses from edge copies are processed/reduced along the way from edge to center​
Approach

Collaboratively design a component set similar to those found in current KCP TMC implementation (dedicated Workspace type, scheduler, syncer-like mechanism, edgeplacement object definition, status collection strategy, and etc.)
Specify a multi-phased proof-of-concept inclusive of component architecture, interfaces, and example workloads
Validate phases of proof-of-concept with KCP, Kube SIG-Multicluster, and CNCF community members interested in Edge
