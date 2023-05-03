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

# Usage: $0 --create-folder --verbose

# This script installs KCP-Edge workspace pluginsto a folder of choice
#
# Arguments:
# [--folder installation_folder] sets the installation folder, default: $PWD/kcp-edge
# [--create-folder] create the instllation folder, if it does not exist
# [-V|--verbose] verbose output

set -e

kcp_edge_folder=""
kcp_edge_create_folder="false"
verbose="false"

while (( $# > 0 )); do
    case "$1" in
	(--folder)
	    if (( $# > 1 ));
	    then { kcp_edge_folder="$2"; shift; }
	    else { echo "$0: missing installation folder" >&2; $0 -h; exit 1; }
	    fi;;
	(--create-folder)
        kcp_edge_create_folder="true";;
	(--verbose|-V)
        verbose="true";;
	(-h|--help)
	    echo "Usage: $0 [--folder installation_folder] [--create-folder] [-V|--verbose]"
	    exit 0;;
	(-*)
	    echo "$0: unknown flag" >&2; $0 -h ; exit 1;
	    exit 1;;
	(*)
	    echo "$0: unknown positional argument" >&2; $0 -h; exit 1;
	    exit 1;;
    esac
    shift
done

if [ "$kcp_edge_folder" == "" ]; then
    kcp_edge_folder="$PWD/kcp-edge"
fi

if [ ! -d "$kcp_edge_folder" ]
then
    if [ "$kcp_edge_create_folder" == "true" ]
    then
		if [ $verbose == "true" ]
		then
			echo "Creating folder: $kcp_edge_folder"
		fi
        mkdir -p "$kcp_edge_folder"
    else
        echo "Specified folder does not exist: $kcp_edge_folder" >&2; $0 -h; exit 1;
    fi
fi

if [ $verbose == "true" ]
then
    git clone https://github.com/yana1205/kcp kcp-edge-ws-plugins
    pushd kcp-edge-ws-plugins > /dev/null
    git checkout emc
    echo "Building the plugins..."
    make
else
    git clone -q https://github.com/yana1205/kcp kcp-edge-ws-plugins
    pushd kcp-edge-ws-plugins > /dev/null
    git checkout -q emc
    make > /dev/null
fi

if [ $verbose == "true" ]
then
	echo "Copying the plugins to $kcp_edge_folder"
fi
sudo cp ./bin/kubectl-* $kcp_edge_folder

if [ $verbose == "true" ]
then
	echo "Cleaning up..."
fi
popd > /dev/null
rm -rf kcp-edge-ws-plugins

if [[ ! ":$PATH:" == *":$kcp_edge_folder:"* ]]; then
	echo "Add KCP-Edge folder to your path: export PATH="\$PATH:$kcp_edge_folder""
fi