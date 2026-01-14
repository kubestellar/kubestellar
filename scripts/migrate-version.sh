#!/bin/bash
# migrate-version.sh - Migrates docs from kubestellar release branch to docs branch
#
# Usage: ./scripts/migrate-version.sh <version> <source-branch>
# Example: ./scripts/migrate-version.sh 0.28.0 release-0.28.0
#
# This script:
# 1. Clones kubestellar repo at the specified release branch
# 2. Creates a new docs/{version} branch in the docs repo
# 3. Copies the documentation content
# 4. Updates CURRENT_VERSION in versions.ts
# 5. Commits and pushes the new branch

set -e

VERSION=$1
SOURCE_BRANCH=$2

if [ -z "$VERSION" ] || [ -z "$SOURCE_BRANCH" ]; then
    echo "Usage: $0 <version> <source-branch>"
    echo "Example: $0 0.28.0 release-0.28.0"
    exit 1
fi

DOCS_REPO_DIR=$(pwd)
TEMP_DIR="/tmp/ks-migrate-$$"

echo "=== Migrating documentation for version $VERSION from $SOURCE_BRANCH ==="

# Ensure we're in the docs repo
if [ ! -f "package.json" ] || [ ! -d "src/config" ]; then
    echo "Error: This script must be run from the root of the docs repository"
    exit 1
fi

# Clean up temp directory if it exists
rm -rf "$TEMP_DIR"
mkdir -p "$TEMP_DIR"

echo "1. Cloning kubestellar repo at branch $SOURCE_BRANCH..."
git clone --branch "$SOURCE_BRANCH" --depth 1 \
    https://github.com/kubestellar/kubestellar.git "$TEMP_DIR/kubestellar"

# Check if the source docs exist
if [ ! -d "$TEMP_DIR/kubestellar/docs/content" ]; then
    echo "Error: docs/content directory not found in source branch"
    rm -rf "$TEMP_DIR"
    exit 1
fi

echo "2. Creating new branch docs/$VERSION..."
# Ensure we're on main and up to date
git checkout main
git pull origin main

# Create the version branch
git checkout -b "docs/$VERSION"

echo "3. Copying documentation content..."
# Remove existing content (preserve directory)
rm -rf docs/content/*

# Copy new content
cp -r "$TEMP_DIR/kubestellar/docs/content/"* docs/content/

# Also copy mkdocs.yml for reference (navigation structure)
if [ -f "$TEMP_DIR/kubestellar/docs/mkdocs.yml" ]; then
    cp "$TEMP_DIR/kubestellar/docs/mkdocs.yml" docs/mkdocs.yml
fi

echo "4. Updating version configuration..."
# Update CURRENT_VERSION in versions.ts
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    sed -i '' "s/export const CURRENT_VERSION = .*/export const CURRENT_VERSION = \"$VERSION\"/" src/config/versions.ts
else
    # Linux
    sed -i "s/export const CURRENT_VERSION = .*/export const CURRENT_VERSION = \"$VERSION\"/" src/config/versions.ts
fi

echo "5. Committing changes..."
git add .
git commit -s -m "docs: migrate documentation for version $VERSION

Migrated from kubestellar/kubestellar branch $SOURCE_BRANCH

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

echo "6. Pushing branch to origin..."
git push origin "docs/$VERSION"

# Clean up
rm -rf "$TEMP_DIR"

echo ""
echo "=== Migration complete ==="
echo "Branch docs/$VERSION has been created and pushed."
echo ""
echo "Next steps:"
echo "1. Verify the branch builds correctly on Netlify"
echo "2. Add this version to src/config/versions.ts on main branch"
echo "3. Test the version dropdown navigation"
