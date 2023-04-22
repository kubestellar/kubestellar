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

# Usage: $0 objname

# Purpose: ensure the SyncTarget and Location with the given name do not exist.

# Invoke this with `kubectl` configured to manipulate your chosen
# inventory management workspace.

case "$1" in
    (-h|--help)
	echo "$0 usage: objname"
	exit
esac

if (( $# != 1 )); then
    echo "$0: must be given one argument, an object name" >&2
    exit 1
fi

set -e

if kubectl get synctargets.workload.kcp.io "$1" &> /dev/null
then kubectl delete synctargets.workload.kcp.io "$1"
fi

if kubectl get locations.scheduling.kcp.io "$1" &> /dev/null
then kubectl delete locations.scheduling.kcp.io "$1"
fi
