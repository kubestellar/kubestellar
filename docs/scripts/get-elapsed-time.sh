#!/usr/bin/env bash

# Copyright 2023 The KubeStellar Authors.
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
# set -o xtrace

REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
# cd "$REPO_ROOT/docs"

FILE_LIST=()
SAVEIFS=$IFS

IFS=',' read -r -a FILE_LIST <<< "${FILENAME}"
if [ -z "${FILE_LIST[0]+x}" ]; then
    echo empty filename, exiting
    exit
else
  option=${FILE_LIST[0]}
fi

IFS=$SAVEIFS


workflow_id=""
filename_ext=".dmd"

case $option in
    placement)
        echo "Selected placement option"
        workflow_id="59166207"
        ;;
    scheduler)
        echo "Selected scheduler option"
        workflow_id="59159715"
        ;;
    mailbox)
        echo "Selected mailbox option"
        workflow_id="59164731"
        ;;
    example1)
        echo "Selected example1 option"
        workflow_id="59184957"
        ;;
    quickstart)
        echo "Selected quickstart option"
        workflow_id="59223574"
        ;;
    syncer)
        echo "Selected syncer option"
        workflow_id="?"
        ;;
    *)
        echo "Invalid option"
        ;;
esac

if [ -n "$option" ]; then
  createdAt=$(gh api -H "Accept: application/vnd.github+json" -H "X-GitHub-Api-Version: 2022-11-28" /repos/kcp-dev/edge-mc/actions/workflows/$workflow_id/runs | jq '.workflow_runs[0].created_at')
  updatedAt=$(gh api -H "Accept: application/vnd.github+json" -H "X-GitHub-Api-Version: 2022-11-28" /repos/kcp-dev/edge-mc/actions/workflows/$workflow_id/runs | jq '.workflow_runs[0].updated_at')
fi

createdAt=${createdAt//\"/}
updatedAt=${updatedAt//\"/}

# Convert timestamps to date objects
date1=$(date -d "$createdAt" "+%s")
echo $date1
# date1=$(date -ju -f "%Y-%m-%dT%H:%M:%SZ" "$createdAt" "+%s")
date2=$(date -d "$updatedAt" "+%s")
echo $date2
# date2=$(date -ju -f "%Y-%m-%dT%H:%M:%SZ" "$updatedAt" "+%s")

# Compute the time difference in seconds
time_diff=$((date2 - date1))
minutes=$((time_diff / 60))

docs-ecutable-filename=$option$filename_ext
echo -e "<!--elapsed-time-start-->\n$minutes\n<!--elapsed-time-end-->" > /docs/overrides/$docs-ecutable-filename

