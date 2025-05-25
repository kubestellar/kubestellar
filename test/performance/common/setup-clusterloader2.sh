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

env="k8s"
exp_type="s" # accepted values: a) l = long-running test; b) s = short-running test

while [ $# != 0 ]; do
    case "$1" in
        (-h|--help) echo "$0 usage: (--env ( kind | ocp ) | --exp ( s | l ))*"
                    exit;;
        (--env)
          if (( $# > 1 )); then
            env="$2"
            shift
          else
            echo "Missing env value" >&2
            exit 1;
          fi;;
        (--exp)
          if (( $# > 1 )); then
            exp_type="$2"
            shift
          else
            echo "Missing exp value" >&2
            exit 1;
          fi;;
    esac
    shift
done

:
: -------------------------------------------------------------------------
: "Configuring clusterloader2 to generate KS performance workloads"

if [ $CL2_DIR == "" ];then
   echo "Set the variable CL2_DIR to the path of the subdirectory clusterloader2/ of a cloned `https://github.com/kubernetes/perf-tests/` repo."
   exit 1;
fi

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)
cp $SCRIPT_DIR/cleanup.sh $CL2_DIR/

if [ $exp_type == "s" ];then
   cp $SCRIPT_DIR/../short-running-tests/workloads/modules/*  $CL2_DIR/testing/load/modules
   cp $SCRIPT_DIR/../short-running-tests/workloads/performance-test-config.yaml  $CL2_DIR/testing/load

   if [ $env == "k8s" ];then
      cp $SCRIPT_DIR/../short-running-tests/workloads/cluster-density-k8s/*  $CL2_DIR/testing/load

   elif [ $env == "ocp" ];then
      cp $SCRIPT_DIR/../short-running-tests/workloads/cluster-density-ocp/*  $CL2_DIR/testing/load
   fi

elif [ $exp_type == "l" ];then
   cp $SCRIPT_DIR/../long-running-tests/workloads/long-duration-exp-config.yaml  $CL2_DIR/testing/load
   cp $SCRIPT_DIR/../long-running-tests/workloads/long-duration-exp-workload.yaml $CL2_DIR/testing/load
   cp $SCRIPT_DIR/../long-running-tests/workloads/long-duration-exp-setup.yaml $CL2_DIR/testing/load
   cp $SCRIPT_DIR/../long-running-tests/workloads/modules/*  $CL2_DIR/testing/load/modules
else
   echo "$0: --exp must be 's' or 'l'" >&2
   exit 1;
fi
