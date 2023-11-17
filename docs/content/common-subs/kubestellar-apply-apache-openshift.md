<!--kubestellar-apply-apache-openshift-start-->
KubeStellar's helm chart automatically creates a Workload Management
Workspace (WMW) for you to store kubernetes workload descriptions and KubeStellar control objects in. The automatically created WMW is at `root:wmw1`.

Create an EdgePlacement control object to direct where your workload runs using the 'location-group=edge' label selector. This label selector's value ensures your workload is directed to both clusters, as they were labeled with 'location-group=edge' when you issued the 'kubestellar prep-for-cluster' command above.

This EdgePlacement includes downsync of a `RoleBinding` that grants
privileges that let the httpd pod run in an OpenShift cluster.

In the `root:wmw1` workspace create the following `EdgePlacement` object: 
```shell hl_lines="13 14 18 27 28 35"
KUBECONFIG=ks-core.kubeconfig curl -sSL \
  https://raw.githubusercontent.com/openshift/router/master/deploy/route_crd.yaml | kubectl apply -f -

KUBECONFIG=ks-core.kubeconfig kubectl ws root:wmw1

KUBECONFIG=ks-core.kubeconfig kubectl apply -f - <<EOF
apiVersion: edge.kubestellar.io/v2alpha1
kind: EdgePlacement
metadata:
  name: my-first-edge-placement
spec:
  locationSelectors:
  - matchLabels: {"location-group":"edge"}
  downsync:
  - apiGroup: ""
    resources: [ configmaps, services ]
    namespaces: [ my-namespace ]
    objectNames: [ "*" ]
  - apiGroup: route.openshift.io
    namespaces:
    - my-namespace
    objectNames: [ "*" ]
    resources: [ routes ]
  - apiGroup: apps
    resources: [ deployments ]
    namespaces: [ my-namespace ]
    objectNames: [ my-first-kubestellar-deployment ]
  - apiGroup: apis.kcp.io
    resources: [ apibindings ]
    namespaceSelectors: []
    objectNames: [ "bind-kubernetes", "bind-apps" ]
  - apiGroup: rbac.authorization.k8s.io
    resources: [ rolebindings ]
    namespaces: [ my-namespace ]
    objectNames: [ let-it-be ]
EOF
```

check if your edgeplacement was applied to the **ks-core** `kubestellar` namespace correctly
```shell
KUBECONFIG=ks-core.kubeconfig kubectl ws root:wmw1
KUBECONFIG=ks-core.kubeconfig kubectl get edgeplacements -n kubestellar -o yaml
```

Now, apply the HTTP server workload definition into the WMW on **ks-core**. Note the namespace label matches the label in the namespaceSelector for the EdgePlacement (`my-first-edge-placement`) object created above. 
```shell hl_lines="5 10 24 25"
KUBECONFIG=ks-core.kubeconfig kubectl apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: my-namespace
---
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: my-namespace
  name: httpd-htdocs
data:
  index.html: |
    <!DOCTYPE html>
    <html>
      <body>
        This web site is hosted on ks-edge-cluster1 and ks-edge-cluster2.
      </body>
    </html>
---
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: my-namespace
  name: my-first-kubestellar-deployment
spec:
  selector: {matchLabels: {app: common} }
  template:
    metadata:
      labels: {app: common}
    spec:
      containers:
      - name: httpd
        image: registry.redhat.io/rhel8/httpd-24
        ports:
        - name: http
          containerPort: 8080
          protocol: TCP
        volumeMounts:
        - name: htdocs
          readOnly: true
          mountPath: /var/www/html/
      volumes:
      - name: htdocs
        configMap:
          name: httpd-htdocs
          optional: false
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: let-it-be
  namespace: my-namespace
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:openshift:scc:privileged
subjects:
- kind: ServiceAccount
  name: default
  namespace: my-namespace
---
apiVersion: v1
kind: Service
metadata:
  name: my-service
  namespace: my-namespace
spec:
  ports:
    - protocol: TCP
      port: 8080
      targetPort: 8080
  selector:
    app: common
---
apiVersion: route.openshift.io/v1
kind: Route
metadata:
  name: my-route
  namespace: my-namespace
spec:
  port:
    targetPort: 8080
  to:
    kind: Service
    name: my-service
EOF
```

check if your configmap, deployment, service, and route was applied to the **ks-core** `my-namespace` namespace correctly
```shell

KUBECONFIG=ks-core.kubeconfig kubectl ws root:wmw1
KUBECONFIG=ks-core.kubeconfig kubectl get deployments/my-first-kubestellar-deployment -n my-namespace -o yaml
KUBECONFIG=ks-core.kubeconfig kubectl get deployments,cm,service,route -n my-namespace
```

<!--kubestellar-apply-apache-openshift-end-->
