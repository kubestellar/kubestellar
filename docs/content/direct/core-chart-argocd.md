# Using Argo CD with KubeStellar Core chart

## Table of Contents
- [Overview](#overview)
- [Pre-requisites](#pre-requisites)
- [Installing Argo CD using KubeStellar Core chart](#installing-argo-cd-using-kubestellar-core-chart)
- [Step-by-Step Installation Example](#step-by-step-installation-example)
- [Deploying Argo CD applications](#deploying-argo-cd-applications)

## Overview

This documents explains how to use the KubeStellar core Helm chart to:

- deploy Argo CD in KubeFlex hosting cluster;
- register every WDS as a target cluster in Argo CD; and
- create Argo CD applications as specified by the chart values.

## Pre-requisites

The prerequisites are the same as [installing KubeStellar using the Core chart](core-chart.md#pre-requisites).

The settings described in this document are an extesnion of the KubeStellar Core chart settings described [here](core-chart.md#kubestellar-core-chart-values).

## Installing Argo CD using KubeStellar Core chart

To enable the installation of Argo CD by the KubeStellar Core chart, use the flag `--set argocd.install=true`. Besides deploying an instance of Argo CD, KubeStellar Core chart will take care of registering all the WDSes installed by the chart as Argo CD target clusters.

When deploying in an **OpenShift** cluster, add the flag `--set argocd.openshift.enabled=true`.

When deploying in a **Kubernetes** cluster, use the flag `--set argocd.global.domain=<url>` to provide the URL for the **nginx** ingress, which defaults to `argocd.localtest.me`.

Note that when creating a local **Kubernetes** cluster using our scripts for **Kind** or **k3s**, the **nginx** ingress will be accessible on host port `9443`; therefore the Argo CD UI can be accessed at the address `https://argocd.localtest.me:9443`.

The initial password for the `admin` user can be retrieved using the command:

```shell
kubectl -n kubeflex-system get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```

## Step-by-Step Installation Example

This section shows how to add Argo CD to an existing KubeStellar installation with actual command outputs.

### Prerequisites

Ensure you have a working KubeStellar installation as described in the [core chart documentation](core-chart.md). Make sure you're in the correct kubectl context:

```bash
kubectl config use-context kind-kubeflex
```

**Output:**
```
Switched to context "kind-kubeflex".
```

### Step 1: Upgrade KubeStellar with Argo CD

```bash
helm upgrade --install ks-core ./core-chart \
  --namespace kubeflex-system \
  --set "kubeflex-operator.hostContainer=kubeflex-control-plane" \
  --set "kubeflex-operator.externalPort=9443" \
  --set-json='ITSes=[{"name":"its1"}]' \
  --set-json='WDSes=[{"name":"wds1","ITSName":"its1"}]' \
  --set argocd.install=true
```

**Output:**
```
Release "ks-core" has been upgraded. Happy Helming!
NAME: ks-core
LAST DEPLOYED: Sun May 25 11:15:52 2025
NAMESPACE: kubeflex-system
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

Access Argo CD UI at https://argocd.localtest.me (append :9443 for Kind or k3s installations).
Obtain Argo CD admin password using the command:
kubectl -n kubeflex-system get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```

### Step 2: Verify Argo CD Installation

Check for Argo CD namespaces:
```bash
kubectl get namespaces | grep -i argo
```

**Output:**
```
(No dedicated argocd namespace - Argo CD is installed in kubeflex-system)
```

Check Argo CD pods across all namespaces:
```bash
kubectl get pods -A | grep -i argo
```

**Output:**
```
kubeflex-system      ks-core-argocd-application-controller-0                           1/1     Running     0             2m5s
kubeflex-system      ks-core-argocd-applicationset-controller-68dd5f4859-dvvcp         1/1     Running     0             2m6s
kubeflex-system      ks-core-argocd-dex-server-844d5898ff-f8v9h                        1/1     Running     0             2m6s
kubeflex-system      ks-core-argocd-notifications-controller-844596b886-l4nqs          1/1     Running     0             2m6s
kubeflex-system      ks-core-argocd-redis-76c6b4db57-qw25s                             1/1     Running     0             2m6s
kubeflex-system      ks-core-argocd-repo-server-655684556b-jdsxl                       1/1     Running     0             2m6s
kubeflex-system      ks-core-argocd-server-c46b7d8f6-8dv9b                             1/1     Running     0             2m6s
```

Check all pods in the kubeflex-system namespace:
```bash
kubectl get pods -n kubeflex-system
```

**Output:**
```
NAME                                                        READY   STATUS      RESTARTS   AGE
ks-core-argocd-application-controller-0                     1/1     Running     0          2m31s
ks-core-argocd-applicationset-controller-68dd5f4859-dvvcp   1/1     Running     0          2m32s
ks-core-argocd-dex-server-844d5898ff-f8v9h                  1/1     Running     0          2m32s
ks-core-argocd-notifications-controller-844596b886-l4nqs    1/1     Running     0          2m32s
ks-core-argocd-redis-76c6b4db57-qw25s                       1/1     Running     0          2m32s
ks-core-argocd-repo-server-655684556b-jdsxl                 1/1     Running     0          2m32s
ks-core-argocd-server-c46b7d8f6-8dv9b                       1/1     Running     0          2m32s
ks-core-k46jt                                               0/1     Completed   0          81m
kubeflex-controller-manager-6c7f49588-h24d7                 2/2     Running     0          81m
postgres-postgresql-0                                       1/1     Running     0          81m
```

### Step 3: Get Argo CD Admin Password

```bash
kubectl -n kubeflex-system get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```

**Output:**
```
lxyMWsIcdtoUY3Py
```

### Step 4: Access Argo CD UI

- **URL**: `https://argocd.localtest.me:9443`
- **Username**: `admin`
- **Password**: `lxyMWsIcdtoUY3Py` (change this with yours)

> **Note:** All Argo CD components are installed in the `kubeflex-system` namespace rather than a dedicated `argocd` namespace. This is the expected behavior when using the KubeStellar Core chart with Argo CD integration.

## Deploying Argo CD applications

The KubeStellar Core chart can also be used to deploy Argo CD applications as specified by chart values. The example below shows the relevant fragment of the chart values that could be used for deploying an application corresponding to `scenario-6` in [KubeStellar docs](example-scenarios.md#scenario-6---multi-cluster-workload-deployment-of-app-with-serviceaccount-with-argocd).

```yaml
argocd:
  applications: # list of Argo CD applications to be create
  - name: scenario-6 # required, must be unique
    project: default # default: default
    repoURL: https://github.com/pdettori/sample-apps.git
    targetRevision: HEAD # default: HEAD
    path: nginx
    destinationWDS: wds1
    destinationNamespace: nginx-sa # default: default
    syncPolicy: auto # default: manual
```

Alternatively, the same result can be achieved from Helm CLI by using the followig minimal argument (note that the default values are not explicitely set):

```shell
--set-json='argocd.applications=[ { "name": "scenario-6", "repoURL": "https://github.com/pdettori/sample-apps.git", "path": "nginx", "destinationWDS": "wds1", "destinationNamespace": "nginx-sa" } ]'
```
