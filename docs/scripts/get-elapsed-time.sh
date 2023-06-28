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

os_type=""

get_os_type() {
  case "$OSTYPE" in
      linux*)   echo "linux" ;;
      darwin*)  echo "darwin" ;;
      *)        echo "Unsupported operating system type: $OSTYPE" >&2 ; exit 1 ;;
  esac
}

if [ "$os_type" == "" ]; then
    os_type=$(get_os_type)
fi

REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
# cd "$REPO_ROOT/docs"

FILE_LIST=()
SAVEIFS=$IFS

IFS=',' read -r -a FILE_LIST <<< "${FILENAME}"
if [ -z "${FILE_LIST[0]+x}" ]; then
    echo empty filename, exiting
    exit
else
  docs_ecutable_filename=${FILE_LIST[0]}
fi

IFS=$SAVEIFS


filename_ext=".dmd"

workflow_name=""
workload_id=""

echo doc_name: $docs_ecutable_filename
if [ -n "$docs_ecutable_filename" ]; then
    x=0;
    workflow_name=$(gh api -H "Accept: application/vnd.github+json" -H "X-GitHub-Api-Version: 2022-11-28" /repos/kubestellar/kubestellar/actions/workflows | jq ".workflows[$x].name") 
    while [[ ${workflow_name,,} != *"${docs_ecutable_filename,,}"* ]]
    do
        x=$((x+1))
        workflow_name=$(gh api -H "Accept: application/vnd.github+json" -H "X-GitHub-Api-Version: 2022-11-28" /repos/kubestellar/kubestellar/actions/workflows | jq ".workflows[$x].name") 

        if [ "$workflow_name" == "null" ];then echo 'no match'; exit; fi
    done
    echo workflow_name: $workflow_name 
    workflow_id=$(gh api -H "Accept: application/vnd.github+json" -H "X-GitHub-Api-Version: 2022-11-28" /repos/kubestellar/kubestellar/actions/workflows | jq ".workflows[$x].id")
    echo workflow_id: $workflow_id
    run_number=$(gh api -H "Accept: application/vnd.github+json" -H "X-GitHub-Api-Version: 2022-11-28" /repos/kubestellar/kubestellar/actions/workflows/$workflow_id/runs | jq ".workflow_runs[0].run_number")
    echo run_number: $run_number
    # workflow_status="" 
    # y=0;
    # workflow_status=$(gh api -H "Accept: application/vnd.github+json" -H "X-GitHub-Api-Version: 2022-11-28" /repos/kubestellar/kubestellar/actions/workflows/$x/runs | jq ".workflows[$y].status") 
    # echo $workflow_status
    # while [ "workflow_status" != "completed" ]
    # do
    #     y=$((x+1))
    #     workflow_status=$(gh api -H "Accept: application/vnd.github+json" -H "X-GitHub-Api-Version: 2022-11-28" /repos/kubestellar/kubestellar/actions/workflows | jq ".workflows[$y].status") 
    # done
    # workflow_id=$(gh api -H "Accept: application/vnd.github+json" -H "X-GitHub-Api-Version: 2022-11-28" /repos/kubestellar/kubestellar/actions/workflows | jq ".workflows[$x].id")
    # echo workflow_id: $workflow_id
    createdAt=$(gh api -H "Accept: application/vnd.github+json" -H "X-GitHub-Api-Version: 2022-11-28" /repos/kubestellar/kubestellar/actions/workflows/$workflow_id/runs | jq '.workflow_runs[0].created_at')
    updatedAt=$(gh api -H "Accept: application/vnd.github+json" -H "X-GitHub-Api-Version: 2022-11-28" /repos/kubestellar/kubestellar/actions/workflows/$workflow_id/runs | jq '.workflow_runs[0].updated_at')
fi

createdAt=${createdAt//\"/}
echo createdAt: $createdAt
updatedAt=${updatedAt//\"/}
echo updatedAt: $updatedAt

# Convert timestamps to date objects
if [ "$os_type" == "darwin" ]; then
    date1=$(date -ju -f "%Y-%m-%dT%H:%M:%SZ" "$createdAt" "+%s")
else
    date1=$(date -d "$createdAt" "+%s")
fi

if [ "$os_type" == "darwin" ]; then
    date2=$(date -ju -f "%Y-%m-%dT%H:%M:%SZ" "$updatedAt" "+%s")
else
    date2=$(date -d "$updatedAt" "+%s")
fi

# Compute the time difference in seconds
time_diff=$((date2 - date1))
minutes=$(( (time_diff + 60 )/ 60))

if [ $minutes -gt 1 ]; then
    echo -e '<!--elapsed-time-start-->!!! tip "Estimated time to complete this example:" 
    ~' $minutes ' minutes<!--elapsed-time-end-->'
else
    echo -e '<!--elapsed-time-start-->!!! tip "Estimated time to complete this example:" 
    no estimate available at this time<!--elapsed-time-end-->'
fi

