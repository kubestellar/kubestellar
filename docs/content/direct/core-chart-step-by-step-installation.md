# KubeStellar Core Chart: Step-by-Step Installation Guide

This guide provides a complete installation process with actual command outputs. The outputs shown are examples and your actual output may vary slightly but should be similar to what's presented here.

## Prerequisites

- [Helm](https://helm.sh/) installed
- Docker installed and running
- `kubectl` installed
- Internet connection for downloading images and charts

## Step 1: Create Kind Cluster with SSL Passthrough

create a new local Kind cluster that satisfies the requirements for KubeStellar setup:

```bash
bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v$KUBESTELLAR_VERSION/scripts/create-kind-cluster-with-SSL-passthrough.sh) --name kubeflex --port 9443
```
Or if you have local codebase then 
```bash
    cd scripts/
    ./create-kind-cluster-with-SSL-passthrough.sh --name kubeflex --port 9443
```

**Expected output :**
```
Creating "kubeflex" kind cluster with SSL passthrougn and 9443 port mapping...
No kind clusters found.
Creating cluster "kubeflex" ...
 ✓ Ensuring node image (kindest/node:v1.29.2) 🖼
 ✓ Preparing nodes 📦  
 ✓ Writing configuration 📜 
 ✓ Starting control-plane 🕹️ 
 ✓ Installing CNI 🔌 
 ✓ Installing StorageClass 💾 
Set kubectl context to "kind-kubeflex"
You can now use your cluster with:

kubectl cluster-info --context kind-kubeflex

Have a question, bug, or feature request? Let us know! https://kind.sigs.k8s.io/#community 🙂
Installing an nginx ingress...
namespace/ingress-nginx created
serviceaccount/ingress-nginx created
serviceaccount/ingress-nginx-admission created
role.rbac.authorization.k8s.io/ingress-nginx created
role.rbac.authorization.k8s.io/ingress-nginx-admission created
clusterrole.rbac.authorization.k8s.io/ingress-nginx created
clusterrole.rbac.authorization.k8s.io/ingress-nginx-admission created
rolebinding.rbac.authorization.k8s.io/ingress-nginx created
rolebinding.rbac.authorization.k8s.io/ingress-nginx-admission created
clusterrolebinding.rbac.authorization.k8s.io/ingress-nginx created
clusterrolebinding.rbac.authorization.k8s.io/ingress-nginx-admission created
configmap/ingress-nginx-controller created
service/ingress-nginx-controller created
service/ingress-nginx-controller-admission created
deployment.apps/ingress-nginx-controller created
job.batch/ingress-nginx-admission-create created
job.batch/ingress-nginx-admission-patch created
ingressclass.networking.k8s.io/nginx created
validatingwebhookconfiguration.admissionregistration.k8s.io/ingress-nginx-admission created
Patching nginx ingress to enable SSL passthrough...
deployment.apps/ingress-nginx-controller patched
Waiting for nginx ingress with SSL passthrough to be ready...
pod/ingress-nginx-controller-68494d6d4-4qhmv condition met
Setting context to "kind-kubeflex"...
Switched to context "kind-kubeflex".
```

## Step 2: Check Container Name

verify the container name that will be needed for the Helm chart configuration:

```bash
docker ps --filter "name=kubeflex" --format "{{.Names}}"
```

**Expected output (similar to):**
```
kubeflex-control-plane
```

## Step 3: Update Chart Dependencies

navigate to your KubeStellar directory and update the chart dependencies:

```bash
helm dependency update core-chart
```

**Expected output (similar to):**
```
Hang tight while we grab the latest from your chart repositories...
...Successfully got an update from the "antrea" chart repository
...Successfully got an update from the "prometheus-community" chart repository
...Successfully got an update from the "istio" chart repository
Update Complete. ⎈Happy Helming!⎈
Saving 2 charts
Downloading kubeflex-operator from repo oci://ghcr.io/kubestellar/kubeflex/chart
Pulled: ghcr.io/kubestellar/kubeflex/chart/kubeflex-operator:v0.8.9
Digest: sha256:2be43de71425ad682edca6544f6c3a5864afbfad09a4b7e1e57bde6dae664334
Downloading argo-cd from repo oci://ghcr.io/argoproj/argo-helm
Pulled: ghcr.io/argoproj/argo-helm/argo-cd:7.8.5
Digest: sha256:662f4687e8e525f86ff9305020632b337a09ffacb7b61b7c42a841922c91da7b
Deleting outdated charts
```

## Step 4: Initial Chart Installation

first, install the basic chart to set up KubeFlex:

```bash
helm upgrade --install ks-core core-chart
```

**Expected output (similar to):**
```
Release "ks-core" does not exist. Installing it now.
NAME: ks-core
LAST DEPLOYED: Tue Jun  3 09:15:08 2025
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES:
For your convenience you will probably want to add contexts to your
kubeconfig named after the non-host-type control planes (WDSes and
ITSes) that you just created (a host-type control plane is just an
alias for the KubeFlex hosting cluster). You can do that with the
following `kflex` commands; each creates a context and makes it the
current one. See
https://github.com/kubestellar/kubestellar/blob/0.28.0-alpha.2/docs/content/direct/core-chart.md#kubeconfig-files-and-contexts-for-control-planes
for a way to do this without using `kflex`.
Start by setting your current kubeconfig context to the one you used
when installing this chart.

kubectl config use-context $the_one_where_you_installed_this_chart
kflex ctx --set-current-for-hosting # make sure the KubeFlex CLI's hidden state is right for what the Helm chart just did


Finally, you can use `kflex ctx` to switch back to the kubeconfig
context for your KubeFlex hosting cluster.
```

## Step 5: Verify Current Context

check your current kubectl context:

```bash
kubectl config get-contexts
```

**Expected output (similar to):**
```
CURRENT   NAME            CLUSTER         AUTHINFO        NAMESPACE
*         kind-kubeflex   kind-kubeflex   kind-kubeflex   
```

## Step 6: Install KubeStellar Core with Control Planes

now upgrade the chart to include the ITS and WDS control planes:

```bash
helm upgrade --install ks-core ./core-chart \
--set "kubeflex-operator.hostContainer=kubeflex-control-plane" \
--set "kubeflex-operator.externalPort=9443" \
--set-json='ITSes=[{"name":"its1"}]' \
--set-json='WDSes=[{"name":"wds1","ITSName":"its1"}]'
```

**Expected output (similar to):**
```
Release "ks-core" has been upgraded. Happy Helming!
NAME: ks-core
LAST DEPLOYED: Tue Jun  3 09:23:35 2025
NAMESPACE: default
STATUS: deployed
REVISION: 2
TEST SUITE: None
NOTES:
For your convenience you will probably want to add contexts to your
kubeconfig named after the non-host-type control planes (WDSes and
ITSes) that you just created (a host-type control plane is just an
alias for the KubeFlex hosting cluster). You can do that with the
following `kflex` commands; each creates a context and makes it the
current one. See
https://github.com/kubestellar/kubestellar/blob/0.28.0-alpha.2/docs/content/direct/core-chart.md#kubeconfig-files-and-contexts-for-control-planes
for a way to do this without using `kflex`.
Start by setting your current kubeconfig context to the one you used
when installing this chart.

kubectl config use-context $the_one_where_you_installed_this_chart
kflex ctx --set-current-for-hosting # make sure the KubeFlex CLI's hidden state is right for what the Helm chart just did

kflex ctx --overwrite-existing-context its1
kflex ctx --overwrite-existing-context wds1

Finally, you can use `kflex ctx` to switch back to the kubeconfig
context for your KubeFlex hosting cluster.
```

## Step 7: Verify Control Planes

check that the control planes have been created successfully:

```bash
kubectl get controlplane
```

**Expected output (similar to):**
```
NAME   SYNCED   READY   TYPE       AGE
its1   True     True    vcluster   6m5s
wds1   True     True    k8s        6m5s
```

> **Note:** It may take a few minute for the control planes to become ready. If you see `READY: False`, wait a moment and check again.

## Step 8: Setup Kubeconfig Contexts

configure the kubectl contexts for accessing the control planes:

```bash
kubectl config use-context kind-kubeflex
kflex ctx --set-current-for-hosting
kflex ctx --overwrite-existing-context its1
kflex ctx --overwrite-existing-context wds1
```

**Expected output (similar to):**
```
Switched to context "kind-kubeflex".
✔ Checking for saved hosting cluster context...
✔ Switching to hosting cluster context...
no kubeconfig context for its1 was found: context its1 not found for control plane its1
✔ Overwriting existing context for control plane
trying to load new context its1 from server...
✔ Switching to context its1...
no kubeconfig context for wds1 was found: context wds1 not found for control plane wds1
✔ Overwriting existing context for control plane
trying to load new context wds1 from server...
✔ Switching to context wds1...
```

> **Note:** The "no kubeconfig context found" messages are normal when setting up contexts for the first time.

## Step 9: Verify Setup

test that all contexts are working correctly:

```bash
kubectl --context its1 get pods
```

**Expected output (similar to):**
```
No resources found in default namespace.
```

```bash
kubectl --context wds1 get pods
```

**Expected output (similar to):**
```
No resources found in default namespace.
```

```bash
kubectl --context kind-kubeflex get pods
```

**Expected output (similar to):**
```
NAME            READY   STATUS      RESTARTS   AGE
ks-core-924p2   0/1     Completed   0          35m
```

> **Note:** the "No resources found in default namespace" output for ITS and WDS contexts is expected and normal. This indicates that the control planes are working correctly but no workloads have been deployed yet.

## Step 10: Verify Namespaces

check the namespaces in each control plane to confirm proper setup:

```bash
kubectl --context wds1 get namespaces
```

**Expected output (similar to):**
```
NAME                 STATUS   AGE
kube-system          Active   26m
kube-public          Active   26m
kube-node-lease      Active   26m
default              Active   26m
kubestellar-report   Active   23m
```

```bash
kubectl --context its1 get namespaces
```

**Expected output (similar to):**
```
NAME                          STATUS   AGE
kube-system                   Active   25m
kube-public                   Active   25m
default                       Active   25m
kube-node-lease               Active   25m
open-cluster-management       Active   25m
open-cluster-management-hub   Active   24m
customization-properties      Active   23m
```

## Troubleshooting

### Common Issues

1. **Control planes not becoming ready**: wait a few minutes and check again. Large container images may take time to download.

2. **Context switching errors**: ensure you're using the correct context names and that the control planes are in "Ready" state.

3. **Port conflicts**: if port 9443 is already in use, choose a different port and update the `externalPort` setting accordingly.

### Verification Commands

to verify your installation is working correctly:

```bash
# Check all contexts
kubectl config get-contexts

# Verify control planes
kubectl --context kind-kubeflex get controlplane

# Test each context
kubectl --context its1 cluster-info
kubectl --context wds1 cluster-info
kubectl --context kind-kubeflex cluster-info
```
