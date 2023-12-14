#!/usr/bin/env bash

# Create a kind cluster which will be used to host the space manager and kubeflex. 
# We could create two separate clusters, one for each, but for convenience we create one. 
# Note that kubeflex requires ingress, so we also need to set that up.

kind create cluster --name sm-mgt --config - <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 443
    hostPort: 1024
    protocol: TCP
EOF

kubectl create -f https://raw.githubusercontent.com/kubestellar/kubestellar/main/example/kind-nginx-ingress-with-SSL-passthrough.yaml
sleep 20
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=180s
