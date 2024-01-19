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

set -x # echo so users can understand what is happening
set -e # exit on error
SRC_DIR="$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)"

:
: -------------------------------------------------------------------------
: Create a Kind hosting cluster with nginx ingress controller and KubeFlex operator
:
kflex init --create-kind
: Kubeflex kind cluster created.

:
: -------------------------------------------------------------------------
: 'Create an inventory & mailbox space of type vcluster running OCM (Open Cluster Management) directly in KubeFlex. Note that -p ocm runs a post-create hook on the vcluster control plane which installs OCM on it.'
:
kflex create imbs1 --type vcluster -p ocm
: imbs1 created.

:
: -------------------------------------------------------------------------
: Install singleton status return addon in IMBS1
:
wait-for-cmd kubectl --context imbs1 api-resources "|" grep managedclusteraddons
helm --kube-context imbs1 upgrade --install status-addon -n open-cluster-management oci://quay.io/pdettori/status-addon-chart --version 0.1.0


:
: -------------------------------------------------------------------------
: Create a Workload Description Space wds1 directly in KubeFlex.
:
kflex create wds1
kubectl --context kind-kubeflex label cp wds1 kflex.kubestellar.io/cptype=wds

cd "${SRC_DIR}/../../.."
pwd
make ko-build-local
make install-local-chart KUBE_CONTEXT=kind-kubeflex
cd -
echo "wds1 created."

:
: -------------------------------------------------------------------------
: Create clusters and register with OCM
:
function create_cluster() {
  cluster=$1
  kind create cluster --name $cluster
  kubectl config rename-context kind-${cluster} $cluster
  clusteradm --context imbs1 get token | grep '^clusteradm join' | sed "s/<cluster_name>/${cluster}/" | awk '{print $0 " --context '${cluster}' --force-internal-endpoint-lookup"}' | sh
}

create_cluster cluster1
create_cluster cluster2

: Wait for csrs in imbs1
wait-for-cmd '(($(kubectl --context imbs1 get csr 2>/dev/null | grep -c Pending) >= 2))'

clusteradm --context imbs1 accept --clusters cluster1
clusteradm --context imbs1 accept --clusters cluster2

kubectl --context imbs1 get managedclusters
kubectl --context imbs1 label managedcluster cluster1 location-group=edge name=cluster1
kubectl --context imbs1 label managedcluster cluster2 location-group=edge name=cluster2

:
: -------------------------------------------------------------------------
: Get all deployments and statefulsets running in the hosting cluster
:
kubectl --context kind-kubeflex get deployments,statefulsets --all-namespaces

:
: -------------------------------------------------------------------------
: "Get available clusters with label location-group=edge and check there are two of them"
:
if ! expect-cmd-output 'kubectl --context imbs1 get managedclusters -l location-group=edge' 'wc -l | grep -wq 3'
then
    echo "Failed to see two clusters."
    exit 1
fi

