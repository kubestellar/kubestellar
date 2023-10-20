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

# Purpose: ensure the existence of a pair of SyncTarget and Location,
# and delete the Location named `default` if it appears.

# Usage: $0 ($kubectl_flag | --imw ws_path | -X)* objname labelname=labelvalue ...

# This will ensure that a SyncTarget and a Location exist with that
# name and having at least the given labels.  The SyncTarget will also
# have another label `id=$objname`, and the Location will select its
# SyncTarget based on just that one label.

imw=.
objname=""
labels=()
kubectl_flags=()

function add_label() {
    key=$(cut -d= -f1 <<<"$1")
    val=$(cut -d= -f2 <<<"$1")
    if [ "$1" != "${key}=${val}" ]; then
	echo "$0: syntax error for label name=value arg $1" >&2
	exit 1
    fi
    if ! [[ "$key" =~ ^[a-zA-Z0-9][a-zA-Z0-9_./-]*$ ]]; then
	echo "$0: syntax error in key, must match '^[a-zA-Z0-9][a-zA-Z0-9_./-]*$' and Kubernetes restrictions" >&2
	exit 1
    fi
    if ! [[ "$val" =~ ^[a-zA-Z0-9]([a-zA-Z0-9_.-]{0,61}[a-zA-Z0-9])?$ ]]; then
	echo "$0: syntax error in value, must match '^[a-zA-Z0-9]([a-zA-Z0-9_.-]{0,61}[a-zA-Z0-9])?$'" >&2
	exit 1
    fi
    if [ "$key" == id ]; then
	echo "$0: 'id' label may not be specified" >&2
	exit 1
    fi
    labels[${#labels[*]}]="$key"
    labels[${#labels[*]}]="$val"
}

while (( $# > 0 )); do
    case "$1" in
	(-h|--help)
	    echo "Usage: kubectl kubestellar ensure location (\$kubectl_flag | --imw ws_path | -X)* objname labelname=labelvalue ..."
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
	    kubectl_flags[${#kubectl_flags[*]}]="$1"
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
	    else add_label "$1"
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

if [ "$objname" == default ]; then
    echo "$0: Sorry, 'default' is not an acceptable name for these objects" >&2
    exit 1
fi

if [ ${#labels[*]} == 0 ]; then
    echo "$0: must be given at least one label key=value" >&2
    exit 1
fi

set -e

echo "--- current directory is $PWD"
cwsi=$(kubectl ws "${kubectl_flags[@]}" .)
cwsi=${cwsi#*'"'}; cwsi=${cwsi%'"'*}
echo "current ws is $cwsi"
echo "imw is $imw"
if [ "$imw" != "." ]
then kubectl ws "${kubectl_flags[@]}" "$imw"
     cwsi=$imw
fi
echo "current cwsi is $cwsi"
if ! kubectl "${kubectl_flags[@]}" get apibinding "edge.kubestellar.io" &> /dev/null; then
    kubectl kcp "${kubectl_flags[@]}" bind apiexport root:espw:edge.kubestellar.io
    echo "bound apiexport root:espw:edge.kubestellar.io into $cwsi:edge.kubestellar.io"
else echo "edge.kubestellar.io apibinding exists in workspace $cwsi"
fi

if ! kubectl "${kubectl_flags[@]}" get synctargets.edge.kubestellar.io "$objname" &> /dev/null; then
(cat <<EOF
apiVersion: edge.kubestellar.io/v2alpha1
kind: SyncTarget
metadata:
  name: "$objname"
  labels:
    id: "$objname"
EOF
) | kubectl "${kubectl_flags[@]}" create -f - || {
    echo "$0: Creation of SyncTarget failed" >&2
    exit 3
}
fi

if ! kubectl "${kubectl_flags[@]}" get locations.edge.kubestellar.io "$objname" &> /dev/null; then
(cat <<EOF
apiVersion: edge.kubestellar.io/v2alpha1
kind: Location
spec:
  resource: {group: edge.kubestellar.io, version: v2alpha1, resource: synctargets}
  instanceSelector:
    matchLabels: {"id":"$objname"}
metadata:
  name: "$objname"
EOF
) | kubectl "${kubectl_flags[@]}" create -f - || {
    echo "$0: Creation of SyncTarget failed" >&2
    exit 3
}
fi

if kubectl "${kubectl_flags[@]}" get locations.edge.kubestellar.io default &> /dev/null; then
    kubectl "${kubectl_flags[@]}" delete locations.edge.kubestellar.io default &> /dev/null
fi


stlabs=$(kubectl "${kubectl_flags[@]}" get synctargets.edge.kubestellar.io  "$objname" -o json | jq .metadata.labels)
loclabs=$(kubectl "${kubectl_flags[@]}" get locations.edge.kubestellar.io "$objname" -o json | jq .metadata.labels)

let index=0 1
while (( index < ${#labels[*]} )); do
    key=${labels[index]}
    val=${labels[index+1]}
    let index=index+2
    stval=$(jq -r <<<"$stlabs" ".[\"${key}\"]")
    if [ "$stval" != "$val" ]; then
	    if ! kubectl "${kubectl_flags[@]}" label --overwrite synctargets.edge.kubestellar.io "$objname" "${key}=${val}" ; then
	        echo "$0: failed to add label to SyncTarget" >&2
	        exit 4
	    fi
    fi
    locval=$(jq -r <<<"$loclabs" ".[\"${key}\"]")
    if [ "$locval" != "$val" ]; then
	    if ! kubectl "${kubectl_flags[@]}" label --overwrite locations.edge.kubestellar.io "$objname" "${key}=${val}" ; then
	        echo "$0: failed to add label to Location" >&2
	        exit 4
	   fi
    fi
done
