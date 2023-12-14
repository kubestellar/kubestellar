# Space Framework

## Terminology 
**Physical Space**, **pSpace** An entity that behaves like a Kubernetes kube-apiserver (including the persistent storage behind it) and the subset of controllers in the kube-controller-manager that are concerned with API machinery generalities (without the management of containerized workloads). A kcp logical cluster is an example. A regular Kubernetes cluster is another example.

**Space Object**, **Space** - A kind of Kubernetes API object that _represents_ a pSpace.  Its "kind" is `Space`, we sometimes include the "object" part when writing specifically about the API objects to clearly distinguish them from what they represent.

**Space Provider** - A Thing that manages the life cycle of multiple pSpaces. This includes at minimal create/delete/List operations on pSpaces. KIND is an example for Space Provider (that manages KIND clusters), KCP is another example - KCP manages KCP's workspaces. 

## High level architecture  & Main components
![SF overall view](images/SF-All.drawio.svg "SF overall view")

The Space Framework (SF) is a generic management framework for space providers and pSpaces. The framework defines an abstraction layer for space providers and pSpace management that allows clients (both script/kubectl based and client-go based clients) to use spaces while maintaining the clients decoupled from the specific pSpace and space provider that is being used. 

The space framework uses two CRDs:
- **SpaceProviderDesc:** This CRD represents a space provider it holds the information needed in order to interact with a specific Space  provider. 
- **Space:** This CRD represents a space as defined above. Each Space includes reference to two secrets that allow clients to connect to the pSpace that it represents:
  1. Secret for in-cluster access - accessing the pSpace from within the same cluster that hosts the pSpace
  2. Secret for External access:  accessing the pSpace from outside the cluster that hosts the pSpace
  
  It's the responsibility of the space provider to be able to supply these access secrets. The details on how the SM retrieves the information is described in the following sections


Spaces can be divided into two categories:
1. Provider based spaces: The Space Manager (SM) uses a space provider to get information (or perform operations) on the spaces. We can think of such pSpaces as "belonging" to the space provider. There are two types of spaces in this category which controls the way these spaces are added to the space framework: 
   * **Managed Space**: The pSpace is created through an explicit request to the management api server to create a new Space. For Managed spaces the space framework initiates a request to the Space provider to create the pSpace, the space framework will then continuously reconcile between the Space and the pSpace states. For managed spaces the desired state is controlled through the Space object. For example, in order to delete the pSpace the client need to delete the corresponding Space object and the SF will interact with the space provider in order to delete the pSpace.
   * **Unmanaged Space**: The SF supports discovery of existing pSpaces on a space provider and adding them as "Unmanaged Spaces" into the framework. This mode is similar to the import mode in terms of the "source of truth" - the desired state is defined by the pSpace. However, as opposed to import, here the pSpaces are automatically discovered through the space provider regular interface and the pSpace access information is retrieved from the  space provider similarly to how it is done in the managed space case. The SF supports pattern based filtering rules to limit the discovery of pSpaces. 
  2. Providerless spaces: This spaces don't belong to a space provider and are called "Imported spaces". Such Space is created by importing an existing pSpace. In this case the Space is not linked to any space provider and the client which imports that Space needs to supply the access information that allows other clients to connect to the pSpace. For imported spaces the desired state is derived by the pSpace. This means that the SF is not responsible for the life cycle management of the pSpace but only update the corresponding Space object. For example, when deleting the space object of an existing imported pSpace the space framework will recreate the space object. 
### Space Manager 
The Space Manager (SM) is a Kubernetes controller that is responsible for maintaining the state of the Space objects. The SM reconciles the Space and SpaceProviderDesc objects, and monitors the state of the pSpaces through the space provider.   
The SM uses a library of space provider adaptors to communicate with the space providers. Currently the SF supports 3 space providers KCP, KIND, and KubeFlex, and includes 3 space provider adaptors that the SM uses.

### Space provider adaptors
The space manager uses a set of space provider adaptors that interact with the space provider. All space provider adaptors implement a simple "space client interface" that includes basic pSpace life cycle operations. In addition the interface also defines the events that are sent from the adaptors to the space manager. Currently the space provider adaptors are implemented as library inside the Space Manager.

### Space aware client (SAC)
The Space aware client (SAC) allows clients/controllers to easily get access to the underlying space by simply using the space name and optionally the SpaceProviderDesc name. There is no direct interaction between the SAC and the Space  provider, and therefore the SAC is transparent to the specific provider of the spaces. 
![Space Aware Client](images/SAC.drawio.svg "Space Aware Client")
The only requirement is that the space can be accessed through regular Kube APIs when using the appropriate kubeconfig information. 

#### Main features
- Constantly watch for changes in the available spaces. The SAC is using an informer on Space objects using the SF Space client-go APIs.
    - Holds a cache of the pSpace access info for each Space 
- Exposes utility functions to get the [rest.Config](https://pkg.go.dev/k8s.io/client-go@v0.28.3/rest#Config) data structure that gives access to a physical space with the identity of a user that is authorized to do everything. This convenient function only needs the space name, and optionally the provider name (if the provider name is omitted the default provider is being used. The namespace where the Space resides can be derived from the name of the provider. In the future we will add a SAC space aware ClientSets as well.

## Architecture details 
When a new SpaceProviderDesc is created the SM creates a new namespace that will be used to host all the Space objects representing pSpaces on that space provider.  
The name of the generated namespace is `spaceprovider-<provider name>`

![Space, SpaceProviderDesc and the Space provider namespace](images/SF-NS.drawio.svg "SF Namespaces")

Note: In the future the spaces will be created in their own namespace and not in namespace per provider. 

The Space also holds a reference to the SpaceProviderDesc associated with this Space and some capabilities (e.g., the space type). The status of the space holds references to secrets holding the access information for that pSpace (currently there are two secrets, one for accessing the pSpace from within the Kubernetes cluster that pSpace resides on and one for accessing th pSpace from outside that cluster). For managed spaces the SF retrieves the access information from the space provider, and then creates the relevant secrets and updates the Space object with the secret reference

**Discovery**  
As mentioned before, the Space manager also supports discovery of spaces created out of band (i.e., not through creating Space objects).

Each provider adaptor constantly watches for pSpace creation/deletion events and feeds the SM with those events. In case a pSpace that matches the discovery filtering is detected and there is no corresponding Space object the SM creates a Space object representing this new detected pSpace and sets its mode to “unmanaged” .
When the pSpace is deleted the SM deletes the corresponding Space object.

**Creation & Deletion flow of managed Spaces**  
The SM sets the status of the Space according to the actual status of the pSpace. When a space is created a creation command is sent to the space provider and the state of the Space is set to `Initializing`. Only when the pSpace is becoming available (each space provider validate this in its own way) the SM sets the Space state to `Ready`.   
The SM uses finalizers for the space deletion flow - the Space object is not removed until the corresponding pSpace is deleted by the space provider.

## Examples

**SpaceProviderDesc**
The following is a SpaceProviderDesc for for a KubeFlex provider type. The SpaceProviderDesc will become "Ready" only after the secret referenced by `secretRef` is successfully retrieved.
```
apiVersion: space.kubestellar.io/v1alpha1
kind: SpaceProviderDesc
metadata:
  name: pkflex
spec:
  ProviderType: "kubeflex"
  SpacePrefixForDiscovery: "ks-"
  secretRef:
    namespace: default
    name: pkflex
```

**Space**
The following is a Space YAML that is used to create a managed space on the previously defined KubeFlex provider. 
The use need to specify the correct namespace for the space
```
apiVersion: space.kubestellar.io/v1alpha1
kind: Space
metadata:
  name: space5
  namespace: "spaceprovider-pkflex"
spec:
  SpaceProviderDescName: "pkflex"
  Type: "managed"
```

## Space Provider adaptors (SPA)
The space provider adaptor is responsible for all the interaction with the Space Provider (used by the SM). The SM today includes implementation of 3 provider adaptors for 3 space provider types - KCP, KubeFlex and KIND. 
When a new SpaceProviderDesc is created, the SM creates an instance of an SPA of the corresponding provider type. 
SPA implements the [ProviderClient](https://github.com/kubestellar/kubestellar/blob/main/space-framework/pkg/space-manager/providerclient/client_interface.go) interface. This interface is relatively simple and includes basic CRUD+Watch operations. 

```
type ProviderClient interface {
	Create(name string, opts Options) error
	Delete(name string, opts Options) error

	// List returns a list of spaces.
	// This method is used to discover the initial set of spaces
	// and to refresh the list of spaces periodically.
	ListSpaces() ([]SpaceInfo, error)

	// List returns a list of space names.
	// This method is used to discover the initial set of spaces
	// and to refresh the list of spaces periodically.
	ListSpacesNames() ([]string, error)

	// Get returns a space info.
	Get(name string) (SpaceInfo, error)

	// Watch returns a Watcher that watches for changes to a list of spaces
	// and react to potential changes.
	Watch() (Watcher, error)
}
```

When the SM creates an instance of the SPA, it also starts a background loop using the  ```Watch()``` interface of the SPA to continuously monitor the status of the pSpace.

### KubeFlex SPA
The KubeFlex space provider exposes KubeFlex's ```ConrolPlane``` as the pSpaces. The KubeFlex SPA interacts with the KubeFlex server to manage these workspaces. 

### KCP SPA
The KCP space provider exposes KCP's workspaces as the pSpaces. The KCP SPA interacts with the KCP server to manage these workspaces. 
We don't support workspace hierarchy and all workspaces are created under the ```root``` workspace and therefor the KCP SPA uses the Space name as the unique identifier for the workspace.  
Example: When the user creates a Space with name ```mySpace``` the SPA will create a workspace named ```mySpace``` with path of ```root:mySpace```

### KIND SPA

#### Example: Create a Space on the KCP space provider
Suppose we created a SpaceProviderDesc of type KCP named sp-kcp1, and we now want to create a Space. As mentioned before when the SpaceProviderDesc was created the SM also created a corresponding SPA instance. 
1. Create a Space object with a reference to sp-kcp1
2. The SM finds the SPA the corresponding SPA 
3. The SM sends a ```Create()``` command through this SPA instance. 
4. The SPA sends a command to create KCP workspace on that KCP server. 
5. The SM sets the status of the Space to ```initialzing``` 
6. When the workspace is "ready" the SPA generates the appropriate kubeconfig information to access this specific workspace and send back the relevant pSpace info back to the SM as a  [SpaceInfo](https://github.com/kubestellar/kubestellar/blob/main/space-framework/pkg/space-manager/providerclient/client_interface.go) object. 
7. The SM creates the relevant access secrets and place a reference to those secrets in the Space status. The SM then move the Status of the Space to ```Ready```
