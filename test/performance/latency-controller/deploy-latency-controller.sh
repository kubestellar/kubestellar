#!/bin/bash

# Copyright 2023 The KubeStellar Authors.
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

set -euo pipefail

function usage() {
  cat <<EOF
Usage: $0 \
  --latency_controller_image <container image for latency controller> \
  --binding-policy-name           <cluster binding name used by controller> \
  --monitored-namespace     <namespace the controller will monitor> \
  --host-context            <kubeconfig context of host/WDS cluster> \
  --wds-context             <ControlPlane context name for WDS extraction> \
  --its-context             <ControlPlane context name for ITS extraction> \
  [--wds-incluster-file]    <optional path to wds in-cluster kubeconfig> \
  [--its-incluster-file]    <optional path to its in-cluster kubeconfig> \
  --wec-files              <comma-separated pairs name:path (e.g. cluster1:./c1.kubeconfig)> \
  [--image-pull-policy <Always|IfNotPresent|...>] \
  [--controller-verbosity <number>]

Notes:
  - The script will deploy resources into namespace: <auto-detected>-system
  - THE SCRIPT WILL NOT CREATE the namespace. It will fail if the namespace does not exist.
  - WDS and ITS names are now automatically extracted from their respective kubeconfig files
  - If --wds-incluster-file or --its-incluster-file are omitted, the script will extract
    the in-cluster kubeconfigs from ControlPlane resources and apply them as secrets
  - --wec-files is required and will be applied as secrets.
EOF
  exit 1
}

# --- Parse args ---
LATENCY_IMAGE="" BINDING_POLICY_NAME="" MONITORED_NS=""
HOST_CTX="" WDS_CONTEXT="" ITS_CONTEXT=""
WDS_IN_FILE="" ITS_IN_FILE=""
WEC_FILES=""
IMAGE_PULL_POLICY="" CONTROLLER_VERBOSITY=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --latency_controller_image) LATENCY_IMAGE="$2"; shift;;
    --binding-policy-name)      BINDING_POLICY_NAME="$2"; shift;;
    --monitored-namespace)      MONITORED_NS="$2"; shift;;
    --host-context)             HOST_CTX="$2"; shift;;
    --wds-context)              WDS_CONTEXT="$2"; shift;;
    --its-context)              ITS_CONTEXT="$2"; shift;;
    --wds-incluster-file)       WDS_IN_FILE="$2"; shift;;
    --its-incluster-file)       ITS_IN_FILE="$2"; shift;;
    --wec-files)                WEC_FILES="$2"; shift;;
    --image-pull-policy)        IMAGE_PULL_POLICY="$2"; shift;;
    --controller-verbosity)     CONTROLLER_VERBOSITY="$2"; shift;;
    -h|--help) usage;;
    *) echo "ERROR: Unknown arg: $1"; usage;;
  esac
  shift
done

# Required checks
if [[ -z "$LATENCY_IMAGE" || -z "$BINDING_POLICY_NAME" || -z "$MONITORED_NS" || -z "$HOST_CTX" || -z "$WDS_CONTEXT" || -z "$ITS_CONTEXT" || -z "$WEC_FILES" ]]; then
  echo "ERROR: Missing required argument."
  usage
fi

# Switch to host cluster (needed for extracting ControlPlane secrets)
echo "INFO: switching kubectl context to ${HOST_CTX}"
kubectl config use-context "$HOST_CTX"

# Helper function to extract context name from kubeconfig
function extract_context_name_from_kubeconfig() {
  local kubeconfig_file="$1"
  if [[ ! -f "$kubeconfig_file" ]]; then
    echo "ERROR: Kubeconfig file not found: $kubeconfig_file" >&2
    exit 2
  fi
  local context_name
  context_name=$(kubectl --kubeconfig="$kubeconfig_file" config current-context 2>/dev/null || true)
  if [[ -z "$context_name" ]]; then
    context_name=$(kubectl --kubeconfig="$kubeconfig_file" config get-contexts -o name | head -n1 || true)
  fi
  if [[ -z "$context_name" ]]; then
    echo "ERROR: Could not extract context name from kubeconfig: $kubeconfig_file" >&2
    exit 2
  fi
  echo "$context_name"
}

# Helper: fetch kubeconfig from ControlPlane secret
function get_cp_kubeconfig() {
  local cpname="$1" outfile="$2"
  local key secret_name secret_namespace
  key=$(kubectl get controlplane "$cpname" -o=jsonpath='{.status.secretRef.inClusterKey}')
  secret_name=$(kubectl get controlplane "$cpname" -o=jsonpath='{.status.secretRef.name}')
  secret_namespace=$(kubectl get controlplane "$cpname" -o=jsonpath='{.status.secretRef.namespace}')
  kubectl -n "$secret_namespace" get secret "$secret_name" -o=jsonpath="{.data.$key}" | base64 -d > "$outfile"
}

# Helper function to extract context name from ControlPlane
function extract_context_name_from_controlplane() {
  local cpname="$1"
  echo "INFO: extracting context name for ControlPlane '$cpname'..." >&2
  if ! kubectl get controlplane "$cpname" >/dev/null 2>&1; then
    echo "ERROR: ControlPlane '$cpname' not found in current context ($HOST_CTX)." >&2
    exit 5
  fi
  local tmpf context_name
  tmpf=$(mktemp /tmp/${cpname}.kubeconfig.XXXX)
  trap '[[ -f "$tmpf" ]] && rm -f "$tmpf"' RETURN
  get_cp_kubeconfig "$cpname" "$tmpf"
  context_name=$(extract_context_name_from_kubeconfig "$tmpf")
  rm -f "$tmpf" || true
  trap - RETURN
  echo "$context_name"
}

# Auto-detect WDS_NAME and ITS_NAME
if [[ -n "${WDS_IN_FILE:-}" ]]; then
  WDS_NAME=$(extract_context_name_from_kubeconfig "$WDS_IN_FILE")
else
  WDS_NAME=$(extract_context_name_from_controlplane "$WDS_CONTEXT")
fi
if [[ -n "${ITS_IN_FILE:-}" ]]; then
  ITS_NAME=$(extract_context_name_from_kubeconfig "$ITS_IN_FILE")
else
  ITS_NAME=$(extract_context_name_from_controlplane "$ITS_CONTEXT")
fi
echo "INFO: auto-detected WDS name: $WDS_NAME"
echo "INFO: auto-detected ITS name: $ITS_NAME"

# Namespace check (do not create)
NAMESPACE="${WDS_NAME}-system"
if ! kubectl get ns "$NAMESPACE" >/dev/null 2>&1; then
  echo "ERROR: Namespace '$NAMESPACE' does not exist on context '$HOST_CTX'."
  exit 3
fi

# Secret helpers
function apply_secret() {
  local name="$1" file="$2"
  kubectl -n "$NAMESPACE" create secret generic "$name" \
    --from-file=kubeconfig="$file" \
    --dry-run=client -o yaml | kubectl apply -f -
}

function extract_cp_incluster_secret() {
  local cpname="$1" target_secret="$2"
  local tmpf
  tmpf=$(mktemp /tmp/${target_secret}.kubeconfig.XXXX)
  trap '[[ -f "$tmpf" ]] && rm -f "$tmpf"' RETURN
  get_cp_kubeconfig "$cpname" "$tmpf"
  apply_secret "$target_secret" "$tmpf"
  rm -f "$tmpf" || true
  trap - RETURN
}

# Apply in-cluster secrets
if [[ -n "${WDS_IN_FILE:-}" ]]; then
  apply_secret "${WDS_NAME}-incluster" "$WDS_IN_FILE"
else
  extract_cp_incluster_secret "$WDS_CONTEXT" "${WDS_NAME}-incluster"
fi
if [[ -n "${ITS_IN_FILE:-}" ]]; then
  apply_secret "${ITS_NAME}-incluster" "$ITS_IN_FILE"
else
  extract_cp_incluster_secret "$ITS_CONTEXT" "${ITS_NAME}-incluster"
fi

# Apply WEC secrets
declare -a WEC_NAMES=()
IFS=',' read -ra pairs <<< "$WEC_FILES"
for pair in "${pairs[@]}"; do
  name="${pair%%:*}"
  file="${pair#*:}"
  apply_secret "wec-${name}-incluster" "$file"
  WEC_NAMES+=("$name")
done

# Build volumes
VOLUME_SOURCES=""
append_secret_yaml() {
  local secret_name="$1" file_name="$2"
  VOLUME_SOURCES="${VOLUME_SOURCES}
              - secret:
                  name: ${secret_name}
                  items:
                    - key: kubeconfig
                      path: ${file_name}"
}
append_secret_yaml "${WDS_NAME}-incluster" "${WDS_NAME}-incluster.kubeconfig"
append_secret_yaml "${ITS_NAME}-incluster" "${ITS_NAME}-incluster.kubeconfig"
for n in "${WEC_NAMES[@]}"; do
  append_secret_yaml "wec-${n}-incluster" "${n}-incluster.kubeconfig"
done

WECS_CONTEXTS_CSV="$(IFS=,; echo "${WEC_NAMES[*]}")"
declare -a kube_paths=()
for n in "${WEC_NAMES[@]}"; do
  kube_paths+=("/etc/kubeconfigs/${n}-incluster.kubeconfig")
done
WECS_KUBECONFIGS_CSV="$(IFS=,; echo "${kube_paths[*]}")"

# Deploy latency-collector only
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  name: latency-controller-sa
  namespace: ${NAMESPACE}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: latency-controller-cr
rules:
  - apiGroups: ["*"]
    resources: ["*"]
    verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: latency-controller-crb
subjects:
  - kind: ServiceAccount
    name: latency-controller-sa
    namespace: ${NAMESPACE}
roleRef:
  kind: ClusterRole
  name: latency-controller-cr
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: latency-controller
  namespace: ${NAMESPACE}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: latency-controller
  template:
    metadata:
      labels:
        app: latency-controller
    spec:
      serviceAccountName: latency-controller-sa
      containers:
      - name: latency-controller
        image: ${LATENCY_IMAGE}
        args:
          - "--wds-context=${WDS_NAME}"
          - "--its-context=${ITS_NAME}"
$( if [[ -n "${WECS_CONTEXTS_CSV}" ]]; then echo "          - \"--wec-contexts=${WECS_CONTEXTS_CSV}\""; fi )
          - "--wds-kubeconfig=/etc/kubeconfigs/${WDS_NAME}-incluster.kubeconfig"
          - "--its-kubeconfig=/etc/kubeconfigs/${ITS_NAME}-incluster.kubeconfig"
$( if [[ -n "${WECS_KUBECONFIGS_CSV}" ]]; then echo "          - \"--wec-kubeconfigs=${WECS_KUBECONFIGS_CSV}\""; fi )
          - "--binding-policy-name=${BINDING_POLICY_NAME}"
          - "--monitored-namespace=${MONITORED_NS}"
        env:
          - name: KUBECONFIG
            value: ""
        volumeMounts:
          - name: kubeconfig-volume
            mountPath: /etc/kubeconfigs
      volumes:
        - name: kubeconfig-volume
          projected:
            sources:
${VOLUME_SOURCES}
EOF

echo "INFO: Latency controller Deployment and RBAC applied into namespace ${NAMESPACE}"
echo "Run: kubectl get pods -n ${NAMESPACE} to check controller pods."
