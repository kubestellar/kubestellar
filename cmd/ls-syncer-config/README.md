This command demonstrates the difficulty with all-custer access to
resources not built into kcp.

To see the demo, start in the `edge-mc` directory and any convenient
kcp workspace and do the following commands.

```shell
kubectl create -f config/crds

kubectl create -f config/exports

kubectl create -f - <<EOF
apiVersion: edge.kcp.io/v1alpha1
kind: SyncerConfig
metadata:
  name: another-one
spec:
  namespaceScope:
    namespaces:
    - commonstuff
    resources:
    - apiVersion: v1
      group: apps
      resource: deployments
EOF
```

The demo then goes like this; here `root:espw` is the workspace where
I did those commands.

```console
$ go run ./cmd/ls-syncer-config --cluster-name root:compute
E0418 14:46:48.134765   40858 main.go:79] "Failed to .EdgeV1alpha1().SyncerConfigs().Cluster().List()" err="the server could not find the requested resource (get syncerconfigs.edge.kcp.io)"
E0418 14:46:48.135285   40858 main.go:89] "Failed to .Cluster().EdgeV1alpha1().SyncerConfigs().List()" err="the server could not find the requested resource (get syncerconfigs.edge.kcp.io)"
E0418 14:46:48.135682   40858 main.go:96] "Failed to .EdgeV1alpha1().SyncerConfigs().List()" err="the server could not find the requested resource (get syncerconfigs.edge.kcp.io)"

$ go run ./cmd/ls-syncer-config --cluster-name root:espw
I0418 14:47:10.927145   40887 main.go:81] "api-then-cluster list succeeded" list=&{TypeMeta:{Kind: APIVersion:} ListMeta:{SelfLink: ResourceVersion:760 Continue: RemainingItemCount:<nil>} Items:[{TypeMeta:{Kind:SyncerConfig APIVersion:edge.kcp.io/v1alpha1} ObjectMeta:{Name:another-one GenerateName: Namespace: SelfLink: UID:4cfedb1a-1bf0-49b7-9365-48d9fb6ea0af ResourceVersion:758 Generation:1 CreationTimestamp:2023-04-18 14:33:53 -0400 EDT DeletionTimestamp:<nil> DeletionGracePeriodSeconds:<nil> Labels:map[] Annotations:map[kcp.io/cluster:1orrokxlkpd2f96m] OwnerReferences:[] Finalizers:[] ZZZ_DeprecatedClusterName: ManagedFields:[{Manager:kubectl-create Operation:Update APIVersion:edge.kcp.io/v1alpha1 Time:2023-04-18 14:33:53 -0400 EDT FieldsType:FieldsV1 FieldsV1:{"f:spec":{".":{},"f:namespaceScope":{".":{},"f:namespaces":{},"f:resources":{}}}} Subresource:}]} Spec:{NamespaceScope:{Namespaces:[commonstuff] Resources:[{GroupResource:{Group:apps Resource:deployments} APIVersion:v1}]} ClusterScope:[] Upsync:[]} Status:{LastSyncerHeartbeatTime:<nil>}}]}
I0418 14:47:10.928631   40887 main.go:91] "cluster-then-api list succeeded" list=&{TypeMeta:{Kind: APIVersion:} ListMeta:{SelfLink: ResourceVersion:760 Continue: RemainingItemCount:<nil>} Items:[{TypeMeta:{Kind:SyncerConfig APIVersion:edge.kcp.io/v1alpha1} ObjectMeta:{Name:another-one GenerateName: Namespace: SelfLink: UID:4cfedb1a-1bf0-49b7-9365-48d9fb6ea0af ResourceVersion:758 Generation:1 CreationTimestamp:2023-04-18 14:33:53 -0400 EDT DeletionTimestamp:<nil> DeletionGracePeriodSeconds:<nil> Labels:map[] Annotations:map[kcp.io/cluster:1orrokxlkpd2f96m] OwnerReferences:[] Finalizers:[] ZZZ_DeprecatedClusterName: ManagedFields:[{Manager:kubectl-create Operation:Update APIVersion:edge.kcp.io/v1alpha1 Time:2023-04-18 14:33:53 -0400 EDT FieldsType:FieldsV1 FieldsV1:{"f:spec":{".":{},"f:namespaceScope":{".":{},"f:namespaces":{},"f:resources":{}}}} Subresource:}]} Spec:{NamespaceScope:{Namespaces:[commonstuff] Resources:[{GroupResource:{Group:apps Resource:deployments} APIVersion:v1}]} ClusterScope:[] Upsync:[]} Status:{LastSyncerHeartbeatTime:<nil>}}]}
E0418 14:47:10.929254   40887 main.go:96] "Failed to .EdgeV1alpha1().SyncerConfigs().List()" err="the server could not find the requested resource (get syncerconfigs.edge.kcp.io)"

$ go run ./cmd/ls-syncer-config --cluster-name '*'
panic: A specific cluster must be provided when scoping, not the wildcard.

goroutine 1 [running]:
github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned/cluster/typed/edge/v1alpha1.(*syncerConfigsClusterInterface).Cluster(0xc00048b1a0, {{0x7ff7bfeff8dd?, 0x1ee5320?}})
	/Users/mspreitz/go/src/github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned/cluster/typed/edge/v1alpha1/syncerconfig.go:58 +0xe5
main.main()
	/Users/mspreitz/go/src/github.com/kcp-dev/edge-mc/cmd/ls-syncer-config/main.go:76 +0x3d8
exit status 2
```

And, a little more directly, try the command

```shell
kubectl get --context system:admin --raw https://$your_ip:6443/clusters/%2A/apis | jq .
```

and observe that the `edge.kcp.io` group is _not_ included, while

```shell
kubectl get --context system:admin --raw https://$your_ip:6443/clusters/root:espw/apis | jq .
```

produces a listing that _does_ include `edge.kcp.io`.
