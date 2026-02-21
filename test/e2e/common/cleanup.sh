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
: "Cleaning up from previous run of an e2e test"

if [ $env == "kind" ];then
    kind delete cluster --name cluster1
    kind delete cluster --name cluster2
    kind delete cluster --name kubeflex
    kubectl config delete-context cluster1 || true
    kubectl config delete-context cluster2 || true

elif [ $env == "k3d" ];then
    k3d cluster delete cluster1 || true
    k3d cluster delete cluster2 || true
    k3d cluster delete kubeflex || true
    kubectl config delete-context cluster1 || true
    kubectl config delete-context cluster2 || true

elif [ $env == "ocp" ];then
    # Unregister the managed clusters
    function unregister_cluster() {
        cluster=$1

        kubectl --context $cluster delete ns nginx --ignore-not-found
        clusteradm unjoin --cluster-name $cluster  2> /dev/null
        kubectl --context $cluster delete ns open-cluster-management open-cluster-management-agent open-cluster-management-agent-addon --ignore-not-found
    }

    unregister_cluster cluster1
    unregister_cluster cluster2

    # To uninstall KubeFlex, first ensure you remove all you control planes:
    kubectl config use-context kscore
    if kubectl get cps; then
       kubectl delete cps --all
    fi

    helm delete -n kubeflex-system kubeflex-operator --ignore-not-found
    helm delete -n kubeflex-system postgres --ignore-not-found
    kubectl -n kubeflex-system delete pvc data-postgres-postgresql-0 --ignore-not-found
    kubectl delete ns kubeflex-system --ignore-not-found

    # Unset the kubeconfig contexts
    kubectl config unset contexts.its1
    kubectl config unset contexts.wds1
    kubectl config unset contexts.wds2
else
   echo "$0: unknown flag option" >&2 ;
   echo "Usage: $0 [--env kind | k3d | ocp]" >& 2
   exit 1
fi






