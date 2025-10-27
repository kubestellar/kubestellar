# kubectl-rbac-flatten command

The kubectl-rbac-flatten command reads `RoleBinding` and
`ClusterRoleBinding` objects and computes and prints the outer product
of the subjects and the rules. The intent is to produce output that is
easy to search for an explanation of why something is allowed.

The kubectl-rbac-flatten command takes all the standard arguments for
a Kubernetes command-line tool, and so can be used as a [kubectl
plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/).

## Data Model

The semantics of the collection of RBAC objects in a cluster is a
collection of _privilege grants_, each of which has the following
attributes.  Here we expand the semantics of the RBAC objects to
atomic tuples; that is, each one identifies just one subject, verb,
object, role-ism, and binding --- with the exception of wildcards.

- **Subject**: who the privilege is granted to. A user (ordinary or
  ServiceAccount) or group. Kubernetes authentication of a request
  identifies the user and the set of groups that the user belongs
  to. This set always includes either `system:authenticated` or
  `system:unauthenticated`.

- **Verb**: the action being allowed. When the _object_ is a
  _resource_, the verb is one of: `create`, `update`, `patch`,
  `delete`, `deletecollection`, `get`, `list`, `watch`.  When the
  object is not a resource, the verb is one of: `get`, `put`,
  `delete`, `post`.  The wildcard `*` grants usage of all verbs.

- **Object**: the thing that the subject is allowed to invoke the verb on.
  There are two cases here: API object, and "non-resource" URL path.

    A non-resource URL path is just the path part of a URL; the schema
    and authority of the authorized URL are implicitly those of the
    apiserver.

    An API object is identified by a "resource", "object name"
    (misnamed "ResourceName" in the RBAC API), and possibly
    "namespace".  A resource exists within an _api group_. A resource
    is identified by a string of the form
    `<base>.<apigroup>/<subresource>`, where `<base>` is the lowercase
    plural way of identifying a kind of object. The `.<apigroup>` is
    empty for the core API group.  The `/<subresource>` is optional.
    Example resources: `nodes`, `jobs.batch`,
    `clusterroles.rbac.authorization.k8s.io`,
    `certificatesigningrequests/approval.certificates.k8s.io`. The
    wildcard `*` can stand for all resources, for all namespaces, and
    for all object names.

    Each resource has a _scope_, which is either "cluster" or
    "namespace".

    Only a ClusterRoleBinding grants access to a
    cluster-scoped resource. Nothing stops a RoleBinding from
    referring to a ClusterRole that refers to cluster-scoped
    resources, but that RoleBinding grants no actual privileges
    regarding those resources.

- **Role-ism**: the Role or ClusterRole object that defines the verb
  and object.

- **Binding**: the RoleBinding or ClusterRoleBinding object that
  associates the role-ism with the subject.

## Query semantics

This command can be directed to filter the reported grants to include
only those that involve a given set of subjects, set of verbs, and/or
set of objects.

## Usage

No filesystem I/O is done except that involved in reading the
kubeconfig file.

The `kubectl-rbac-flatten` command takes all the standard Kubernetes
client command line flags (e.g., `--kubeconfig`, `--context`) plus the
additional ones listed below.

```
  -o, --output-format string             output format, either json or table (default "table")
      --resources strings                comma-separated list of resources to include; resource syntax is plural.group/subresource; .group and /subresource are omitted when appropriate; '*' means all resources (default [*])
      --show-role                        include role in listing for resource grans (default true)
```


## Output

Warnings about ineffective RBAC are logged to stderr, and as much
processing as possible is done.

There are two possible output formats, chosen by a command line flag.

### JSON Output


The JSON output does a minimal outer product, producing a tuple per
(Subject, PolicyRule). The tuples also include some additional fields
describing their source. The precise data type is as follows, where
`NamespacedName` is imported from `k8s.io/apimachinery/pkg/types`.

```go
type Tuple struct {
	Binding       NamespacedName
	RoleInCluster bool
	RoleName      string
	Subject       rbac.Subject
	Rule          PolicyRule
}

type PolicyRule struct {
	Verbs            []string
	Resources        []string
	ObjectNames      []string
	NonResourcePaths []string
}
```

The output is one JSON value, an array. Each tuple is written on one
line, with the comma separators on other lines, making it easy to use
text-based tools for simple searches. For non-trivial searching, you
can use a real JSON processor.

Following is an example of JSON output.

```json
[
{"Binding":{"Namespace":"","Name":"cluster-admin"},"RoleInCluster":true,"RoleName":"cluster-admin","Subject":{"kind":"Group","apiGroup":"rbac.authorization.k8s.io","name":"system:masters"},"Rule":{"Verbs":["*"],"Resources":null,"ObjectNames":null,"NonResourcePaths":["*"]}}
,
{"Binding":{"Namespace":"","Name":"dpctlr-node-viewer"},"RoleInCluster":true,"RoleName":"node-viewer","Subject":{"kind":"ServiceAccount","name":"dual-pods-controller","namespace":"default"},"Rule":{"Verbs":["get","list","watch"],"Resources":["nodes"],"ObjectNames":null,"NonResourcePaths":null}}
,
...
]
```

### Tabular Output

The tabular output does a maximal cross product, so that no tuple in
the result has a list/slice/array-structured field. The tabular output
has two tables, one for resource-based rules and one for non-resource
URLs. The second one is omitted if `--api-goups` is given anything
other than `*`.

The fields in one row are separated by TABs. The empty string API
group is output as an empty string (thus between two TABs).

The resource-based table omits the column for the role, for brevity.

Following are excerpts from example tabular output, with `--show-role=false`.

```
BINDING                                                          SUBJECT                                                        VERB               RESOURCE                                                          OBJNAME
/cluster-admin                                                   G:system:masters                                               *                  apiservices.apiregistration.k8s.io                                *
/cluster-admin                                                   G:system:masters                                               *                  apiservices/status.apiregistration.k8s.io                         *
...

BINDING                                    ROLE                                      SUBJECT                    VERB   PATH
/cluster-admin                             cluster-admin                             G:system:masters           *      *
/kubeadm:cluster-admins                    cluster-admin                             G:kubeadm:cluster-admins   *      *
/system:discovery                          system:discovery                          G:system:authenticated     get    /api
...
```

The columns are as follows.

- **BINDING:** The namespace/name for the `ClusterRoleBinding` or `RoleBinding`; namespace is empty fo a `ClusterRoleBinding`.
- **ROLE:** The name of the `ClusterRole` or `Role` object referenced by the BINDING.
- **SUBJECT:** A slightly compacted representation of the `rbac.Subject`.
- **VERB**
- **RESOURCE:** Including API group and subresource. Omitted if exactly 1 resource is being queried for.
- **OBJNAME:** Name of an individual object. `*` matches all names.
- **NRURL:** A URL path usable in Non-Resource URLs; `*` matches all paths.
