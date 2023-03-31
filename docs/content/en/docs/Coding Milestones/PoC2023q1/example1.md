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
$ kind create cluster --name florin --config florin-config.yaml
$ kind create cluster --name guilder --config guilder-config.yaml
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

```shell
$ kubectl ws root
$ kubectl ws create inv1 --enter
```

### Create a SyncTarget object to represent the florin cluster

Use `kubectl` to create the following SyncTarget object:

```shell
cat <<EOF | kubectl apply -f -
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

Use `kubectl` to create the following Location object.

```shell
cat <<EOF | kubectl apply -f -
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

### Create a SyncTarget object describing the guilder cluster

Use `kubectl` to create the following SyncTarget object.

```shell
cat <<EOF | kubectl apply -f -
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

Use `kubectl` to create the following Location object.

```shell
cat <<EOF | kubectl apply -f -
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

```shell
$ kubectl ws root
$ kubectl ws create espw --enter
```

### Populate the edge service provider workspace

This creates the controllers and APIExports from the edge service
provider workspace.

This will be wrapped up into a single command, still to be designed.
It will include creating the objects in the `config/crds` and
`config/exports` directories of the edge-mc repo.

### The mailbox controller will create two mailbox workspaces

That is, one for each SyncTarget.  After that is done (TODO: show how),
check it out as follows.

```shell
$ kubectl get workspaces
NAME                                                       TYPE        REGION   PHASE   URL                                                     AGE
niqdko2g2pwoadfb-mb-f99e773f-3db2-439e-8054-827c4ac55368   universal            Ready   https://192.168.58.123:6443/clusters/0ay27fcwuo2sv6ht   22s
niqdko2g2pwoadfb-mb-c5820696-016b-41f6-b676-d7c0ef02fc5a   universal            Ready   https://192.168.58.123:6443/clusters/dead3333beef3333   22s
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
commands.

```shell
$ kubectl ws root
$ kubectl ws create work-c --enter
```

Next, make sure that the Kubernetes workload APIs are available in
this workspace.  TODO: show how.  If they are not then use `kubectl`
to create the following APIBinding object --- which enables use of
those Kubernetes APIs.

```shell
cat <<EOF | kubectl apply -f -
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
cat <<EOF | kubectl apply -f -
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
cat <<EOF | kubectl apply -f -
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
cat <<EOF | kubectl apply -f -
apiVersion: edge.kcp.io/v1alpha1
kind: EdgePlacement
metadata:
  name: edge-placement-c
spec:
  locationSelectors:
  - matchLabels: {"env":"prod"}
  namespaceSelector:
    matchLabels: {"common":"si"}
EOF
```

### Create and populate the workload management workspace for the special workload

Use `kubectl` to create the workload management workspace for the
special workload, using the following commands.

```shell
kubectl ws root
kubectl ws create work-s --enter
```

Next, make sure that the Kubernetes workload APIs are available in
this workspace.  TODO: show how.  If they are not then use `kubectl`
to create the following APIBinding object --- which enables use of
those Kubernetes APIs.

```shell
cat <<EOF | kubectl apply -f -
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
cat <<EOF | kubectl apply -f -
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
cat <<EOF | kubectl apply -f -
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
cat <<EOF | kubectl apply -f -
apiVersion: edge.kcp.io/v1alpha1
kind: EdgePlacement
metadata:
  name: edge-placement-s
spec:
  locationSelectors:
  - matchLabels: {"env":"prod","extended":"si"}
  namespaceSelector: 
    matchLabels: {"special":"si"}
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

Check out the SinglePlacementSlice objects as follows.

```shell
$ kubectl ws root:work-c
$ kubectl get SinglePlacementSlice
(TODO: show what it looks like)
$ kubectl ws root:work-s
$ kubectl get SinglePlacementSlice
(TODO: show what it looks like)
```

## Stage 3

![Placement translation](/docs/coding-milestones/poc2023q1/Edge-PoC-2023q1-Scenario-1-stage-3.svg "Stage
3 summary")

In Stage 3, in response to the EdgePlacement and SinglePlacementSlice
objects, the placement translator will copy the workload prescriptions
into the mailbox workspaces and create TMC placement objects there.

TODO later: add customization

The florin cluster gets only the common workload.  Examine florin's TMC
Placement object and common workload as follows.

```shell
$ kubectl ws root:espw:niqdko2g2pwoadfb-mb-f99e773f-3db2-439e-8054-827c4ac55368
$ kubectl get Placement -o yaml
TODO: show what it looks like
$ kubectl get ns
(will list the syncer workspace and specialstuff)
$ kubectl get Deployment -A
(will list the syncer and speciald Deployments)
```

The guilder cluster gets both the common and special workloads.
Examine guilder's TMC Placement object and workloads as follows.

```shell
$ kubectl ws root:espw:niqdko2g2pwoadfb-mb-c5820696-016b-41f6-b676-d7c0ef02fc5a
$ kubectl get Placement -o yaml
TODO: show what it looks like
$ kubectl get ns
(will list the syncer workspace, commonstuff, and specialstuff)
$ kubectl get Deployment -A
(will list the syncer, commond, and speciald Deployments)
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

```shell
$ kubectl get --context kind-florin ns | grep stuff         
commonstuff          Active   8h

$ kubectl get --context kind-florin Deployment -n commonstuff
NAME      READY   UP-TO-DATE   AVAILABLE   AGE
commond   1/1     1            1           8h
```

Examine the guilder cluster.  Find both workload namespaces and both
Deployments.

```shell
$ kubectl get --context kind-guilder ns | grep stuff
commonstuff          Active   8h
specialstuff         Active   8h

$ kubectl get --context kind-guilder Deployment -A | grep stuff
commonstuff          commond                  1/1     1            1           8h
specialstuff         speciald                 1/1     1            1           8h
```

Check that the common workload on the florin cluster is working.

```shell
$ curl http://localhost:8081
<!DOCTYPE html>
<html>
  <body>
    This is a common web site.
  </body>
</html>
```

Check that the special workload on the guilder cluster is working.

```shell
$ curl http://localhost:8082
<!DOCTYPE html>
<html>
  <body>
    This is a special web site.
  </body>
</html>
```

Check that the common workload on the guilder cluster is working.

```shell
$ curl http://localhost:8083
<!DOCTYPE html>
<html>
  <body>
    This is a common web site.
  </body>
</html>
```

## Stage 6

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
