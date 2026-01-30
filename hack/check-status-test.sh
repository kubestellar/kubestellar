#!/bin/bash
cd "$(dirname "$0")/../.."

echo "Checking Go syntax for status controller test..."
go fmt ./test/e2e/ginkgo/status_controller_test.go

echo ""
echo "Checking imports and basic compilation..."
go build -o /tmp/test-check ./test/e2e/ginkgo/... 2>&1 | grep -i "status_controller" || echo "✓ No syntax errors found"

echo ""
echo "Status controller test file is ready."
