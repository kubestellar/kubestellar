# kubestellar-list-syncing-objects

This is a simple utility and demo that exercises the mailboxwatch
package to make and use an informer on a given kind of objects in
the mailbox workspaces.

This utility/demo can either do a one-shot listing or watch indefinitely.
When doing a watch, a field is added to each object to indicate the kind
of notification being presented; the field is named "Action" and its value
will be either "add", "update", or "delete".

This utility/demo normally outputs YAML but can alternatively output
JSON, one line per object.

This utility/demo is given two Kubernetes client configurations.
One, called "all", is for reading the chosen objects from all workspaces.
The other, called "parent", is for reading the mailbox Workspace objects
from their parent Workspace.

Aside from the usual, the command line arguments are as follows.

```console
$ go run ./cmd/kubestellar-list-syncing-objects -h
Usage of kubestellar-list-syncing-objects:
...
      --api-group string                 API group of objects to watch (default is Kubernetes core group)
      --api-kind string                  kind of objects to watch
      --api-resource string              API resource (lowercase plural) of objects to watch (defaults to lowercase(kind)+'s')
      --api-version string               API version (just version, no group) of objects to watch (default "v1")
      --json                             indicates whether to output as lines of JSON rather than YAML
      --watch                            indicates whether to inform rather than just list
...
      --all-cluster string               The name of the kubeconfig cluster to use for access to the chosen objects in all clusters
      --all-context string               The name of the kubeconfig context to use for access to the chosen objects in all clusters (default "system:admin")
      --all-kubeconfig string            Path to the kubeconfig file to use for access to the chosen objects in all clusters
      --all-user string                  The name of the kubeconfig user to use for access to the chosen objects in all clusters
...
      --parent-cluster string              The name of the kubeconfig cluster to use for access to the parent of mailbox workspaces
      --parent-context string            The name of the kubeconfig context to use for access to the parent of mailbox workspaces (default "root")
      --parent-kubeconfig string           Path to the kubeconfig file to use for access to the parent of mailbox workspaces
      --parent-user string                 The name of the kubeconfig user to use for access to the parent of mailbox workspaces
...
pflag: help requested
exit status 2
```

Following is an example of its usage.

```console
$ go run ./cmd/kubestellar-list-syncing-objects --api-group apps --api-kind ReplicaSet
---
apiVersion: apps/v1
kind: ReplicaSet
metadata:
...
status:
  availableReplicas: 1
  fullyLabeledReplicas: 1
  observedGeneration: 1
  readyReplicas: 1
  replicas: 1

---
apiVersion: apps/v1
kind: ReplicaSet
metadata:
...
status:
  availableReplicas: 1
  fullyLabeledReplicas: 1
  observedGeneration: 1
  readyReplicas: 1
  replicas: 1
```
