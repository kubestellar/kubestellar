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

set -euo pipefail
set -x


CONTEXT="kind-kubeflex"     
NAMESPACE="argocd"

echo "⏳ Waiting for Argo CD pods to be ready..."
kubectl --context "$CONTEXT" wait --for=condition=Ready pods --all -n "$NAMESPACE" --timeout=120s

ARGOCD_POD=$(kubectl --context "$CONTEXT" -n "$NAMESPACE" get pod \
  -l app.kubernetes.io/name=argocd-server -o jsonpath='{.items[0].metadata.name}')

ARGOCD_PASSWORD=$(kubectl --context "$CONTEXT" -n "$NAMESPACE" get secret argocd-initial-admin-secret \
  -o jsonpath='{.data.password}' | base64 -d)

kubectl --context "$CONTEXT" -n "$NAMESPACE" exec "$ARGOCD_POD" -- \
  argocd login --username admin \
               --password "$ARGOCD_PASSWORD" \
               --insecure argocd-server."$NAMESPACE".svc.cluster.local

kubectl --context "$CONTEXT" -n "$NAMESPACE" exec "$ARGOCD_POD" -- \
  argocd cluster list

APP_NAME=smoke-guestbook
REPO_URL=https://github.com/argoproj/argocd-example-apps.git
APP_PATH=guestbook
DEST_SERVER=https://kubernetes.default.svc
DEST_NAMESPACE=default

kubectl --context "$CONTEXT" -n "$NAMESPACE" exec "$ARGOCD_POD" -- \
  argocd app create "$APP_NAME" \
    --repo "$REPO_URL" \
    --path "$APP_PATH" \
    --dest-server "$DEST_SERVER" \
    --dest-namespace "$DEST_NAMESPACE" \
    --directory-recurse \
    --sync-policy automated \
    --auto-prune

kubectl --context "$CONTEXT" -n "$NAMESPACE" exec "$ARGOCD_POD" -- \
  argocd app sync "$APP_NAME" --timeout 300

kubectl --context "$CONTEXT" -n "$NAMESPACE" exec "$ARGOCD_POD" -- \
  argocd app wait "$APP_NAME" --health --sync --timeout 300

echo "✅ SUCCESS: Argo CD application reconciled correctly!"

kubectl --context "$CONTEXT" -n "$NAMESPACE" exec "$ARGOCD_POD" -- \
  argocd app delete "$APP_NAME" --yes --cascade


echo "🧹 Cleanup complete. Argo CD smoke-test finished successfully."