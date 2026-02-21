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

use_release=false
KUBESTELLAR_CONTROLLER_MANAGER_VERBOSITY=5
TRANSPORT_CONTROLLER_VERBOSITY=5
CLUSTER_SOURCE=kind
HOSTING_CONTEXT=kind-kubeflex
DEPLOYMENT_CONFIG=standard

while [ $# != 0 ]; do
    case "$1" in
        (-h|--help)
            echo "$0 usage: (--released | --kubestellar-controller-manager-verbosity \$num | --transport-controller-verbosity \$num | --env \$kind_or_k3d_or_ocp | --deployment-config \$config)*"
            echo ""
            echo "Deployment configurations:"
            echo "  standard       - Separate vcluster ControlPlanes for ITS (its1) and WDS (wds1) [default]"
            echo "  host-its-wds   - Hosting cluster plays both ITS and WDS roles (type: host)"
            echo "  shared-its-wds - A single ControlPlane (wds1) plays both ITS and WDS roles"
            exit;;
        (--released)
            use_release=true;;
        (--kubestellar-controller-manager-verbosity)
          if (( $# > 1 )); then
            KUBESTELLAR_CONTROLLER_MANAGER_VERBOSITY="$2"
            shift
          else
            echo "Missing kubestellar-controller-manager-verbosity value" >&2
            exit 1;
          fi;;
        (--transport-controller-verbosity)
          if (( $# > 1 )); then
            TRANSPORT_CONTROLLER_VERBOSITY="$2"
            shift
          else
            echo "Missing transport-controller-verbosity value" >&2
            exit 1;
          fi;;
        (--env)
          if (( $# < 1 )); then
            echo "Missing --env value" >&2
            exit 1
          fi
          case "$2" in
            (kind) CLUSTER_SOURCE=kind;     HOSTING_CONTEXT=kind-kubeflex;;
            (k3d)  CLUSTER_SOURCE=k3d;      HOSTING_CONTEXT=k3d-kubeflex;;
            (ocp)  CLUSTER_SOURCE=existing; HOSTING_CONTEXT=kscore;;
            (*) echo "--env must be given 'kind', 'k3d', or 'ocp'" >&2
                exit 1;;
          esac
          shift;;
        (--deployment-config)
          if (( $# > 1 )); then
            DEPLOYMENT_CONFIG="$2"
            shift
          else
            echo "Missing --deployment-config value" >&2
            exit 1;
          fi;;
        (*) echo "$0: unrecognized argument/flag '$1'" >&2
            exit 1
    esac
    shift
done

case "$DEPLOYMENT_CONFIG" in
    (standard|host-its-wds|shared-its-wds) ;;
    (*) echo "$0: --deployment-config must be 'standard', 'host-its-wds', or 'shared-its-wds'" >&2
        exit 1;;
esac

if ! [[ "$KUBESTELLAR_CONTROLLER_MANAGER_VERBOSITY" =~ [0-9]+ ]]
then echo "$0: \$KUBESTELLAR_CONTROLLER_MANAGER_VERBOSITY must be an integer" >&2
     exit 1
fi

if ! [[ "$TRANSPORT_CONTROLLER_VERBOSITY" =~ [0-9]+ ]]
then echo "$0: \$TRANSPORT_CONTROLLER_VERBOSITY must be an integer" >&2
     exit 1
fi


if [[ "$KFLEX_DISABLE_CHATTY" = true ]] ; then
   disable_chatty_status="--chatty-status=false"
   echo "disable_chatty_status = $disable_chatty_status"
fi

SRC_DIR="$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)"
COMMON_SRCS="${SRC_DIR}/../common"
source "$COMMON_SRCS/setup-shell.sh"

:
: -------------------------------------------------------------------------
: Create the KubeFlex hosting cluster, if necessary.
:
case "$CLUSTER_SOURCE" in
    (kind)
        ${SRC_DIR}/../../../scripts/create-kind-cluster-with-SSL-passthrough.sh --name kubeflex
        : Kubeflex kind cluster created.
        ;;
    (k3d)
        k3d cluster create -p "9443:443@loadbalancer" --k3s-arg "--disable=traefik@server:*" kubeflex
        kubectl --context k3d-kubeflex wait --for=condition=Ready node --all --timeout=600s
        helm install ingress-nginx ingress-nginx \
            --set "controller.extraArgs.enable-ssl-passthrough=" \
            --repo https://kubernetes.github.io/ingress-nginx \
            --version 4.12.1 \
            --namespace ingress-nginx --create-namespace \
            --kube-context k3d-kubeflex \
            --wait --wait-for-jobs \
            --timeout 24h
        : k3d kubeflex cluster created with nginx ingress.
        ;;
    (existing)
        kubectl config use-context "$HOSTING_CONTEXT"
        : kubectl configured to use existing cluster in "$HOSTING_CONTEXT" context
        ;;
esac


:
: -------------------------------------------------------------------------
: Build Helm set flags based on deployment configuration
:
case "$DEPLOYMENT_CONFIG" in
    (standard)
        # Default: separate vcluster ControlPlanes for ITS and WDS
        HELM_ITSES='ITSes=[{"name":"its1"}]'
        HELM_WDSES='WDSes=[{"name":"wds1"}]'
        ITS_NAME=its1
        WDS_NAME=wds1
        ITS_SYSTEM_NS=its1-system
        WDS_SYSTEM_NS=wds1-system
        ;;
    (host-its-wds)
        # Hosting cluster plays both ITS and WDS roles (type: host)
        HELM_ITSES='ITSes=[{"name":"its1","type":"host"}]'
        HELM_WDSES='WDSes=[{"name":"wds1","type":"host"}]'
        ITS_NAME=its1
        WDS_NAME=wds1
        ITS_SYSTEM_NS=its1-system
        WDS_SYSTEM_NS=wds1-system
        ;;
    (shared-its-wds)
        # A single ControlPlane plays both ITS and WDS roles
        HELM_ITSES='ITSes=[{"name":"its1"}]'
        HELM_WDSES='WDSes=[{"name":"wds1","ITSName":"its1"}]'
        ITS_NAME=its1
        WDS_NAME=wds1
        ITS_SYSTEM_NS=its1-system
        WDS_SYSTEM_NS=wds1-system
        ;;
esac

echo "Deployment configuration: $DEPLOYMENT_CONFIG"
echo "  ITS: $ITS_NAME (namespace: $ITS_SYSTEM_NS)"
echo "  WDS: $WDS_NAME (namespace: $WDS_SYSTEM_NS)"

:
: -------------------------------------------------------------------------
: Install the core-chart
:

pushd "${SRC_DIR}/../../.."
if [ "$use_release" = true ] ; then
  helm upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart \
    --version $(yq .KUBESTELLAR_VERSION core-chart/values.yaml) \
    --kube-context $HOSTING_CONTEXT \
    --set-json="$HELM_ITSES" \
    --set-json="$HELM_WDSES" \
    --set verbosity.kubestellar=${KUBESTELLAR_CONTROLLER_MANAGER_VERBOSITY} \
    --set verbosity.transport=${TRANSPORT_CONTROLLER_VERBOSITY}
else
  if [ "$CLUSTER_SOURCE" = "k3d" ]; then
    make k3d-load-image
    EXTRA_HELM_FLAGS="--set kubeflex-operator.hostContainer=k3d-kubeflex-server-0"
  else
    make kind-load-image
    EXTRA_HELM_FLAGS=""
  fi
  helm dependency update core-chart/
  helm upgrade --install ks-core core-chart/ \
    --set KUBESTELLAR_VERSION=$(git rev-parse --short HEAD) \
    --kube-context $HOSTING_CONTEXT \
    --set-json="$HELM_ITSES" \
    --set-json="$HELM_WDSES" \
    --set verbosity.kubestellar=${KUBESTELLAR_CONTROLLER_MANAGER_VERBOSITY} \
    --set verbosity.transport=${TRANSPORT_CONTROLLER_VERBOSITY} \
    $EXTRA_HELM_FLAGS
  fi
popd

: Waiting for OCM hub to be ready...
kubectl wait controlplane.tenancy.kflex.kubestellar.org/${ITS_NAME} --for 'jsonpath={.status.postCreateHooks.its-hub-init}=true' --timeout 800s
kubectl wait -n ${ITS_SYSTEM_NS} job.batch/its-hub-init --for condition=Complete --timeout 800s
kubectl wait controlplane.tenancy.kflex.kubestellar.org/${ITS_NAME} --for 'jsonpath={.status.postCreateHooks.install-status-addon}=true' --timeout 800s
kubectl wait -n ${ITS_SYSTEM_NS} job.batch/install-status-addon --for condition=Complete --timeout 800s

# The update-cluster-info job is only created for non-host ITS types (e.g., vcluster)
# because host-type ITS uses the hosting cluster's API endpoint directly.
if [ "$DEPLOYMENT_CONFIG" != "host-its-wds" ]; then
    kubectl wait -n ${ITS_SYSTEM_NS} job.batch/update-cluster-info --for condition=Complete --timeout 400s
fi

wait_timeout=800s
kubectl --context "$HOSTING_CONTEXT" -n ${WDS_SYSTEM_NS} wait --for=condition=Ready pod -l name=transport-controller --timeout ${wait_timeout}
kubectl --context "$HOSTING_CONTEXT" -n ${WDS_SYSTEM_NS} wait --for=condition=Ready pod -l control-plane=controller-manager --timeout ${wait_timeout}
echo "transport-controller and kubestellar-controller-manager are running."

kubectl config use-context "$HOSTING_CONTEXT"
kflex ctx --set-current-for-hosting

case "$DEPLOYMENT_CONFIG" in
    (host-its-wds)
        # For host-type control planes, kflex ctx cannot create separate contexts
        # because there is no separate API server. Instead, create alias contexts
        # that point to the same cluster/authinfo as the hosting cluster.
        hosting_cluster=$(kubectl config view -o jsonpath="{.contexts[?(@.name==\"$HOSTING_CONTEXT\")].context.cluster}")
        hosting_authinfo=$(kubectl config view -o jsonpath="{.contexts[?(@.name==\"$HOSTING_CONTEXT\")].context.user}")
        kubectl config set-context ${WDS_NAME} --cluster="$hosting_cluster" --user="$hosting_authinfo"
        kubectl config set-context ${ITS_NAME} --cluster="$hosting_cluster" --user="$hosting_authinfo"
        ;;
    (*)
        # For vcluster/k8s-type control planes, kflex ctx creates proper contexts
        kflex ctx --overwrite-existing-context ${WDS_NAME}
        kflex ctx --overwrite-existing-context ${ITS_NAME}
        ;;
esac

# Ensure BindingPolicy CRD is established before proceeding
wait-for-cmd "kubectl --context ${WDS_NAME} get crd bindingpolicies.control.kubestellar.io >/dev/null 2>&1"

kflex ctx

wait-for-cmd "kubectl --context ${ITS_NAME} get ns customization-properties"

:
: -------------------------------------------------------------------------
: Create clusters and register with OCM
:
function add_wec() {
    cluster=$1
    case "$CLUSTER_SOURCE" in
        (kind)
            kind create cluster --name $cluster
            kubectl config rename-context kind-${cluster} $cluster
            joinflags="--force-internal-endpoint-lookup";;
        (k3d)
            k3d cluster create --network k3d-kubeflex $cluster
            kubectl config rename-context k3d-${cluster} $cluster
            joinflags="--force-internal-endpoint-lookup";;
        (existing)
            joinflags="";;
    esac
    clusteradm --context ${ITS_NAME} get token | grep '^clusteradm join' | sed "s/<cluster_name>/${cluster}/" | awk '{print $0 " --context '${cluster}' --singleton '${joinflags}'"}' | sh
}

"${SRC_DIR}/../../../scripts/check_pre_req.sh" --assert --verbose ocm

add_wec cluster1
add_wec cluster2

: Wait for csrs in ${ITS_NAME}
wait-for-cmd '(($(kubectl --context '${ITS_NAME}' get csr 2>/dev/null | grep -c Pending) >= 2))'

clusteradm --context ${ITS_NAME} accept --clusters cluster1
clusteradm --context ${ITS_NAME} accept --clusters cluster2

kubectl --context ${ITS_NAME} get managedclusters
kubectl --context ${ITS_NAME} label managedcluster cluster1 location-group=edge name=cluster1 region=east
kubectl --context ${ITS_NAME} create cm -n customization-properties cluster1 --from-literal clusterURL=https://my.clusters/1001-abcd
kubectl --context ${ITS_NAME} label managedcluster cluster2 location-group=edge name=cluster2 region=west
kubectl --context ${ITS_NAME} create cm -n customization-properties cluster2 --from-literal clusterURL=https://my.clusters/2002-cdef

:
: -------------------------------------------------------------------------
: Get all deployments and statefulsets running in the hosting cluster.
: Expect to see the controller-manager and transport-controller created in the wds-system
: namespace and the its statefulset created in the its-system namespace.
:
case "$DEPLOYMENT_CONFIG" in
    (standard)
        # vcluster ITS + k8s WDS: expect 5 resources (kubestellar-controller-manager, transport-controller in wds1-system + vcluster etc in its1-system)
        wait-for-cmd '(( $(kubectl --context '"'$HOSTING_CONTEXT'"' get deployments,statefulsets --all-namespaces | grep -e '"${WDS_NAME}"' -e '"${ITS_NAME}"' | wc -l) == 5 ))'
        ;;
    (host-its-wds)
        # host type ITS + host type WDS: fewer resources since no vcluster; expect at least the controller-manager and transport-controller
        wait-for-cmd '(( $(kubectl --context '"'$HOSTING_CONTEXT'"' get deployments,statefulsets --all-namespaces | grep -e '"${WDS_NAME}"' -e '"${ITS_NAME}"' | wc -l) >= 2 ))'
        ;;
    (shared-its-wds)
        # vcluster ITS + separate WDS pointing to the same ITS: expect resources for both
        wait-for-cmd '(( $(kubectl --context '"'$HOSTING_CONTEXT'"' get deployments,statefulsets --all-namespaces | grep -e '"${WDS_NAME}"' -e '"${ITS_NAME}"' | wc -l) >= 3 ))'
        ;;
esac

:
: -------------------------------------------------------------------------
: "Get available clusters with label location-group=edge and check there are two of them"
:
if ! expect-cmd-output "kubectl --context ${ITS_NAME} get managedclusters -l location-group=edge" 'wc -l | grep -wq 3'
then
    echo "Failed to see two clusters."
    exit 1
fi
