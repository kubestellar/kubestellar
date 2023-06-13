<!--example1-stage-2-start-->
## Stage 2

![Placement and scheduling](../Edge-PoC-2023q1-Scenario-1-stage-2.svg
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
``` {.bash .hide-me}
sleep 15
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
``` {.bash .hide-me}
sleep 10
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
``` {.bash .hide-me}
sleep 10
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
``` {.bash .hide-me}
sleep 10
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
``` {.bash .hide-me}
sleep 10
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
ESPW and invoke the KubeStellar command that runs the scheduler.

```shell
kubectl ws root:espw
```
``` { .bash .no-copy }
Current workspace is "root:espw".
```
```shell
go run ./cmd/kubestellar-scheduler &
sleep 45
```
``` { .bash .no-copy }
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

```shell
kubectl ws root:my-org:wmw-c
```
``` { .bash .no-copy }
Current workspace is "root:my-org:wmw-c".
```

```shell
kubectl get SinglePlacementSlice -o yaml
```
``` { .bash .no-copy }
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
<!--example1-stage-2-end-->