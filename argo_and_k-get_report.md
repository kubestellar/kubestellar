# How Kubectl decides what fields to display while running `kubectl get ...` command ?

Whenever we run `kubectl get ...` command, Kubernetes API server returns a `metav1.Table` object and kubectl does formatting stuff to display neatly on the console. 

### For Native Kuberenetes Resources
The `AddHandlers function` at https://github.com/kubernetes/kubernetes/blob/master/pkg/printers/internalversion/printers.go#L95 defines column headers for all native resource types. Column headers means the fields that we see for an object in the console while running `kubectl get ...`. 

Example: For pods, headers are defined as:
```shell
podColumnDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Ready", Type: "string", Description: "The aggregate readiness state of this pod for accepting traffic."},
		{Name: "Status", Type: "string", Description: "The aggregate status of the containers in this pod."},
		{Name: "Restarts", Type: "string", Description: "The number of times the containers in this pod have been restarted and when the last container in this pod has restarted."},
		{Name: "Age", Type: "string", Description: metav1.ObjectMeta{}.SwaggerDoc()["creationTimestamp"]},
		{Name: "IP", Type: "string", Priority: 1, Description: apiv1.PodStatus{}.SwaggerDoc()["podIP"]},
		{Name: "Node", Type: "string", Priority: 1, Description: apiv1.PodSpec{}.SwaggerDoc()["nodeName"]},
		{Name: "Nominated Node", Type: "string", Priority: 1, Description: apiv1.PodStatus{}.SwaggerDoc()["nominatedNodeName"]},
		{Name: "Readiness Gates", Type: "string", Priority: 1, Description: apiv1.PodSpec{}.SwaggerDoc()["readinessGates"]},
	}
```
NOTE: Columns with `Priority=1` are displayed only in `kubectl get ... -o wide` mode for any native resource type.

For some common native resources, the columnheader definitions are at:
- [Deployment](https://github.com/kubernetes/kubernetes/blob/master/pkg/printers/internalversion/printers.go#L346)
- [ReplicaSet](https://github.com/kubernetes/kubernetes/blob/master/pkg/printers/internalversion/printers.go#L144)
- [DaemonSet](https://github.com/kubernetes/kubernetes/blob/master/pkg/printers/internalversion/printers.go#L157)
- [StatefulSet](https://github.com/kubernetes/kubernetes/blob/master/pkg/printers/internalversion/printers.go#L234)
- [Job](https://github.com/kubernetes/kubernetes/blob/master/pkg/printers/internalversion/printers.go#L173)
- [CronJob](https://github.com/kubernetes/kubernetes/blob/master/pkg/printers/internalversion/printers.go#L186)
- [Service](https://github.com/kubernetes/kubernetes/blob/master/pkg/printers/internalversion/printers.go#L201)
- [Namespace](https://github.com/kubernetes/kubernetes/blob/master/pkg/printers/internalversion/printers.go#L283)
- [Secret](https://github.com/kubernetes/kubernetes/blob/master/pkg/printers/internalversion/printers.go#L291)
- [Configmap](github.com/kubernetes/kubernetes/blob/master/pkg/printers/internalversion/printers.go#L371)


The corresponding print handlers ( functions that are responsible for calculating the field values for each object ) are also implemented in the same file. 

Example: 
For pod, the handler logic is implemented at: https://github.com/kubernetes/kubernetes/blob/master/pkg/printers/internalversion/printers.go#L891  which computes how to populate values for the headers either directly from object manifest or deriving from multiple fields.

### For CRDs
There are not any predefined print handlers implementation for CRDs in Kubernetes. So, API Server uses `.spec.versions[].additionalPrinterColumns` field from CRD spec to dynamically build a `metav1.Table` object and return it. The `additionalPrinterColumns` defines the name of the field, value type ( string, int, etc... ) and jsonPath ( defines where to look for the field value inside CRD object ).
 
Example: For `BindingPolicy` CRD, the additionalfields are defined at https://github.com/kubestellar/kubestellar/blob/main/config/crd/bases/control.kubestellar.io_bindingpolicies.yaml#L19. 

If there is no `additionalPrinterColumns` field in CRD spec, then only `name` and `age` is shown(default) for the CRD object while running `kubectl get ...`.


#
# How ArgoCD determines Health Status of resources deployed by Argocd Application ?

As of now, ArgoCD provides built-in health check functionality for most of the native Kubernetes resources. For any CRDs or any resource that doesn't have built-in health check implemented, ArgoCD allows to define [custom health checks](https://argo-cd.readthedocs.io/en/latest/operator-manual/health/#custom-health-checks) through `Lua` scripts ( though this is not our concern ).

Here are the minimal fields that need to be inside the object manifests for different Resource Kinds so that ArgoCD can perform health checks for that resource. The health check logic for how all these fields are utilized by ArgoCD is already implemented; the resources only need to have those fields. Those resources which have built-in health check require the following fields:

### Deployment
Spec Fields:
- generation
- replicas

Status Fields:
- observedGeneration
- updatedReplicas
- replicas
- availableReplicas

### DaemonSet
Spec Fields:
- generation
- updateStrategy.type

Status Fields:
- observedGeneration
- updatedNumberScheduled
- desiredNumberScheduled
- numberAvailable

### StatefulSet
Spec Fields:
- generation
- replicas
- updateStrategy.type

Status Fields:
- observedGeneration
- readyReplicas
- updatedReplicas
- updateRevision
- currentRevision

### ReplicaSet
Spec Fields:
- generation
- replicas

Status Fields:
- observedGeneration
- availableReplicas

### Pod
Spec Fields:
- restartPolicy

Status Fields:
- containerStatuses
- initContainerStatuses ( if initContainers are used )
- phase

### Job
Spec Fields:

Status Fields:
- conditions
- conditions[*].type
- conditions[*].status

### Service
Spec Fields:
- Not Needed

Status Fields:
- loadBalancer.ingress (required only for LoadBalancer type services; for ClusterIP, Headless, NodePort, or ExternalName services, status is not used for health checks)

### PVC( PersistentVolumeClaim )
Spec Fields:

Status Fields:
- phase

#
This information should be enough to start the implementation by considering the fields required for Argo.The `.spec` fields are not under our control;, we are only responsible to propagate the  `.status` fields necessary for ArgoCD to perform health checks for particular Resource Kind.

For the passive Resources like:
- Secrets
- ConfigMaps
- Roles, ClusterRoles
- RoleBindings, ClusterRoleBindings
- ServiceAccount
- Namespaces
, etc...

ArgoCD doesn't have built-in health checks. They are considered healthy by default as soon as they are created in the target cluster ( unless user has not defined custom health check Lua scripts, again this is not our concern ). Also, these resources generally don't have meaningful status fields; so we might not have to propagate `status` back during the upsync process.
