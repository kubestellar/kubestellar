<!--quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters-start-->
KubeStellar will have automatically created a Workload Management
Workspace (WMW) for the user to store workload descriptions and KubeStellar
control objects in. The automatically created WMW is at `root:wmw1`.

Create the `EdgePlacement` object for your workload. Note that the locationSelectors element has one label selector that matches the Location objects (`edge-cluster1` and `edge-cluster2`) because they were labeled with 'location_group=edge' when created above.  This match which will direct workload to both edge clusters.

In the `root:wmw1` workspace create the following `EdgePlacement` object: 
  
```shell linenums="1" hl_lines="10 15 20"
kubectl ws root:wmw1

kubectl apply -f - <<EOF
apiVersion: edge.kubestellar.io/v2alpha1
kind: EdgePlacement
metadata:
  name: my-first-edge-placement
spec:
  locationSelectors:
  - matchLabels: {"location_group":"edge"}
  downsync:
  - apiGroup: ""
    resources: [ configmaps ]
    namespaceSelectors:
    - matchLabels: {"common":"sure_is"}
    objectNames: [ "*" ]
  - apiGroup: apps
    resources: [ deployments ]
    namespaceSelectors:
    - matchLabels: {"common":"sure_is"}
    objectNames: [ commond ]
  - apiGroup: apis.kcp.io
    resources: [ apibindings ]
    namespaceSelectors: []
    objectNames: [ "bind-kubernetes", "bind-apps" ]
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

Put the prescription of the HTTP server workload into the WMW. Note the namespace label matches the label in the namespaceSelector for the EdgePlacement (`my-first-edge-placement`) object created above. 


```shell linenums="1" hl_lines="6"
kubectl apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: commonstuff
  labels: {common: "sure_is"}
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

Now, let's check that the deployment was created in the kind `edge-cluster1` cluster - it may take a few 10s of seconds to appear:

```shell
kubectl --context kind-edge-cluster1 get deployments -A
```

which should yield something like:

``` { .sh .no-copy }
NAMESPACE                            NAME                                 READY   UP-TO-DATE   AVAILABLE   AGE
commonstuff                          commond                              1/1     1            1           6m48s
kubestellar-syncer-florin-2upj1awn   kubestellar-syncer-edge-cluster1-2upj1awn   1/1     1            1           16m
kube-system                          coredns                              2/2     2            2           28m
local-path-storage                   local-path-provisioner               1/1     1            1           28m
```

Also, let's check that the deployment was created in the kind `edge-cluster2` cluster:

```shell
kubectl --context kind-edge-cluster2 get deployments -A
```

which should yield something like:

``` { .sh .no-copy }
NAMESPACE                             NAME                                  READY   UP-TO-DATE   AVAILABLE   AGE
commonstuff                           commond                               1/1     1            1           7m54s
kubestellar-syncer-guilder-6tuay5d6   kubestellar-syncer-edge-cluster2-6tuay5d6   1/1     1            1           12m
kube-system                           coredns                               2/2     2            2           27m
local-path-storage                    local-path-provisioner                1/1     1            1           27m
```

Lastly, let's check that the workload is working in both clusters:

For `edge-cluster1`:

```shell
while [[ $(kubectl --context kind-edge-cluster1 get pod -l "app=common" -n commonstuff -o jsonpath='{.items[0].status.phase}') != "Running" ]]; do sleep 5; done;curl http://localhost:8094
```

which should eventually yield:

```html
<!DOCTYPE html>
<html>
  <body>
    This is a common web site.
  </body>
</html>
```

For `guilder`:

```shell
while [[ $(kubectl --context kind-edge-cluster2 get pod -l "app=common" -n commonstuff -o jsonpath='{.items[0].status.phase}') != "Running" ]]; do sleep 5; done;curl http://localhost:8096
```
which should eventually yield:

```html
<!DOCTYPE html>
<html>
  <body>
    This is a common web site.
  </body>
</html>
```


Congratulations, you’ve just deployed a workload to two edge clusters using kubestellar! To learn more about kubestellar please visit our [User Guide](../../user-guide)
<!--quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters-end-->
