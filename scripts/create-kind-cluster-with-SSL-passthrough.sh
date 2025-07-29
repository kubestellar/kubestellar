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

set -o errexit

# Script info
SCRIPT_NAME="create-kind-cluster-with-SSL-passthrough.sh"

NGINX_INGRESS_URL="https://raw.githubusercontent.com/kubernetes/ingress-nginx/refs/tags/controller-v1.12.1/deploy/static/provider/kind/deploy.yaml"
name=kubestellar
port=9443
wait=true
prefetch=false
setcontext=true


display_help() {
  cat << EOF
Usage: ${SCRIPT_NAME} [options]

--name name     set a specific name of the kind cluster (default: kubestellar)
--port port     map the specified host port to the kind cluster port 443 (default: 9443)
--nowait        when set to true, the script proceeds without waiting for the nginx ingress patching to complete
--prefetch      prefetch the nginx ingress images for the kind cluster while the kind cluster is being created
--nosetcontext  when set to true, the script does not change the current kubectl context to the newly created cluster
-X              enable verbose execution of the script for debugging
EOF
}

while (( $# > 0 )); do
  case "$1" in
  (-n|--name)
    if (( $# > 1 ));
    then { name="$2"; shift; }
    else { echo "${SCRIPT_NAME}: missing name for the kind cluster" >&2; exit 1; }
    fi;;
  (-p|--port)
    if (( $# > 1 ));
    then { port="$2"; shift; }
    else { echo "$: missing port number" >&2; exit 1; }
    fi;;
  (--nowait)
    wait=false;;
  (--prefetch)
    prefetch=true;;
  (--nosetcontext)
    setcontext=false;;
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


###############################################################################
# Prefetch the images required by nginx ingress
###############################################################################
prefetch_images=()
prefetch_pids=()
if [[ "$prefetch" == "true" ]] ; then
  nginx_ingress="$(curl --no-progress-meter "$NGINX_INGRESS_URL" | sed 's/@sha256.*$//')"
  IFS=' ' read -r -a prefetch_images <<< "$(echo "$nginx_ingress" | grep 'image: ' | awk '{print $2}' | uniq | tr '\n' ' ')"
  for image in "${prefetch_images[@]}" ; do
    echo -n "Prefetching image: \"${image}\"... "
    docker pull "${image}" -q &
    prefetch_pids+=("$!")
    echo "pid=${prefetch_pids[${#prefetch_pids[@]}-1]}"
  done
fi


###############################################################################
# Create the kind cluster
###############################################################################
echo "Creating \"${name}\" kind cluster with SSL passthrougn and ${port} port mapping..."
if [[ -z "$(kind get clusters | grep "^${name}$")" ]] ; then
  kind create cluster --name "${name}" --config - <<EOF
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
    hostPort: ${port}
    protocol: TCP
EOF
else
  echo "Skipping... \"${name}\" kind cluster already exists."
fi


###############################################################################
# Waiting for prefetching to complete... load images into cluster
###############################################################################
if [[ "$prefetch" == "true" ]] ; then
  for i in "${!prefetch_pids[@]}" ; do
    echo -n "Waiting for prefetch process with pid=${prefetch_pids[i]} to complete... "
    wait ${prefetch_pids[i]}
    exitcode="$?"
    echo  "exitcode=$exitcode"
    if [[ "$exitcode" != "0" ]] ; then
      >&2 echo "ERROR: prefetch process failed!"
    fi
    echo "Loading image \"${prefetch_images[i]}\"to the cluster..."
    kind load docker-image "${prefetch_images[i]}" --name "$name"
  done
fi


###############################################################################
# Installing nginx ingress
###############################################################################
echo "Installing an nginx ingress..."
if [[ "$prefetch" == "true" ]] ; then
  echo "$nginx_ingress" | kubectl --context "kind-${name}" apply -f -
else
  kubectl --context "kind-${name}" apply -f "$NGINX_INGRESS_URL"
fi


echo "Patching nginx ingress to enable SSL passthrough..."
# Base SSL passthrough patch (original functionality)
kubectl --context "kind-${name}" patch deployment ingress-nginx-controller -n ingress-nginx -p '{"spec":{"template":{"spec":{"containers":[{"name":"controller","args":["/nginx-ingress-controller","--election-id=ingress-nginx-leader","--controller-class=k8s.io/ingress-nginx","--ingress-class=nginx","--configmap=$(POD_NAMESPACE)/ingress-nginx-controller","--validating-webhook=:8443","--validating-webhook-certificate=/usr/local/certificates/cert","--validating-webhook-key=/usr/local/certificates/key","--watch-ingress-without-class=true","--publish-status-address=localhost","--enable-ssl-passthrough"]}]}}}}'

echo "Applying demo optimizations to nginx ingress..."
# Add valid demo optimization flags in separate patches for reliability
# Note: Removed invalid flags that were causing crashes

# Disable admission webhooks for faster startup
kubectl --context "kind-${name}" patch validatingwebhookconfiguration ingress-nginx-admission -p '{"webhooks":[{"name":"validate.nginx.ingress.kubernetes.io","admissionReviewVersions":["v1","v1beta1"],"clientConfig":{"service":{"name":"ingress-nginx-controller-admission","namespace":"ingress-nginx","path":"/networking/v1/ingresses"}},"rules":[{"operations":["CREATE","UPDATE"],"apiGroups":["networking.k8s.io"],"apiVersions":["v1"],"resources":["ingresses"]}],"failurePolicy":"Ignore","sideEffects":"None"}]}'

# Add resource limits for faster scheduling
kubectl --context "kind-${name}" patch deployment ingress-nginx-controller -n ingress-nginx -p '{"spec":{"template":{"spec":{"containers":[{"name":"controller","resources":{"requests":{"cpu":"100m","memory":"90Mi"},"limits":{"cpu":"1000m","memory":"500Mi"}}}]}}}}'

# Optimize for demo environment by disabling unnecessary features via ConfigMap
kubectl --context "kind-${name}" patch configmap ingress-nginx-controller -n ingress-nginx -p '{"data":{"enable-real-ip":"false","proxy-protocol":"false","server-tokens":"false","ssl-protocols":"TLSv1.2 TLSv1.3"}}'

if [[ "$wait" == "true" ]] ; then
  echo "Waiting for nginx ingress with SSL passthrough to be ready..."
  
  # First, wait for deployment rollout (from upstream)
  echo "Waiting for nginx ingress controller deployment to be rolled out..."
  kubectl --context "kind-${name}" rollout status deployment/ingress-nginx-controller \
    --namespace ingress-nginx \
    --timeout=300s
  
  echo "Waiting for nginx ingress controller pod to be ready..."
  
  # More efficient wait logic with shorter intervals and better detection
  max_attempts=150  # 150 * 2 seconds = 5 minutes for progress feedback
  attempt=0
  
  while [ $attempt -lt $max_attempts ]; do
      sleep 2
      attempt=$((attempt + 1))
      
      # Check if controller pod exists and is ready
      pods=$(kubectl --context kind-${name} get pod -n ingress-nginx -l app.kubernetes.io/component=controller -o jsonpath='{.items[*].metadata.name}' 2>/dev/null)
      if [ -z "$pods" ]; then 
          echo "Waiting for nginx controller pod to be created... (attempt $attempt/$max_attempts)"
          continue
      fi
      
      # If multiple pods exist, we're in the middle of a rolling update
      if [[ "$pods" =~ [[:space:]] ]]; then
          echo "Rolling update in progress... (attempt $attempt/$max_attempts)"
          continue
      fi
      
      # Check if the pod has the SSL passthrough argument
      args=$(kubectl --context kind-${name} get pod -n ingress-nginx -l app.kubernetes.io/component=controller -o jsonpath='{.items[0].spec.containers[0].args}' 2>/dev/null)
      if [[ ! $args =~ enable-ssl-passthrough ]]; then
          echo "Waiting for SSL passthrough configuration... (attempt $attempt/$max_attempts)"
          continue
      fi
      
      # Check if pod is actually ready
      if kubectl --context kind-${name} get pod -n ingress-nginx -l app.kubernetes.io/component=controller -o jsonpath='{.items[0].status.conditions[?(@.type=="Ready")].status}' 2>/dev/null | grep -q "True"; then
          echo "Nginx ingress controller is ready with SSL passthrough enabled"
          break
      else
          echo "Pod exists but not ready yet... (attempt $attempt/$max_attempts)"
      fi
  done
  
  if [ $attempt -eq $max_attempts ]; then
      echo "Warning: Maximum wait time reached. Nginx may not be fully ready."
  fi
  
  # Final wait with timeout as backup
  kubectl --context "kind-${name}" wait --namespace ingress-nginx \
    --for=condition=ready pod \
    --selector=app.kubernetes.io/component=controller \
    --timeout=24h
fi


###############################################################################
# Setting context
###############################################################################
if [[ "$setcontext" == "true" ]] ; then
  echo "Setting context to \"kind-${name}\"..."
  kubectl config use-context "kind-${name}"
fi
