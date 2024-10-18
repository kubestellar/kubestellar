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

# Deploys Kubestellar environment for demo purposes.

# Exit immediately if a command exits with a non-zero status
set -e

# Check if required software is installed
kubestellar_version=0.25.0-rc.1

echo -e "Checking that pre-req softwares are installed..."
curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v${kubestellar_version}/hack/check_pre_req.sh | bash -s -- -V kflex ocm helm kubectl docker kind


# Run a more detailed check and store the output
output=$(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v${kubestellar_version}/hack/check_pre_req.sh | bash -s -- -A -V kflex ocm helm kubectl docker kind


# Start environment cleanup
echo -e "\nStarting environment clean up..."
echo -e "Starting cluster clean up..."

# Function to clean up clusters
cluster_clean_up() {
    error_message=$(eval "$1" 2>&1)
    if [ $? -ne 0 ]; then
        echo "clean up failed. Error:"
        echo "$error_message"
    fi
}

# Clean up specific clusters
cluster_clean_up "kind delete cluster --name kubeflex"
cluster_clean_up "kind delete cluster --name cluster1"
cluster_clean_up "kind delete cluster --name cluster2"
echo -e "Cluster space clean up has been completed"

# Start context cleanup
echo -e "\nStarting context clean up..."

# Function to clean up contexts
context_clean_up() {
    output=$(kubectl config get-contexts -o name)

    while IFS= read -r line; do
        if [ "$line" == "kind-kubeflex" ]; then 
            echo "Deleting kind-kubeflex context..."
            kubectl config delete-context kind-kubeflex

        elif [ "$line" == "cluster1" ]; then
            echo "Deleting cluster1 context..."
            kubectl config delete-context cluster1

        elif [ "$line" == "cluster2" ]; then
            echo "Deleting cluster2 context..."
            kubectl config delete-context cluster2
        
        elif [ "$line" == "its1" ]; then
            echo "Deleting its1 context..."
            kubectl config delete-context its1
        
        elif [ "$line" == "wds1" ]; then
            echo "Deleting wds1 context..."
            kubectl config delete-context wds1
        
        fi

    done <<< "$output"
}

# Run context cleanup
context_clean_up
echo "Context space clean up completed"

# Install KubeStellar core
echo -e "\nStarting the process to install KubeStellar core: kind-kubeflex..."

curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v${kubestellar_version}/scripts/create-kind-cluster-with-SSL-passthrough.sh | bash -s -- --name kubeflex --port 9443

helm upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart --version $kubestellar_version --set-json='ITSes=[{"name":"its1"}]' --set-json='WDSes=[{"name":"wds1"}]'


# Wait for non-host KubeFlex Control Planes to be ready
echo -e "\nWaiting for new non-host KubeFlex Control Planes to be Ready:"
for cpname in its1 wds1; do
  while [[ `kubectl get cp $cpname -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}'` != "True" ]]; do
    echo "Waiting for \"$cpname\"..."
    sleep 5
  done
  echo "\"$cpname\" is ready."
done

# Set up contexts for its1 and wds1
kubectl config delete-context its1 || true
kflex ctx its1
kubectl config delete-context wds1 || true
kflex ctx wds1
kflex ctx

# Create clusters and contexts for cluster1 and cluster2
echo -e "\nCreating cluster and context for cluster 1 and 2..."

# Set flags for kind clusters (not for OpenShift clusters)
flags="--force-internal-endpoint-lookup"
clusters=(cluster1 cluster2);
for cluster in "${clusters[@]}"; do
   kind create cluster --name ${cluster}
   kubectl config rename-context kind-${cluster} ${cluster}
   clusteradm --context its1 get token | grep '^clusteradm join' | sed "s/<cluster_name>/${cluster}/" | awk '{print $0 " --context '${cluster}' --singleton '${flags}'"}' | sh
done

# Check and approve CSRs for cluster1 and cluster2
echo -e "Checking that the CSR for cluster 1 and 2 appears..."

checking_cluster() {
    found=false
    
    while true; do

        output=$(kubectl --context its1 get csr)

        while IFS= read -r line; do

            if echo "$line" | grep -q $1; then
                echo "$1 has been found, approving CSR"
                clusteradm --context its1 accept --clusters "$1"
                found=true
                break
            fi

        done <<< "$output"

        if [ "$found" = true ]; then
            break
            
        else
            echo "$1 not found. Trying again..."
            sleep 5
        fi

    done
}

echo""
echo "Approving CSR for cluster1 and cluster2..."
checking_cluster cluster1
checking_cluster cluster2

# Check and label the new clusters in the OCM inventory
echo""
echo "Checking the new clusters are in the OCM inventory and label them"
kubectl --context its1 get managedclusters
kubectl --context its1 label managedcluster cluster1 location-group=edge name=cluster1
kubectl --context its1 label managedcluster cluster2 location-group=edge name=cluster2
