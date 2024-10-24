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

set -e

export kubestellar_version=0.25.0-rc.1
echo -e "KubeStellar Version: ${kubestellar_version}"

echo -e "Checking that pre-req softwares are installed..."
curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v${kubestellar_version}/hack/check_pre_req.sh | bash -s -- -V kflex ocm helm kubectl docker kind

output=$(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v${kubestellar_version}/hack/check_pre_req.sh | bash -s -- -A -V kflex ocm helm kubectl docker kind)

##########################################
function wait-for-cmd() (
    cmd="$@"
    wait_counter=0
    while ! (eval "$cmd") ; do
        if (($wait_counter > 100)); then
            echo "Failed to ${cmd}."
            exit 1
        fi
        ((wait_counter += 1))
        sleep 5
    done
)

cluster_clean_up() {
    error_message=$(eval "$1" 2>&1)
    if [ $? -ne 0 ]; then
        echo "clean up failed. Error:"
        echo "$error_message"
    fi
}

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
##########################################

echo -e "\nStarting environment clean up..."
echo -e "Starting cluster clean up..."

cluster_clean_up "kind delete cluster --name kubeflex"
cluster_clean_up "kind delete cluster --name cluster1"
cluster_clean_up "kind delete cluster --name cluster2"
echo -e "Cluster space clean up has been completed"

echo -e "\nStarting context clean up..."

context_clean_up
echo "Context space clean up completed"

echo -e "\nStarting the process to install KubeStellar core: kind-kubeflex..."

curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v${KUBESTELLAR_VERSION}/scripts/create-kind-cluster-with-SSL-passthrough.sh | bash -s -- --name kubeflex --port 9443

helm upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart \
    --version $KUBESTELLAR_VERSION \
    --set-json='ITSes=[{"name":"its1"}]' \
    --set-json='WDSes=[{"name":"wds1"}]' \
    --set-json='verbosity.default=5'

kubectl config delete-context its1 || true
kflex ctx its1
kubectl config delete-context wds1 || true
kflex ctx wds1

# switch back to its1 context to crete 2 remote clusters, add OCM agents to them, and register them back to the KubeStellar core
kflex ctx its1

echo -e "\nWaiting for OCM cluster manager to be ready..."

wait-for-cmd "[[ \$(kubectl --context its1 get deployments.apps -n open-cluster-management -o jsonpath='{.status.readyReplicas}' cluster-manager 2>/dev/null) -ge 1 ]]"
echo -e "OCM cluster manager is ready"

echo -e "\nCreating cluster and context for cluster 1 and 2..."

: set flags to "" if you have installed KubeStellar on an OpenShift cluster
flags="--force-internal-endpoint-lookup"
clusters=(cluster1 cluster2);
for cluster in "${clusters[@]}"; do
   kind create cluster --name ${cluster}
   kubectl config rename-context kind-${cluster} ${cluster}
   clusteradm --context its1 get token | grep '^clusteradm join' | sed "s/<cluster_name>/${cluster}/" | awk '{print $0 " --context '${cluster}' --singleton '${flags}'"}' | sh
done

echo -e "Checking that the CSR for cluster 1 and 2 appears..."

echo""
echo "Waiting for cluster1 and cluster2 to be ready and then approve their CSRs"
checking_cluster cluster1
checking_cluster cluster2

echo""
echo "Checking the new clusters are in the OCM inventory and label them"
kubectl --context its1 get managedclusters
kubectl --context its1 label managedcluster cluster1 location-group=edge name=cluster1
kubectl --context its1 label managedcluster cluster2 location-group=edge name=cluster2

