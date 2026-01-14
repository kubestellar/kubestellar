#!/bin/bash
# create-version-branches.sh - Batch migrate all stable versions
#
# Usage: ./scripts/create-version-branches.sh [--dry-run]
#
# This script migrates all stable release versions from kubestellar/kubestellar
# to this docs repository.

set -e

DRY_RUN=false
if [ "$1" == "--dry-run" ]; then
    DRY_RUN=true
    echo "=== DRY RUN MODE - No changes will be made ==="
fi

# Stable versions to migrate (version:source-branch format)
# Listed from newest to oldest
STABLE_VERSIONS=(
    "0.28.0:release-0.28.0"
    "0.27.2:release-0.27.2"
    "0.27.1:release-0.27.1"
    "0.27.0:release-0.27.0"
    "0.26.0:release-0.26.0"
    "0.25.1:release-0.25.1"
    "0.25.0:release-0.25.0"
    "0.24.0:release-0.24.0"
    "0.23.1:release-0.23.1"
    "0.23.0:release-0.23.0"
    "0.22.0:release-0.22.0"
    "0.21.2:release-0.21.2"
    "0.21.1:release-0.21.1"
    "0.21.0:release-0.21.0"
)

SCRIPT_DIR=$(dirname "$0")
FAILED_VERSIONS=()
SUCCESSFUL_VERSIONS=()

echo "=== Starting batch migration of ${#STABLE_VERSIONS[@]} versions ==="
echo ""

for entry in "${STABLE_VERSIONS[@]}"; do
    VERSION="${entry%%:*}"
    SOURCE_BRANCH="${entry##*:}"

    echo "----------------------------------------"
    echo "Processing version $VERSION from $SOURCE_BRANCH"
    echo "----------------------------------------"

    if [ "$DRY_RUN" == "true" ]; then
        echo "[DRY RUN] Would migrate version $VERSION from $SOURCE_BRANCH"
        SUCCESSFUL_VERSIONS+=("$VERSION")
        continue
    fi

    # Check if branch already exists
    if git ls-remote --heads origin "docs/$VERSION" | grep -q "docs/$VERSION"; then
        echo "Branch docs/$VERSION already exists, skipping..."
        continue
    fi

    # Run migration script
    if "$SCRIPT_DIR/migrate-version.sh" "$VERSION" "$SOURCE_BRANCH"; then
        SUCCESSFUL_VERSIONS+=("$VERSION")
        echo "Successfully migrated version $VERSION"
    else
        FAILED_VERSIONS+=("$VERSION")
        echo "Failed to migrate version $VERSION"
    fi

    # Return to main branch for next iteration
    git checkout main

    echo ""
done

echo "=========================================="
echo "=== Migration Summary ==="
echo "=========================================="
echo ""
echo "Successful: ${#SUCCESSFUL_VERSIONS[@]} versions"
for v in "${SUCCESSFUL_VERSIONS[@]}"; do
    echo "  - $v"
done
echo ""
if [ ${#FAILED_VERSIONS[@]} -gt 0 ]; then
    echo "Failed: ${#FAILED_VERSIONS[@]} versions"
    for v in "${FAILED_VERSIONS[@]}"; do
        echo "  - $v"
    done
fi
echo ""
echo "Next steps:"
echo "1. Update src/config/versions.ts on main branch with all migrated versions"
echo "2. Test version dropdown navigation"
echo "3. Verify each branch builds correctly on Netlify"
