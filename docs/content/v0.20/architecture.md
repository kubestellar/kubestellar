# KubeStellar Architecture

**NOTE**: This is only a very rudimentary start.

## Higher Level

At the higher level there are three modules.

1. KubeFlex
2. [the status addon](https://github.ibm.com/dettori/status-addon)
3. KubeStellar

The first two are independent and KubeStellar depends on them.

The status addon defines the `WorkStatus` Kind of objects and some (what?) associated stuff.

## KubeStellar

KubeStellar currently has two major parts.

1. OCM for inventory and transport
2. A central controller-manager that provides the added functionalities

We plan to make the transport pluggable in the near future.

KubeStellar currently exposes users directly to OCM inventory management and WEC registration.

KubeSteller currently prereqs KubeFlex to provide one or more spaces. We plan to make this optional in the near future.

There are two roles for spaces: Inventory and Mailbox Space (IMBS) and Workload Description Space (WDS).

An IMBS holds OCM inventory (`ManagedCluster`) objects and mailbox namespaces. The mailbox namespaces and their contents are implementation details that users do not deal with. Each mailbox workspace corresponds 1:1 with a WEC and holds `ManifestWork` objects managed by the central KubeStellar controllers.

A WDS holds user workload objects and the user's objects that form the interface to KubeStellar control. Currently the only control objects are `Placement` objects. We plan to add `Customization` objects soon, and later objects to specify summarization.

There is a central controller-manager per WDS. Currently this controller-manager runs in the KubeFlex hosting cluster.
