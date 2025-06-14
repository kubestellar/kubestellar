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

# This script applies the necessary RBAC permissions to enable StorageClass management
# by the OCM klusterlet-work-sa service account in a WEC cluster.

set -e

# Default to no specific context
CONTEXT=""

function print_usage() {
  echo "Usage: $0 [--context CLUSTER_CONTEXT]"
  echo "Apply OCM storage permissions to enable StorageClass management in WECs"
  echo ""
  echo "Options:"
  echo "  --context CLUSTER_CONTEXT  Kubernetes context to apply permissions to"
  echo "  -h, --help                 Display this help message"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  key="$1"
  case $key in
    --context)
      CONTEXT="$2"
      shift
      shift
      ;;
    -h|--help)
      print_usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      print_usage
      exit 1
      ;;
  esac
done

CONTEXT_ARG=""
if [[ -n "${CONTEXT}" ]]; then
  CONTEXT_ARG="--context ${CONTEXT}"
  echo "Applying permissions to context: ${CONTEXT}"
fi

# Apply the ClusterRole and ClusterRoleBinding
kubectl ${CONTEXT_ARG} apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: klusterlet-work-sa-storage
rules:
- apiGroups: ["storage.k8s.io"]
  resources: ["storageclasses"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: klusterlet-work-sa-storage
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: klusterlet-work-sa-storage
subjects:
- kind: ServiceAccount
  name: klusterlet-work-sa
  namespace: open-cluster-management-agent
EOF

echo "Successfully applied StorageClass permissions for klusterlet-work-sa"
