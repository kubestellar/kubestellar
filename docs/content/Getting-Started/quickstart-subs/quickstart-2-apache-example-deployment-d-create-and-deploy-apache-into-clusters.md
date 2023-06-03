<!--quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters-start-->
### d. Create and deploy the Apache Server workload into florin and guilder clusters

Create the `EdgePlacement` object for your workload. Its “where predicate” (the locationSelectors array) has one label selector that matches the Location objects (`florin` and `guilder`) created earlier, thus directing the workload to both edge clusters.

In the `example-wmw` workspace create the following `EdgePlacement` object: 
  
```shell linenums="1"
kubectl ws root:my-org:example-wmw

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

Put the prescription of the HTTP server workload into the WMW. Note the namespace label matches the label in the namespaceSelector for the EdgePlacement (`edge-placement-c`) object created above. 


```shell linenums="1"
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

Now, let's check that the deployment was created in the `florin` edge cluster - it may take a few 10s of seconds to appear:

```shell
kubectl --context kind-florin get deployments -A
```

which should yield something like:

```console
NAMESPACE                         NAME                              READY   UP-TO-DATE   AVAILABLE   AGE
commonstuff                       commond                           1/1     1            1           6m48s
kcp-edge-syncer-florin-2upj1awn   kcp-edge-syncer-florin-2upj1awn   1/1     1            1           16m
kube-system                       coredns                           2/2     2            2           28m
local-path-storage                local-path-provisioner            1/1     1            1           28m
```

Also, let's check that the deployment was created in the `guilder` edge cluster:

```shell
kubectl --context kind-guilder get deployments -A
```

which should yield something like:

```console
NAMESPACE                          NAME                               READY   UP-TO-DATE   AVAILABLE   AGE
commonstuff                        commond                            1/1     1            1           7m54s
kcp-edge-syncer-guilder-6tuay5d6   kcp-edge-syncer-guilder-6tuay5d6   1/1     1            1           12m
kube-system                        coredns                            2/2     2            2           27m
local-path-storage                 local-path-provisioner             1/1     1            1           27m
```

Lastly, let's check that the workload is working in both clusters:

For `florin`:

```shell
while [[ $(kubectl --context kind-florin get pod -l "app=common" -n commonstuff -o jsonpath='{.items[0].status.phase}') != "Running" ]]; do sleep 5; done;curl http://localhost:8094
```

which may yield the error below, depending on how long it takes for the Apache HTTP Server pod to get synchronized and running:

```console
curl: (52) Empty reply from server
```

If you are getting the error then wait 1-2 minutes and run `curl` again to see the expected result:

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
while [[ $(kubectl --context kind-guilder get pod -l "app=common" -n commonstuff -o jsonpath='{.items[0].status.phase}') != "Running" ]]; do sleep 5; done;curl http://localhost:8096
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


Congratulations, you’ve just deployed a workload to two edge clusters using kubestellar! To learn more about kubestellar please visit our [User Guide](user-guide.md)
<!--quickstart-2-apache-example-deployment-d-create-and-deploy-apache-into-clusters-end-->