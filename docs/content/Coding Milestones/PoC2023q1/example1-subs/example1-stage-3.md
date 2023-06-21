<!--example1-stage-3-start-->
## Stage 3

![Placement translation](../Edge-PoC-2023q1-Scenario-1-stage-3.svg "Stage
3 summary")

In Stage 3, in response to the EdgePlacement and SinglePlacementSlice
objects, the placement translator will copy the workload prescriptions
into the mailbox workspaces and create `SyncerConfig` objects there.

Eventually there will be convenient automation running the placement
translator.  In the meantime, you can run it manually: switch to the
ESPW and use the KubeStellar command that runs the placement translator.

```shell
kubectl ws root:espw
```
``` { .bash .no-copy }
Current workspace is "root:espw".
```
```shell
go run ./cmd/placement-translator &
sleep 120
```
``` { .bash .no-copy }
I0423 01:39:56.362722   11644 shared_informer.go:282] Waiting for caches to sync for placement-translator
...
```

After it stops logging stuff, wait another minute and then you can ^C
it or use another shell to continue exploring.

The florin cluster gets only the common workload.  Examine florin's
`SyncerConfig` as follows.  Utilize florin's name (which you stored in Stage 1) here.

```shell
kubectl ws $FLORIN_WS
```

``` { .bash .no-copy }
Current workspace is "root:espw:1t82bk54r6gjnzsp-mb-1a045336-8178-4026-8a56-5cd5609c0ec1" (type root:universal).
```

```shell
kubectl get SyncerConfig the-one -o yaml
```

``` { .bash .no-copy }
apiVersion: edge.kcp.io/v1alpha1
kind: SyncerConfig
metadata:
  annotations:
    kcp.io/cluster: 12299slctppnhjnn
  creationTimestamp: "2023-04-23T05:39:56Z"
  generation: 3
  name: the-one
  resourceVersion: "1323"
  uid: 8840fee6-37dc-407e-ad01-2ad59389d4ff
spec:
  namespaceScope:
    namespaces:
    - commonstuff
    resources:
    - apiVersion: v1
      group: ""
      resource: secrets
    - apiVersion: v1
      group: rbac.authorization.k8s.io
      resource: roles
    - apiVersion: v1
      group: batch
      resource: jobs
    - apiVersion: v1
      group: ""
      resource: serviceaccounts
    - apiVersion: v1
      group: ""
      resource: limitranges
    - apiVersion: v1
      group: apps
      resource: daemonsets
    - apiVersion: v1
      group: storage.k8s.io
      resource: csistoragecapacities
    - apiVersion: v1
      group: rbac.authorization.k8s.io
      resource: rolebindings
    - apiVersion: v1
      group: ""
      resource: pods
    - apiVersion: v1
      group: ""
      resource: resourcequotas
    - apiVersion: v1
      group: discovery.k8s.io
      resource: endpointslices
    - apiVersion: v1
      group: ""
      resource: replicationcontrollers
    - apiVersion: v1
      group: networking.k8s.io
      resource: ingresses
    - apiVersion: v1
      group: ""
      resource: services
    - apiVersion: v1
      group: apps
      resource: deployments
    - apiVersion: v1
      group: ""
      resource: persistentvolumeclaims
    - apiVersion: v1
      group: coordination.k8s.io
      resource: leases
    - apiVersion: v1
      group: policy
      resource: poddisruptionbudgets
    - apiVersion: v1
      group: apps
      resource: statefulsets
    - apiVersion: v1
      group: networking.k8s.io
      resource: networkpolicies
    - apiVersion: v1
      group: batch
      resource: cronjobs
    - apiVersion: v1
      group: ""
      resource: endpoints
    - apiVersion: v2
      group: autoscaling
      resource: horizontalpodautoscalers
    - apiVersion: v1
      group: apps
      resource: replicasets
    - apiVersion: v1
      group: ""
      resource: configmaps
    - apiVersion: v1
      group: ""
      resource: podtemplates
  upsync:
  - apiGroup: group1.test
    names:
    - george
    - cosmo
    namespaces:
    - orbital
    resources:
    - sprockets
    - flanges
  - apiGroup: group2.test
    names:
    - william
    resources:
    - cogs
status: {}
```

You can check that the workload got there too.

```shell
kubectl get ns
```
``` { .bash .no-copy }
NAME          STATUS   AGE
commonstuff   Active   6m34s
default       Active   32m
```

```shell
kubectl get replicasets -A
```
``` { .bash .no-copy }
NAMESPACE     NAME      DESIRED   CURRENT   READY   AGE
commonstuff   commond   0         1         1       10m
```

The guilder cluster gets both the common and special workloads.
Examine guilder's `SyncerConfig` object and workloads as follows, using
the name that you stored in Stage 1.

```shell
kubectl ws root:espw
```
``` { .bash .no-copy }
Current workspace is "root:espw".
```

```shell
kubectl ws $GUILDER_WS
```
``` { .bash .no-copy }
Current workspace is "root:espw:1t82bk54r6gjnzsp-mb-f0a82ab1-63f4-49ea-954d-3a41a35a9f1c" (type root:universal).
```

```shell
kubectl get SyncerConfig the-one -o yaml
```
``` { .bash .no-copy }
apiVersion: edge.kcp.io/v1alpha1
kind: SyncerConfig
metadata:
  annotations:
    kcp.io/cluster: yk9a66vjms1pi8hu
  creationTimestamp: "2023-04-23T05:39:56Z"
  generation: 4
  name: the-one
  resourceVersion: "1325"
  uid: 3da056c7-0d5c-45a3-9d91-d04f04415f30
spec:
  namespaceScope:
    namespaces:
    - commonstuff
    - specialstuff
    resources:
    - apiVersion: v1
      group: apps
      resource: replicasets
    - apiVersion: v1
      group: storage.k8s.io
      resource: csistoragecapacities
    - apiVersion: v1
      group: networking.k8s.io
      resource: networkpolicies
    - apiVersion: v1
      group: apps
      resource: daemonsets
    - apiVersion: v1
      group: ""
      resource: services
    - apiVersion: v1
      group: ""
      resource: replicationcontrollers
    - apiVersion: v1
      group: networking.k8s.io
      resource: ingresses
    - apiVersion: v2
      group: autoscaling
      resource: horizontalpodautoscalers
    - apiVersion: v1
      group: policy
      resource: poddisruptionbudgets
    - apiVersion: v1
      group: rbac.authorization.k8s.io
      resource: roles
    - apiVersion: v1
      group: ""
      resource: serviceaccounts
    - apiVersion: v1
      group: batch
      resource: jobs
    - apiVersion: v1
      group: batch
      resource: cronjobs
    - apiVersion: v1
      group: ""
      resource: persistentvolumeclaims
    - apiVersion: v1
      group: apps
      resource: deployments
    - apiVersion: v1
      group: rbac.authorization.k8s.io
      resource: rolebindings
    - apiVersion: v1
      group: apps
      resource: statefulsets
    - apiVersion: v1
      group: ""
      resource: configmaps
    - apiVersion: v1
      group: ""
      resource: podtemplates
    - apiVersion: v1
      group: ""
      resource: resourcequotas
    - apiVersion: v1
      group: coordination.k8s.io
      resource: leases
    - apiVersion: v1
      group: ""
      resource: endpoints
    - apiVersion: v1
      group: discovery.k8s.io
      resource: endpointslices
    - apiVersion: v1
      group: ""
      resource: secrets
    - apiVersion: v1
      group: ""
      resource: limitranges
    - apiVersion: v1
      group: ""
      resource: pods
  upsync:
  - apiGroup: group3.test
    names:
    - '*'
    resources:
    - widgets
  - apiGroup: group1.test
    names:
    - george
    - cosmo
    namespaces:
    - orbital
    resources:
    - sprockets
    - flanges
  - apiGroup: group2.test
    names:
    - william
    resources:
    - cogs
status: {}
```

```shell
kubectl get deployments,replicasets -A
```
``` { .bash .no-copy }
NAMESPACE      NAME                       READY   UP-TO-DATE   AVAILABLE   AGE
specialstuff   deployment.apps/speciald   0/0     1            0           12m

NAMESPACE     NAME                      DESIRED   CURRENT   READY   AGE
commonstuff   replicaset.apps/commond   0         1         1       7m4s
```
<!--example1-stage-3-stop-->
