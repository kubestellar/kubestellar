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

ns="ks-monitoring"
ctx="kind-kubeflex"



while [ $# != 0 ]; do
    case "$1" in
        (-h|--help) echo "$0 usage: (--cluster-context (e.g., --cluster-context core-cluster (default value: kind-kubeflex)) | --ns (e.g., --ns monitoring (default value: ks-monitoring)))*"
                    exit;;
        (--ns)
          if (( $# > 1 )); then
            ns="$2"
            shift
          else
            echo "Missing ns value" >&2
            exit 1;
          fi;;
        (--cluster-context)
          if (( $# > 1 )); then
            ctx="$2"
            shift
          else
            echo "Missing cluster-context value" >&2
            exit 1;
          fi;;
    esac
    shift
done

helm --kube-context $ctx delete -n $ns grafana --ignore-not-found
helm --kube-context $ctx delete -n $ns pyroscope --ignore-not-found
helm --kube-context $ctx delete -n $ns prometheus --ignore-not-found
helm --kube-context $ctx delete -n $ns minio --ignore-not-found
helm --kube-context $ctx delete -n $ns thanos --ignore-not-found

kubectl --context $ctx -n $ns delete deployment minio --ignore-not-found
kubectl --context $ctx -n $ns delete pvc --all
kubectl --context $ctx delete ns $ns  --ignore-not-found
