<!--kubestellar-apply-apache-start-->
<span class="Space-Bd-BT">KUBESTELLAR</span>'s helm chart automatically creates a Workload Management
Workspace (WMW) for you to store kubernetes workload descriptions and <span class="Space-Bd-BT">KUBESTELLAR</span> control objects in. The automatically created WMW is at `root:wmw1`.

Create an EdgePlacement control object to direct where your workload runs using the 'location-group=edge' label selector. This label selector's value ensures your workload is directed to both clusters, as they were labeled with 'location-group=edge' when you issued the 'kubestellar prep-for-cluster' command above.

In the `root:wmw1` workspace create the following `EdgePlacement` object: 
```shell linenums="1" hl_lines="10 11 16 21"
export KUBECONFIG=ks-core.kubeconfig
kubectl ws root:wmw1

kubectl apply -f - <<EOF
apiVersion: edge.kubestellar.io/v2alpha1
kind: EdgePlacement
metadata:
  name: my-first-edge-placement
spec:
  locationSelectors:
  - matchLabels: {"location-group":"edge"}
  downsync:
  - apiGroup: ""
    resources: [ configmaps ]
    namespaceSelectors:
    - matchLabels: {"common":"sure-is"}
    objectNames: [ "*" ]
  - apiGroup: apps
    resources: [ deployments ]
    namespaceSelectors:
    - matchLabels: {"common":"sure-is"}
    objectNames: [ my-first-kubestellar-deployment ]
  - apiGroup: apis.kcp.io
    resources: [ apibindings ]
    namespaceSelectors: []
    objectNames: [ "bind-kubernetes", "bind-apps" ]
EOF
```

check if your edgeplacement was applied to the **ks-core** `kubestellar` namespace correctly
```shell
export KUBECONFIG=ks-core.kubeconfig
kubectl ws root:wmw1
kubectl get edgeplacements -n kubestellar -o yaml
```

Now, apply the HTTP server workload definition into the WMW on **ks-core**. Note the namespace label matches the label in the namespaceSelector for the EdgePlacement (`my-first-edge-placement`) object created above. 
```shell linenums="1" hl_lines="7 14 29"
export KUBECONFIG=ks-core.kubeconfig
kubectl apply -f - <<EOF

apiVersion: v1
kind: Namespace
metadata:
  name: my-namespace
  labels: {common: "sure-is"}
---
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: my-namespace
  name: httpd-htdocs
  labels: {common: "sure-is"}
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
  labels: {common: "sure-is"}
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

check if your configmap and deployment was applied to the **ks-core** `my-namespace` namespace correctly
```shell
export KUBECONFIG=ks-core.kubeconfig
kubectl ws root:wmw1
kubectl get deployments/my-first-kubestellar-deployment -n my-namespace -o yaml
kubectl get deployments,cm -n my-namespace
```

<!--kubestellar-apply-apache-end-->