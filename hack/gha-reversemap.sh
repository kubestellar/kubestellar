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

# Purpose: Harmonise '$GITHUB_WORKFLOWS_PATH' and the reversemap '$REVERSEMAP_FILE'

# Usage: see help function
# Working directory must be the project root

# Define the required version of yq
GITHUB_TOKEN=
GITHUB_WORKFLOWS_PATH="./.github/workflows"
REVERSEMAP_FILE=".gha-reversemap.yml"
YQ_REQUIRED_VERSION="v4.45.1"
GIT_COMMITSHA_LENGTH=40
RE_ACTION_REF_AT="([a-zA-Z0-9 _/. \-]+)@"
RE_ACTION_FULLREF="$RE_ACTION_REF_AT([a-zA-Z0-9 _/. \-]+)"

ERR_YQ_DOWNLOAD_FAILED=50
ERR_YQ_NOT_INSTALLED=60
ERR_GITHUB_TOKEN_INVALID=70
ERR_ARCH_UNSUPPORTED=80

help() {
    cat <<EOF
Harmonise '$GITHUB_WORKFLOWS_PATH' and the reversemap '$REVERSEMAP_FILE'

Usage:
    $0  <operation>

Example:
    # Update actions/checkout to its latest release
    $0  update actions/checkout

Operations:
    apply                 -   Apply the commit sha of every actions listed in the reversemap as the version to use in all github workflows
    update <action-ref>   -   Update the version of the given action reference to its latest release within the reverse map (sha, tag, urls)
    sync                  -   Sync the reverse map values (sha, tag, urls) from the actions used in all github workflows

Parameters:
    <action-ref>:   {gh_owner}/{gh_repo}
EOF
}

# Log information
_loginfo() {
    echo -e "$(date -Iseconds);INFO;$1"
}

# Exit with an error
_error_exit() {
    echo -e "$(date -Iseconds);ERROR;$2" >&2
    exit "$1"
}

# Internal - indicates proper return of value within a function
_return() {
    echo "$1"
}

# Function to check if yq is installed and has the required version
_check_yq_version() {
    if command -v yq >/dev/null 2>&1; then
        INSTALLED_VERSION=$(yq --version 2>/dev/null)
        if ! [[ "$INSTALLED_VERSION" =~ $YQ_REQUIRED_VERSION ]]; then
            _error_exit $ERR_YQ_NOT_INSTALLED "yq is installed but the version is $INSTALLED_VERSION. Required version is $YQ_REQUIRED_VERSION."
        fi
    else
        _error_exit $ERR_YQ_NOT_INSTALLED "yq is not installed."
    fi
}

# Function to download and install yq
_install_yq() {
    os=$(uname | tr '[:upper:]' '[:lower:]')
    arch=$(uname -m)

    if [[ "$arch" == "x86_64" ]]; then
        arch="amd64"
    elif [[ "$arch" == "aarch64" ]]; then
        arch="arm64"
    elif [[ "$arch" == "arm64" ]]; then
        arch="arm64"
    else
        _error_exit $ERR_ARCH_UNSUPPORTED "Unsupported architecture: $arch"
    fi

    yq_binary="yq_${os}_${arch}"
    yq_binary_url="https://github.com/mikefarah/yq/releases/download/${YQ_REQUIRED_VERSION}/${yq_binary}.tar.gz"

    _loginfo "Downloading yq $YQ_REQUIRED_VERSION for $os/$arch ..."
    curl -L -o "/tmp/${yq_binary}.tar.gz" "$yq_binary_url"

    if [[ $? -ne 0 ]]; then
        _error_exit $ERR_YQ_DOWNLOAD_FAILED "Failed to download yq."
    fi

    tar xzf "/tmp/${yq_binary}.tar.gz" -C /tmp
    sudo chmod +x "/tmp/${yq_binary}"
    sudo mv "/tmp/${yq_binary}" /usr/local/bin/yq

    echo "yq $YQ_REQUIRED_VERSION installed successfully."
}

# Get commitsha from an action ref upstream
_fetch_sha_from_upstream_ref() {
    action_ref=$1
    tag_or_branch=$2
    action_ref_safe=$(echo "$action_ref" | cut -d '/' -f 1,2)
    API_GITHUB_BRANCH="https://api.github.com/repos/${action_ref_safe}/git/refs/heads/${tag_or_branch}"
    API_GITHUB_TAG="https://api.github.com/repos/${action_ref_safe}/git/refs/tags/${tag_or_branch}"
    HTTP_STATUS=$(curl -o /dev/null -s -w "%{http_code}\n" -H "Authorization: Bearer $GITHUB_TOKEN" "$API_GITHUB_TAG")
    if [[ $HTTP_STATUS -ge 200 && $HTTP_STATUS -lt 300 ]]; then
        commit_sha=$(curl -s -H "Authorization: Bearer $GITHUB_TOKEN" "$API_GITHUB_TAG" | jq -r '.object.sha')
        # _loginfo "tag api $commit_sha"
    else
        commit_sha=$(curl -s -H "Authorization: Bearer $GITHUB_TOKEN" "$API_GITHUB_BRANCH" | jq -r '.object.sha')
        # _loginfo "branch api $commit_sha"
    fi
    _return "$commit_sha"
}

# Update reverse map KV using yq
_yq_update_reversemap() {
    action_ref=$1
    action_tag=$2
    action_sha=$3
    action_ref_safe=$(echo "$action_ref" | cut -d '/' -f 1,2) # In case the action is a sub-action of a repo (ie. user/action/subaction@v1)
    yq ".${action_ref}.sha = \"${action_sha}\"" -i $REVERSEMAP_FILE
    yq ".${action_ref}.sha-url = \"https://github.com/${action_ref_safe}/commit/${action_sha}\"" -i $REVERSEMAP_FILE
    yq ".${action_ref}.tag = \"${action_tag}\"" -i $REVERSEMAP_FILE
    yq ".${action_ref}.tag-url = \"https://github.com/${action_ref_safe}/tree/${action_tag}\"" -i $REVERSEMAP_FILE
}

# Synchronise reverse map values with an action full ref (ie. actions/checkout@v4.2.0)
_sync_reversemap_with() {
    filename=$1
    for action_fullref in $(yq '.jobs[].steps[] | select(has("uses")) | .uses ' "$filename"); do
        action_ref=$(echo "$action_fullref" | cut -d '@' -f 1)
        action_tag=$(echo "$action_fullref" | cut -d '@' -f 2)
        length=${#action_tag}
        _loginfo "ref=$action_ref and tag=$action_tag len=$length"
        if [[ $length -ne $GIT_COMMITSHA_LENGTH ]]; then
            action_sha=$(_fetch_sha_from_upstream_ref "$action_ref" "$action_tag")
            _loginfo "taking commit of action=${action_ref} tag=${action_tag} sha=${action_sha}"
            _yq_update_reversemap "$action_ref" "$action_tag" "$action_sha"
        fi
    done
}

# Fetch latest release tag of an action upstream
_fetch_latest_tag() {
    action_ref=$1
    API_GITHUB_LATEST_RELEASE=https://api.github.com/repos/$action_ref/releases/latest
    _return "$(curl -s -H "Authorization: Bearer $GITHUB_TOKEN" "$API_GITHUB_LATEST_RELEASE" | jq -r '.tag_name')"
}

# Update action version within a file
_update_action_version_infile() {
    file=$1
    action_ref=$2
    action_sha=$3
    sed -E -i '' "s;$action_ref@.+;$action_ref@$action_sha;g" "$file"
}

# Get commit sha of an action ref stored in the reverse map
_get_sha_from_reversemap() {
    action_ref=$1
    query=$(yq ".\"$action_ref\".sha" $REVERSEMAP_FILE)
    if [[ $? -ne 0 ]]; then
        _error_exit "$query"
    fi
    _return "$query"
}

# Run operation apply
run_apply() {
    for file in "$GITHUB_WORKFLOWS_PATH"/*.yml; do
        _loginfo "applying '$REVERSEMAP_FILE' commit sha to be used in '$file' ..."
        action_refs=$(grep -E -o "$RE_ACTION_REF_AT" "$file")
        for action_ref in $action_refs; do
            action_ref=${action_ref%?}
            action_sha=$(_get_sha_from_reversemap "$action_ref")
            _loginfo "found $action_ref in reversemap with sha=$action_sha"
            _update_action_version_infile "$file" "$action_ref" "$action_sha"
        done
    done
}

# Run operation update
run_update() {
    dependency_arg=$1
    _loginfo "updating dependency '$dependency_arg' tag to latest version available inside reverse map '$REVERSEMAP_FILE'"
    latest_tag=$(_fetch_latest_tag "$dependency_arg")
    action_sha=$(_fetch_sha_from_upstream_ref "$dependency_arg" "$latest_tag")
    _yq_update_reversemap "$dependency_arg" "$latest_tag" "$action_sha"
}

# Run operation sync
run_sync() {
    for file in "$GITHUB_WORKFLOWS_PATH"/*.yml; do
        _loginfo "syncing $REVERSEMAP_FILE with actions from '$file' ..."
        _sync_reversemap_with "$file"
    done
}

# Run the CLI with its operations
run_cli() {
    operation_arg=$1
    case $operation_arg in
    "apply")
        _loginfo "running $1"
        run_apply
        ;;
    "update")
        dependency_arg=$2
        if [[ -z $dependency_arg ]]; then
            _error_exit -1 "missing dependency name to update."
        fi
        _loginfo "running $1"
        run_update "$dependency_arg"
        ;;
    "sync")
        _loginfo "running $1"
        run_sync
        ;;
    *)
        help
        exit 2
        ;;
    esac

}

# Execute the main program
main() {
    if [[ -z $GITHUB_TOKEN ]]; then
        _error_exit $ERR_GITHUB_TOKEN_INVALID "no github pat"
    fi
    _check_yq_version
    if [[ $? -eq $ERR_YQ_NOT_INSTALLED ]]; then
        _install_yq
    fi
    run_cli $@

}

main $@
