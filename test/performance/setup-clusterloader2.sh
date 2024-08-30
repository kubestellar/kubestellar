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

set -x # echo so that users can understand what is happening
set -e # exit on error

env="kind"

if [ "$1" == "--env" ]; then
    env="$2"
    shift 2
fi

:
: -------------------------------------------------------------------------
: "Configuring clusterloader2 to generate KS performance workloads"

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)

cp $SCRIPT_DIR/cleanup.sh $CL2_DIR/
cp $SCRIPT_DIR/workloads/modules/*  $CL2_DIR/testing/load/modules
cp $SCRIPT_DIR/workloads/performance-test-config.yaml  $CL2_DIR/testing/load

if [ $env == "kind" ];then
   cp $SCRIPT_DIR/workloads/cluster-density-k8s/*  $CL2_DIR/testing/load

elif [ $env == "ocp" ];then
   cp $SCRIPT_DIR/workloads/cluster-density-ocp/*  $CL2_DIR/testing/load
else
   echo "$0: unknown flag option" >&2 ;
   echo "Usage: $0 [--env ( kind | ocp )]" >& 2
   exit 1
fi