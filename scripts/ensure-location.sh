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

# Purpose: ensure the existence of a pair of SyncTarget and Location,
# and delete the Location named `default` if it appears.

# Usage: $0 objname labelname=labelvalue ...

# Invoke this with your kubectl configuration pointing at your chosen
# inventory management workspace.

# This will ensure that a SyncTarget and a Location exist with that
# name and having at least the given labels.  The SyncTarget will also
# have another label `id=$objname`, and the Location will select its
# SyncTarget based on just that one label.

case "$1" in
    (-h|--help)
	echo "$0 usage: objname labelname=labelvalue..."
	exit
esac

if (( $# < 2 )); then
    echo "$0: must be given an object name and at least one labelname=labelvalue pair" >&2
    exit 1
fi

objname="$1"
shift

if ! [[ "$objname" =~ ^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$ ]]; then
    echo "$0: objname not valid, must match POSIX extended re '^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$'" >& 2
    exit 1
fi

if [ "$objname" == default ]; then
    echo "$0: Sorry, 'default' is not an acceptable name for these objects" >&2
    exit 1
fi

if ! kubectl get SyncTarget "$objname" &> /dev/null; then
(cat <<EOF
apiVersion: workload.kcp.io/v1alpha1
kind: SyncTarget
metadata:
  name: "$objname"
  labels:
    id: "$objname"
EOF
) | kubectl create -f - || {
    echo "$0: Creation of SyncTarget failed" >&2
    exit 3
}
fi

if ! kubectl get Location "$objname" &> /dev/null; then
(cat <<EOF
apiVersion: scheduling.kcp.io/v1alpha1
kind: Location
spec:
  resource: {group: workload.kcp.io, version: v1alpha1, resource: synctargets}
  instanceSelector:
    matchLabels: {"id":"$objname"}
metadata:
  name: "$objname"
EOF
) | kubectl create -f - || {
    echo "$0: Creation of SyncTarget failed" >&2
    exit 3
}
fi

if kubectl get Location default &> /dev/null; then
    kubectl delete Location default &> /dev/null
fi


stlabs=$(kubectl get SyncTarget "$objname" -o json | jq .metadata.labels)
loclabs=$(kubectl get Location "$objname" -o json | jq .metadata.labels)

while (( $# > 0 )); do
    key=$(cut -d= -f1 <<<"$1")
    val=$(cut -d= -f2 <<<"$1")
    if [ "$1" != "${key}=${val}" ]; then
	echo "$0: syntax error for label name=value arg $1" >&2
	exit 4
    fi
    if ! [[ "$key" =~ ^[a-zA-Z0-9][a-zA-Z0-9_./-]*$ ]]; then
	echo "$0: syntax error in key, must match '^[a-zA-Z0-9][a-zA-Z0-9_./-]*$' and Kubernetes restrictions" >&2
	exit 4
    fi
    if ! [[ "$val" =~ ^[a-zA-Z0-9]([a-zA-Z0-9_.-]{0,61}[a-zA-Z0-9])?$ ]]; then
	echo "$0: syntax error in value, must match '^[a-zA-Z0-9]([a-zA-Z0-9_.-]{0,61}[a-zA-Z0-9])?$'" >&2
	exit 4
    fi
        if [ "$key" == id ]; then
	echo "$0: 'id' label may not be specified" >&2
	exit 4
    fi
    stval=$(jq -r <<<"$stlabs" ".[\"${key}\"]")
    if [ "$stval" != "$val" ]; then
	if ! kubectl label SyncTarget "$objname" "$1" ; then
	    echo "$0: failed to add label to SyncTarget" >&2
	    exit 4
	fi
    fi
    locval=$(jq -r <<<"$loclabs" ".[\"${key}\"]")
    if [ "$locval" != "$val" ]; then
	if ! kubectl label Location "$objname" "$1" ; then
	    echo "$0: failed to add label to Location" >&2
	    exit 4
	fi
    fi
    shift
done
