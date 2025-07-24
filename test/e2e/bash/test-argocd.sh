#!/usr/bin/env bash
# Copyright 2024 The KubeStellar Authors.
# Licensed under the Apache License, Version 2.0 (see LICENSE).

# -----------------------------------------------------------------------------
# Smoke-test: verify that Argo CD is fully operational (Helm-based install).
# The test will:
#   1. Ensure all Argo CD system pods are Running.
#   2. Log in to the Argo CD API using the CLI.
#   3. Create a disposable "guestbook" sample app.
#   4. Sync the app and wait until it is Synced + Healthy.
#   5. Clean up by deleting the app.
# -----------------------------------------------------------------------------

set -euo pipefail
set -x

CONTEXT="kind-argocd"       # 🔁 Replace with your kube context
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
  argocd app delete "$APP_NAME" --yes --cascade --timeout 120

echo "🧹 Cleanup complete. Argo CD smoke-test finished successfully."
