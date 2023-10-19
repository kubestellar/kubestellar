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

# Purpose: ensure that a workload management workspace exists and has
# the needed APIBindings.

# Usage: $0 [-X] kubectl_flag... wmw_name

wmw_name=""

while (( $# > 0 )); do
    case "$1" in
	(-h|--help)
	    echo "Usage: kubectl kubestellar remove wmw [-X] kubectl_flag... wm_workspace_name"
	    exit 0;;
	(-X) set -o xtrace;;
	(--context*)
	    # TODO: support --context
	    echo "$0: --context flag not supported" >&2; exit 1;;
	(--*=*|-?=*)
	    kubectl_flags[${#kubectl_flags[*]}]="$1";;
	(--*|-?)
	    kubectl_flags[${#kubectl_flags[*]}]="$1"
	    if (( $# > 1 )); then 
		 kubectl_flags[${#kubectl_flags[*]}]="$2"
		 shift
	    fi;;
	(-*)
	    echo "$0: flag syntax error" >&2
	    exit 1;;
	(*)
	    if [ -z "$wmw_name" ]
	    then wmw_name="$1"
	    else echo "$0: too many positional arguments" >&2
		 exit 1
	    fi
    esac
    shift
done

if [ "$wmw_name" == "" ]; then
    echo "$0: workload management workspace name not specified" >&2
    exit 1
fi

set -e

kubectl ws "${kubectl_flags[@]}" root
if kubectl "${kubectl_flags[@]}" get workspaces.tenancy.kcp.io "$wmw_name" &>/dev/null
then kubectl "${kubectl_flags[@]}" delete workspaces.tenancy.kcp.io "$wmw_name"
fi

