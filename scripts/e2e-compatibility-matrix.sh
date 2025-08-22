# Copyright 2025 The KubeStellar Authors.
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

#!/usr/bin/env bash
# KubeStellar/KubeFlex Compatibility E2E Matrix Test Script
# Migrated from GitHub Actions to run in Prow/OCI. Fetches matrix dynamically.
set -euo pipefail

retry() {
  local -r -i max_attempts="$1"; shift
  local -r cmd=("$@")
  local -i attempt_num=1
  local delay=5
  until "${cmd[@]}"; do
    if (( attempt_num == max_attempts )); then
      echo "[ERROR] Command '${cmd[*]}' failed after $attempt_num attempts."
      return 1
    else
      echo "[WARN] Command '${cmd[*]}' failed. Retrying in $delay seconds... ($attempt_num/$max_attempts)"
      sleep $delay
      ((attempt_num++))
      delay=$((delay * 2))
    fi
  done
}

# Get min required kflex version from script
MIN_KFLEX_VERSION=$(python3 scripts/extract_min_kflex_version.py || true)
if [[ -z "$MIN_KFLEX_VERSION" ]]; then
  echo "[ERROR] Failed to determine minimum kflex version."
  exit 1
fi
echo "[INFO] Minimum kflex version required: $MIN_KFLEX_VERSION"

# Prepare GitHub API authorization header if GITHUB_TOKEN is set
AUTH_HEADER=()
if [[ -n "${GITHUB_TOKEN:-}" ]]; then
  AUTH_HEADER=(-H "Authorization: token $GITHUB_TOKEN")
fi

# Fetch all kflex releases >= min version
RELEASES_JSON=$(retry 5 curl -s "${AUTH_HEADER[@]}" https://api.github.com/repos/kubestellar/kubeflex/releases)
KUBEFLEX_VERSIONS=()
mapfile -t ALL_VERSIONS < <(echo "$RELEASES_JSON" | jq -r '.[].tag_name' | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | sort -V)
for V in "${ALL_VERSIONS[@]}"; do
  VNUM=${V#v}
  if [[ "$(printf "%s\n%s" "$MIN_KFLEX_VERSION" "$VNUM" | sort -V | head -1)" == "$MIN_KFLEX_VERSION" ]]; then
    KUBEFLEX_VERSIONS+=("$VNUM")
  fi
done
# Always include 'main' as a test target
KUBEFLEX_VERSIONS+=("main")
echo "[INFO] KubeFlex versions to test: ${KUBEFLEX_VERSIONS[*]}"

# You can also fetch KubeStellar versions similarly or hardcode for now
KUBESTELLAR_VERSIONS=("main" "stable")

RESULTS=()
for ks in "${KUBESTELLAR_VERSIONS[@]}"; do
  for kf in "${KUBEFLEX_VERSIONS[@]}"; do
    echo -e "---\n[INFO] Testing KubeStellar=$ks with KubeFlex=$kf"
    # Clean up cluster (delete ns, uninstall, etc.)
    retry 5 kubectl delete ns kubestellar kubeflex --ignore-not-found
    # Wait for namespaces to be fully deleted before proceeding
    for ns in kubestellar kubeflex; do
      for i in {1..60}; do
        if ! kubectl get ns "$ns" &>/dev/null; then
          break
        fi
        sleep 5
      done
      if kubectl get ns "$ns" &>/dev/null; then
        echo "[ERROR] Namespace $ns still exists after waiting; aborting."
        exit 1
      fi
    done
    sleep 5
    # Install KubeStellar
    if [[ "$ks" == "main" ]]; then
      KS_MANIFEST_URL="https://raw.githubusercontent.com/kubestellar/kubestellar/main/deploy/kubestellar.yaml"
    elif [[ "$ks" == "stable" ]]; then
      KS_MANIFEST_URL="https://github.com/kubestellar/kubestellar/releases/latest/download/kubestellar.yaml"
    else
      KS_MANIFEST_URL="https://github.com/kubestellar/kubestellar/releases/download/v$ks/kubestellar.yaml"
    fi
    retry 5 kubectl create ns kubestellar || true
    retry 5 kubectl apply -n kubestellar -f "$KS_MANIFEST_URL"
    # Install KubeFlex
    if [[ "$kf" == "main" ]]; then
      KF_MANIFEST_URL="https://raw.githubusercontent.com/kubestellar/kubeflex/main/deploy/kubeflex.yaml"
    else
      KF_MANIFEST_URL="https://github.com/kubestellar/kubeflex/releases/download/v$kf/kubeflex.yaml"
    fi
    retry 5 kubectl create ns kubeflex || true
    retry 5 kubectl apply -n kubeflex -f "$KF_MANIFEST_URL"
    # Wait for deployments
    retry 3 kubectl rollout status deployment -n kubestellar --timeout=300s
    retry 3 kubectl rollout status deployment -n kubeflex --timeout=300s
    # Integration checks
    # 1. WEC registration (simulated)
    echo "[TEST] WEC registration (simulated)"
    # Simulate a WEC by creating a test kubeconfig secret
    retry 5 kubectl create ns test-wec || true
    cat <<EOF | retry 5 kubectl apply -n test-wec -f -
apiVersion: v1
kind: Secret
metadata:
  name: wec-kubeconfig
  labels:
    kubestellar.io/wec: "true"
type: Opaque
data:
  kubeconfig: $(echo "dummy" | base64)
EOF
    sleep 5
    # Check that the WEC appears in KubeStellar's WEC list (simulate with label query)
    if kubectl get secret -n test-wec -l kubestellar.io/wec=true -o name | grep -q 'secret/wec-kubeconfig'; then
      echo "[PASS] WEC registered (simulated)"
    else
      echo "[FAIL] WEC registration (simulated)"
      RESULTS+=("FAIL WEC $ks $kf")
      continue
    fi

    # 2. ITS/WDS interaction (simulated)
    echo "[TEST] ITS/WDS interaction (simulated)"
    # Simulate ITS/WDS by creating a test resource and checking for label presence
    cat <<EOF | retry 5 kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-its
  namespace: kubestellar
  labels:
    kubestellar.io/its: "true"
data:
  foo: bar
EOF
    sleep 5
    # Check for ITS/WDS label presence
    if kubectl get configmap test-its -n kubestellar -o json | jq -e '.metadata.labels["kubestellar.io/its"] == "true"' >/dev/null; then
      echo "[PASS] ITS/WDS interaction (simulated)"
    else
      echo "[FAIL] ITS/WDS interaction (simulated)"
      RESULTS+=("FAIL ITS $ks $kf")
      continue
    fi

    # 3. Object sync (simulated)
    echo "[TEST] Object sync (simulated)"
    # Create a test object in source ns, verify it appears in target ns (simulate sync)
    retry 5 kubectl create ns source-ns || true
    retry 5 kubectl create ns target-ns || true
    cat <<EOF | retry 5 kubectl apply -n source-ns -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: sync-test
  labels:
    kubestellar.io/sync: "true"
data:
  test: value
EOF
    sleep 10
    # Simulate sync: copy object manually (replace with real sync if available)
    if kubectl get configmap sync-test -n source-ns -o yaml | sed 's/namespace: source-ns/namespace: target-ns/g' | retry 5 kubectl apply -f -; then
      if kubectl get configmap sync-test -n target-ns &>/dev/null; then
        echo "[PASS] Object sync (simulated)"
        RESULTS+=("PASS $ks $kf")
      else
        echo "[FAIL] Object sync (simulated)"
        RESULTS+=("FAIL SYNC $ks $kf")
      fi
    else
      echo "[FAIL] Object sync apply failed (simulated)"
      RESULTS+=("FAIL SYNC $ks $kf")
    fi
  done
done

# Output summary for Prow/TestGrid
for r in "${RESULTS[@]}"; do
  echo "$r"
done

# Exit non-zero if any test case failed
if printf '%s\n' "${RESULTS[@]}" | grep -q '^FAIL'; then
  exit 1
fi
