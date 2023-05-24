---
title: "2023q1 PoC Example Scenario"
linkTitle: "2023q1 PoC Example Scenario"
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

![Boxes and arrows. Two kind clusters exist, named florin and guilder. The Inventory Management workspace contains two pairs of SyncTarget and Location objects. The Edge Service Provider workspace contains the PoC controllers; the mailbox controller reads the SyncTarget objects and creates two mailbox workspaces.](Edge-PoC-2023q1-Scenario-1-stage-1.svg "Stage 1 Summary")

Stage 1 creates the infrastructure and the edge service provider
workspace and lets that react to the inventory.  Then the edge syncers
are deployed, in the edge clusters and configured to work with the
corresponding mailbox workspaces.  This stage has the following steps.

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

Download and build or install [kcp](https://github.com/kcp-dev/kcp),
according to your preference.

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

### Get edge-mc

Download and build or install
[edge-mc](https://github.com/kcp-dev/edge-mc), according to your
preference.  That is, either (a) `git clone` the repo and then `make
build` to populate its `bin` directory, or (b) fetch the binary
archive appropriate for your machine from a release and unpack it
(creating a `bin` directory).  In the following exhibited command
lines, the commands described as "edge-mc commands" and the commands
that start with `kubectl kubestellar` rely on the edge-mc `bin` directory
being on the `$PATH`.  Alternatively you could invoke them with
explicit pathnames.  The kubectl plugin lines use fully specific
executables (e.g., `kubectl kubestellar prep-for-syncer` corresponds to
`bin/kubectl-kubestellar-prep_for_syncer`).

### Create SyncTarget and Location objects to represent the florin and guilder clusters

Use the following two commands. They label both florin and guilder
with `env=prod`, and also label guilder with `extended=si`.

```shell
kubectl kubestellar ensure location florin  loc-name=florin  env=prod
kubectl kubestellar ensure location guilder loc-name=guilder env=prod extended=si
```

Those two script invocations are equivalent to creating the following
four objects.

```yaml
apiVersion: workload.kcp.io/v1alpha1
kind: SyncTarget
metadata:
  name: florin
  labels:
    id: florin
    loc-name: florin
    env: prod
---
apiVersion: scheduling.kcp.io/v1alpha1
kind: Location
metadata:
  name: florin
  labels:
    loc-name: florin
    env: prod
spec:
  resource: {group: workload.kcp.io, version: v1alpha1, resource: synctargets}
  instanceSelector:
    matchLabels: {id: florin}
---
apiVersion: workload.kcp.io/v1alpha1
kind: SyncTarget
metadata:
  name: guilder
  labels:
    id: guilder
    loc-name: guilder
    env: prod
    extended: si
---
apiVersion: scheduling.kcp.io/v1alpha1
kind: Location
metadata:
  name: guilder
  labels:
    loc-name: guilder
    env: prod
    extended: si
spec:
  resource: {group: workload.kcp.io, version: v1alpha1, resource: synctargets}
  instanceSelector:
    matchLabels: {id: guilder}
```

That script also deletes the Location named `default`, which is not
used in this PoC, if it shows up.

### Create the edge service provider workspace

Use the following commands.

```shell
kubectl ws root
kubectl ws create espw --enter
```

### Populate the edge service provider workspace

This puts the definition and export of the edge-mc API in the edge
service provider workspace.

Use the following command.

```shell
kubectl create -f config/exports
```

### The mailbox controller

Running the mailbox controller will be conveniently automated.
Eventually.  In the meantime, you can use the edge-mc command shown
here.

```console
$ mailbox-controller -v=2
...
I0423 01:09:37.991080   10624 main.go:196] "Found APIExport view" exportName="workload.kcp.io" serverURL="https://192.168.58.123:6443/services/apiexport/root/workload.kcp.io"
...
I0423 01:09:38.449395   10624 controller.go:299] "Created APIBinding" worker=1 mbwsName="apmziqj9p9fqlflm-mb-bf452e1f-45a0-4d5d-b35c-ef1ece2879ba" mbwsCluster="yk9a66vjms1pi8hu" bindingName="bind-edge" resourceVersion="914"
...
I0423 01:09:38.842881   10624 controller.go:299] "Created APIBinding" worker=3 mbwsName="apmziqj9p9fqlflm-mb-b8c64c64-070c-435b-b3bd-9c0f0c040a54" mbwsCluster="12299slctppnhjnn" bindingName="bind-edge" resourceVersion="968"
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
1t82bk54r6gjnzsp-mb-1a045336-8178-4026-8a56-5cd5609c0ec1   universal            Ready   https://192.168.58.123:6443/clusters/1najcltzt2nqax47   50s
1t82bk54r6gjnzsp-mb-f0a82ab1-63f4-49ea-954d-3a41a35a9f1c   universal            Ready   https://192.168.58.123:6443/clusters/1y7wll1dz806h3sb   50s
```

More usefully, using custom columns you can get a listing that shows
the _name_ of the associated SyncTarget.

```console
$ kubectl get Workspace -o "custom-columns=NAME:.metadata.name,SYNCTARGET:.metadata.annotations['edge\.kcp\.io/sync-target-name'],CLUSTER:.spec.cluster"
NAME                                                       SYNCTARGET   CLUSTER
1t82bk54r6gjnzsp-mb-1a045336-8178-4026-8a56-5cd5609c0ec1   florin       1najcltzt2nqax47
1t82bk54r6gjnzsp-mb-f0a82ab1-63f4-49ea-954d-3a41a35a9f1c   guilder      1y7wll1dz806h3sb
```

Also: if you ever need to look up just one mailbox workspace by
SyncTarget name, you could do it as follows.

```console
$ kubectl get Workspace -o json | jq -r '.items | .[] | .metadata | select(.annotations ["edge.kcp.io/sync-target-name"] == "guilder") | .name'
1t82bk54r6gjnzsp-mb-f0a82ab1-63f4-49ea-954d-3a41a35a9f1c
```

### Connect guilder edge cluster with its mailbox workspace

The following command will (a) create, in the mailbox workspace for
guilder, an identity and authorizations for the edge syncer and (b)
write a file containing YAML for deploying the syncer in the guilder
cluster.

```console
$ kubectl kubestellar prep-for-syncer --imw root:imw-1 guilder
Current workspace is "root:imw-1".
Current workspace is "root:espw".
Current workspace is "root:espw:1t82bk54r6gjnzsp-mb-f0a82ab1-63f4-49ea-954d-3a41a35a9f1c" (type root:universal).
Creating service account "kcp-edge-syncer-guilder-wfeig2lv"
Creating cluster role "kcp-edge-syncer-guilder-wfeig2lv" to give service account "kcp-edge-syncer-guilder-wfeig2lv"

 1. write and sync access to the synctarget "kcp-edge-syncer-guilder-wfeig2lv"
 2. write access to apiresourceimports.

Creating or updating cluster role binding "kcp-edge-syncer-guilder-wfeig2lv" to bind service account "kcp-edge-syncer-guilder-wfeig2lv" to cluster role "kcp-edge-syncer-guilder-wfeig2lv".

Wrote physical cluster manifest to guilder-syncer.yaml for namespace "kcp-edge-syncer-guilder-wfeig2lv". Use

  KUBECONFIG=<pcluster-config> kubectl apply -f "guilder-syncer.yaml"

to apply it. Use

  KUBECONFIG=<pcluster-config> kubectl get deployment -n "kcp-edge-syncer-guilder-wfeig2lv" kcp-edge-syncer-guilder-wfeig2lv

to verify the syncer pod is running.
Current workspace is "root:espw".
```

The file written was, as mentioned in the output,
`guilder-syncer.yaml`.  Next `kubectl apply` that to the guilder
cluster.  That will look something like the following; adjust as
necessary to make kubectl manipulate **your** guilder cluster.

```console
$ KUBECONFIG=~/.kube/config kubectl --context kind-guilder apply -f guilder-syncer.yaml
namespace/kcp-edge-syncer-guilder-wfeig2lv created
serviceaccount/kcp-edge-syncer-guilder-wfeig2lv created
secret/kcp-edge-syncer-guilder-wfeig2lv-token created
clusterrole.rbac.authorization.k8s.io/kcp-edge-syncer-guilder-wfeig2lv created
clusterrolebinding.rbac.authorization.k8s.io/kcp-edge-syncer-guilder-wfeig2lv created
secret/kcp-edge-syncer-guilder-wfeig2lv created
deployment.apps/kcp-edge-syncer-guilder-wfeig2lv created
```

You might check that the syncer is running, as follows.

```console
$ KUBECONFIG=~/.kube/config kubectl --context kind-guilder get deploy -A
NAMESPACE                          NAME                               READY   UP-TO-DATE   AVAILABLE   AGE
kcp-edge-syncer-guilder-saaywsu5   kcp-edge-syncer-guilder-saaywsu5   1/1     1            1           52s
kube-system                        coredns                            2/2     2            2           35m
local-path-storage                 local-path-provisioner             1/1     1            1           35m
```

### Connect florin edge cluster with its mailbox workspace

Do the analogous stuff for the florin cluster.

```console
$ kubectl kubestellar prep-for-syncer --imw root:imw-1 florin
Current workspace is "root:imw-1".
Current workspace is "root:espw".
Current workspace is "root:espw:1t82bk54r6gjnzsp-mb-1a045336-8178-4026-8a56-5cd5609c0ec1" (type root:universal).
Creating service account "kcp-edge-syncer-florin-32uaph9l"
Creating cluster role "kcp-edge-syncer-florin-32uaph9l" to give service account "kcp-edge-syncer-florin-32uaph9l"

 1. write and sync access to the synctarget "kcp-edge-syncer-florin-32uaph9l"
 2. write access to apiresourceimports.

Creating or updating cluster role binding "kcp-edge-syncer-florin-32uaph9l" to bind service account "kcp-edge-syncer-florin-32uaph9l" to cluster role "kcp-edge-syncer-florin-32uaph9l".

Wrote physical cluster manifest to florin-syncer.yaml for namespace "kcp-edge-syncer-florin-32uaph9l". Use

  KUBECONFIG=<pcluster-config> kubectl apply -f "florin-syncer.yaml"

to apply it. Use

  KUBECONFIG=<pcluster-config> kubectl get deployment -n "kcp-edge-syncer-florin-32uaph9l" kcp-edge-syncer-florin-32uaph9l

to verify the syncer pod is running.
Current workspace is "root:espw".
```

And deploy the syncer in the florin cluster.

```console
$ KUBECONFIG=~/.kube/config kubectl --context kind-florin apply -f florin-syncer.yaml 
namespace/kcp-edge-syncer-florin-32uaph9l created
serviceaccount/kcp-edge-syncer-florin-32uaph9l created
secret/kcp-edge-syncer-florin-32uaph9l-token created
clusterrole.rbac.authorization.k8s.io/kcp-edge-syncer-florin-32uaph9l created
clusterrolebinding.rbac.authorization.k8s.io/kcp-edge-syncer-florin-32uaph9l created
secret/kcp-edge-syncer-florin-32uaph9l created
deployment.apps/kcp-edge-syncer-florin-32uaph9l created
```

## Stage 2

![Placement and scheduling](Edge-PoC-2023q1-Scenario-1-stage-2.svg
"Stage 2 summary")

Stage 2 creates two workloads, called "common" and "special", and lets
the scheduler react.  It has the following steps.

### Create and populate the workload management workspace for the common workload

One of the workloads is called "common", because it will go to both
edge clusters.  The other one is called "special".

In this example, each workload description goes in its own workload
management workspace (WMW).  Start by creating a common parent for
those two workspaces, with the following commands.

```shell
kubectl ws root
kubectl ws create my-org --enter
```

Next, create the WMW for the common workload.  The following command
will do that, if issued while "root:my-org" is the current workspace.

```shell
kubectl kubestellar ensure wmw wmw-c
```

This is equivalent to creating that workspace and then entering it and
creating the following two `APIBinding` objects.

```yaml
apiVersion: apis.kcp.io/v1alpha1
kind: APIBinding
metadata:
  name: bind-espw
spec:
  reference:
    export:
      path: root:espw
      name: edge.kcp.io
---
apiVersion: apis.kcp.io/v1alpha1
kind: APIBinding
metadata:
  name: bind-kube
spec:
  reference:
    export:
      path: "root:compute"
      name: kubernetes
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
  annotations:
    edge.kcp.io/expand-parameters: "true"
data:
  index.html: |
    <!DOCTYPE html>
    <html>
      <body>
        This is a common web site.
        Running in %(loc-name).
      </body>
    </html>
---
apiVersion: edge.kcp.io/v1alpha1
kind: Customizer
metadata:
  namespace: commonstuff
  name: example-customizer
  annotations:
    edge.kcp.io/expand-parameters: "true"
replacements:
- path: "$.spec.template.spec.containers.0.env.0.value"
  value: '"env is %(env)"'
---
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: commonstuff
  name: commond
  annotations:
    edge.kcp.io/customizer: example-customizer
spec:
  selector: {matchLabels: {app: common} }
  template:
    metadata:
      labels: {app: common}
    spec:
      containers:
      - name: httpd
        env:
        - name: EXAMPLE_VAR
          value: example value
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
    names: ["william"]
EOF
```

### Create and populate the workload management workspace for the special workload

Use the following `kubectl` commands to create the WMW for the special
workload.

```shell
kubectl ws root:my-org
kubectl kubestellar ensure wmw wmw-s
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
  annotations:
    edge.kcp.io/expand-parameters: "true"
data:
  index.html: |
    <!DOCTYPE html>
    <html>
      <body>
        This is a special web site.
        Running in %(loc-name).
      </body>
    </html>
---
apiVersion: edge.kcp.io/v1alpha1
kind: Customizer
metadata:
  namespace: specialstuff
  name: example-customizer
  annotations:
    edge.kcp.io/expand-parameters: "true"
replacements:
- path: "$.spec.template.spec.containers.0.env.0.value"
  value: '"in %(env) env"'
---
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: specialstuff
  name: speciald
  annotations:
    edge.kcp.io/customizer: example-customizer
spec:
  selector: {matchLabels: {app: special} }
  template:
    metadata:
      labels: {app: special}
    spec:
      containers:
      - name: httpd
        env:
        - name: EXAMPLE_VAR
          value: example value
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

In response to each EdgePlacement, the scheduler will create a
corresponding SinglePlacementSlice object.  These will indicate the
following resolutions of the "where" predicates.

| EdgePlacement | Resolved Where |
| ------------- | -------------: |
| edge-placement-c | florin, guilder |
| edge-placement-s | guilder |

Eventually there will be automation that conveniently runs the
scheduler.  In the meantime, you can run it by hand: switch to the
ESPW and invoke the edge-mc command that runs the scheduler.

```console
$ kubectl ws root:espw
Current workspace is "root:espw".
$ kubestellar-scheduler
I0423 01:33:37.036752   11305 kubestellar-scheduler.go:212] "Found APIExport view" exportName="edge.kcp.io" serverURL="https://192.168.58.123:6443/services/apiexport/7qkse309upzrv0fy/edge.kcp.io"
...
I0423 01:33:37.320859   11305 reconcile_on_location.go:192] "updated SinglePlacementSlice" controller="kubestellar-scheduler" triggeringKind=Location key="apmziqj9p9fqlflm|florin" locationWorkspace="apmziqj9p9fqlflm" location="florin" workloadWorkspace="10l175x6ejfjag3e" singlePlacementSlice="edge-placement-c"
...
I0423 01:33:37.391772   11305 reconcile_on_location.go:192] "updated SinglePlacementSlice" controller="kubestellar-scheduler" triggeringKind=Location key="apmziqj9p9fqlflm|guilder" locationWorkspace="apmziqj9p9fqlflm" location="guilder" workloadWorkspace="10l175x6ejfjag3e" singlePlacementSlice="edge-placement-c"
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
  - cluster: apmziqj9p9fqlflm
    locationName: florin
    syncTargetName: florin
    syncTargetUID: b8c64c64-070c-435b-b3bd-9c0f0c040a54
  - cluster: apmziqj9p9fqlflm
    locationName: guilder
    syncTargetName: guilder
    syncTargetUID: bf452e1f-45a0-4d5d-b35c-ef1ece2879ba
  kind: SinglePlacementSlice
  metadata:
    annotations:
      kcp.io/cluster: 10l175x6ejfjag3e
    creationTimestamp: "2023-04-23T05:33:37Z"
    generation: 4
    name: edge-placement-c
    ownerReferences:
    - apiVersion: edge.kcp.io/v1alpha1
      kind: EdgePlacement
      name: edge-placement-c
      uid: 199cfe1e-48d9-4351-af5c-e66c83bf50dd
    resourceVersion: "1316"
    uid: b5db1f9d-1aed-4a25-91da-26dfbb5d8879
kind: List
metadata:
  resourceVersion: ""
```

Also check out the SinglePlacementSlice objects in
`root:my-org:wmw-s`.  It should go similarly, but the `destinations`
should include only the entry for guilder.

## Stage 3

![Placement translation](Edge-PoC-2023q1-Scenario-1-stage-3.svg "Stage
3 summary")

In Stage 3, in response to the EdgePlacement and SinglePlacementSlice
objects, the placement translator will copy the workload prescriptions
into the mailbox workspaces and create `SyncerConfig` objects there.

Eventually there will be convenient automation running the placement
translator.  In the meantime, you can run it manually: switch to the
ESPW and use the edge-mc command that runs the placement translator.

```console
$ kubectl ws root:espw
Current workspace is "root:espw".
$ placement-translator
I0423 01:39:56.362722   11644 shared_informer.go:282] Waiting for caches to sync for placement-translator
...
```

After it stops logging stuff, wait another minute and then you can ^C
it or use another shell to continue exploring.

The florin cluster gets only the common workload.  Examine florin's
`SyncerConfig` as follows.

```shell
$ kubectl ws 1t82bk54r6gjnzsp-mb-1a045336-8178-4026-8a56-5cd5609c0ec1
Current workspace is "root:espw:1t82bk54r6gjnzsp-mb-1a045336-8178-4026-8a56-5cd5609c0ec1" (type root:universal).

$ kubectl get SyncerConfig the-one -o yaml
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
      group: networking.k8s.io
      resource: ingresses
    - apiVersion: v1
      group: rbac.authorization.k8s.io
      resource: roles
    - apiVersion: v1
      group: ""
      resource: configmaps
    - apiVersion: v1
      group: ""
      resource: limitranges
    - apiVersion: v1
      group: ""
      resource: secrets
    - apiVersion: v1
      group: rbac.authorization.k8s.io
      resource: rolebindings
    - apiVersion: v1
      group: apps
      resource: deployments
    - apiVersion: v1
      group: ""
      resource: pods
    - apiVersion: v1
      group: ""
      resource: serviceaccounts
    - apiVersion: v1
      group: ""
      resource: services
    - apiVersion: v1
      group: ""
      resource: resourcequotas
    - apiVersion: v1
      group: coordination.k8s.io
      resource: leases
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
Examine guilder's `SyncerConfig` object and workloads as follows.

```console
$ kubectl ws root:espw
Current workspace is "root:espw".

$ kubectl ws 1t82bk54r6gjnzsp-mb-f0a82ab1-63f4-49ea-954d-3a41a35a9f1c
Current workspace is "root:espw:1t82bk54r6gjnzsp-mb-f0a82ab1-63f4-49ea-954d-3a41a35a9f1c" (type root:universal).

$ kubectl get SyncerConfig the-one -o yaml
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
      group: ""
      resource: services
    - apiVersion: v1
      group: apps
      resource: deployments
    - apiVersion: v1
      group: ""
      resource: pods
    - apiVersion: v1
      group: coordination.k8s.io
      resource: leases
    - apiVersion: v1
      group: networking.k8s.io
      resource: ingresses
    - apiVersion: v1
      group: ""
      resource: limitranges
    - apiVersion: v1
      group: ""
      resource: serviceaccounts
    - apiVersion: v1
      group: rbac.authorization.k8s.io
      resource: rolebindings
    - apiVersion: v1
      group: ""
      resource: configmaps
    - apiVersion: v1
      group: ""
      resource: secrets
    - apiVersion: v1
      group: rbac.authorization.k8s.io
      resource: roles
    - apiVersion: v1
      group: ""
      resource: resourcequotas
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

$ kubectl get deployments -A
NAMESPACE      NAME       READY   UP-TO-DATE   AVAILABLE   AGE
commonstuff    commond    0/0     0            0           6m1s
specialstuff   speciald   0/0     0            0           5m58s
```

## Stage 4

![Syncer effects](Edge-PoC-2023q1-Scenario-1-stage-4.svg "Stage 4 summary")

In Stage 4, the edge syncer does its thing.  Actually, it should have
done it as soon as the relevant inputs became available in stage 3.
Now we examine what happened.

You can check that the workloads are running in the edge clusters as
they should be.

The syncer does its thing between the florin cluster and its mailbox
workspace.  This is driven by the `SyncerConfig` object named
`the-one` in that mailbox workspace.

The syncer does its thing between the guilder cluster and its mailbox
workspace.  This is driven by the `SyncerConfig` object named
`the-one` in that mailbox workspace.

Using the kubeconfig that `kind` modified, examine the florin cluster.
Find just the `commonstuff` namespace and the `commond` Deployment.

```console
$ KUBECONFIG=~/.kube/config kubectl --context kind-florin get ns
NAME                              STATUS   AGE
commonstuff                       Active   6m51s
default                           Active   57m
kcp-edge-syncer-florin-1t9zgidy   Active   17m
kube-node-lease                   Active   57m
kube-public                       Active   57m
kube-system                       Active   57m
local-path-storage                Active   57m

$ KUBECONFIG=~/.kube/config kubectl --context kind-florin get deploy -A | egrep 'NAME|stuff'
NAMESPACE                         NAME                              READY   UP-TO-DATE   AVAILABLE   AGE
commonstuff                       commond                           1/1     1            1           7m59s
```

Examine the guilder cluster.  Find both workload namespaces and both
Deployments.

```console
$ KUBECONFIG=~/.kube/config kubectl --context kind-guilder get ns | egrep NAME\|stuff
NAME                               STATUS   AGE
commonstuff                        Active   8m33s
specialstuff                       Active   8m33s

$ KUBECONFIG=~/.kube/config kubectl --context kind-guilder get deploy -A | egrep NAME\|stuff
NAMESPACE                          NAME                               READY   UP-TO-DATE   AVAILABLE   AGE
commonstuff                        commond                            1/1     1            1           8m37s
specialstuff                       speciald                           1/1     1            1           8m55s
```

Examining the common workload in the guilder cluster, for example,
will show that the replacement-style customization happened.

```console
$ kubectl --context kind-guilder get deploy -n commonstuff commond -o yaml
...
      containers:
      - env:
        - name: EXAMPLE_VAR
          value: env is prod
        image: library/httpd:2.4
        imagePullPolicy: IfNotPresent
        name: httpd
...
```

Check that the common workload on the florin cluster is working.

```console
$ curl http://localhost:8081
<!DOCTYPE html>
<html>
  <body>
    This is a common web site.
    Running in florin.
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
    Running in guilder.
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
    Running in guilder.
  </body>
</html>
```


## Stage 5

![Summarization for special](Edge-PoC-2023q1-Scenario-1-stage-5s.svg "Status summarization for special")

The status summarizer, driven by the EdgePlacement and
SinglePlacementSlice for the special workload, creates a status
summary object in the specialstuff namespace in the special workload
workspace holding a summary of the corresponding Deployment objects.
In this case there is just one such object, in the mailbox workspace
for the guilder cluster.

![Summarization for common](Edge-PoC-2023q1-Scenario-1-stage-5c.svg "Status summarization for common")

The status summarizer, driven by the EdgePlacement and
SinglePlacementSlice for the common workload, creates a status summary
object in the commonstuff namespace in the common workload workspace
holding a summary of the corresponding Deployment objects.  Those are
the `commond` Deployment objects in the two mailbox workspaces.

