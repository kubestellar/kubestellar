# cluster-watchall

This is a demonstration of a way to monitor all objects in all
workspaces in a kcp server.  This is: a demonstration of how to use
`pkg/apiwatch`, a demonstration of how to monitor objects, and a
utility to investigate what API resources are actually defined in a
kcp server (more precisely than `kubectl api-resources`).

This demonstration only reports object creations.  Note also that it
is workspaces, not other kinds of logical clusters, that matter here.

This demonstrates, among other things, use of the informer on "API
resources" in [pkg/apiwatch](../../pkg/apiwatch).

The command line flags include all the usual ones for `kubectl` (see [the kubectl common flags doc](https://v1-24.docs.kubernetes.io/docs/reference/kubectl/kubectl/); I am not sure whether [the in-cluster overrides](https://v1-24.docs.kubernetes.io/docs/reference/kubectl/#in-cluster-authentication-and-namespace-overrides) is also relevant), of which a few prominent ones are shown below, plus one for configuring the metrics & debug endpoint.

```console
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

```shell
cluster-watchall --context system:admin
```

This demo outputs what it discovered as three interleaved CSV tables.
The tables have rows of the following formats.

```console
CLUSTER,$clusterName
RESOURCE,$clusterName,$apiGroup,$apiVersion,$kind,$informable
OBJECT,$clusterName,$apiGroupVersion,$kind,$namespace,$name
```

Following is some sample output.

```console
CLUSTER,kvdk2spgmbix
RESOURCE,1p4q3bat75x35ez3,,v1,ConfigMap,true
RESOURCE,root,authentication.k8s.io,v1,TokenReview,false
OBJECT,1p4q3bat75x35ez3,v1,Namespace,,kcp-system
OBJECT,kvdk2spgmbix,v1,Secret,default,default-token-2fjw9
OBJECT,kvdk2spgmbix,rbac.authorization.k8s.io/v1,ClusterRoleBinding,,workspace-admin
```
