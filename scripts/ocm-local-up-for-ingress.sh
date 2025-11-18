#!/usr/bin/env bash
# Copyright 2024 The KubeStellar Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Create a kind cluster for KubeStellar deployment

set -e

hub=${CLUSTER1:-hub}
c1=${CLUSTER1:-cluster1}
c2=${CLUSTER2:-cluster2}

hubctx="kind-${hub}"
c1ctx="kind-${c1}"
c2ctx="kind-${c2}"

kind create cluster --name "${hub}" --config - <<EOF
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
    hostPort: 9443
    protocol: TCP
EOF

echo "Installing an nginx ingress controller into the hub cluster..."
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/refs/tags/controller-v1.12.1/deploy/static/provider/kind/deploy.yaml

echo "Patching nginx ingress controller to enable SSL passthrough..."
kubectl --context "${hubctx}" patch deployment ingress-nginx-controller -n ingress-nginx -p '{"spec":{"template":{"spec":{"containers":[{"name":"controller","args":["/nginx-ingress-controller","--election-id=ingress-nginx-leader","--controller-class=k8s.io/ingress-nginx","--ingress-class=nginx","--configmap=$(POD_NAMESPACE)/ingress-nginx-controller","--validating-webhook=:8443","--validating-webhook-certificate=/usr/local/certificates/cert","--validating-webhook-key=/usr/local/certificates/key","--watch-ingress-without-class=true","--publish-status-address=localhost","--enable-ssl-passthrough"]}]}}}}'

kind create cluster --name "${c1}"
kind create cluster --name "${c2}"

echo "Waiting for nginx ingress controller with SSL passthrough to be ready..."
# Wait for the deployment to roll out with the new configuration
echo "Waiting for nginx ingress controller deployment to be rolled out..."
kubectl --context "${hubctx}" rollout status deployment/ingress-nginx-controller \
  --namespace ingress-nginx \
  --timeout=300s

# Wait for the new pod to be ready
echo "Waiting for nginx ingress controller pod to be ready..."
kubectl --context "${hubctx}" wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s


get_bootstrap_sa_name() {
    local v="$1"
    if [ -z "$v" ]; then
        echo "cluster-bootstrap"
        return
    fi
    if [ "$(printf '%s\n' "$v" "0.11.0" | sort -V | head -1)" = "0.11.0" ]; then
        echo "cluster-bootstrap"
    else
        echo "agent-registration-bootstrap"
    fi
}

echo "Initialize the ocm hub cluster\n"
clusteradm init --wait --context ${hubctx}

CLUSTERADM_VERSION=$(clusteradm version --short 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
sa_name=$(get_bootstrap_sa_name "$CLUSTERADM_VERSION")

joincmd=$(clusteradm get token --context ${hubctx} | grep clusteradm)

echo "Join cluster1 to hub\n"
$(echo ${joincmd} --force-internal-endpoint-lookup --wait --context ${c1ctx} | sed "s/<cluster_name>/$c1/g")

echo "Join cluster2 to hub\n"
$(echo ${joincmd} --force-internal-endpoint-lookup --wait --context ${c2ctx} | sed "s/<cluster_name>/$c2/g")

echo "Accept join of cluster1 and cluster2"
clusteradm accept --context ${hubctx} --clusters ${c1},${c2} --wait

kubectl get managedclusters --all-namespaces --context ${hubctx}
