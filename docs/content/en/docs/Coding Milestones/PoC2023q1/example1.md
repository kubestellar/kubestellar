---
title: "2023q1 PoC Example Scenario"
linkTitle: "2023q1 PoC Example Scenario"
weight: 100
description: >-
  
---

This doc attempts to show a simple example usage of the 2023q1 PoC.
This doc is a work in progress.

This example involves two edge clusters and two workloads.  One
workload goes on both edge clusters and one workload goes on only one
edge cluster.  Nothing changes after the initial activity.

This example is presented in stages.  The controllers involved are
always maintaining relationships.  This document focuses on changes as
they appear in this example.

## Stage 1

![Boxes and arrows. Two kind clusters exist, named kind1 and kind3. The Inventory Management workspace contains two pairs of SyncTarget and Location objects. The Edge Service Provider workspace contains the PoD controllers; the mailbox controller reads the SyncTarget objects and creates two mailbox workspaces.](/docs/coding-milestones/poc2023q1/Edge-PoC-2023q1-Scenario-1-stage-1.svg "Stage 1 Summary")

Stage 1 creates the infrastructure and the edge service provider
workspace and lets that react to the inventory.  It has the following
steps.

### Create two kind clusters.

This example uses two [kind](https://kind.sigs.k8s.io/) clusters as
edge clusters.  We will call them "florin" and "guilder".

This example uses extremely simple workloads, which
use `hostPort` networking in Kubernetes.  To make those ports easily
reachable from your host, this example uses an explicit `kind`
configuration for each edge cluster.

For the florin cluster, which will get only one workload, create a
file named `florin-config.yaml` with the following contents.  In a
`kind` config file, `containerPort` is about the container that is
also a host (a Kubernetes node), while the `hostPort` is about the
host that hosts that container.

```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 8081
    hostPort: 8081
```

For the guilder cluster, which will get two workloads, create a file
named `guilder-config.yaml` with the following contents.  The workload
that uses hostPort 8081 goes in both clusters, while the workload that
uses hostPort 8082 goes only in the guilder cluster.

```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 8081
    hostPort: 8083
  - containerPort: 8082
    hostPort: 8082
```

Finally, create the two clusters with the following two commands,
paying attention to `$KUBECONFIG` and, if that's empty,
`~/.kube/config`: `kind create` will inject/replace the relevant
"context" in your active kubeconfig.

```shell
kind create cluster --name florin --config florin-config.yaml
kind create cluster --name guilder --config guilder-config.yaml
```

### Start kcp

In some shell that will be used only for this purpose, issue the `kcp
start` command.  If you have junk from previous runs laying around,
you should probably `rm -rf .kcp` first.

In the shell commands in all the following steps it is assumed that
`kcp` is running and `$KUBECONFIG` is set to the
`.kcp/admin.kubeconfig` that `kcp` produces, except where explicitly
noted that the florin or guilder cluster is being accessed.

It is also assumed that you have the usual kcp kubectl plugins on your
`$PATH`.

### Create an inventory management workspace.

Use the following commands.

```shell
kubectl ws root
kubectl ws create imw-1 --enter
```

### Create a SyncTarget object to represent the florin cluster

Use `kubectl` to create the SyncTarget object, as in the following
command.

```shell
kubectl apply -f - <<EOF
apiVersion: workload.kcp.io/v1alpha1
kind: SyncTarget
metadata:
  name: sync-target-f
  labels:
    example: si
    extended: non
spec:
  cells:
    foo: bar
EOF
```

### Create a Location object describing the florin cluster

Use `kubectl` to create the Location object, as in the following
command.

```shell
kubectl apply -f - <<EOF
apiVersion: scheduling.kcp.io/v1alpha1
kind: Location
metadata:
  name: location-f
  labels:
    env: prod
spec:
  resource: {group: workload.kcp.io, version: v1alpha1, resource: synctargets}
  instanceSelector:
    matchLabels: {"example":"si", "extended":"non"}
EOF
```

### Delete default Location

You will probably find that something automatically created a
`Location` named `default` for your convenience.  It is actually a
nuisance in this scenario.  Delete that Location, such as with the
following command.

```shell
kubectl delete locations default
```

### Create a SyncTarget object describing the guilder cluster

Use `kubectl` to create the SyncTarget object, like in the following
command.

```shell
kubectl apply -f - <<EOF
apiVersion: workload.kcp.io/v1alpha1
kind: SyncTarget
metadata:
  name: sync-target-g
  labels:
    example: si
    extended: si
spec:
  cells:
    bar: baz
EOF
```

### Create a Location object describing the guilder cluster

Use `kubectl` to create the Location object, such as with the
following command.

```shell
kubectl apply -f - <<EOF
apiVersion: scheduling.kcp.io/v1alpha1
kind: Location
metadata:
  name: location-g
  labels:
    env: prod
    extended: si
spec:
  resource: {group: workload.kcp.io, version: v1alpha1, resource: synctargets}
  instanceSelector:
    matchLabels: {"example":"si", "extended":"si"}
EOF
```

### Create the edge service provider workspace

Use the following commands.

```shell
kubectl ws root
kubectl ws create espw --enter
```

### Populate the edge service provider workspace

This puts the definition and export of the edge-mc API in the edge
service provider workspace.

Use the following commands.

```shell
kubectl create -f config/crds
kubectl create -f config/exports
```

### The mailbox controller

Running the mailbox controller will be conveniently automated.
Eventually.  In the meantime, you can run it by hand as follows.

```console
$ go run ./cmd/mailbox-controller -v=2
...
I0418 00:06:33.600117    6576 main.go:196] "Found APIExport view" exportName="workload.kcp.io" serverURL="https://192.168.58.123:6443/services/apiexport/root/workload.kcp.io"
...
I0418 00:06:34.361128    6576 controller.go:299] "Created APIBinding" worker=1 mbwsName="2rp1gztc6m5b8b7r-mb-31e5fa4d-a84e-4397-a523-63fa62d16dad" mbwsCluster="1b0eso4ld8np1d3z" bindingName="bind-edge" resourceVersion="1303"
I0418 00:06:34.361216    6576 controller.go:299] "Created APIBinding" worker=2 mbwsName="2rp1gztc6m5b8b7r-mb-58f7e799-4653-422b-adba-b1e5e85a7fac" mbwsCluster="2gqno7cdbsthqsmz" bindingName="bind-edge" resourceVersion="1305"
^C
```

You need a `-v` setting of 2 or numerically higher to get log messages
about individual mailbox workspaces.

This controller creates a mailbox workspace for each SyncTarget and
puts an APIBinding to the edge API in each of those mailbox
workspaces.  For this simple scenario, you do not need to keep this
controller running after it does those things (hence the `^C` above);
normally it would run continuously.

You can get a listing of those mailbox workspaces as follows.

```console
$ kubectl get Workspaces
NAME                                                       TYPE        REGION   PHASE   URL                                                     AGE
2rp1gztc6m5b8b7r-mb-31e5fa4d-a84e-4397-a523-63fa62d16dad   universal            Ready   https://192.168.58.123:6443/clusters/1b0eso4ld8np1d3z   4m56s
2rp1gztc6m5b8b7r-mb-58f7e799-4653-422b-adba-b1e5e85a7fac   universal            Ready   https://192.168.58.123:6443/clusters/2gqno7cdbsthqsmz   4m56s
```

## Stage 2

![Placement and scheduling](/docs/coding-milestones/poc2023q1/Edge-PoC-2023q1-Scenario-1-stage-2.svg
"Stage 2 summary")

Stage 2 creates two workloads, called "common" and "special", and lets
the edge scheduler react.  It has the following steps.

### Create and populate the workload management workspace for the common workload

In this example, each workload description goes in its own workload
management workspace.

One of the workloads is called "common", because it will go to both
edge clusters.

Create the "common" workload management workspace with the following
commands.  For the sake of orderliness we choose to keep the two
workload management workspaces under a common parent.

```shell
kubectl ws root
kubectl ws create my-org --enter
kubectl ws create wmw-c --enter
```

Next, make sure that the Kubernetes workload APIs are available in
this workspace.  In a freshly created workspace of the expected type
(`root:universal` in this case), the Kube workload APIs would not yet
be available.  Use `kubectl` to create the following APIBinding object
--- which enables use of those Kubernetes APIs.

```shell
kubectl apply -f - <<EOF
apiVersion: apis.kcp.io/v1alpha1
kind: APIBinding
metadata:
  name: bind-kube
spec:
  reference:
    export:
      path: "root:compute"
      name: kubernetes
EOF
```

Next, use `kubectl` to create the following workload objects in that
workspace.  The workload in this example in an Apache httpd server
that serves up a very simple web page, conveyed via a Kubernetes
ConfigMap that is mounted as a volume for the httpd pod.

```shell
kubectl apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: commonstuff
  labels: {common: "si"}
---
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: commonstuff
  name: httpd-htdocs
data:
  index.html: |
    <!DOCTYPE html>
    <html>
      <body>
        This is a common web site.
      </body>
    </html>
---
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: commonstuff
  name: commond
spec:
  selector: {matchLabels: {app: common} }
  template:
    metadata:
      labels: {app: common}
    spec:
      containers:
      - name: httpd
        image: library/httpd:2.4
        ports:
        - name: http
          containerPort: 80
          hostPort: 8081
          protocol: TCP
        volumeMounts:
        - name: htdocs
          readOnly: true
          mountPath: /usr/local/apache2/htdocs
      volumes:
      - name: htdocs
        configMap:
          name: httpd-htdocs
          optional: false
EOF
```

Before or after the previous step, use `kubectl` to create the
following APIBinding object --- which enables use of the edge API.

```shell
kubectl apply -f - <<EOF
apiVersion: apis.kcp.io/v1alpha1
kind: APIBinding
metadata:
  name: bind-espw
spec:
  reference:
    export:
      path: root:espw
      name: edge.kcp.io
EOF
```

Finally, use `kubectl` to create the following EdgePlacement object.
Its "where predicate" (the `locationSelectors` array) has one label
selector that matches both Location objects created earlier, thus
directing the common workload to both edge clusters.
   
```shell
kubectl apply -f - <<EOF
apiVersion: edge.kcp.io/v1alpha1
kind: EdgePlacement
metadata:
  name: edge-placement-c
spec:
  locationSelectors:
  - matchLabels: {"env":"prod"}
  namespaceSelector:
    matchLabels: {"common":"si"}
  nonNamespacedObjects:
  - apiGroup: apis.kcp.io
    resources: [ "apibindings" ]
    resourceNames: [ "bind-kube" ]
  upsync:
  - apiGroup: "group1.test"
    resources: ["sprockets", "flanges"]
    namespaces: ["orbital"]
    names: ["george", "cosmo"]
  - apiGroup: "group2.test"
    resources: ["cogs"]
    names: ["William"]
EOF
```

### Create and populate the workload management workspace for the special workload

Use `kubectl` to create the workload management workspace for the
special workload, using the following commands.

```shell
kubectl ws root:my-org
kubectl ws create wmw-s --enter
```

Next, make sure that the Kubernetes workload APIs are available in
this workspace.  Use `kubectl` to create the following APIBinding
object --- which enables use of those Kubernetes APIs.

```shell
kubectl apply -f - <<EOF
apiVersion: apis.kcp.io/v1alpha1
kind: APIBinding
metadata:
  name: bind-kube
spec:
  reference:
    export:
      path: "root:compute"
      name: kubernetes
EOF
```

Next, use `kubectl` to create the following workload objects in that workspace.

```shell
kubectl apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: specialstuff
  labels: {special: "si"}
---
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: specialstuff
  name: httpd-htdocs
data:
  index.html: |
    <!DOCTYPE html>
    <html>
      <body>
        This is a special web site.
      </body>
    </html>
---
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: specialstuff
  name: speciald
spec:
  selector: {matchLabels: {app: special} }
  template:
    metadata:
      labels: {app: special}
    spec:
      containers:
      - name: httpd
        image: library/httpd:2.4
        ports:
        - name: http
          containerPort: 80
          hostPort: 8082
          protocol: TCP
        volumeMounts:
        - name: htdocs
          readOnly: true
          mountPath: /usr/local/apache2/htdocs
      volumes:
      - name: htdocs
        configMap:
          name: httpd-htdocs
          optional: false
EOF
```

Before or after the previous step, use `kubectl` to create the
following APIBinding object --- which enables use of the edge API.

```shell
kubectl apply -f - <<EOF
apiVersion: apis.kcp.io/v1alpha1
kind: APIBinding
metadata:
  name: bind-espw
spec:
  reference:
    export:
      path: root:espw
      name: edge.kcp.io
EOF
```

Finally, use `kubectl` to create the following EdgePlacement object.
Its "where predicate" (the `locationSelectors` array) has one label
selector that matches only one of the Location objects created
earlier, thus directing the special workload to just one edge cluster.
   
```shell
kubectl apply -f - <<EOF
apiVersion: edge.kcp.io/v1alpha1
kind: EdgePlacement
metadata:
  name: edge-placement-s
spec:
  locationSelectors:
  - matchLabels: {"env":"prod","extended":"si"}
  namespaceSelector: 
    matchLabels: {"special":"si"}
  nonNamespacedObjects:
  - apiGroup: apis.kcp.io
    resources: [ "apibindings" ]
    resourceNames: [ "bind-kube" ]
  upsync:
  - apiGroup: "group1.test"
    resources: ["sprockets", "flanges"]
    namespaces: ["orbital"]
    names: ["george", "cosmo"]
  - apiGroup: "group3.test"
    resources: ["widgets"]
    names: ["*"]
EOF
```

### Edge scheduling

In response to each EdgePlacement, the edge scheduler will create a
corresponding SinglePlacementSlice object.  These will indicate the
following resolutions of the "where" predicates.

| EdgePlacement | Resolved Where |
| ------------- | -------------: |
| edge-placement-c | florin, guilder |
| edge-placement-s | guilder |

Eventually there will be automation that conveniently runs the edge
scheduler.  In the meantime, you can run it by hand wit a command like
the following.

```console
$ kubectl ws root:espw
Current workspace is "root:espw".
$ go run ./cmd/scheduler
I0418 00:23:58.516391    6754 scheduler.go:212] "Found APIExport view" exportName="edge.kcp.io" serverURL="https://192.168.58.123:6443/services/apiexport/2mk49qlwexsf0dbl/edge.kcp.io"
...
I0418 00:23:58.787650    6754 reconcile_on_location.go:192] "updated SinglePlacementSlice" controller="edge-scheduler" triggeringKind=Location key="2rp1gztc6m5b8b7r|location-g" locationWorkspace="2rp1gztc6m5b8b7r" location="location-g" workloadWorkspace="29qof7rv6tzefi7b" singlePlacementSlice="edge-placement-c"
...
I0418 00:23:58.790090    6754 reconcile_on_location.go:192] "updated SinglePlacementSlice" controller="edge-scheduler" triggeringKind=Location key="2rp1gztc6m5b8b7r|location-f" locationWorkspace="2rp1gztc6m5b8b7r" location="location-f" workloadWorkspace="29qof7rv6tzefi7b" singlePlacementSlice="edge-placement-c"
^C
```

In this simple scenario you do not need to keep the scheduler running
after it gets its initial work done; normally it would run
continually.

Check out the SinglePlacementSlice objects as follows.

```console
$ kubectl ws root:my-org:wmw-c
Current workspace is "root:my-org:wmw-c".
$ kubectl get SinglePlacementSlice -o yaml
apiVersion: v1
items:
- apiVersion: edge.kcp.io/v1alpha1
  destinations:
  - cluster: y2vvakjnwxns951p
    locationName: location-g
    syncTargetName: sync-target-g
    syncTargetUID: 6be95d47-e85e-474d-9367-bb00e1074718
  - cluster: y2vvakjnwxns951p
    locationName: location-f
    syncTargetName: sync-target-f
    syncTargetUID: f42b9446-97c5-42a7-a514-0af32350289d
  kind: SinglePlacementSlice
  metadata:
    annotations:
      kcp.io/cluster: 1em0rxri71aqmd41
    creationTimestamp: "2023-04-07T02:52:09Z"
    generation: 11
    name: edge-placement-c
    ownerReferences:
    - apiVersion: edge.kcp.io/v1alpha1
      kind: EdgePlacement
      name: edge-placement-c
      uid: 598bbf90-2487-42f6-bf02-0c9f8511acce
    resourceVersion: "1154"
    uid: 4e69bd90-ce13-4d62-8deb-f500216757eb
kind: List
metadata:
  resourceVersion: ""
```

Also check out the SinglePlacementSlice objects in
`root:my-org:wmw-s`.  It should go similarly, but the `destinations`
should include only the entry for florin.

## Stage 3

![Placement translation](/docs/coding-milestones/poc2023q1/Edge-PoC-2023q1-Scenario-1-stage-3.svg "Stage
3 summary")

In Stage 3, in response to the EdgePlacement and SinglePlacementSlice
objects, the placement translator will copy the workload prescriptions
into the mailbox workspaces and create `SyncerConfig` objects there.

TODO later: add customization

Eventually there will be convenient automation running the placement
translator.  In the meantime, you can run it manually as follows.

```console
$ kubectl ws root:espw
Current workspace is "root:espw".
$ go run ./cmd/placement-translator
I0418 00:32:49.789575    6849 shared_informer.go:282] Waiting for caches to sync for placement-translator
...
```

After it stops logging stuff, wait another minute and then you can ^C
it or use another shell to continue exploring.

The florin cluster gets only the common workload.  Examine florin's
`SyncerConfig` as follows.

```shell
$ kubectl ws 2rp1gztc6m5b8b7r-mb-58f7e799-4653-422b-adba-b1e5e85a7fac
Current workspace is "root:espw:2rp1gztc6m5b8b7r-mb-58f7e799-4653-422b-adba-b1e5e85a7fac" (type root:universal).

$ kubectl get SyncerConfig the-one -o yaml
apiVersion: edge.kcp.io/v1alpha1
kind: SyncerConfig
metadata:
  annotations:
    kcp.io/cluster: 2gqno7cdbsthqsmz
  creationTimestamp: "2023-04-18T04:32:50Z"
  generation: 4
  name: the-one
  resourceVersion: "1653"
  uid: 47843727-bbe8-4660-8c25-4566fc0b6d70
spec:
  namespaceScope:
    namespaces:
    - commonstuff
    resources:
    - apiVersion: v1
      group: rbac.authorization.k8s.io
      resource: roles
    - apiVersion: v1
      group: apps
      resource: deployments
    - apiVersion: v1
      group: ""
      resource: configmaps
    - apiVersion: v1
      group: ""
      resource: secrets
    - apiVersion: v1
      group: ""
      resource: serviceaccounts
    - apiVersion: v1
      group: networking.k8s.io
      resource: ingresses
    - apiVersion: v1
      group: ""
      resource: resourcequotas
    - apiVersion: v1
      group: ""
      resource: services
    - apiVersion: v1
      group: ""
      resource: limitranges
    - apiVersion: v1
      group: ""
      resource: pods
    - apiVersion: v1
      group: rbac.authorization.k8s.io
      resource: rolebindings
    - apiVersion: v1
      group: coordination.k8s.io
      resource: leases
  upsync:
  - apiGroup: group2.test
    names:
    - William
    resources:
    - cogs
  - apiGroup: group1.test
    names:
    - george
    - cosmo
    namespaces:
    - orbital
    resources:
    - sprockets
    - flanges
status: {}
```

You can check that the workload got there too.

```console
$ kubectl get ns
NAME          STATUS   AGE
commonstuff   Active   6m34s
default       Active   32m

$ kubectl get deployments -A
NAMESPACE     NAME      READY   UP-TO-DATE   AVAILABLE   AGE
commonstuff   commond   0/0     0            0           6m44s
```

The guilder cluster gets both the common and special workloads.
Examine guilder's TMC Placement object and workloads as follows.

```console
$ kubectl ws root:espw
Current workspace is "root:espw".

$ kubectl ws 2rp1gztc6m5b8b7r-mb-31e5fa4d-a84e-4397-a523-63fa62d16dad
Current workspace is "root:espw:2rp1gztc6m5b8b7r-mb-31e5fa4d-a84e-4397-a523-63fa62d16dad" (type root:universal).

$ kubectl get deployments -A
NAMESPACE      NAME       READY   UP-TO-DATE   AVAILABLE   AGE
commonstuff    commond    0/0     0            0           8m26s
specialstuff   speciald   0/0     0            0           8m26s
```

## Stage 4

![TMC kicks in](/docs/coding-milestones/poc2023q1/Edge-PoC-2023q1-Scenario-1-stage-4.svg "Stage 4 summary")

In Stage 4, TMC does its thing.  You can check that the workloads are
running in the edge clusters as they should be.

TMC does its thing between the florin cluster and its mailbox
workspace.  This is driven by the one TMC Placement object in that
mailbox workspace.

TMC does its thing between the guilder cluster and its mailbox
workspace.  This is driven by the two TMC Placement objects in that
mailbox workspace, one for the common workload and one for the special
workload.

Using the kubeconfig that `kind` modified, examine the florin cluster.
Find just the `commonstuff` namespace and the `commond` Deployment.

```console
$ kubectl get --context kind-florin ns | grep stuff         
commonstuff          Active   8h

$ kubectl get --context kind-florin Deployment -n commonstuff
NAME      READY   UP-TO-DATE   AVAILABLE   AGE
commond   1/1     1            1           8h
```

Examine the guilder cluster.  Find both workload namespaces and both
Deployments.

```console
$ kubectl get --context kind-guilder ns | grep stuff
commonstuff          Active   8h
specialstuff         Active   8h

$ kubectl get --context kind-guilder Deployment -A | grep stuff
commonstuff          commond                  1/1     1            1           8h
specialstuff         speciald                 1/1     1            1           8h
```

Check that the common workload on the florin cluster is working.

```console
$ curl http://localhost:8081
<!DOCTYPE html>
<html>
  <body>
    This is a common web site.
  </body>
</html>
```

Check that the special workload on the guilder cluster is working.

```console
$ curl http://localhost:8082
<!DOCTYPE html>
<html>
  <body>
    This is a special web site.
  </body>
</html>
```

Check that the common workload on the guilder cluster is working.

```console
$ curl http://localhost:8083
<!DOCTYPE html>
<html>
  <body>
    This is a common web site.
  </body>
</html>
```

## Stage 5

![Summarization for special](/docs/coding-milestones/poc2023q1/Edge-PoC-2023q1-Scenario-1-stage-5s.svg "Status summarization for special")

The status summarizer, driven by the EdgePlacement and
SinglePlacementSlice for the special workload, creates a status
summary object in the specialstuff namespace in the special workload
workspace holding a summary of the corresponding Deployment objects.
In this case there is just one such object, in the mailbox workspace
for the guilder cluster.

![Summarization for common](/docs/coding-milestones/poc2023q1/Edge-PoC-2023q1-Scenario-1-stage-5c.svg "Status summarization for common")

The status summarizer, driven by the EdgePlacement and
SinglePlacementSlice for the common workload, creates a status summary
object in the commonstuff namespace in the common workload workspace
holding a summary of the corresponding Deployment objects.  Those are
the `commond` Deployment objects in the two mailbox workspaces.
