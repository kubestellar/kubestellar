#!/usr/bin/env bash

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

# Purpose: update the reverse map $REVERSEMAP_FILE from actions used in $GITHUB_WORKFLOWS_PATH

# Usage: $0
# Working directory does not matter.

# Define the required version of yq
GITHUB_PAT=                                                 # API GITHUB TOKEN
GITHUB_WORKFLOWS_PATH="./.github/workflows/"
REVERSEMAP_FILE="./.github/.reversemap.yml"
YQ_REQUIRED_VERSION="v4.45.1"
YQ_IS_INSTALLED=0
GIT_COMMITSHA_LENGTH=40

# Function to check if yq is installed and has the required version
check_yq_version() {
    if command -v yq >/dev/null 2>&1; then
        INSTALLED_VERSION=$(yq --version 2>/dev/null)
        if [[ "$INSTALLED_VERSION" =~ "$YQ_REQUIRED_VERSION" ]]; then
            echo "yq $YQ_REQUIRED_VERSION is already installed."
            YQ_IS_INSTALLED=1
        else
            echo "yq is installed but the version is $INSTALLED_VERSION. Required version is $YQ_REQUIRED_VERSION."
        fi
    else
        echo "yq is not installed."
    fi
}

# Function to download and install yq
install_yq() {
    OS=$(uname | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    if [[ "$ARCH" == "x86_64" ]]; then
        ARCH="amd64"
    elif [[ "$ARCH" == "aarch64" ]]; then
        ARCH="arm64"
    elif [[ "$ARCH" == "arm64" ]]; then
        ARCH="arm64"
    else
        echo "Unsupported architecture: $ARCH"
        exit 1
    fi

    YQ_BINARY="yq_${OS}_${ARCH}"
    DOWNLOAD_URL="https://github.com/mikefarah/yq/releases/download/${YQ_REQUIRED_VERSION}/${YQ_BINARY}.tar.gz"

    echo "Downloading yq $YQ_REQUIRED_VERSION for $OS/$ARCH..."
    curl -L -o /tmp/${YQ_BINARY}.tar.gz "$DOWNLOAD_URL"

    if [[ $? -ne 0 ]]; then
        echo "Failed to download yq."
        exit 1
    fi

    tar xzf /tmp/${YQ_BINARY}.tar.gz -C /tmp
    sudo chmod +x /tmp/${YQ_BINARY}
    sudo mv /tmp/${YQ_BINARY} /usr/local/bin/yq

    echo "yq $YQ_REQUIRED_VERSION installed successfully."
}

# Get commitsha from an action ref
_get_commitsha_from_ref() {
    action_ref=$1
    tag_or_branch=$2
    API_GITHUB_BRANCH="https://api.github.com/repos/${action_ref}/git/refs/heads/${tag_or_branch}"
    API_GITHUB_TAG="https://api.github.com/repos/${action_ref}/git/refs/tags/${tag_or_branch}"
    HTTP_STATUS=$(curl -o /dev/null -s -w "%{http_code}\n" -H "Authorization: $GITHUB_PAT" "$API_GITHUB_TAG")
    if [[ $HTTP_STATUS -ge 200 && $HTTP_STATUS -lt 300 ]]; then
        commit_sha=$(curl -s -H "Authorization: $GITHUB_PAT" $API_GITHUB_TAG | jq -r '.object.sha')
    else
        commit_sha=$(curl -s -H "Authorization: $GITHUB_PAT" $API_GITHUB_BRANCH | jq -r '.object.sha')
    fi
    echo $commit_sha
}

# Extract actions tag and update reversemap with commit
extract_actions_info() {
    filename=$1
    outfile=$2
    for action_fullref in $(yq '.jobs[].steps[] | select(has("uses")) | .uses ' $filename); do
        action_ref=$(echo $action_fullref | cut -d '@' -f 1)
        action_tag=$(echo $action_fullref | cut -d '@' -f 2)
        length=${#action_tag}
        echo "debug: ref=$action_ref and tag=$action_tag len=$length"
        if [[ $length -ne $GIT_COMMITSHA_LENGTH ]]; then
            action_commitsha=$(_get_commitsha_from_ref $action_ref $action_tag)
            echo "info: sha=$action_commitsha"
            yq ".${action_ref}.commit-sha = \"${action_commitsha}\"" -i $2
            yq ".${action_ref}.tag = \"${action_tag}\"" -i $2
        fi
    done
}

# Execute the main program
main() {
    check_yq_version
    if [[ $YQ_IS_INSTALLED -eq 0 ]]; then
        install_yq
    fi
    for file in $(ls $GITHUB_WORKFLOWS_PATH); do
        echo "info: fetching $file..."
        extract_actions_info "$GITHUB_WORKFLOWS_PATH/$file" $REVERSEMAP_FILE
    done
}

main;
