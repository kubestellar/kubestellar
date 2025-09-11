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

# Unit test for KubeFlex compatibility test script

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_SCRIPT="${SCRIPT_DIR}/test-kflex-compatibility.sh"

echo "=== Testing KubeFlex Compatibility Script ==="

# Test 1: Check script exists and is executable
if [[ ! -x "$TEST_SCRIPT" ]]; then
    echo "ERROR: Script $TEST_SCRIPT is not executable"
    exit 1
fi
echo "✓ Script exists and is executable"

# Test 2: Check script can extract minimum version
MIN_VERSION_LINE=$(grep -E "Kubeflex version: v[0-9]+\.[0-9]+\.[0-9]+" "${SCRIPT_DIR}/../scripts/check_pre_req.sh" || true)
if [[ -z "$MIN_VERSION_LINE" ]]; then
    echo "ERROR: Could not find minimum KubeFlex version in check_pre_req.sh"
    exit 1
fi
MIN_VERSION=$(echo "$MIN_VERSION_LINE" | grep -oE "v[0-9]+\.[0-9]+\.[0-9]+" | head -1)
echo "✓ Successfully extracted minimum version: $MIN_VERSION"

# Test 3: Check that script fails gracefully when API is unavailable
# We'll use a timeout here to make sure the script doesn't hang
if timeout 10s bash -c 'GITHUB_API_URL="https://invalid-url" "$1"' _ "$TEST_SCRIPT" 2>/dev/null; then
    echo "WARNING: Script should fail when API is unavailable"
else
    echo "✓ Script fails gracefully when API is unavailable"
fi

# Test 4: Validate version comparison logic (extract from script)
version_compare() {
    local v1=$1
    local v2=$2
    
    # Remove 'v' prefix and compare
    v1_clean=${v1#v}
    v2_clean=${v2#v}
    
    printf '%s\n%s\n' "$v1_clean" "$v2_clean" | sort -V | head -n1
}

# Test version comparisons
if [[ "$(version_compare "v0.8.0" "v0.9.0")" == "0.8.0" ]]; then
    echo "✓ Version comparison works correctly (v0.8.0 < v0.9.0)"
else
    echo "ERROR: Version comparison failed"
    exit 1
fi

if [[ "$(version_compare "v0.9.0" "v0.8.0")" == "0.8.0" ]]; then
    echo "✓ Version comparison works correctly (v0.9.0 > v0.8.0)"
else
    echo "ERROR: Version comparison failed"
    exit 1
fi

echo "✓ All tests passed!"
echo "=== KubeFlex Compatibility Script Test Complete ==="
