This document includes the text from the space-framework design document

The full design document can be found here: https://docs.google.com/document/d/1WZCxlGoaCZVymqMhLCa5JZeURdCTzju8C5ABS8V4lW0/edit?usp=sharing


# Space support for KubeStellar

## Background

KubeStellar architecture leverages spaces as the container building block for various purposes. For example, the separation between the KubeStellar service provider and the inventory is done by using different spaces for the inventory and the service provider. 

The KubeStellar design currently 

1. Separate different user workloads into different spaces (workload management spaces)
2. Synchronise each EdgeCluster into a dedicated space (called the mailbox space ). 

In the future some of the above design may change (e.g., the need for  1:1 mapping  of MBLC:EdgeCluster or the need for separate WMLC for each workload), however we believe some usage of spaces will still be needed.

Today KubeStellar is using KCP workspaces as spaces and is practically built on top of KCP. 

This proposal tries to disconnect KubeStellar architecture and internal design and the underlying support for spaces. 

The main concept of the proposal is to define a space provider interface that the KubeStellar will be able to consume without being exposed to the internal implementation details of the Space provider itself. 

In addition, in order to allow an easy usage of the spaces by clients (e.g., controllers, etc..) we introduce a cluster-aware client that will be able to retrieve the needed cluster information by simply specifying the cluster ID (basically the cluster name). 


## Space provider framework

ill be created/deleted by the

The diagram illustrates the high level architecture of the space framework. The framework supports two types of spaces:

1. Managed Space : This cluster is created through an explicit request to the management api server to create a new Space. As a result the Space  manager will initiate a request to the Space  provider to create the actual cluster. The desired state of the managed Space  cluster is defined through the Space object. For example, in order to delete this cluster you need to simply delete the corresponding Space object. 
2. Unmanaged Space : A cluster can be created directly on the provider and then imported (through discovery) into the space framework. Such a cluster is defined as “unmanaged” and its life cycle is controlled out-of-bound through the provider. For example, the Space object will be created/deleted by the Space  manager according to the existence of the cluster on the provider.

The Space  framework uses the following API resources and namespaces:

- **SpaceProviderDesc:** Holds the information needed in order to interact with a specific Space  provider. For stage 1, we have different Space  provider clients for each type of provider (e.g., KIND, k3s, etc..), so the ClusterProviderDesc also includes the specific provider type. For phase 2, we plan to have a common provider REST interface, so there will not be different provider types. This is a cluster level CRD.
- **Provider namespace:** For each provider we create and use a dedicated namespace to hold the Space objects. The namespace is created by the Space  manager when a new ClusterProvideDesc is created.
- **Space:** Represent the space created at the provider. As mentioned before, there are two types of Spaces - managed and unmanaged. Space objects resides in the namespace created for their provider

The Space also holds a reference to the SpaceProviderDesc associated with this Space . This can be defined when creating the Space object. If no SpaceProviderDesc reference is supplied then the default provider is used (the framework always has a provider that is defined as the default provider)

### space manager 

The Space  manager is a Kuberentes controller that is responsible for reconciling the Space and the SpaceProviderDesc resources. For managed spaces the Space  manager will use the provider client to perform the needed operations (create/delete) through the provider. The framework supports using multiple provider types and multiple provider instances of the same type (e.g., KCP, remote-KIND, etc.) 

The Space  manager get events from two sources:

1. From the management API server: Changes of the Space and SpaceProviderDesc resources.
2. From the Client Provider: Changes on the status of the space itself

The following diagram illustrates the state diagram of the Space

**Discovery**

As mentioned before, the Space  manager also supports discovery of spaces created out of band (i.e., not through creating a Space object. 

This is done through the discovery process - each discovered cluster that doesn’t have a corresponding Space object is considered as a space that should be imported into the framework. When detecting such a new cluster the cluster-manager creates a Space object representing this new cluster and sets its mode to “unmanaged” .

The assumption behind this flow is that there shouldn’t be orphan spaces (i.e. spaces without corresponding Space object), so that each such space needs to be imported. The Space Space manager uses finalizers for the space deletion flow (the Space object is not deleted until the corresponding space is removed)  so no orphan Space(s) are expected.

The SpaceProvideDesc resource allows the provider to define a predefined prefix for the cluster names that should be discovered and imported. 

**Provider Client Interface**

Note: For stage 2, the interface will be defined as a set of REST APIs that the provider needs to expose.

Defines the set of operations all Space  provider clients should support (Note - this is just a preliminary sample )


## Space aware client 

The Space aware client (msclient) allows clients/controllers to easily get access to the underlying space by simply using the space name. There is no direct interaction between the msclient and the Space  provider, and therefore the msclient is transparent to the specific provider of the spaces. 

The only requirement is that the space can be accessed through regular Kube APIs when using the appropriate kubeconfig information. 

**MCClient**

- Holds a cache of Space’s access info (e.g.,  RestConfig)

  - Constantly watch for changes in the available spaces

- Exposes utility functions to get RestConfig and/or ClientSet for a specific space (according to cluster name) 

  - For ClientSet there are currently 3 types: Kube, KubeStellar, and Dynamic

****

**Next steps & future investigations & todos**

1. Inject a the MC library calls at the request (or REST) level 
2. Combine the 3 different clientSet functions into single one using generics


### Cross cluster list() & watch() operations 
**Note**: This is still work in progress

The msclient also exposes the ability to perform watch() and list() on resources across clusters. It will do that by returning a cross-cluster lister-watcher and cross-cluster informer that the user can use for these operations. 

The cross cluster List() is done by simply performing a List() on each space and combining all results into a single reply.

For the cross cluster Watch() the msclient opens Watch() into each of the spaces and merges the incoming events into a result event channel that is returned to the user.

**Open questions (some under investigations)**

1. Conflict in object names (or even RV): Depending on the Space provider, the spaces can be completely independent so we can’t assume any rules regarding conflicts between the clusters. 

2. How to supply a global cross cluster list watcher implementation while still using the strongly typed client-go generated resource functions

   1. While the resource type can be managed with go generics this won’t help with the need to have different clientsets for kube and kubestaller resources.
   2. Dynamics clientsets and Dynamic informers can solve the single implementation issue, however we lose the advantage of strongly typed code (e.g., catch errors in compilation time) 

3. Do we need to add a cluster identification to the returned objects so that the user can see which cluster each object was coming from ? If so this will require wrapping the resource or adding a new annotation ?

4. Should we support dynamic add/remove of Space during cross-cluster watch(). This goes into a grey area of events coming from a cluster that is not there anymore. 
