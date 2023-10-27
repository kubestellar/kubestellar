<!--kubestellar-list-syncing-start-->

Every object subject to downsync or upsync has a full per-WEC copy in
the core. These include reported state from the WECs. If you are using
release 0.10 or later of KubeStellar then you can list these copies of
your httpd `Deployment` objects with the following command.

``` { .bash .no-copy }
kubestellar-list-syncing-objects --api-group apps --api-kind Deployment
```

``` { .bash .no-copy }
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    ... (lots of other details) ...
  name: my-first-kubestellar-deployment
  namespace: my-namespace
spec:
  ... (the spec) ...
status:
  availableReplicas: 1
  conditions:
  - lastTransitionTime: "2023-10-27T07:00:19Z"
    lastUpdateTime: "2023-10-27T07:00:19Z"
    message: Deployment has minimum availability.
    reason: MinimumReplicasAvailable
    status: "True"
    type: Available
  - lastTransitionTime: "2023-10-27T07:00:19Z"
    lastUpdateTime: "2023-10-27T07:00:19Z"
    message: ReplicaSet "my-first-kubestellar-deployment-76f6fc4cfc" has successfully progressed.
    reason: NewReplicaSetAvailable
    status: "True"
    type: Progressing
  observedGeneration: 618
  readyReplicas: 1
  replicas: 1
  updatedReplicas: 1

---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    ... (lots of other details) ...
  name: my-first-kubestellar-deployment
  namespace: my-namespace
spec:
  ... (the spec) ...
status:
  ... (another happy status) ...
```
<!--kubestellar-list-syncing-end-->