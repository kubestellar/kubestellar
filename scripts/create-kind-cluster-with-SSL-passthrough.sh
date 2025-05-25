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
    wait ${prefetch_pidss[i]}
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
kubectl --context "kind-${name}" patch deployment ingress-nginx-controller -n ingress-nginx -p '{"spec":{"template":{"spec":{"containers":[{"name":"controller","args":["/nginx-ingress-controller","--election-id=ingress-nginx-leader","--controller-class=k8s.io/ingress-nginx","--ingress-class=nginx","--configmap=$(POD_NAMESPACE)/ingress-nginx-controller","--validating-webhook=:8443","--validating-webhook-certificate=/usr/local/certificates/cert","--validating-webhook-key=/usr/local/certificates/key","--watch-ingress-without-class=true","--publish-status-address=localhost","--enable-ssl-passthrough"]}]}}}}'

if [[ "$wait" == "true" ]] ; then
  echo "Waiting for nginx ingress with SSL passthrough to be ready..."
  while true; do
      sleep 5
      pods=$(kubectl --context kind-${name} get pod -n ingress-nginx -l app.kubernetes.io/component=controller -o jsonpath='{.items[*].metadata.name}')
      if [ -z "$pods" ]; then continue; fi
      if [[ "$pods" =~ [[:space:]] ]]
      then # both pre-patch and post-patch Pods are present
          continue
      fi
      args=$(kubectl --context kind-${name} get pod -n ingress-nginx -l app.kubernetes.io/component=controller -o jsonpath='{.items[0].spec.containers[0].args}')
      if [[ $args =~ enable-ssl-passthrough ]]
      then break
      # Otherwise this Pod is from before the patch
      fi
  done
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
