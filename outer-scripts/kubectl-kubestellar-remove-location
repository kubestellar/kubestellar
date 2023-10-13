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

# Usage: $0 ($kubectl_flag | --imw ws_path | -X)* objname

# Purpose: ensure the SyncTarget and Location with the given name do not exist.

imw=.
objname=""
kubectl_flags=()

while (( $# > 0 )); do
    case "$1" in
	(-h|--help)
	    echo "Usage: kubectl kubestellar remove location (\$kubectl_flag | --imw ws_path | -X)* objname"
	    exit 0;;
	(-X) set -o xtrace;;
	(--imw)
	    if (( $# >1 ))
	    then imw="$2"; shift
	    else echo "$0: missing IMW pathname" >&2; exit 1
	    fi;;
	(--context*)
	    # TODO: support --context
	    echo "$0: --context flag not supported" >&2; exit 1;;
	(--*=*|-?=*)
	    kubectl_flags[${#kubectl_flags[*]}]="$1";;
	(--*|-?)
	    kubectl_flags[${#kubectl_flags[*]}]="$1";
	    if (( $# > 1 )); then 
		 kubectl_flags[${#kubectl_flags[*]}]="$2"
		 shift
	    fi;;
	(-*)
	    echo "$0: flag syntax error" >&2
	    exit 1;;
	(*)
	    if [ -z "$objname" ]
	    then objname="$1"
	    else echo "$0: only one positional argument is allowed" >&2
		 exit 1
	    fi
    esac
    shift
done

if [ -z "$objname" ]; then
    echo "$0: must be given a non-empty object name" >&2
    exit 1
fi

if ! [[ "$objname" =~ ^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$ ]]; then
    echo "$0: objname not valid, must match POSIX extended re '^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$'" >& 2
    exit 1
fi

set -e

# this assumes that an APIBinding exists for edge.kubestellar.io objects
if [ "$imw" != "." ]
then kubectl ws "${kubectl_flags[@]}" "$imw"
fi

if kubectl "${kubectl_flags[@]}" get synctargets.edge.kubestellar.io "$objname" &> /dev/null
then kubectl "${kubectl_flags[@]}" delete synctargets.edge.kubestellar.io "$objname"
fi

if kubectl "${kubectl_flags[@]}" get locations.edge.kubestellar.io "$objname" &> /dev/null
then kubectl "${kubectl_flags[@]}" delete locations.edge.kubestellar.io "$objname"
fi
