#!/usr/bin/env bash
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

# Test configuration scenarios for KubeStellar
# Each scenario defines how WDS and ITS roles are deployed

# Default scenario (current setup)
declare -A DEFAULT_SCENARIO=(
    [name]="default"
    [description]="Three new kind clusters with separate WDS and ITS"
    [wds_type]="new"
    [its_type]="new"
    [cluster_source]="kind"
    [combined_control_plane]="false"
)

# Scenario: WDS on hosting cluster
declare -A WDS_ON_HOST_SCENARIO=(
    [name]="wds-on-host"
    [description]="WDS on hosting cluster, new ITS"
    [wds_type]="hosting"
    [its_type]="new"
    [cluster_source]="kind"
    [combined_control_plane]="false"
)

# Scenario: ITS on hosting cluster
declare -A ITS_ON_HOST_SCENARIO=(
    [name]="its-on-host"
    [description]="ITS on hosting cluster, new WDS"
    [wds_type]="new"
    [its_type]="hosting"
    [cluster_source]="kind"
    [combined_control_plane]="false"
)

# Scenario: Combined WDS and ITS
declare -A COMBINED_SCENARIO=(
    [name]="combined"
    [description]="Single control plane for both WDS and ITS"
    [wds_type]="new"
    [its_type]="new"
    [cluster_source]="kind"
    [combined_control_plane]="true"
)

# Function to get scenario configuration
get_scenario_config() {
    local scenario_name=$1
    case "$scenario_name" in
        "default")
            echo "${DEFAULT_SCENARIO[@]}"
            ;;
        "wds-on-host")
            echo "${WDS_ON_HOST_SCENARIO[@]}"
            ;;
        "its-on-host")
            echo "${ITS_ON_HOST_SCENARIO[@]}"
            ;;
        "combined")
            echo "${COMBINED_SCENARIO[@]}"
            ;;
        *)
            echo "Unknown scenario: $scenario_name" >&2
            return 1
            ;;
    esac
}

# Function to list available scenarios
list_scenarios() {
    echo "Available test scenarios:"
    echo "1. default - ${DEFAULT_SCENARIO[description]}"
    echo "2. wds-on-host - ${WDS_ON_HOST_SCENARIO[description]}"
    echo "3. its-on-host - ${ITS_ON_HOST_SCENARIO[description]}"
    echo "4. combined - ${COMBINED_SCENARIO[description]}"
} 