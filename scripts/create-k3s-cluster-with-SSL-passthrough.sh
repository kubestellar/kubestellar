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

# Create a k3s cluster for KubeStellar deployment

set -o errexit

# Script info
SCRIPT_NAME="create-k3s-cluster-with-SSL-passthrough.sh" 

port=9443
wait=true

display_help() {
  cat << EOF
Usage: ${SCRIPT_NAME} [--port port] [--nowait][-X]

--port port     map the specified host port to the kind cluster port 443 (default: 9443)
--nowait        when set to true, the script proceeds without waiting for the nginx ingress patching to complete
-X              enable verbose execution of the script for debugging
EOF
}

while (( $# > 0 )); do
    case "$1" in
        (-p|--port)
            if (( $# > 1 ));
            then { port="$2"; shift; }
            else { echo "${SCRIPT_NAME}: missing port number" >&2; exit 1; }
            fi;;
        (--nowait)
            wait=false;;
        (-X)
            set -x;;
        (-h|--help)
            display_help
            exit 0;;
        (-*)
            echo "${SCRIPT_NAME}: unknown flag" >&2
            exit 1;;
        (*)
            echo "${SCRIPT_NAME}: unknown positional argument" >&2
            exit 1;;
    esac
    shift
done

echo "Creating a k3s cluster with SSL passthrough and ${port} port mapping..."

if which kubectl > /dev/null ; then
    if test -f ~/.kube/config; then
        if $(kubectl cluster-info 2> /dev/null | grep -qe "control plane"); then
            echo kubernetes cluster is already up!
            kubectl cluster-info | head -3
            exit 1
        fi
    fi
fi

# Check if k3s is already installed
if ! command -v k3s &> /dev/null
then
    echo "k3s could not be found, installing..."
    sudo curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="--disable traefik,servicelb" sh -
else
    echo "k3s is already installed"
fi

mkdir -p ~/.kube
export KUBECONFIG=~/.kube/config
sudo kubectl --kubeconfig=/etc/rancher/k3s/k3s.yaml config view --raw > "$KUBECONFIG"

# Wait for k3s service to be ready and kubectl to work
echo "Waiting for k3s service to be ready..."
timeout=300
elapsed=0
while [ $elapsed -lt $timeout ]; do
    if sudo systemctl is-active --quiet k3s && kubectl get nodes >/dev/null 2>&1; then
        echo "k3s service is ready!"
        break
    fi
    echo "k3s service not ready yet, waiting... (${elapsed}s/${timeout}s)"
    sleep 5
    elapsed=$((elapsed + 5))
done

if [ $elapsed -ge $timeout ]; then
    echo "Timeout waiting for k3s service to be ready"
    exit 1
fi

echo "Waiting for k3s cluster nodes to be ready..."
kubectl wait --for=condition=ready nodes --all --timeout=300s

kubectl describe endpoints kubernetes

echo "Installing an nginx ingress with SSL passthrough enabled..."

if $(kubectl get svc -n ingress-nginx | grep -qe "ingress-nginx-controller"); then
    echo ingress-nginx is already installed!
    exit 2
fi

helm upgrade --install ingress-nginx ingress-nginx \
    --repo https://kubernetes.github.io/ingress-nginx \
    --namespace ingress-nginx --create-namespace \
    --set controller.updateStrategy.type=RollingUpdate \
    --set controller.updateStrategy.rollingUpdate.maxUnavailable=1 \
    --set controller.service.type=NodePort \
    --set controller.publishService.enabled=false \
    --set controller.hostPort.enabled=true \
    --set controller.hostPort.ports.https=${port} \
    --set controller.extraArgs.enable-ssl-passthrough=true

if [[ "$wait" == "true" ]] ; then
  echo "Waiting for nginx ingress with SSL passthrough to complete..."
  
  # Wait for nginx ingress controller pod to be created
  echo "Waiting for nginx ingress controller pod to be created..."
  kubectl wait --namespace ingress-nginx \
    --for=create pod \
    --selector=app.kubernetes.io/component=controller \
    --timeout=300s
  
  # Wait for the pod to be ready
  echo "Waiting for nginx ingress controller pod to be ready..."
  kubectl wait --namespace ingress-nginx \
    --for=condition=ready pod \
    --selector=app.kubernetes.io/component=controller \
    --timeout=90s
fi