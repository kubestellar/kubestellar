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
(
  KUBECONFIG=$SM_CONFIG placement-translator &
  sleep 10
)
```

The following commands wait for the placement translator to get its
job done for this example.

```shell
FLORIN_MB_CONFIG="${PWD}/temp-space-config/${FLORIN_SPACE}"
GUILDER_MB_CONFIG="${PWD}/temp-space-config/${GUILDER_SPACE}"
kubectl-kubestellar-get-config-for-space --space-name $FLORIN_SPACE --provider-name default --sm-core-config $SM_CONFIG --sm-context $SM_CONTEXT --output $FLORIN_MB_CONFIG
kubectl-kubestellar-get-config-for-space --space-name $GUILDER_SPACE --provider-name default --sm-core-config $SM_CONFIG --sm-context $SM_CONTEXT --output $GUILDER_MB_CONFIG

# wait until SyncerConfig, ReplicaSets and Deployments are ready
mbxws=($FLORIN_SPACE $GUILDER_SPACE)
for ii in "${mbxws[@]}"; do
  (
    export KUBECONFIG="${PWD}/temp-space-config/$ii"
    # wait for SyncerConfig resource
    while ! kubectl get SyncerConfig the-one &> /dev/null; do
      sleep 10
    done
    echo "* SyncerConfig resource exists in mailbox $ii"
    # wait for ReplicaSet resource
    while ! kubectl get rs &> /dev/null; do
      sleep 10
    done
    echo "* ReplicaSet resource exists in mailbox $ii"
    # wait until ReplicaSet in mailbox
    while ! kubectl get rs -n commonstuff commond; do
      sleep 10
    done
    echo "* commonstuff ReplicaSet in mailbox $ii"
  )
done
(
  export KUBECONFIG=$GUILDER_MB_CONFIG
  # check for deployment in guilder
  while ! kubectl get deploy -A &> /dev/null; do
    sleep 10
  done
  echo "* Deployment resource exists"
  while ! kubectl get deploy -n specialstuff speciald; do
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
)
```

You can check that the common workload's ReplicaSet objects got to
their mailbox workspaces with the following command. It will list the
two copies of that object, each with an annotation whose key is
`kcp.io/cluster` and whose value is the kcp `logicalcluster.Name` of
the mailbox workspace; those names appear in the "CLUSTER" column of
the custom-columns listing near the end of [the section above about
the mailbox controller](../#the-mailbox-controller).

```shell
# TODO: kubestellar-list-syncing-objects has kcp dependencies. Will remove when controllers support spaces.
kubestellar-list-syncing-objects --api-group apps --api-kind ReplicaSet
```

``` { .bash .no-copy }
---
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  annotations:
    edge.kubestellar.io/customizer: example-customizer
    kcp.io/cluster: 1y7wll1dz806h3sb
    ... (lots of other details) ...
  name: commond
  namespace: commonstuff
spec:
  ... (the customized spec) ...
status:
  ... (may be filled in by the time you look) ...

---
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  annotations:
    edge.kubestellar.io/customizer: example-customizer
    kcp.io/cluster: 1najcltzt2nqax47
    ... (lots of other details) ...
  name: commond
  namespace: commonstuff
spec:
  ... (the customized spec) ...
status:
  ... (may be filled in by the time you look) ...
```

That display should show objects in two different mailbox workspaces;
the following command checks that.

```shell
# TODO: kubestellar-list-syncing-objects has kcp dependencies. Will remove when controllers support spaces.
test $(kubestellar-list-syncing-objects --api-group apps --api-kind ReplicaSet | grep "^ *kcp.io/cluster: [0-9a-z]*$" | sort | uniq | wc -l) -ge 2
```

The various APIBinding and CustomResourceDefinition objects involved
should also appear in the mailbox workspaces.

```shell
# TODO: kubestellar-list-syncing-objects has kcp dependencies. Will remove when controllers support spaces.
test $(kubestellar-list-syncing-objects --api-group apis.kcp.io --api-version v1alpha1 --api-kind APIBinding | grep -cw "name: bind-apps") -ge 2
kubestellar-list-syncing-objects --api-group apis.kcp.io --api-version v1alpha1 --api-kind APIBinding | grep -w "name: bind-kubernetes"
kubestellar-list-syncing-objects --api-group apiextensions.k8s.io --api-kind CustomResourceDefinition | fgrep -w "name: crontabs.stable.example.com"
```

The `APIService` of the special workload should also appear, along
with some error messages about `APIService` not being known in the
other mailbox workspaces.

```shell
# TODO: kubestellar-list-syncing-objects has kcp dependencies. Will remove when controllers support spaces.
kubestellar-list-syncing-objects --api-group apiregistration.k8s.io --api-kind APIService 2>&1 | grep -v "APIService.*the server could not find the requested resource" | fgrep -w "name: v1090.example.my"
```

The florin cluster gets only the common workload.  Examine florin's
`SyncerConfig` as follows.  Utilize the name of the mailbox workspace
for florin (which you stored in Stage 1) here.

```shell
KUBECONFIG=$FLORIN_MB_CONFIG kubectl get SyncerConfig the-one -o yaml
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

The guilder cluster gets both the common and special workloads.
Examine guilder's `SyncerConfig` object and workloads as follows,
using the mailbox workspace name that you stored in Stage 1.

```shell
KUBECONFIG=$GUILDER_MB_CONFIG kubectl get SyncerConfig the-one -o yaml
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
    group: ""
    objects:
    - specialstuff
    resource: namespaces
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

You can check for specific workload objects here with the following
command.

```shell
KUBECONFIG=$GUILDER_MB_CONFIG kubectl get deployments,replicasets -A
```
``` { .bash .no-copy }
NAMESPACE      NAME                       READY   UP-TO-DATE   AVAILABLE   AGE
specialstuff   deployment.apps/speciald   0/0     1            0           12m

NAMESPACE     NAME                      DESIRED   CURRENT   READY   AGE
commonstuff   replicaset.apps/commond   0         1         1       7m4s
```
<!--example1-stage-3-stop-->
