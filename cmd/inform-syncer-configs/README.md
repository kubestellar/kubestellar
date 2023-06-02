# inform-syncer-configs

This is a simple diagnostic and demo that exercises the mailboxwatch
package to make and use an informer on `SyncerConfig` objects.

Aside from the usual, the command line arguments are as follows.

```console
$ go run ./cmd/inform-syncer-configs -h
Usage of inform-syncer-configs:
...
      --all-cluster string               The name of the kubeconfig cluster to use for access to the SyncerConfig objects in all clusters
      --all-context string               The name of the kubeconfig context to use for access to the SyncerConfig objects in all clusters (default "system:admin")
      --all-kubeconfig string            Path to the kubeconfig file to use for access to the SyncerConfig objects in all clusters
      --all-user string                  The name of the kubeconfig user to use for access to the SyncerConfig objects in all clusters
...
      --espw-cluster string              The name of the kubeconfig cluster to use for access to the edge service provider workspace
      --espw-context string              The name of the kubeconfig context to use for access to the edge service provider workspace
      --espw-kubeconfig string           Path to the kubeconfig file to use for access to the edge service provider workspace
      --espw-user string                 The name of the kubeconfig user to use for access to the edge service provider workspace
...
pflag: help requested
exit status 2
```

This command requires two kube client configurations.  One points at
the edge service provider workspace (ESPW) with authorization to list
and watch Workspace objects (the mailbox Workspaces) in that
workspace.  The other client config points at the server base with
authorization to list and watch `SyncerConfig` objects from all
clusters.

Currently the output is very simple logging.

Following is an example of its usage.

```console
$ kubectl ws root:espw
Current workspace is "root:espw".

$ go run ./cmd/inform-syncer-configs 
I0419 00:32:51.695348   94959 main.go:103] "Running"
I0419 00:32:51.699433   94959 main.go:112] "Notified" action="add" object=&{TypeMeta:{Kind:SyncerConfig APIVersion:edge.kcp.io/v1alpha1} ObjectMeta:{Name:the-one GenerateName: Namespace: SelfLink: UID:d0432bef-fcbe-4b1b-8ec6-c15555bd13b0 ResourceVersion:1513 Generation:3 CreationTimestamp:2023-04-19 00:04:40 -0400 EDT DeletionTimestamp:<nil> DeletionGracePeriodSeconds:<nil> Labels:map[] Annotations:map[kcp.io/cluster:2m0dtpobj6tuzdst] OwnerReferences:[] Finalizers:[] ZZZ_DeprecatedClusterName: ManagedFields:[{Manager:placement-translator Operation:Update APIVersion:edge.kcp.io/v1alpha1 Time:2023-04-19 00:04:41 -0400 EDT FieldsType:FieldsV1 FieldsV1:{"f:spec":{".":{},"f:namespaceScope":{".":{},"f:namespaces":{},"f:resources":{}},"f:upsync":{}},"f:status":{}} Subresource:}]} Spec:{NamespaceScope:{Namespaces:[commonstuff] Resources:[{GroupResource:{Group:rbac.authorization.k8s.io Resource:roles} APIVersion:v1} {GroupResource:{Group:rbac.authorization.k8s.io Resource:rolebindings} APIVersion:v1} {GroupResource:{Group:networking.k8s.io Resource:ingresses} APIVersion:v1} {GroupResource:{Group: Resource:resourcequotas} APIVersion:v1} {GroupResource:{Group: Resource:secrets} APIVersion:v1} {GroupResource:{Group: Resource:services} APIVersion:v1} {GroupResource:{Group: Resource:limitranges} APIVersion:v1} {GroupResource:{Group: Resource:configmaps} APIVersion:v1} {GroupResource:{Group:apps Resource:deployments} APIVersion:v1} {GroupResource:{Group: Resource:pods} APIVersion:v1} {GroupResource:{Group: Resource:serviceaccounts} APIVersion:v1} {GroupResource:{Group:coordination.k8s.io Resource:leases} APIVersion:v1}]} ClusterScope:[] Upsync:[{APIGroup:group1.test Resources:[sprockets flanges] Namespaces:[orbital] Names:[george cosmo]} {APIGroup:group2.test Resources:[cogs] Namespaces:[] Names:[william]}]} Status:{LastSyncerHeartbeatTime:<nil>}}
I0419 00:32:51.699639   94959 main.go:112] "Notified" action="add" object=&{TypeMeta:{Kind:SyncerConfig APIVersion:edge.kcp.io/v1alpha1} ObjectMeta:{Name:the-one GenerateName: Namespace: SelfLink: UID:ff627d73-b8b6-469c-8340-47e6662b01fd ResourceVersion:1546 Generation:3 CreationTimestamp:2023-04-19 00:04:40 -0400 EDT DeletionTimestamp:<nil> DeletionGracePeriodSeconds:<nil> Labels:map[] Annotations:map[kcp.io/cluster:2caac9u6nrytp2ro] OwnerReferences:[] Finalizers:[] ZZZ_DeprecatedClusterName: ManagedFields:[{Manager:placement-translator Operation:Update APIVersion:edge.kcp.io/v1alpha1 Time:2023-04-19 00:04:40 -0400 EDT FieldsType:FieldsV1 FieldsV1:{"f:spec":{".":{},"f:namespaceScope":{".":{},"f:resources":{}},"f:upsync":{}},"f:status":{}} Subresource:} {Manager:kubectl-edit Operation:Update APIVersion:edge.kcp.io/v1alpha1 Time:2023-04-19 00:25:56 -0400 EDT FieldsType:FieldsV1 FieldsV1:{"f:spec":{"f:namespaceScope":{"f:namespaces":{}}}} Subresource:}]} Spec:{NamespaceScope:{Namespaces:[specialstuff otherstuff commonstuff] Resources:[{GroupResource:{Group: Resource:serviceaccounts} APIVersion:v1} {GroupResource:{Group:rbac.authorization.k8s.io Resource:rolebindings} APIVersion:v1} {GroupResource:{Group: Resource:secrets} APIVersion:v1} {GroupResource:{Group: Resource:services} APIVersion:v1} {GroupResource:{Group: Resource:pods} APIVersion:v1} {GroupResource:{Group: Resource:configmaps} APIVersion:v1} {GroupResource:{Group:apps Resource:deployments} APIVersion:v1} {GroupResource:{Group: Resource:resourcequotas} APIVersion:v1} {GroupResource:{Group:rbac.authorization.k8s.io Resource:roles} APIVersion:v1} {GroupResource:{Group:coordination.k8s.io Resource:leases} APIVersion:v1} {GroupResource:{Group: Resource:limitranges} APIVersion:v1} {GroupResource:{Group:networking.k8s.io Resource:ingresses} APIVersion:v1}]} ClusterScope:[] Upsync:[{APIGroup:group3.test Resources:[widgets] Namespaces:[] Names:[*]} {APIGroup:group1.test Resources:[sprockets flanges] Namespaces:[orbital] Names:[george cosmo]} {APIGroup:group2.test Resources:[cogs] Namespaces:[] Names:[william]}]} Status:{LastSyncerHeartbeatTime:<nil>}}
```
