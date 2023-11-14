<!--example1-stage-2-start-->
## Stage 2

![Placement and Where Resolving](../Edge-PoC-2023q1-Scenario-1-stage-2.svg
"Stage 2 summary")

Stage 2 creates two workloads, called "common" and "special", and lets
the Where Resolver react.  It has the following steps.

### Create and populate the workload management workspace for the common workload

One of the workloads is called "common", because it will go to both
edge clusters.  The other one is called "special".

In this example, each workload description goes in its own workload
management workspace (WMW).  Start by creating a WMW for the common
workload, with the following commands.

```shell
IN_CLUSTER=false SPACE_MANAGER_KUBECONFIG=~/.kube/config kubectl kubestellar ensure wmw wmw-c
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
      name: edge.kubestellar.io
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
---
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: commonstuff
  name: httpd-htdocs
  annotations:
    edge.kubestellar.io/expand-parameters: "true"
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
apiVersion: edge.kubestellar.io/v2alpha1
kind: Customizer
metadata:
  namespace: commonstuff
  name: example-customizer
  annotations:
    edge.kubestellar.io/expand-parameters: "true"
replacements:
- path: "$.spec.template.spec.containers.0.env.0.value"
  value: '"env is %(env)"'
---
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  namespace: commonstuff
  name: commond
  annotations:
    edge.kubestellar.io/customizer: example-customizer
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
apiVersion: edge.kubestellar.io/v2alpha1
kind: EdgePlacement
metadata:
  name: edge-placement-c
spec:
  locationSelectors:
  - matchLabels: {"env":"prod"}
  downsync:
  - apiGroup: ""
    resources: [ configmaps ]
    namespaces: [ commonstuff ]
    objectNames: [ httpd-htdocs ]
  - apiGroup: apps
    resources: [ replicasets ]
    namespaces: [ commonstuff ]
  wantSingletonReportedState: true
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
IN_CLUSTER=false SPACE_MANAGER_KUBECONFIG=~/.kube/config kubectl kubestellar ensure wmw wmw-s
```

In this workload we will also demonstrate how to downsync objects
whose kind is defined by a `CustomResourceDefinition` object. We will
use the one from [the Kubernetes documentation for
CRDs](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/),
modified so that the resource it defines is in the category
`all`. First, create the definition object with the following command.

```shell
kubectl apply -f - <<EOF
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: crontabs.stable.example.com
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: stable.example.com
  # list of versions supported by this CustomResourceDefinition
  versions:
    - name: v1
      # Each version can be enabled/disabled by Served flag.
      served: true
      # One and only one version must be marked as the storage version.
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                cronSpec:
                  type: string
                image:
                  type: string
                replicas:
                  type: integer
  # either Namespaced or Cluster
  scope: Namespaced
  names:
    # plural name to be used in the URL: /apis/<group>/<version>/<plural>
    plural: crontabs
    # singular name to be used as an alias on the CLI and for display
    singular: crontab
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: CronTab
    # shortNames allow shorter string to match your resource on the CLI
    shortNames:
    - ct
    categories:
    - all
EOF
```

Next, use the following command to wait for the apiserver to process
that definition.

```shell
kubectl wait --for condition=Established crd crontabs.stable.example.com
```

Next, use `kubectl` to create the following workload objects in that
workspace. The `APIService` object included here does not contribute
to the httpd workload but is here to demonstrate that `APIService`
objects can be downsynced.

```shell
kubectl apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: specialstuff
  labels: {special: "yes"}
  annotations: {just-for: fun}
---
apiVersion: "stable.example.com/v1"
kind: CronTab
metadata:
  name: my-new-cron-object
  namespace: specialstuff
spec:
  cronSpec: "* * * * */5"
  image: my-awesome-cron-image
---
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: specialstuff
  name: httpd-htdocs
  annotations:
    edge.kubestellar.io/expand-parameters: "true"
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
apiVersion: edge.kubestellar.io/v2alpha1
kind: Customizer
metadata:
  namespace: specialstuff
  name: example-customizer
  annotations:
    edge.kubestellar.io/expand-parameters: "true"
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
    edge.kubestellar.io/customizer: example-customizer
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
---
apiVersion: apiregistration.k8s.io/v1
kind: APIService
metadata:
  name: v1090.example.my
spec:
  group: example.my
  groupPriorityMinimum: 360
  service:
    name: my-service
    namespace: my-example
  version: v1090
  versionPriority: 42
EOF
```
``` {.bash .hide-me}
sleep 10
```

Finally, use `kubectl` to create the following EdgePlacement object.
Its "where predicate" (the `locationSelectors` array) has one label
selector that matches only one of the Location objects created
earlier, thus directing the special workload to just one edge cluster.

The "what predicate" explicitly includes the `Namespace` object named
"specialstuff", which causes all of its desired state (including
labels and annotations) to be downsynced. This contrasts with the
common EdgePlacement, which does not explicitly mention the
`commonstuff` namespace, relying on the implicit creation of
namespaces as needed in the WECs.
   
```shell
kubectl apply -f - <<EOF
apiVersion: edge.kubestellar.io/v2alpha1
kind: EdgePlacement
metadata:
  name: edge-placement-s
spec:
  locationSelectors:
  - matchLabels: {"env":"prod","extended":"yes"}
  downsync:
  - apiGroup: ""
    resources: [ configmaps ]
    namespaceSelectors:
    - matchLabels: {"special":"yes"}
  - apiGroup: apps
    resources: [ deployments ]
    namespaceSelectors:
    - matchLabels: {"special":"yes"}
    objectNames: [ speciald ]
  - apiGroup: apiregistration.k8s.io
    resources: [ apiservices ]
    objectNames: [ v1090.example.my ]
  - apiGroup: stable.example.com
    resources: [ crontabs ]
    namespaces: [ specialstuff ]
    objectNames: [ my-new-cron-object ]
  - apiGroup: ""
    resources: [ namespaces ]
    objectNames: [ specialstuff ]
  wantSingletonReportedState: true
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

### Where Resolver

In response to each EdgePlacement, the Where Resolver will create a
corresponding SinglePlacementSlice object.  These will indicate the
following resolutions of the "where" predicates.

| EdgePlacement | Resolved Where |
| ------------- | -------------: |
| edge-placement-c | florin, guilder |
| edge-placement-s | guilder |

If you have deployed the KubeStellar core in a Kubernetes cluster then
the where resolver is running in a pod there. If instead you are
running the core controllers are bare processes then you can use the
following commands to launch the where-resolver; it requires the ESPW
to be the current kcp workspace at start time.

```shell
kubectl ws root:espw
kubestellar-where-resolver &
sleep 10
```

The following commands wait until the where-resolver has done its job
for the common and special `EdgePlacement` objects.

```shell
kubectl ws root:wmw-c
while ! kubectl get SinglePlacementSlice &> /dev/null; do
  sleep 10
done
kubectl ws root:wmw-s
while ! kubectl get SinglePlacementSlice &> /dev/null; do
  sleep 10
done
```

If things are working properly then you will see log lines like the
following (among many others) in the where-resolver's log.

``` { .bash .no-copy }
I0423 01:33:37.036752   11305 main.go:212] "Found APIExport view" exportName="edge.kubestellar.io" serverURL="https://192.168.58.123:6443/services/apiexport/7qkse309upzrv0fy/edge.kubestellar.io"
...
I0423 01:33:37.320859   11305 reconcile_on_location.go:192] "updated SinglePlacementSlice" controller="kubestellar-where-resolver" triggeringKind=Location key="apmziqj9p9fqlflm|florin" locationWorkspace="apmziqj9p9fqlflm" location="florin" workloadWorkspace="10l175x6ejfjag3e" singlePlacementSlice="edge-placement-c"
...
I0423 01:33:37.391772   11305 reconcile_on_location.go:192] "updated SinglePlacementSlice" controller="kubestellar-where-resolver" triggeringKind=Location key="apmziqj9p9fqlflm|guilder" locationWorkspace="apmziqj9p9fqlflm" location="guilder" workloadWorkspace="10l175x6ejfjag3e" singlePlacementSlice="edge-placement-c"
```

Check out a SinglePlacementSlice object as follows.

```shell
kubectl ws root:wmw-c
```
``` { .bash .no-copy }
Current workspace is "root:wmw-c".
```

```shell
kubectl get SinglePlacementSlice -o yaml
```
``` { .bash .no-copy }
apiVersion: v1
items:
- apiVersion: edge.kubestellar.io/v2alpha1
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
    - apiVersion: edge.kubestellar.io/v2alpha1
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
`root:wmw-s`.  It should go similarly, but the `destinations`
should include only the entry for guilder.
<!--example1-stage-2-end-->
