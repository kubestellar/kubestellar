
cd /Users/andan02/projects/edge-mc/docs/scripts/
set -o errexit
set -o nounset
set -o pipefail
set -o xtrace
cat > florin-config.yaml << EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 8081
    hostPort: 8094
EOF
cat > guilder-config.yaml << EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 8081
    hostPort: 8096
  - containerPort: 8082
    hostPort: 8097
EOF
kind create cluster --name florin --config florin-config.yaml
kind create cluster --name guilder --config guilder-config.yaml
git clone https://github.com/kcp-dev/edge-mc KubeStellar
git clone -b v0.11.0 https://github.com/kcp-dev/kcp kcp
cd kcp
make build
export KUBECONFIG=$(pwd)/.kcp/admin.kubeconfig
export PATH=$(pwd)/bin:$PATH
kcp start &> /dev/null &
sleep 30 
kubectl ws root
kubectl ws create imw-1 --enter
cd ../KubeStellar
make build
export PATH=$(pwd)/bin:$PATH
kubectl kubestellar ensure location florin  loc-name=florin  env=prod
kubectl kubestellar ensure location guilder loc-name=guilder env=prod extended=si
kubectl ws root
kubectl ws create espw --enter
kubectl create -f config/exports
go run ./cmd/mailbox-controller -v=2 &
sleep 45
kubectl get Workspaces
kubectl get Workspace -o "custom-columns=NAME:.metadata.name,SYNCTARGET:.metadata.annotations['edge\.kcp\.io/sync-target-name'],CLUSTER:.spec.cluster"
GUILDER_WS=$(kubectl get Workspace -o json | jq -r '.items | .[] | .metadata | select(.annotations ["edge.kcp.io/sync-target-name"] == "guilder") | .name')
FLORIN_WS=$(kubectl get Workspace -o json | jq -r '.items | .[] | .metadata | select(.annotations ["edge.kcp.io/sync-target-name"] == "florin") | .name')
kubectl kubestellar prep-for-syncer --imw root:imw-1 guilder
KUBECONFIG=~/.kube/config kubectl --context kind-guilder apply -f guilder-syncer.yaml
KUBECONFIG=~/.kube/config kubectl --context kind-guilder get deploy -A
kubectl kubestellar prep-for-syncer --imw root:imw-1 florin
KUBECONFIG=~/.kube/config kubectl --context kind-florin apply -f florin-syncer.yaml 
kubectl ws root
kubectl ws create my-org --enter
kubectl kubestellar ensure wmw wmw-c
sleep 15
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
sleep 10
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
sleep 10
kubectl ws root:my-org
kubectl kubestellar ensure wmw wmw-s
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
sleep 10
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
sleep 10
kubectl ws root:espw
go run ./cmd/kubestellar-scheduler &
sleep 45
kubectl ws root:my-org:wmw-c
kubectl get SinglePlacementSlice -o yaml
kubectl ws root:espw
go run ./cmd/placement-translator &
sleep 45
kubectl ws $FLORIN_WS
kubectl get SyncerConfig the-one -o yaml
kubectl get ns
kubectl get deployments -A
kubectl ws root:espw
kubectl ws $GUILDER_WS
kubectl get SyncerConfig the-one -o yaml
kubectl get deployments -A
KUBECONFIG=~/.kube/config kubectl --context kind-florin get ns
KUBECONFIG=~/.kube/config kubectl --context kind-florin get deploy -A | egrep 'NAME|stuff'
KUBECONFIG=~/.kube/config kubectl --context kind-guilder get ns | egrep NAME\|stuff
KUBECONFIG=~/.kube/config kubectl --context kind-guilder get deploy -A | egrep NAME\|stuff
kubectl --context kind-guilder get deploy -n commonstuff commond -o yaml
sleep 20
curl http://localhost:8094
sleep 10
curl http://localhost:8097
curl http://localhost:8096
