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

# Usage: $0 (-o file_pathname | --syncer-image container_image_ref )* synctarget_name

# Purpose: For the given SyncTarget, (a) prepare the corresponding
# mailbox workspace for the syncer and (b) output the YAML that needs
# to be created in the edge cluster to install the syncer there.

# Assumption: only one SyncTarget has the given name.

# Invoke this with the edge service provider workspace (ESPW) current,
# and it will (if successful) terminate with the mailbox workspace current.

case "$1" in
    (-h|--help)
	echo "$0 usage: (-o file_pathname | --syncer-image container_image_ref )* synctarget_name"
	exit
esac

stname=""
output=""
syncer_image="quay.io/kcpedge/syncer:dev-2023-04-18"

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

if ! kubectl get APIExport edge.kcp.io &> /dev/null ; then
    echo "$0: it looks like the ESPW is not current" >&2
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

set -e

kubectl ws "$mbws"

scriptdir="$(dirname "$0")"

"$scriptdir/ensure-syncer-plugin.sh"

kcpdir="$scriptdir/../build/syncer-kcp"

"$kcpdir/bin/kubectl-kcp" workload edge-sync "$stname" --syncer-image "$syncer_image" -o "$output"
