# kubectl-rbac-flatten command

The kubectl-rbac-flatten command reads `RoleBinding` and
`ClusterRoleBinding` objects and computes and prints the outer product
of the subjects and the rules. The intent is to produce output that is
easy to search for an explanation of why something is allowed.

The kubectl-rbac-flatten command takes all the standard arguments for
a Kubernetes command-line tool, and so can be used as a [kubectl
plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/).

## Usage

No filesystem I/O is done except that involved in reading the
kubeconfig file.

The `kubectl-rbac-flatten` command takes all the standard Kubernetes
client command line flags (e.g., `--kubeconfig`, `--context`) plus the
additional ones listed below.

```
      --api-groups strings               comma-separated list of API groups to include; '*' means all (default [*])
  -o, --output-format string             output format, either json or table (default "table")
      --resources strings                comma-separated list of resources to include; '*' means all (default [*])
```

The syntax shown for the default value of `--api-groups` and
`--resources` is wrong. The correct syntax is a comma-separated
collection of strings. For example, `--api-groups
networking.k8s.io,certificates.k8s.io`.

When filtering on API group or "resource", an RBAC rule that matches
"*" is included in the result.

## Output

Errors are logged to stderr, and as much processing as possible is
done.

There are two possible output formats, chosen by a command line flag.

### JSON Output


The JSON output does a minimal outer product, producing a tuple per
(Subject, PolicyRule). The tuples also include some additional fields
describing their source. The precise data type is as follows, where
`NamespacedName` is imported from `k8s.io/apimachinery/pkg/types`.

```go
type Tuple struct {
	Binding  NamespacedName
	RoleName string
	Subject  rbac.Subject
	Rule     rbac.PolicyRule
}
```

The output is one JSON value, an array. Each tuple is written on one
line, with the comma separators on other lines, making it easy to use
text-based tools for simple searches. For non-trivial searching, you
can use a real JSON processor.

Following is an example of JSON output.

```json
[
{"Binding":{"Namespace":"","Name":"cluster-admin"},"RoleName":"cluster-admin","Subject":{"kind":"Group","apiGroup":"rbac.authorization.k8s.io","name":"system:masters"},"Rule":{"verbs":["*"],"apiGroups":["*"],"resources":["*"]}}
,
{"Binding":{"Namespace":"","Name":"cluster-admin"},"RoleName":"cluster-admin","Subject":{"kind":"Group","apiGroup":"rbac.authorization.k8s.io","name":"system:masters"},"Rule":{"verbs":["*"],"nonResourceURLs":["*"]}}
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

Following are excerpts from example tabular output.

```
BINDING                                                         SUBJECT                                                        VERB               APIGROUP                        RESOURCE                                    OBJNAME
/cluster-admin                                                  G:system:masters                                               *                  *                               *                                           *
/ingress-nginx                                                  SA:ingress-nginx/ingress-nginx                                 list                                               configmaps                                  *
/ingress-nginx                                                  SA:ingress-nginx/ingress-nginx                                 list                                               endpoints                                   *
...

BINDING                                       ROLE                                      SUBJECT                                          VERB   NRURL
/cluster-admin                                cluster-admin                             G:system:masters                                 *      *
/kubeadm:cluster-admins                       cluster-admin                             G:kubeadm:cluster-admins                         *      *
...
```

The columns are as follows.

- **BINDING:** The namespace/name for the `ClusterRoleBinding` or `RoleBinding`; namespace is empty fo a `ClusterRoleBinding`.
- **ROLE:** The name of the `ClusterRole` or `Role` object referenced by the BINDING.
- **SUBJECT:** A slightly compacted representation of the `rbac.Subject`.
- **VERB**
- **APIGROUP:** As usual in Kubernetes, the empty string represents the "core" API group. This column is omitted if filtering on API group and exactly one API group is allowed.
- **RESOURCE:** Identifies a collection of objects, similarly to a "kind". This column is omitted if filtering on "resource" and exactly one is allowed.
- **OBJNAME:** Name of an individual object. `*` matches all names.
- **NRURL:** A URL path usable in Non-Resource URLs; `*` matches all paths.
