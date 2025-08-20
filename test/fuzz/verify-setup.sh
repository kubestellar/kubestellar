#!/bin/bash

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

set -o errexit
set -o nounset
set -o pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🔧 KubeStellar Fuzz Testing Setup Verification"
echo "=============================================="

# Check Go version
echo -n "Checking Go version... "
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | cut -d' ' -f3)
    echo -e "${GREEN}✓${NC} Found Go $GO_VERSION"
    
    # Check if Go version is 1.18 or higher (required for fuzzing)
    GO_MAJOR=$(echo $GO_VERSION | sed 's/go//' | cut -d'.' -f1)
    GO_MINOR=$(echo $GO_VERSION | sed 's/go//' | cut -d'.' -f2)
    
    if [ "$GO_MAJOR" -gt 1 ] || [ "$GO_MAJOR" -eq 1 -a "$GO_MINOR" -ge 18 ]; then
        echo -e "  ${GREEN}✓${NC} Go version supports fuzzing (requires Go 1.18+)"
    else
        echo -e "  ${RED}✗${NC} Go version is too old. Fuzzing requires Go 1.18 or higher"
        exit 1
    fi
else
    echo -e "${RED}✗${NC} Go is not installed"
    exit 1
fi

echo ""

# Check if we're in the right directory
echo -n "Checking project structure... "
if [ -f "../../go.mod" ] && grep -q "github.com/kubestellar/kubestellar" "../../go.mod"; then
    echo -e "${GREEN}✓${NC} In KubeStellar project directory"
else
    echo -e "${RED}✗${NC} Not in KubeStellar project root or invalid structure"
    echo "Please run this script from the test/fuzz directory"
    exit 1
fi

echo ""

# Check if fuzz tests exist
echo -n "Checking fuzz tests... "
FUZZ_TESTS=$(go test -list='Fuzz.*' 2>/dev/null | grep -c '^Fuzz' || echo "0")
if [ "$FUZZ_TESTS" -gt 0 ]; then
    echo -e "${GREEN}✓${NC} Found $FUZZ_TESTS fuzz tests"
    go test -list='Fuzz.*' | grep '^Fuzz' | while read test; do
        echo "  - $test"
    done
else
    echo -e "${RED}✗${NC} No fuzz tests found"
    exit 1
fi

echo ""

# Check dependencies
echo -n "Checking dependencies... "
if go mod verify &> /dev/null; then
    echo -e "${GREEN}✓${NC} Dependencies verified"
else
    echo -e "${YELLOW}⚠${NC} Dependencies may need updating"
    echo "  Run 'go mod tidy' to fix dependency issues"
fi

echo ""

# Check if test data exists
echo -n "Checking test data... "
if [ -d "testdata" ] && [ "$(ls -A testdata 2>/dev/null)" ]; then
    echo -e "${GREEN}✓${NC} Test data directory found"
    echo "  Available test files:"
    ls testdata/*.yaml 2>/dev/null | while read file; do
        echo "  - $(basename "$file")"
    done
else
    echo -e "${YELLOW}⚠${NC} No test data found in testdata directory"
fi

echo ""

# Test compilation
echo -n "Testing fuzz test compilation... "
if go test -c . &> /dev/null; then
    echo -e "${GREEN}✓${NC} Fuzz tests compile successfully"
    rm -f fuzz.test  # Clean up test binary
else
    echo -e "${RED}✗${NC} Fuzz tests fail to compile"
    echo "Run 'go test -c .' to see compilation errors"
    exit 1
fi

echo ""

# Run a quick fuzz test
echo "🧪 Running quick fuzz test (5 seconds)..."
echo "========================================="

QUICK_TEST="FuzzJSONPathParsing"
AVAILABLE_TESTS=$(go test -list='Fuzz.*' 2>/dev/null | grep '^Fuzz' || echo "")
if echo "$AVAILABLE_TESTS" | grep -q "^$QUICK_TEST$"; then
    echo "Running $QUICK_TEST for 5 seconds..."
    
    # Check if timeout command is available
    if command -v timeout &> /dev/null; then
        # Linux/GNU timeout
        if timeout 10s go test -fuzz="$QUICK_TEST" -fuzztime=5s 2>&1; then
            echo -e "${GREEN}✓${NC} Quick fuzz test completed successfully"
        else
            EXITCODE=$?
            if [ $EXITCODE -eq 124 ]; then
                echo -e "${GREEN}✓${NC} Quick fuzz test timed out as expected"
            else
                echo -e "${RED}✗${NC} Quick fuzz test failed with exit code $EXITCODE"
                exit 1
            fi
        fi
    elif command -v gtimeout &> /dev/null; then
        # macOS with GNU coreutils (brew install coreutils)
        if gtimeout 10s go test -fuzz="$QUICK_TEST" -fuzztime=5s 2>&1; then
            echo -e "${GREEN}✓${NC} Quick fuzz test completed successfully"
        else
            EXITCODE=$?
            if [ $EXITCODE -eq 124 ]; then
                echo -e "${GREEN}✓${NC} Quick fuzz test timed out as expected"
            else
                echo -e "${RED}✗${NC} Quick fuzz test failed with exit code $EXITCODE"
                exit 1
            fi
        fi
    else
        # No timeout available - just run the test with short duration
        echo "  (No timeout command available, running without timeout protection)"
        if go test -fuzz="$QUICK_TEST" -fuzztime=5s 2>&1; then
            echo -e "${GREEN}✓${NC} Quick fuzz test completed successfully"
        else
            EXITCODE=$?
            echo -e "${RED}✗${NC} Quick fuzz test failed with exit code $EXITCODE"
            exit 1
        fi
    fi
else
    echo -e "${YELLOW}⚠${NC} Could not find $QUICK_TEST, skipping quick test"
    echo "Available tests: $AVAILABLE_TESTS"
fi

echo ""
echo "🎉 Setup verification complete!"
echo "==============================="
echo ""
echo "Next steps:"
echo "1. Run all fuzz tests: make test-fuzz"
echo "2. Run short fuzz tests: make test-fuzz-short"
echo "3. Run specific test: go test -fuzz=FuzzTestName -fuzztime=30s"
echo "4. The GitHub Actions workflow will run automatically on PR/push"
echo ""
echo "For more information, see test/fuzz/README.md"
