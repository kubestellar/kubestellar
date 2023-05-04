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

# Usage: $0 (-o file_pathname | --syncer-image container_image_ref | --espw-path workspace_path)* synctarget_name

# Purpose: For the given SyncTarget, (a) prepare the corresponding
# mailbox workspace for the syncer and (b) output the YAML that needs
# to be created in the edge cluster to install the syncer there.

# Assumption: only one SyncTarget has the given name.

# This script requires the edge-mc variant kubectl-kcp plugin to
# already exist at ../bin/kubectl-kcpforedgesyncer.

case "$1" in
    (-h|--help)
	echo "$0 usage: (-o file_pathname | --syncer-image container_image_ref | --espw-path workspace_path)* synctarget_name"
	exit
esac

scriptdir="$(dirname "$0")"
bindir="$(cd "$scriptdir"; cd ../bin; pwd)"

if ! [ -x "$bindir/kubectl-kcpforedgesyncer" ]; then
    echo "$0: $bindir/kubectl-kcpforedgesyncer does not exist; did you 'make build' or unpack a release archive here?" >&2
    exit 2
fi

stname=""
output=""
syncer_image="quay.io/kcpedge/syncer:v0.1.0"
espw_path="root:espw"

while (( $# > 0 )); do
    case "$1" in
	(-o)
	    if (( $# > 1 ));
	    then { output="$2"; shift; }
	    else { echo "$0: missing output path" >&2; exit 1; }
	    fi;;
	(--syncer-image)
	    if (( $# > 1 ));
	    then { syncer_image="$2"; shift; }
	    else { echo "$0: missing syncer image reference" >&2; exit 1; }
	    fi;;
	(--espw-path)
	    if (( $# > 1 ));
	    then { espw-path="$2"; shift; }
	    else { echo "$0: missing edge service provider workspace path" >&2; exit 1; }
	    fi;;
	(-*)
	    echo "$0: invalid flag $1" >&2
	    exit 1;;
	(*)
	    if [ "$stname" != "" ]; then
		echo "$0: too many positional arguments" >&2
		exit 1
	    fi
	    stname="$1"
    esac
    shift
done

if [ "$stname" == "" ]; then
    echo "$0: SyncTarget name not specified" >&2
    exit 1
fi

if [ "$output" == "" ]; then
    output="$stname-syncer.yaml"
fi

if [ "$espw_path" != "." ] && ! [[ "$espw_path" =~ [a-z0-9].* ]]; then
    echo "$0: espw-path ${espw_path@Q} is not valid" >&2
    exit 1
fi

set -e

kubectl ws "$espw_path"

if ! kubectl get APIExport edge.kcp.io &> /dev/null ; then
    echo "$0: it looks like ${espw_path@Q} is not the edge service provider workspace" >&2
    exit 2
fi

mbws=$(kubectl get Workspace -o json | jq -r ".items | .[] | .metadata | select(.annotations [\"edge.kcp.io/sync-target-name\"] == \"$stname\") | .name")

if [ $(wc -w <<<"$mbws") = 0 ]; then
    sleep 15 : maybe the mailbox controller is slow, give it a chance
    mbws=$(kubectl get Workspace -o json | jq -r ".items | .[] | .metadata | select(.annotations [\"edge.kcp.io/sync-target-name\"] == \"$stname\") | .name")
fi

if [ $(wc -w <<<"$mbws") = 0 ]; then
    echo "$0: did not find a mailbox workspace for SyncTarget $stname; is the mailbox controller running?" >&2
    exit 5
fi

if [ $(wc -w <<<"$mbws") != 1 ]; then
    echo "$0: had trouble identifying the mailbox controller for SyncTarget $stname; what does \`kubectl get Workspace -o \"custom-columns=NAME:.metadata.name,SYNCTARGET:.metadata.annotations['edge\.kcp\.io/sync-target-name']\"\` tell you?" >&2
    exit 6
fi

kubectl ws "$mbws"

"$bindir/kubectl-kcpforedgesyncer" workload edge-sync "$stname" --syncer-image "$syncer_image" -o "$output"
