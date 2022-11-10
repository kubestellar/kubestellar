# edge-mc

## Overview
edge-mc is a subproject of kcp focusing on concerns arising from edge multicluster use cases:

- Hierarchy, infrastructure & platform, roles & responsibilities, integration architecture, security issues
- Runtime in[ter]dependence: An edge location may need to operate independently of the center and other edge locations​
- Non-namespaced objects: need general support
- Cardinality of destinations: A source object may propagate to many thousands of destinations. ​ 

## Goals

- collaboratively design a component set similar to those found in the current kcp TMC implementation (dedicated Workspace type, scheduler, syncer-like mechanism, edge placement object definition, status collection strategy, and etc.)
- Specify a multi-phased proof-of-concept inclusive of component architecture, interfaces, and example workloads
- Validate phases of proof-of-concept with kcp, Kube SIG-Multicluster, and CNCF community members interested in Edge

## Areas of exploration

- Desired placement expression​: Need a way for one center object to express large number of desired copies​
- Scheduling/syncing interface​: Need something that scales to large number of destinations​
- Rollout control​: Client needs programmatic control of rollout, possibly including domain-specific logic​
- Customization: Need a way for one pattern in the center to express how to customize for all the desired destinations​
- Status from many destinations​: Center clients may need a way to access status from individual edge copies
- Status summarization​: Client needs a way to say how statuses from edge copies are processed/reduced along the way from edge to center​.

## Quickstart

TBD :building_construction:

## Next Steps

TBD :building_construction:

## Contributing

We ❤️ our contributors! If you're interested in helping us out, please head over to our [Contributing](CONTRIBUTING.md) guide.

## Getting in touch

You may get in touch with us with the channels availabke in the [kcp-dev community](https://github.com/kcp-dev/kcp#getting-in-touch).  
