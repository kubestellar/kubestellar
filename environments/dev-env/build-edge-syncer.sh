#!/usr/bin/env bash
# Copyright 2023 The KCP Authors.
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

set -e

syncTarget_name=""
syncer_name="the-one"
verbosity=0
img="quay.io/kcpedge/syncer:dev-2023-04-18"

while (( $# > 0 )); do
    if [ "$1" == "--syncTarget" ]; then
        syncTarget_name=$2
        shift
    elif [ "$1" == "--syncerName" ]; then
        syncer_name==$2
        shift
    elif [ "$1" == "--img" ]; then
        img=$2
        shift
    elif [ "$1" == "-v" ]; then
        verbosity=1
    fi 
    shift
done

kubectl ws root:espw > /dev/null 2>&1
workspace=$(kubectl get Workspace -o json | jq  --arg v "$syncTarget_name" -r '.items | .[] | .metadata | select(.annotations ["edge.kcp.io/sync-target-name"] == $v) | .name')

# Create edge-syncer manifest
if [ $verbosity == 1 ]; then
    kubectl ws root:espw:$workspace
    kubectl kcp workload edge-sync $syncer_name --syncer-image $img -o $syncTarget_name-syncer.yaml    
else
    kubectl ws root:espw:$workspace > /dev/null 2>&1
    kubectl kcp workload edge-sync $syncer_name --syncer-image $img -o $syncTarget_name-syncer.yaml > /dev/null 2>&1
fi

echo "-----------------------------------------------------------"
echo "Edge-syncer manifest created:  $syncTarget_name-syncer.yaml"
echo "Current workspace: root:espw:$workspace"
echo "-----------------------------------------------------------"
