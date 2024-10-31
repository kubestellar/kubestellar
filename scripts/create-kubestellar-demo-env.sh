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

kubestellar_version=0.25.0-rc.1
echo -e "KubeStellar Version: ${kubestellar_version}"

echo -e "Checking that pre-req softwares are installed..."
curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v${kubestellar_version}/hack/check_pre_req.sh | bash -s -- -V kflex ocm helm kubectl docker kind

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
            echo "CSR for $1 not found. Trying again..."
            sleep 20
        fi

    done
}
##########################################

echo -e "\nStarting environment clean up..."
echo -e "Starting cluster clean up..."

cluster_clean_up "kind delete cluster --name kubeflex" &
cluster_clean_up "kind delete cluster --name cluster1" &
cluster_clean_up "kind delete cluster --name cluster2" &
wait
echo -e "\033[33m✔\033[0m Cluster space clean up has been completed"

echo -e "\nStarting context clean up..."
context_clean_up
echo -e "\033[33m✔\033[0m Context space clean up completed"

echo -e "\nStarting the process to install KubeStellar core: kind-kubeflex..."
clusters=(cluster1 cluster2)
for cluster in "${clusters[@]}"; do
   (
     echo -e "Creating cluster ${cluster}..."
     kind create cluster --name "${cluster}" >/dev/null 2>&1 &&
     echo -e "\033[33m✔\033[0m ${cluster} creation and context setup complete"
   ) &
done
wait 

for cluster in "${clusters[@]}"; do
   kubectl config rename-context "kind-${cluster}" "${cluster}" >/dev/null 2>&1
done

for cluster in "${clusters[@]}"; do
  if kubectl config get-contexts | grep -w " ${cluster} " >/dev/null 2>&1; then
    echo -e "\033[33m✔\033[0m $cluster context exists."
  else
    if kubectl config rename-context "kind-${cluster}" "${cluster}" >/dev/null 2>&1; then
      echo -e "\033[33m✔\033[0m Renamed context 'kind-${cluster}' to '${cluster}'."
    else
      echo -e "Failed to rename context 'kind-${cluster}' to '${cluster}'. It may not exist."
    fi
  fi
done

echo -e "Creating KubeFlex cluster with SSL Passthrough"
curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v${kubestellar_version}/scripts/create-kind-cluster-with-SSL-passthrough.sh | bash -s -- --name kubeflex --port 9443 
echo -e "\033[33m✔\033[0m Completed KubeFlex cluster with SSL Passthrough"

echo -e "\nPulling container images local..."
images=("ghcr.io/loft-sh/vcluster:0.16.4"
        "rancher/k3s:v1.27.2-k3s1"
        "quay.io/open-cluster-management/registration-operator:v0.13.2"
        "docker.io/bitnami/postgresql:16.0.0-debian-11-r13")

for image in "${images[@]}"; do
    (
        docker pull "$image" && kind load docker-image "$image" --name kubeflex
    ) &
done

helm upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart \
    --version $kubestellar_version \
    --set-json='ITSes=[{"name":"its1"}]' \
    --set-json='WDSes=[{"name":"wds1"}]' \
    --set-json='verbosity.default=5'

kflex ctx --set-current-for-hosting # make sure the KubeFlex CLI's hidden state is right for what the Helm chart just did
kflex ctx --overwrite-existing-context wds1
kflex ctx --overwrite-existing-context its1

echo -e "\nWaiting for OCM cluster manager to be ready..."

wait-for-cmd "[[ \$(kubectl --context its1 get deployments.apps -n open-cluster-management -o jsonpath='{.status.readyReplicas}' cluster-manager 2>/dev/null) -ge 1 ]]"
echo -e "\033[33m✔\033[0m OCM cluster manager is ready"

echo -e "\nRegistering cluster 1 and 2 for remote access with KubeStellar Core..."

: set flags to "" if you have installed KubeStellar on an OpenShift cluster
flags="--force-internal-endpoint-lookup"
clusters=(cluster1 cluster2);
for cluster in "${clusters[@]}"; do
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
echo""
echo -e "\033[33m✔\033[0m Congratulations! Your KubeStellar demo environment is now ready to use."

cat <<"EOF"

Be sure to execute the following commands to set the shell variables expected in the example scenarios.

host_context=kind-kubeflex
its_cp=its1
its_context=its1
wds_cp=wds1
wds_context=wds1
wec1_name=cluster1
wec2_name=cluster2
wec1_context=$wec1_name
wec2_context=$wec2_name
label_query_both=location-group=edge
label_query_one=name=cluster1
EOF
