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
# GITHUB_TOKEN=
set -e

GITHUB_WORKFLOWS_PATH="./.github/workflows"
REVERSEMAP_FILE=".gha-reversemap.yml"
YQ_REQUIRED_VERSION="v4.45"
GIT_COMMITSHA_LENGTH=40
TMP_OUTPUT="/tmp/$(date -u -Iseconds | cut -d '+' -f1).json"

ERR_YQ_DOWNLOAD_FAILED=50
ERR_YQ_NOT_INSTALLED=60
ERR_GITHUB_TOKEN_INVALID=70
ERR_ARCH_UNSUPPORTED=80

help() {
    cat <<EOF
Harmonise '$GITHUB_WORKFLOWS_PATH' and the reversemap '$REVERSEMAP_FILE'

Usage:
    $0  <operation> [ARGUMENTS]

Example:
    # Update actions/checkout to its latest release
    $0  update actions/checkout

Operations:
    apply-reversemap        [WORKFLOW_FILE...]          -   Update the given workflow files with the information in the reversemap
                                                            file; if no workflow files are given then all are updated

    update-action-version   ACTION_REF...               -   Update the version of the given action reference within the reversemap
                                                            (sha, tag, urls) to its latest regular release tag

    update-reversemap       [WORKFLOW_FILE...]          -   Update the reverse map values (sha, tag, urls) with the information 
                                                            in the github workflow; if no workflow files are specified then 
                                                            all workflows are used

EOF
}

# Log information
_loginfo() {
    echo -e "$(date -Iseconds);INFO;$1"
}

# Exit with an error
_exit_with_error() {
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
            _exit_with_error $ERR_YQ_NOT_INSTALLED "yq is installed but the version is $INSTALLED_VERSION. Required version is $YQ_REQUIRED_VERSION."
        fi
    else
        _exit_with_error $ERR_YQ_NOT_INSTALLED "yq is not installed."
    fi
}

# Check github token
_check_github_token(){
    if [[ -z $GITHUB_TOKEN ]]; then
        _exit_with_error $ERR_GITHUB_TOKEN_INVALID "environment variable GITHUB_TOKEN is not set."
    fi
}

# Get commitsha from an action ref upstream
_fetch_sha_from_upstream_ref() {
    action_ref=$1
    tag_or_branch=$2
    action_ref_safe=$(echo "$action_ref" | cut -d '/' -f 1,2)
    API_GITHUB_BRANCH="https://api.github.com/repos/${action_ref_safe}/git/refs/heads/${tag_or_branch}"
    API_GITHUB_TAG="https://api.github.com/repos/${action_ref_safe}/git/refs/tags/${tag_or_branch}"
    HTTP_STATUS=$(curl -o "$TMP_OUTPUT" -s -w "%{http_code}" -H "Authorization: Bearer $GITHUB_TOKEN" "$API_GITHUB_TAG")
    if [[ $HTTP_STATUS -ge 200 && $HTTP_STATUS -lt 300 ]]; then
        commit_sha=$(jq -r '.object.sha' $TMP_OUTPUT)
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

# Update the reverse map values with an action full ref (ie. actions/checkout@v4.2.0)
_update_reversemap_with() {
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
    sed_replace_expr="s;(uses:) $action_ref@[a-zA-Z0-9 \._\-]+;\1 $action_ref@$action_sha ;g"
    if [[ "$(uname -s)" = "Darwin" ]]; then
        # MacOS sed usage of -i option differs from Linux version.
        sed -E -i '' "$sed_replace_expr" "$file"
    else
        sed -E -i "$sed_replace_expr" "$file"
    fi
}

# Get commit sha of an action ref stored in the reverse map
_get_sha_from_reversemap() {
    action_ref=$1
    query=$(yq ".\"$action_ref\".sha" $REVERSEMAP_FILE)
    if [[ $? -ne 0 ]]; then
        _exit_with_error "no sha has been found for $action_ref in $REVERSEMAP_FILE"
    fi
    _return "$query"
}

# Apply the commit sha of every action listed in the reversemap as the version to use in all github workflows
run_apply_reversemap() {
    files=$@
    for file in $files; do
        _loginfo "applying '$REVERSEMAP_FILE' commit sha to be used in '$file' ..."
        action_fullrefs=$(yq '.jobs[].steps[] | select(.uses) | .uses' "$file")
        for action_fullref in $action_fullrefs; do
            action_ref=$(echo "$action_fullref" | cut -d "@" -f 1)
            action_sha=$(_get_sha_from_reversemap "$action_ref")
            _loginfo "found '$action_ref' in reversemap with sha=$action_sha"
            _update_action_version_infile "$file" "$action_ref" "$action_sha"
        done
    done
}

# Update the version of the given action references within the reverse map (sha, tag, urls) to its latest regular release tag
run_update_action_version() {
    action_refs=$@
    for action_ref in $action_refs; do
        _loginfo "updating dependency '$action_ref' tag to latest version available inside reverse map '$REVERSEMAP_FILE'"
        latest_tag=$(_fetch_latest_tag "$action_ref")
        action_sha=$(_fetch_sha_from_upstream_ref "$action_ref" "$latest_tag")
        _yq_update_reversemap "$action_ref" "$latest_tag" "$action_sha"
    done
}

# Update the reverse map values (sha, tag, urls) from the actions used in given github workflows
run_update_reversemap() {
    files=$@
    for file in $files; do
        _loginfo "updating $REVERSEMAP_FILE with actions from '$file' ..."
        _update_reversemap_with "$file"
    done
}

# Run the CLI with its operations
run_cli() {
    operation_arg=$1
    case $operation_arg in
    "help")
        help
        ;;
    "apply-reversemap")
        shift
        _check_yq_version
        files=$@
        if [[ "$#" -eq 0 ]]; then
            files="$GITHUB_WORKFLOWS_PATH/*.yml"
        fi
        _loginfo "running $operation_arg on $files"
        run_apply_reversemap $files
        ;;
    "update-action-version")
        shift
        _check_github_token
        _check_yq_version
        action_refs=$@
        if [[ "$#" -eq 0 ]]; then
            _exit_with_error 1 "missing action ref to update. Format must be '{gh_owner}/{gh_repo}'"
        fi
        _loginfo "running $operation_arg on $action_refs"
        run_update_action_version $action_refs
        ;;
    "update-reversemap")
        shift
        _check_github_token
        _check_yq_version
        files=$@
        if [[ "$#" -eq 0 ]]; then
            files="$GITHUB_WORKFLOWS_PATH/*.yml"
        fi
        _loginfo "running $operation_arg on $files"
        run_update_reversemap $files
        ;;
    *)
        help
        exit 1
        ;;
    esac

}

# Execute the main program
run_cli $@
