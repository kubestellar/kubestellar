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
placement-translator &
# wait until SyncerConfig, ReplicaSets and Deployments are ready
sleep 10
mbxws=($FLORIN_WS $GUILDER_WS)
for ii in "${mbxws[@]}"; do
  kubectl ws root:$ii
  # wait for SyncerConfig resource
  while ! kubectl get SyncerConfig the-one &> /dev/null; do
    sleep 10
  done
  echo "* SyncerConfig resource exists"
  # wait for ReplicaSet resource
  while ! kubectl get rs &> /dev/null; do
    sleep 10
  done
  echo "* ReplicaSet resource exists"
  # wait until ReplicaSet running
  while [ $(kubectl get rs -n commonstuff --field-selector metadata.name=commond | tail -1 | sed -e 's/ \+/ /g' | cut -d " " -f 4) -lt 1 ]; do
    sleep 10
  done
  echo "* commonstuff ReplicaSet running"
done
# check for deployment in guilder
while ! kubectl get deploy -A &> /dev/null; do
  sleep 10
done
echo "* Deployment resource exists"
while [ $(kubectl get deploy -n specialstuff --field-selector metadata.name=speciald | tail -1 | sed -e 's/ \+/ /g' | cut -d " " -f 4) -lt 1 ]; do
  sleep 10
done
echo "* specialstuff Deployment running"
```

After it stops logging stuff, wait another minute and then you can
proceed.

The florin cluster gets only the common workload.  Examine florin's
`SyncerConfig` as follows.  Utilize florin's name (which you stored in Stage 1) here.

```shell
kubectl ws root
```
``` { .bash .no-copy }
Current workspace is "root".
```

```shell
kubectl ws $FLORIN_WS
```

``` { .bash .no-copy }
Current workspace is "root:1t82bk54r6gjnzsp-mb-1a045336-8178-4026-8a56-5cd5609c0ec1" (type root:universal).
```

```shell
kubectl get SyncerConfig the-one -o yaml
```

``` { .bash .no-copy }
apiVersion: edge.kubestellar.io/v2alpha1
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
  namespaceScope: {}
  namespacedObjects:
  - apiVersion: v1
    group: ""
    objectsByNamespace:
    - names:
      - httpd-htdocs
      namespace: commonstuff
    resource: configmaps
  - apiVersion: v1
    group: apps
    objectsByNamespace:
    - names:
      - commond
      namespace: commonstuff
    resource: replicasets
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
kubectl ws root
```
``` { .bash .no-copy }
Current workspace is "root".
```

```shell
kubectl ws $GUILDER_WS
```
``` { .bash .no-copy }
Current workspace is "root:1t82bk54r6gjnzsp-mb-f0a82ab1-63f4-49ea-954d-3a41a35a9f1c" (type root:universal).
```

```shell
kubectl get SyncerConfig the-one -o yaml
```
``` { .bash .no-copy }
apiVersion: edge.kubestellar.io/v2alpha1
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
  clusterScope:
  - apiVersion: v1
    group: apiregistration.k8s.io
    objects:
    - v1090.example.my
    resource: apiservices
  namespaceScope: {}
  namespacedObjects:
  - apiVersion: v1
    group: apps
    objectsByNamespace:
    - names:
      - commond
      namespace: commonstuff
    resource: replicasets
  - apiVersion: v1
    group: apps
    objectsByNamespace:
    - names:
      - speciald
      namespace: specialstuff
    resource: deployments
  - apiVersion: v1
    group: ""
    objectsByNamespace:
    - names:
      - httpd-htdocs
      namespace: commonstuff
    - names:
      - httpd-htdocs
      namespace: specialstuff
    resource: configmaps
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
