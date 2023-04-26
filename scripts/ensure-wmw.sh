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

# Purpose: ensure that a workload management workspace exists and has
# the needed APIBindings.

# Usage: kubectl ws parent; $0 [--with-kube | --no-kube] wmw_name

want_kube=true
wmw_name=""

while (( $# > 0 )); do
    case "$1" in
	(--no-kube)
	    want_kube="false";;
	(--with-kube)
	    want_kube="true";;
	(-h|--help)
	    echo "Usage: kubectl ws parent; $0 [--with-kube | --no-kube] wm_workspace_name"
	    exit 0;;
	(-*)
	    echo "Usage: kubectl ws parent; $0 [--with-kube | --no-kube] wm_workspace_name" >&2
	    exit 1;;
	(*)
	    if [ "$wmw_name" != "" ]; then
		echo "$0: too many positional arguments" >&2
		exit 1
	    fi
	    wmw_name="$1";;
    esac
    shift
done

if [ "$wmw_name" == "" ]; then
    echo "$0: workload management workspace name not specified" >&2
    exit 1
fi

if [ "$(kubectl ws . --short)" != root ]; then
    kubectl ws ..
    if kubectl get Workspace "$wmw_name" &> /dev/null; then
	echo "Warning: parent has a child with the given WMW name; are you sure the right workspace is current?" >&2
	sleep 15 # give a chance for ^C
    fi
    kubectl ws -
fi

set -e

if kubectl get Workspace "$wmw_name" &> /dev/null
then kubectl ws "$wmw_name"
else kubectl ws create "$wmw_name" --enter
fi

if ! kubectl get APIBinding bind-espw &> /dev/null; then
kubectl apply -f - <<EOF
apiVersion: apis.kcp.io/v1alpha1
kind: APIBinding
metadata:
  name: bind-espw
spec:
  reference:
    export:
      path: root:espw
      name: edge.kcp.io
EOF
fi

if [ "$want_kube" == true ] && ! kubectl get APIBinding bind-kube &> /dev/null; then
kubectl apply -f - <<EOF
apiVersion: apis.kcp.io/v1alpha1
kind: APIBinding
metadata:
  name: bind-kube
spec:
  reference:
    export:
      path: "root:compute"
      name: kubernetes
EOF
elif [ "$want_kube" == false ] && kubectl get APIBinding bind-kube &> /dev/null; then
     kubectl delete APIBinding bind-kube
fi
