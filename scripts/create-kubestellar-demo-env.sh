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

# Function to check if a port is free
wait_for_port_free() {
    local port=$1
    while lsof -i $port >/dev/null 2>&1; do
        echo "Waiting for port $port to be free..."
        sleep 5
    done
}

echo -e "Checking container runtime..."
if ! dunsel=$(docker ps 2>&1); then
    echo "Error: The script cannot continue because Docker or Podman is not running. Please start your container runtime before running the script again."
    exit 1
fi
echo "Container runtime is running."

kubestellar_version=0.28.0
echo -e "KubeStellar Version: ${kubestellar_version}"

echo -e "Checking that pre-req softwares are installed..."
if [ "$k8s_platform" == "kind" ]; then
    curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v${kubestellar_version}/scripts/check_pre_req.sh | bash -s -- --assert -V kflex ocm helm kubectl docker kind
else
    curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v${kubestellar_version}/scripts/check_pre_req.sh | bash -s -- --assert -V kflex ocm helm kubectl docker k3d
fi

##########################################
cluster_clean_up() {
    if ! error_message=$(eval "$1" 2>&1); then
        echo "clean up failed. Error:"
        echo "$error_message"
    fi
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
    cluster_clean_up "k3d cluster delete cluster1"
    cluster_clean_up "k3d cluster delete cluster2"
fi
if command_exists "kind"; then
    cluster_clean_up "kind delete cluster --name kubeflex"
    cluster_clean_up "kind delete cluster --name cluster1"
    cluster_clean_up "kind delete cluster --name cluster2"
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

# Create WEC clusters in parallel for maximum speed
wec_pids=()
for cluster in "${clusters[@]}"; do
    echo "Creating cluster $cluster in background..."
    {
      if [ "$k8s_platform" == "kind" ]; then
        if kind create cluster --name "${cluster}" &>"${cluster_log_dir}/${cluster}.log"; then
            echo -e "\033[33m✔\033[0m Cluster $cluster was successfully created"
            kubectl config rename-context "${k8s_platform}-${cluster}" "${cluster}" >/dev/null 2>&1
        else
            echo -e "\033[0;31mX\033[0m Creation of cluster $cluster failed!" >&2
            cat "${cluster_log_dir}/${cluster}.log" >&2
            exit 1
        fi
      else
        if k3d cluster create --network k3d-kubeflex "${cluster}" &>"${cluster_log_dir}/${cluster}.log"; then
            echo -e "\033[33m✔\033[0m Cluster $cluster was successfully created"
            kubectl config rename-context "${k8s_platform}-${cluster}" "${cluster}" >/dev/null 2>&1
        else
            echo -e "\033[0;31mX\033[0m Creation of cluster $cluster failed!" >&2
            cat "${cluster_log_dir}/${cluster}.log" >&2
            exit 1
        fi
      fi
    } &
    wec_pids+=($!)
done

echo "WEC clusters creating in parallel..."

# Wait for all WEC cluster creation to complete
for pid in "${wec_pids[@]}"; do
    wait $pid
done

echo "All WEC clusters created successfully"

# Function to prefetch nginx ingress images in parallel
prefetch_nginx_images() {
    local platform=$1
    echo "Pre-fetching nginx ingress images in background..."
    
    # Get nginx ingress images
    nginx_images=(
        "registry.k8s.io/ingress-nginx/controller:v1.12.1"
        "registry.k8s.io/ingress-nginx/kube-webhook-certgen:v1.5.1"
    )
    
    for image in "${nginx_images[@]}"; do
        docker pull "$image" &
    done
    
    # Load images into cluster based on platform
    for image in "${nginx_images[@]}"; do
        if [ "$platform" == "kind" ]; then
            kind load docker-image "$image" --name kubeflex &
        else
            k3d image import "$image" --cluster kubeflex &
        fi
    done
}

# Function for optimized k3d nginx installation  
install_optimized_nginx_k3d() {
    echo "Installing optimized nginx ingress for k3d..."
    
    # Use optimized Helm values for faster startup
    if helm install ingress-nginx ingress-nginx \
        --repo https://kubernetes.github.io/ingress-nginx \
        --version 4.12.1 \
        --namespace ingress-nginx \
        --create-namespace \
        --set controller.replicaCount=1 \
        --set controller.nodeSelector."kubernetes\.io/os"=linux \
        --set controller.admissionWebhooks.enabled=false \
        --set controller.metrics.enabled=false \
        --set controller.podAnnotations."prometheus\.io/scrape"=false \
        --set controller.service.type=NodePort \
        --set controller.service.nodePorts.http=30080 \
        --set controller.service.nodePorts.https=30443 \
        --set controller.extraArgs.enable-ssl-passthrough=true \
        --set controller.extraArgs.default-ssl-certificate=ingress-nginx/default-ssl-certificate \
        --set controller.lifecycle.preStop.exec.command=["/wait-shutdown"] \
        --set defaultBackend.enabled=false \
        --timeout=24h \
        --wait; then
        echo "Nginx ingress installation completed successfully"
        
        # Quick readiness check
        kubectl --context k3d-kubeflex wait --namespace ingress-nginx \
            --for=condition=ready pod \
            --selector=app.kubernetes.io/component=controller \
            --timeout=60s
    else
        echo "Error: Nginx ingress installation failed" >&2
        return 1
    fi
}

echo -e "Creating KubeFlex cluster with SSL Passthrough"
if [ "$k8s_platform" == "kind" ]; then
    # Start nginx image prefetching in parallel with cluster creation
    prefetch_nginx_images "kind" &
    prefetch_pid=$!
    
    # Use optimized flags for Kind cluster creation
    curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v${kubestellar_version}/scripts/create-kind-cluster-with-SSL-passthrough.sh | bash -s -- --name kubeflex --nosetcontext --prefetch --nowait &
    kind_creation_pid=$!
    
    # Wait for prefetching to complete
    wait $prefetch_pid
    echo "Image prefetching completed"
    
    # Wait for Kind cluster creation to complete
    wait $kind_creation_pid
    echo "Kind cluster creation completed"
    
else
    # Optimized k3d path with parallel operations
    echo "Creating k3d cluster and prefetching nginx images in parallel..."
    
    # Start cluster creation and image prefetching in parallel
    k3d cluster create -p "9443:443@loadbalancer" --k3s-arg "--disable=traefik@server:*" kubeflex &
    cluster_pid=$!
    
    prefetch_nginx_images "k3d" &
    prefetch_pid=$!
    
    # Wait for cluster creation
    wait $cluster_pid
    echo "k3d cluster creation completed"
    
    # Wait a moment for cluster to be fully ready
    sleep 10
    
    # Install optimized nginx while prefetching completes
    install_optimized_nginx_k3d &
    nginx_install_pid=$!
    
    # Wait for prefetching to complete
    wait $prefetch_pid
    echo "Image prefetching completed"
    
    # Wait for nginx installation to complete
    wait $nginx_install_pid
fi
echo -e "\033[33m✔\033[0m Completed KubeFlex cluster with SSL Passthrough"

kubectl config use-context $k8s_platform-kubeflex

echo -e "\nPulling container images local..."
images=("ghcr.io/loft-sh/vcluster:0.16.4"
        "rancher/k3s:v1.27.2-k3s1"
        "quay.io/open-cluster-management/registration-operator:v0.15.2"
        "quay.io/kubestellar/postgresql:16.0.0-debian-11-r13")

# Pull all images in parallel for faster completion
pull_pids=()
for image in "${images[@]}"; do
    echo "Pulling $image in background..."
    docker pull "$image" &
    pull_pids+=($!)
done

# Wait for all pulls to complete
for pid in "${pull_pids[@]}"; do
    wait $pid
done
echo "All image pulls completed"

# Load images into cluster in parallel
load_pids=()
for image in "${images[@]}"; do
    if [ "$k8s_platform" == "kind" ]; then
        echo "Loading $image into kind cluster..."
        kind load docker-image "$image" --name kubeflex &
        load_pids+=($!)
    else
        echo "Importing $image into k3d cluster..."
        k3d image import "$image" --cluster kubeflex &
        load_pids+=($!)
    fi
done

# Wait for all loads to complete
for pid in "${load_pids[@]}"; do
    wait $pid
done
echo "All image loads completed"

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
kubectl --context $k8s_platform-kubeflex wait controlplane.tenancy.kflex.kubestellar.org/its1 --for 'jsonpath={.status.postCreateHooks.its-with-clusteradm}=true' --timeout 24h
kubectl --context $k8s_platform-kubeflex wait -n its1-system job.batch/its-with-clusteradm --for condition=Complete --timeout 24h
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
