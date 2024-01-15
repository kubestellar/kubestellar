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

set -e # exit on error

echo "Create a Kind hosting cluster with nginx ingress controller and KubeFlex operator"
echo "-------------------------------------------------------------------------"
kflex init --create-kind

echo "Create an inventory & mailbox space of type vcluster running OCM (Open Cluster Management) directly in KubeFlex. Note that -p ocm runs a post-create hook on the vcluster control plane which installs OCM on it."
echo "-------------------------------------------------------------------------"
kflex create imbs1 --type vcluster -p ocm

echo "Create a Workload Description Space wds1 directly in KubeFlex."
echo "-------------------------------------------------------------------------"
kflex create wds1 
kubectl config use-context kind-kubeflex
kubectl label cp wds1 kflex.kubestellar.io/cptype=wds

cd ../../../
make ko-build
make install-local-chart
cd -

echo "Create clusters and register with OCM"
echo "-------------------------------------------------------------------------"
function create_cluster() {
  cluster=$1
  kind create cluster --name $cluster
  kubectl config rename-context kind-${cluster} $cluster
  clusteradm --context imbs1 get token | grep '^clusteradm join' | sed "s/<cluster_name>/${cluster}/" | awk '{print $0 " --context '${cluster}' --force-internal-endpoint-lookup"}' | sh
}

create_cluster cluster1
create_cluster cluster2

kubectl config use-context imbs1

echo "Wait for csr on imbs1"
waitCounter=0
while (($(kubectl --context imbs1 get csr 2>/dev/null | grep -c "Pending") < 2)); do
  if (($waitCounter > 180)); then
    echo "Failed to observe deployment on ${cluster}."
    exit 1 
  fi
  ((waitCounter += 1))
  sleep 1
done

clusteradm --context imbs1 accept --clusters cluster1
clusteradm --context imbs1 accept --clusters cluster2

kubectl --context imbs1 get managedclusters
kubectl --context imbs1 label managedcluster cluster1 location-group=edge
kubectl --context imbs1 label managedcluster cluster2 location-group=edge

echo "Get all deployments and statefulsets running in the hosting cluster"
echo "-------------------------------------------------------------------------"
kubectl --context kind-kubeflex get deployments,statefulsets --all-namespaces

echo "Get available clusters with label location-group=edge and check there are two of them"
echo "-------------------------------------------------------------------------"
kubectl --context imbs1 get managedclusters -l location-group=edge | tee out
if (("$(wc -l < out)" != "3")); then
  echo "Failed to see two clusters."
  exit 1
fi

