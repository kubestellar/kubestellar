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

name=kubestellar
port=9443
wait=true
setcontext=true

display_help() {
  cat << EOF
Usage: $0 [--name name] [--port port] [--nowait] [--nosetcontext] [-X]

--name name     set a specific name of the kind cluster (default: kubestellar)
--port port     map the specified host port to the kind cluster port 443 (default: 9443)
--nowait        when set to true, the script proceeds without waiting for the nginx ingress patching to complete
--nosetcontext  when set to true, the script does not change the current kubectl context to the newly created cluster
-X              enable verbose execution of the script for debugging
EOF
}

while (( $# > 0 )); do
  case "$1" in
  (-n|--name)
    if (( $# > 1 ));
    then { name="$2"; shift; }
    else { echo "$0: missing name for the kind cluster" >&2; exit 1; }
    fi;;
  (-p|--port)
    if (( $# > 1 ));
    then { port="$2"; shift; }
    else { echo "$0: missing port number" >&2; exit 1; }
    fi;;
  (--nowait)
    wait=false;;
  (--nosetcontext)
    setcontext=false;;
  (-X)
    set -x;;
  (-h|--help)
    display_help
    exit 0;;
  (-*)
    echo "$0: unknown flag" >&2
    exit 1;;
  (*)
    echo "$0: unknown positional argument" >&2
    exit 1;;
  esac
  shift
done

echo "Creating \"${name}\" kind cluster with SSL passthrougn and ${port} port mapping..."
if [[ "$(kind get clusters | grep "^${name}$")" == "" ]] ; then
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

echo "Installing an nginx ingress..."
kubectl --context "kind-${name}" apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

echo "Pathcing nginx ingress to enable SSL passthrough..."
kubectl --context "kind-${name}" patch deployment ingress-nginx-controller -n ingress-nginx -p '{"spec":{"template":{"spec":{"containers":[{"name":"controller","args":["/nginx-ingress-controller","--election-id=ingress-nginx-leader","--controller-class=k8s.io/ingress-nginx","--ingress-class=nginx","--configmap=$(POD_NAMESPACE)/ingress-nginx-controller","--validating-webhook=:8443","--validating-webhook-certificate=/usr/local/certificates/cert","--validating-webhook-key=/usr/local/certificates/key","--watch-ingress-without-class=true","--publish-status-address=localhost","--enable-ssl-passthrough"]}]}}}}'

if [[ "$wait" == "true" ]] ; then
  echo "Waiting for nginx ingress with SSL passthrough to be ready..."
  while [ -z "$(kubectl --context kind-${name} get pod --namespace ingress-nginx --selector=app.kubernetes.io/component=controller --no-headers -o name 2> /dev/null)" ] ;  do
      sleep 5
  done
  kubectl --context "kind-${name}" wait --namespace ingress-nginx \
    --for=condition=ready pod \
    --selector=app.kubernetes.io/component=controller \
    --timeout=90s
fi

if [[ "$setcontext" == "true" ]] ; then
  echo "Setting context to \"kind-${name}\"..."
  kubectl config use-context "kind-${name}"
fi
