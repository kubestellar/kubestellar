# kubestellar-latency-collector

> A Kubernetes controller to measure and expose latency metrics for KubeStellar’s downsync and upsync operations.

---

## 🚀 Demo Setup

Create a KubeStellar demo environment using Kind:

```bash
bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/refs/tags/v0.27.2/scripts/create-kubestellar-demo-env.sh) --platform kind
```

Set up your environment:

```bash
export host_context=kind-kubeflex
export its_cp=its1
export its_context=its1
export wds_cp=wds1
export wds_context=wds1
export wec1_name=cluster1
export wec2_name=cluster2
export wec1_context=cluster1
export wec2_context=cluster2
export label_query_both="location-group=edge"
export label_query_one="name=cluster1"
```

## 📦 Deploy Example Workloads

Binding Policy

```bash
kubectl --context "$wds_context" apply -f - <<EOF
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: nginx-singleton-bpolicy
spec:
  clusterSelectors:
  - matchLabels: {"name":"cluster1"}
  downsync:
  - objectSelectors:
    - matchLabels: {"app.kubernetes.io/name":"nginx-singleton"}
    wantSingletonReportedState: true
EOF
```

### Deployments

Apply deployments to generate observable latency metrics:

```bash
# Deployment 1
kubectl --context "$wds_context" apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-singleton-deployment
  namespace: default
  labels:
    app.kubernetes.io/name: nginx-singleton
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 80
        readinessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 5
EOF
```

```bash
# Deployment 2 (variation)
kubectl --context "$wds_context" apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-test-deployment
  namespace: default
  labels:
    app.kubernetes.io/name: nginx-singleton
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 80
        readinessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 5
EOF
```

## 🛠️ Build & Run

Build

```bash
make build
```

Run

```bash
make run
```

## 📊 Monitoring Setup

### Another Terminal (Prometheus)

```bash
# Download and extract
wget https://github.com/prometheus/prometheus/releases/download/v2.51.0/prometheus-2.51.0.linux-amd64.tar.gz
tar -xzf prometheus-2.51.0.linux-amd64.tar.gz
cd prometheus-2.51.0.linux-amd64
```

Create `prometheus.yml`:

```bash
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'kubestellar-latency'
    static_configs:
      - targets: ['localhost:2222']
    metrics_path: '/metrics'
```

Run Prometheus:

```bash
./prometheus \
  --config.file=prometheus.yml \
  --storage.tsdb.path=./data \
  --web.listen-address=0.0.0.0:9090
```

### Again In Another Terminal (Grafana)

```bash
# Download and extract
wget https://dl.grafana.com/oss/release/grafana-10.4.1.linux-amd64.tar.gz
tar -xzf grafana-10.4.1.linux-amd64.tar.gz
cd grafana-10.4.1
```

Start Grafana

```bash
./bin/grafana-server web
```

Access Grafana at http://localhost:3000
(Default credentials: admin / admin)

## Import Dashboard

- Navigate to Dashboards > Import
- Upload the file: ./kubestellar-dashboard.json
