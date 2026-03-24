#!/bin/bash
set -e
env="${1:-kind}"
fail_flag=""

if [[ "$2" == "--fail-fast" ]]; then
    fail_flag="--fail-fast"
fi

SRC_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SRC_DIR/.." && pwd)"
cd "$REPO_ROOT"

echo "Running e2e singleton status propagation test..."
echo "Environment: $env"
echo ""

test/e2e/run-test.sh --env "$env" --test-type ginkgo $fail_flag
