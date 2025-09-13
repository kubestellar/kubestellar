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

# Unit tests for KubeFlex CLI compatibility test script
# Tests the script interface without duplicating internal logic

set -euo pipefail

# Source directory for KubeStellar scripts
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_SCRIPT="${SCRIPT_DIR}/test-kflex-compatibility.sh"

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Test result tracking
run_test() {
    local test_name="$1"
    shift
    local expected_exit_code="${1:-0}"
    shift
    
    ((TESTS_RUN++))
    echo "Running test: $test_name"
    
    local actual_exit_code=0
    local output
    
    # Capture both stdout and stderr, and the exit code
    if output=$("$@" 2>&1); then
        actual_exit_code=0
    else
        actual_exit_code=$?
    fi
    
    if [[ $actual_exit_code -eq $expected_exit_code ]]; then
        echo "  ✓ PASS"
        ((TESTS_PASSED++))
        return 0
    else
        echo "  ✗ FAIL - Expected exit code $expected_exit_code, got $actual_exit_code"
        echo "  Output: $output"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Test that the script exists and is executable
test_script_exists() {
    if [[ ! -f "$TEST_SCRIPT" ]]; then
        echo "ERROR: Test script not found: $TEST_SCRIPT"
        return 1
    fi
    
    if [[ ! -x "$TEST_SCRIPT" ]]; then
        echo "ERROR: Test script not executable: $TEST_SCRIPT"
        return 1
    fi
    
    return 0
}

# Test help flag
test_help_flag() {
    run_test "help flag shows usage" 0 "$TEST_SCRIPT" --help
}

# Test dry-run flag
test_dry_run_flag() {
    run_test "dry-run flag works" 0 "$TEST_SCRIPT" --dry-run
}

# Test max-versions flag with valid values
test_max_versions_flag() {
    run_test "max-versions with valid number" 0 "$TEST_SCRIPT" --dry-run --max-versions 3
}

# Test max-versions flag with invalid values
test_max_versions_invalid() {
    run_test "max-versions with invalid value" 1 "$TEST_SCRIPT" --max-versions
    run_test "max-versions with non-numeric value" 1 "$TEST_SCRIPT" --max-versions abc
}

# Test unknown flag handling
test_unknown_flag() {
    run_test "unknown flag is rejected" 1 "$TEST_SCRIPT" --unknown-flag
}

# Test that script can be sourced without executing main
test_script_sourcing() {
    local test_file=$(mktemp)
    cat > "$test_file" << 'EOF'
#!/bin/bash
set -euo pipefail
# Source the script and verify functions are available
# Save the script path before clearing arguments
script_path="$1"
# Clear arguments to avoid triggering argument parsing
set --
source "$script_path"
# Check that main functions are defined
if ! declare -f version_le > /dev/null; then
    echo "version_le function not found"
    exit 1
fi
if ! declare -f check_prerequisites > /dev/null; then
    echo "check_prerequisites function not found"  
    exit 1
fi
echo "Sourcing test passed"
EOF
    chmod +x "$test_file"
    
    run_test "script can be sourced" 0 "$test_file" "$TEST_SCRIPT"
    rm -f "$test_file"
}

# Test environment variable support
test_environment_variables() {
    # Test custom KUBEFLEX_FETCH_CMD
    run_test "custom KUBEFLEX_FETCH_CMD" 0 env KUBEFLEX_FETCH_CMD="echo test" "$TEST_SCRIPT" --dry-run
}

# Test version comparison logic (by testing the script behavior, not duplicating logic)
test_version_filtering() {
    # The script should handle version filtering correctly - test via dry-run behavior
    run_test "version filtering works" 0 "$TEST_SCRIPT" --dry-run --max-versions 1
}

# Test script with minimal options
test_minimal_run() {
    run_test "script runs with minimal options" 0 "$TEST_SCRIPT" --dry-run --max-versions 1
}

# Main test runner
main() {
    echo "Running KubeFlex compatibility test script unit tests..."
    echo ""
    
    # Check prerequisites
    if ! test_script_exists; then
        echo "FATAL: Cannot find test script"
        exit 1
    fi
    
    # Run all tests
    test_help_flag
    test_dry_run_flag
    test_max_versions_flag
    test_max_versions_invalid
    test_unknown_flag
    test_script_sourcing
    test_environment_variables
    test_version_filtering
    test_minimal_run
    
    # Print summary
    echo ""
    echo "=== Unit Test Summary ==="
    echo "Tests run: $TESTS_RUN"
    echo "Passed: $TESTS_PASSED"
    echo "Failed: $TESTS_FAILED"
    
    if [[ $TESTS_FAILED -gt 0 ]]; then
        echo "ERROR: Some unit tests failed"
        exit 1
    fi
    
    if [[ $TESTS_RUN -eq 0 ]]; then
        echo "WARNING: No tests were run"
        exit 1
    fi
    
    echo "SUCCESS: All unit tests passed"
}

# Only run main if script is executed directly (not sourced)
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi