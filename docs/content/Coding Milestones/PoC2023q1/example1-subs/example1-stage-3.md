<!--example1-stage-3-start-->
## Stage 3

![Placement translation](../Edge-PoC-2023q1-Scenario-1-stage-3.svg "Stage
3 summary")

In Stage 3, in response to the EdgePlacement and SinglePlacementSlice
objects, the placement translator will copy the workload prescriptions
into the mailbox workspaces and create `SyncerConfig` objects there.

If you have deployed the KubeStellar core as workload in a Kubernetes
cluster then the placement translator is running in a Pod there. If
instead you are running the core controllers as bare processes then
use the following commands to launch the placement translator; it
requires the ESPW to be current at start time.

```shell
kubectl ws root:espw
placement-translator &
sleep 10
```

The following commands wait for the placement translator to get its
job done for this example.

```shell
# wait until SyncerConfig, ReplicaSets and Deployments are ready
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
  # wait until ReplicaSet in mailbox
  while [ "$(kubectl get rs -n commonstuff commond -o 'jsonpath={.status.readyReplicas}')" != 1 ]; do
    sleep 10
  done
  echo "* commonstuff ReplicaSet in mailbox $ii"
done
# check for deployment in guilder
while ! kubectl get deploy -A &> /dev/null; do
  sleep 10
done
echo "* Deployment resource exists"
while [ "$(kubectl get deploy -n specialstuff speciald -o 'jsonpath={.status.availableReplicas}')" != 1 ]; do
  sleep 10
done
echo "* specialstuff Deployment in its mailbox"
# wait for crontab CRD to be established
while ! kubectl get crd crontabs.stable.example.com; do sleep 10; done
kubectl wait --for condition=Established crd crontabs.stable.example.com
echo "* CronTab CRD is established in its mailbox"
# wait for my-new-cron-object to be in its mailbox
while ! kubectl get ct -n specialstuff my-new-cron-object; do sleep 10; done
echo "* CronTab my-new-cron-object is in its mailbox"
```

The florin cluster gets only the common workload.  Examine florin's
`SyncerConfig` as follows.  Utilize the name of the mailbox workspace
for florin (which you stored in Stage 1) here.

```shell
kubectl ws root:$FLORIN_WS
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
Examine guilder's `SyncerConfig` object and workloads as follows,
using the mailbox workspace name that you stored in Stage 1.

```shell
kubectl ws root:$GUILDER_WS
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
    group: apiextensions.k8s.io
    objects:
    - crontabs.stable.example.com
    resource: customresourcedefinitions
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
    group: stable.example.com
    objectsByNamespace:
    - names:
      - my-new-cron-object
      namespace: specialstuff
    resource: crontabs
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
