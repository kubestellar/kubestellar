#!/bin/bash

# Input version should be X.Y.Z
input_version="$1"

if ! [[ "$1" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Invalid version format. Please provide a version in the format X.Y.Z (all numbers)." >&2
    exit 1
fi

# Replace version string using sed

sed -i "/- --version/{n;s/- \"[0-9]\+\.[0-9]\+\.[0-9]\+\"/- \"$input_version\"/}" config/postcreate-hooks/kubestellar.yaml

sed -i "s/export KUBESTELLAR_VERSION=[0-9]\+\.[0-9]\+\.[0-9]\+/export KUBESTELLAR_VERSION=$input_version/" docs/content/direct/examples.md

sed -i "/The latest release is/s/[0-9]\+\.[0-9]\+\.[0-9]\+/$input_version/g" docs/content/direct/README.md
