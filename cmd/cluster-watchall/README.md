# cluster-watchall

This is a demonstration of a way to monitor all objects in all
workspaces in a kcp server.  A little more carefully: the objects
monitored are those of kinds that support informers --- that is, they
support the `list` and `watch` verbs.  This demonstration only reports
object creations.  Note also that it is workspaces, not other kinds of
logical clusters, that matter here.

This demonstrates, among other things, use of the informer on "API
resources" in [pkg/apiwatch](../../pkg/apiwatch).

The notable command line flags are as follows.

```
      --kubeconfig string                Path to the kubeconfig file to use for CLI requests
      --context string                   The name of the kubeconfig context to use
      --server-bind-address ipport       The IP address with port at which to serve /metrics and /debug/pprof/ (default :10203)

```

This program consumes the `KUBECONFIG` envar and `--kubeconfig`
command line flag as usual.  If the context is specified then that one
is selected from those in the kubeconfig, otherwise the current
kubeconfig context is used.  The chosen kubeconfig context must allow
all-cluster operations; typically this is the context named
`system:admin`.

Thus, a typical minimal invocation might be something like the
following.

```
cluster-watchall --context system:admin
```

This demo outputs what it discovered as three interleaved CSV tables.
The tables have rows of the following formats.

```
CLUSTER,$clusterName
RESOURCE,$clusterName,$apiGroup,$apiVersion,$kind,$informable
OBJECT,$clusterName,$apiGroupVersion,$kind,$namespace,$name
```

Following is some sample output.

```
CLUSTER,kvdk2spgmbix
RESOURCE,1p4q3bat75x35ez3,,v1,ConfigMap,true
RESOURCE,root,authentication.k8s.io,v1,TokenReview,false
OBJECT,1p4q3bat75x35ez3,v1,Namespace,,kcp-system
OBJECT,kvdk2spgmbix,v1,Secret,default,default-token-2fjw9
OBJECT,kvdk2spgmbix,rbac.authorization.k8s.io/v1,ClusterRoleBinding,,workspace-admin
```
