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

set -e

# Script info
SCRIPT_NAME="create-kubestellar-demo-env.sh"

# Default Kubernetes platform parameter
k8s_platform="kind"

# Parse command-line arguments
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --platform) k8s_platform="$2"; shift ;;
        -X) set -x ;;
        -h|--help)
            echo "Usage: ${SCRIPT_NAME} [--platform <kind|k3d>] [-X] [-h|--help]" >&2
            exit 0
            ;;
        *)
            echo "Unknown parameter passed: $1" >&2
            echo "Usage: ${SCRIPT_NAME} [--platform <kind|k3d>] [-X] [-h|--help]" >&2
            exit 1
            ;;
    esac
    shift
done

if [[ "$k8s_platform" != "kind" && "$k8s_platform" != "k3d" ]]; then
    echo "Invalid platform specified: $k8s_platform"
    echo "Supported platforms are: kind, k3d"
    exit 1
fi

echo "Selected Kubernetes platform: $k8s_platform"

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to wait for a kind cluster to be fully deleted
wait_for_kind_cluster_deletion() {
    local cluster_name=$1
    local context_name="kind-${cluster_name}"
    
    # Wait for the cluster context to be unreachable (indicating deletion)
    echo "Waiting for kind cluster $cluster_name to be fully deleted..."
    kubectl wait --for=delete nodes --all --context="$context_name" --timeout=60s 2>/dev/null || true
    
    # Additional check using kind CLI as fallback
    local max_wait=30
    local wait_time=0
    while kind get clusters 2>/dev/null | grep -q "^${cluster_name}$"; do
        if [ $wait_time -ge $max_wait ]; then
            echo "Warning: Timed out waiting for kind cluster $cluster_name to be fully deleted"
            break
        fi
        sleep 1
        wait_time=$((wait_time + 1))
    done
}

# Function to wait for a k3d cluster to be fully deleted
wait_for_k3d_cluster_deletion() {
    local cluster_name=$1
    local context_name="k3d-${cluster_name}"
    
    # Wait for the cluster context to be unreachable (indicating deletion)
    echo "Waiting for k3d cluster $cluster_name to be fully deleted..."
    kubectl wait --for=delete nodes --all --context="$context_name" --timeout=60s 2>/dev/null || true
    
    # Additional check using k3d CLI as fallback
    local max_wait=30
    local wait_time=0
    while k3d cluster list 2>/dev/null | grep -q "^${cluster_name}\\s"; do
        if [ $wait_time -ge $max_wait ]; then
            echo "Warning: Timed out waiting for k3d cluster $cluster_name to be fully deleted"
            break
        fi
        sleep 1
        wait_time=$((wait_time + 1))
    done
}

echo -e "Checking container runtime..."
if ! dunsel=$(docker ps 2>&1); then
    echo "Error: The script cannot continue because Docker or Podman is not running. Please start your container runtime before running the script again."
    exit 1
fi
echo "Container runtime is running."

kubestellar_version=0.29.0
echo -e "KubeStellar Version: ${kubestellar_version}"

echo -e "Checking that pre-req softwares are installed..."
if [ "$k8s_platform" == "kind" ]; then
    curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v${kubestellar_version}/scripts/check_pre_req.sh | bash -s -- --assert -V kflex ocm helm kubectl docker kind
else
    curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v${kubestellar_version}/scripts/check_pre_req.sh | bash -s -- --assert -V kflex ocm helm kubectl docker k3d
fi

##########################################
cluster_clean_up() {
    local cmd=$1
    local cluster_name=$2
    local platform=$3
    
    if ! error_message=$(eval "$cmd" 2>&1); then
        echo "clean up failed. Error:"
        echo "$error_message"
    else
        # Wait for the cluster to be fully deleted
        if [ "$platform" == "kind" ]; then
            wait_for_kind_cluster_deletion "$cluster_name"
        elif [ "$platform" == "k3d" ]; then
            wait_for_k3d_cluster_deletion "$cluster_name"
        fi
    fi
}

# Function to create kind cluster with retry logic
create_kind_cluster_with_retry() {
    local cluster_name=$1
    local config_arg=$2
    local max_attempts=3
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        echo "Attempt $attempt to create kind cluster '$cluster_name'..." >&2
        
        # First ensure any leftover containers or networks are cleaned up
        if docker ps -a | grep -q "$cluster_name"; then
            echo "Cleaning up leftover Docker containers for $cluster_name..." >&2
            docker ps -a | grep "$cluster_name" | awk '{print $1}' | xargs -r docker rm -f
        fi
        
        if docker network ls | grep -q "kind-$cluster_name" || docker network ls | grep -q "$cluster_name"; then
            echo "Cleaning up leftover Docker networks for $cluster_name..." >&2
            docker network ls --format "{{.Name}}" | grep -E "(kind-)?$cluster_name" | xargs -r docker network rm 2>/dev/null || true
        fi
        
        # Wait a moment for Docker cleanup to settle
        sleep 3
        
        # Try to create the cluster
        if kind create cluster --name "$cluster_name" $config_arg; then
            echo "Successfully created kind cluster '$cluster_name'" >&2
            return 0
        else
            echo "Failed to create kind cluster '$cluster_name' on attempt $attempt" >&2
            
            # Clean up any partial creation
            kind delete cluster --name "$cluster_name" 2>/dev/null || true
            wait_for_kind_cluster_deletion "$cluster_name"
            
            if [ $attempt -eq $max_attempts ]; then
                echo "ERROR: Failed to create kind cluster '$cluster_name' after $max_attempts attempts" >&2
                return 1
            fi
            
            # Wait before retrying
            echo "Waiting 10 seconds before retry..." >&2
            sleep 10
        fi
        
        attempt=$((attempt + 1))
    done
}

# Function to check if a port is free
wait_for_port_free() {
    local port=$1
    while lsof -i $port >/dev/null 2>&1; do
        echo "Waiting for port $port to be free..."
        sleep 5
    done
}

context_clean_up() {
    output=$(kubectl config get-contexts -o name)

    while IFS= read -r line; do
        if [ "$line" == "cluster1" ]; then
            echo "Deleting cluster1 context..."
            kubectl config delete-context cluster1

        elif [ "$line" == "cluster2" ]; then
            echo "Deleting cluster2 context..."
            kubectl config delete-context cluster2

        fi

    done <<< "$output"
}

checking_cluster() {
    while IFS= read -r line; do
        # Check for the cluster name and "Pending" status
        if echo "$line" | grep -q "$1" && echo "$line" | grep -q "Pending"; then
            echo "Pending CSR for $1 has been found, approving..."
            clusteradm --context its1 accept --clusters "$1"

            if [ $? -eq 0 ]; then
                echo -e "\033[33m✔\033[0m CSR Approved for $1."
                return 0
            else
                echo -e "\033[0;31mX\033[0m Failed to approve CSR for $1."
                return 1
            fi
        fi
    done < <(kubectl --context its1 get csr --watch)
}

echo -e "\nStarting environment clean up..."
echo -e "Starting cluster clean up..."

if command_exists "k3d"; then
    cluster_clean_up "k3d cluster delete kubeflex"
    cluster_clean_up "k3d cluster delete cluster1" "cluster1" "k3d"
    cluster_clean_up "k3d cluster delete cluster2" "cluster2" "k3d"
fi
if command_exists "kind"; then
    cluster_clean_up "kind delete cluster --name kubeflex"
    cluster_clean_up "kind delete cluster --name cluster1" "cluster1" "kind"
    cluster_clean_up "kind delete cluster --name cluster2" "cluster2" "kind"
fi

echo -e "\033[33m✔\033[0m Cluster space clean up has been completed"

# Wait for port 9443 to be free before proceeding
wait_for_port_free tcp:9443

echo -e "\nStarting context clean up..."
context_clean_up
echo -e "\033[33m✔\033[0m Context space clean up completed"

echo -e "\nCreating two $k8s_platform clusters to serve as example WECs"
clusters=(cluster1 cluster2)
cluster_log_dir=$(mktemp -d)
trap "rm -rf $cluster_log_dir" EXIT
for cluster in "${clusters[@]}"; do
    if [ "$k8s_platform" == "kind" ]; then
        # Use retry function for kind clusters to handle race conditions
        if create_kind_cluster_with_retry "${cluster}" "" &>"${cluster_log_dir}/${cluster}.log"; then
            echo -e "\033[33m✔\033[0m Cluster $cluster was successfully created"
            kubectl config rename-context "kind-${cluster}" "${cluster}" >/dev/null 2>&1
        else
            echo -e "\033[0;31mX\033[0m Creation of cluster $cluster failed!" >&2
            cat "${cluster_log_dir}/${cluster}.log" >&2
            false
        fi
    else
        # For k3d, use the original approach for now
        if k3d cluster create --network k3d-kubeflex "${cluster}" &>"${cluster_log_dir}/${cluster}.log"; then
            echo -e "\033[33m✔\033[0m Cluster $cluster was successfully created"
            kubectl config rename-context "k3d-${cluster}" "${cluster}" >/dev/null 2>&1
        else
            echo -e "\033[0;31mX\033[0m Creation of cluster $cluster failed!" >&2
            cat "${cluster_log_dir}/${cluster}.log" >&2
            false
        fi
    fi
done

echo -e "Creating KubeFlex cluster with SSL Passthrough"
if [ "$k8s_platform" == "kind" ]; then
    curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v${kubestellar_version}/scripts/create-kind-cluster-with-SSL-passthrough.sh | bash -s -- --name kubeflex --nosetcontext
else
    k3d cluster create -p "9443:443@loadbalancer" --k3s-arg "--disable=traefik@server:*" kubeflex
    kubectl wait --for=condition=Ready node --all --timeout=600s #Ensure both API server and nodes are ready
    helm install ingress-nginx ingress-nginx --set "controller.extraArgs.enable-ssl-passthrough=" --repo https://kubernetes.github.io/ingress-nginx --version 4.12.1 --namespace ingress-nginx --create-namespace --timeout 24h
fi
echo -e "\033[33m✔\033[0m Completed KubeFlex cluster with SSL Passthrough"

kubectl config use-context $k8s_platform-kubeflex

echo -e "\nPulling container images local..."
images=("ghcr.io/loft-sh/vcluster:0.16.4"
        "rancher/k3s:v1.27.2-k3s1"
        "quay.io/open-cluster-management/registration-operator:v0.15.2"
        "quay.io/kubestellar/postgresql:16.0.0-debian-11-r13")

for image in "${images[@]}"; do
    if ! docker inspect $image &> /dev/null; then
        docker pull $image
    fi
done
wait

if [ "$k8s_platform" == "kind" ]; then
    echo -e "\nFlatten images to single architecture to fix problems with kind load commands in recent Docker versions..."
    DOCKER_EMPTY_CONTEXT="$(mktemp -d)"
    for image in "${images[@]}"; do
        echo "FROM $image" | docker build -t "$image" -f- "$DOCKER_EMPTY_CONTEXT"
        # NOTE that this simpler solution does not work because it strips ENTRYPOINT
        # docker save "$image" | docker image import - "$image" &
    done
    wait
    rm -rf "$DOCKER_EMPTY_CONTEXT"
fi

for image in "${images[@]}"; do
    if [ "$k8s_platform" == "kind" ]; then
        kind load docker-image "$image" --name kubeflex
    else
        k3d image import "$image" --cluster kubeflex
    fi
done

echo -e "\nStarting the process to install KubeStellar core: $k8s_platform-kubeflex..."
if [ "$k8s_platform" == "k3d" ]
then var_flags="--set kubeflex-operator.hostContainer=k3d-kubeflex-server-0"
else var_flags=""
fi

helm upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart \
        --version $kubestellar_version \
        --set-json='ITSes=[{"name":"its1"}]' \
        --set-json='WDSes=[{"name":"wds1"},{"name":"wds2", "type":"host"}]' \
        --set-json='verbosity.default=5' \
        --set-json='kubeflex-operator.verbosity=5' \
        --timeout=24h \
        $var_flags

kflex ctx --set-current-for-hosting # make sure the KubeFlex CLI's hidden state is right for what the Helm chart just did
kflex ctx --overwrite-existing-context wds1
kflex ctx --overwrite-existing-context wds2
kflex ctx --overwrite-existing-context its1

echo -e "\nWaiting for OCM cluster manager to be ready..."
kubectl --context $k8s_platform-kubeflex wait controlplane.tenancy.kflex.kubestellar.org/its1 --for 'jsonpath={.status.postCreateHooks.its-hub-init}=true' --timeout 24h
kubectl --context $k8s_platform-kubeflex wait -n its1-system job.batch/its-hub-init --for condition=Complete --timeout 24h
kubectl --context $k8s_platform-kubeflex wait controlplane.tenancy.kflex.kubestellar.org/its1 --for 'jsonpath={.status.postCreateHooks.install-status-addon}=true' --timeout 24h
kubectl --context $k8s_platform-kubeflex wait -n its1-system job.batch/install-status-addon --for condition=Complete --timeout 24h
echo -e "\nWaiting for OCM hub cluster-info to be updated..."
kubectl --context $k8s_platform-kubeflex wait -n its1-system job.batch/update-cluster-info --for condition=Complete --timeout 24h
echo -e "\033[33m✔\033[0m OCM hub is ready"

echo -e "\nRegistering cluster 1 and 2 for remote access with KubeStellar Core..."

: set flags to "" if you have installed KubeStellar on an OpenShift cluster
flags="--force-internal-endpoint-lookup"
clusters=(cluster1 cluster2);
if ! joincmd=$(clusteradm --context its1 get token | grep '^clusteradm join')
then echo -e "\033[0;31mX\033[0m get token failed!\n" >&2; echo "$joincmd" >&2; false
fi
for cluster in "${clusters[@]}"; do
   if log=$(${joincmd/<cluster_name>/${cluster}} -v=6 --context ${cluster} --singleton ${flags} 2>&1)
   then echo -e "\033[33m✔\033[0m clusteradm join of $cluster succeeded"
   else echo -e "\033[0;31mX\033[0m clusteradm join of $cluster failed!" >&2; echo "$log" >&2; false
   fi
done

echo -e "Checking that the CSR for cluster 1 and 2 appears..."

echo
echo "Waiting for cluster1 and cluster2 to be ready and then approve their CSRs"
checking_cluster cluster1
checking_cluster cluster2

echo
echo "Checking the new clusters are in the OCM inventory and label them"
kubectl --context its1 get managedclusters
kubectl --context its1 label managedcluster cluster1 location-group=edge name=cluster1
kubectl --context its1 label managedcluster cluster2 location-group=edge name=cluster2

echo
echo Waiting for transport controller to create namespace customization-properties
# We allow versions of kubectl that do not support `kubectl wait --for=create`
wait_counter=0
while ! (kubectl --context its1 get ns customization-properties) ; do
    if (($wait_counter > 20)); then
        echo "Namespace customization-properties failed to appear!" >&2
        exit 1
    fi
    ((wait_counter += 1))
    sleep 10
done

echo
echo -e "\033[33m✔\033[0m Congratulations! Your KubeStellar demo environment is now ready to use."

cat <<EOF

Be sure to execute the following commands to set the shell variables expected in the example scenarios.

host_context=${k8s_platform}-kubeflex
its_cp=its1
its_context=its1
wds_cp=wds1
wds_context=wds1
wec1_name=cluster1
wec2_name=cluster2
wec1_context=\$wec1_name
wec2_context=\$wec2_name
label_query_both=location-group=edge
label_query_one=name=cluster1
EOF
