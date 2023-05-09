# **KCP-Edge** quick start

Table of contents:

- [**KCP-Edge** quick start](#kcp-edge-quick-start)
  - [1. Install **KCP-Edge** pre-requisites](#1-install-kcp-edge-pre-requisites)
    - [a. `kubectl`](#a-kubectl)
    - [b. `kcp`](#b-kcp)
  - [2. Install and run **KCP-Edge**](#2-install-and-run-kcp-edge)
  - [3. Create a **KCP-Edge** Inventory Management Workspace (IMW)](#3-create-a-kcp-edge-inventory-management-workspace-imw)
  - [4. Create a **KCP-Edge** Workload Management Workspace (WMW)](#4-create-a-kcp-edge-workload-management-workspace-wmw)

This guide is intended to show how to quickly bring up a **KCP-Edge** environment with its dependencies from a binary release.

## 1. Install **KCP-Edge** pre-requisites

### a. `kubectl`

Detailed installation instructions for different operative systems are available [here](https://kubernetes.io/docs/tasks/tools/).

### b. `kcp`

Since `KCP-Edge` leverages [`kcp`](kcp.io) logical workspace virtualization capability, we have to install `kcp` v0.11.0 binaries first by following the detailed installation instructions available [here](https://docs.kcp.io/kcp/main/) or by using the convenience script below:

```bash
bash <(curl -s https://raw.githubusercontent.com/kcp-dev/edge-mc/main/hack/install-kcp-with-plugins.sh) --version v0.11.0 --folder $(pwd)/kcp --create-folder
export PATH="$PATH:$(pwd)/kcp/bin"
```

Run **kcp** with the following command:

```bash
kcp start >& kcp_log.txt &
export KUBECONFIG="$(pwd)/.kcp/admin.kubeconfig"
```

It should be noted that, when **kcp** is run with the command above, it listens to all of the host's non-loopback addresses and pick one to put in the generated `admin.kubeconfig` file. While this works fine in most cases, such as when using a kind cluster locally (see later), sometimes it may be necessary to specify a public/external ip address that can be reached by remote clusters. For this purpose the following command can be used:

```bash
kcp start --bind-address <ip_address> >& kcp_log.txt &
export KUBECONFIG="$(pwd)/.kcp/admin.kubeconfig"
```

In this case, the specified `<ip_address>` will appear in the generated `admin.kubeconfig` file the `kcp` server will listen only at that address.

After few seconds, check that `kcp` is running using the command:

```bash
kubectl version --short
```

which should yield something like:

```text
Client Version: v1.25.3
Kustomize Version: v4.5.7
Server Version: v1.24.3+kcp-v0.11.0
```

Additionally, one can check the available virtual workspaces using the command:

```bash
kubectl ws tree
```

which should yield something like:

```text
.
└── root
    └── compute
```

## 2. Install and run **KCP-Edge**

**KCP-Edge** v0.1.0 release binaries and scripts can be easily installed in a `kcp-edge` subfolder of the current working directory using the following command:

```bash
bash <(curl -s https://raw.githubusercontent.com/kcp-dev/edge-mc/main/hack/install-kcp-edge.sh) --version v0.1.0 --folder $(pwd)/kcp-edge --create-folder
export PATH="$PATH:$(pwd)/kcp-edge/bin"
```

After installation, **KCP-Edge** can be run with the following command:

```bash
# NOTE: need to change the command below when merged into edge-mc or included in release binaries
bash <(curl -s https://raw.githubusercontent.com/dumb0002/edge-mc/script/hack/kcp-edge.sh) start --user kit
```

Check that `KCP-Edge` controllers are running with the following command:

```bash
ps aux | grep -e mailbox-controller -e placement-translator -e scheduler
```

which should yield something like:

```text
user     1892  0.0  0.3 747644 29628 pts/1    Sl   10:51   0:00 mailbox-controller --inventory-context=root --mbws-context=base -v=2
user     1902  0.3  0.3 743652 27504 pts/1    Sl   10:51   0:02 scheduler -v 2 --root-user kcp-admin --root-cluster root --sysadm-context system:admin --sysadm-user shard-admin
user     1912  0.3  0.5 760428 41660 pts/1    Sl   10:51   0:02 placement-translator --allclusters-context system:admin -v=2
```

Check that the Edge Service Provider Workspace (`espw`) is created with the following command:

```bash
kubectl ws tree
```

which should yield:

```text
.
└── root
    ├── compute
    └── espw
```

## 3. Create a kcp-edge Inventory Management Workspace (IMW) and Workload Management Workspace (WMW)

**IMW** are used by **KCP-Edge** to store *sync targets* and *placement* objects.
Create an **IMW** named `imw-1` with the following command:

```bash
kubectl ws root
kubectl ws create "imw-1"
```

**WMW** are used by **KCP-Edge** to store *workloads* amd *edge placement* objects.
Create an **WMW** named `wmw-1` in a `my-org` workspace with the following command:

```bash
kubectl ws root
kubectl ws create "my-org"
ensure-wmw.sh "wmw-1"
```


## 4. Example deployment of nginx workload into a *kind local cluster

a. Stand up a local florin kind cluster

```shell
kind create cluster --name florin
```  

b. Create a sync target, placement, and edge syncer for onboarding the created florin edge cluster

Create a syncTarget and location objects to represent florin:

```console
ensure-location.sh florin  env=prod
```

The following commands list the objects that were created:

```console
$ kubectl get locations,synctargets
NAME                                RESOURCE      AVAILABLE   INSTANCES   LABELS   AGE
location.scheduling.kcp.io/florin   synctargets   0           1                    57s

NAME                                AGE
synctarget.workload.kcp.io/florin   58s
```

Generate the edge syncer manifest

```shell
kubectl ws root:espw
mailbox-prep.sh florin
```

```shell
Current workspace is "root:espw:19igldm1mmolruzr-mb-6b0309f0-84f3-4926-9344-81df2f989f69" (type root:universal).

Creating service account "kcp-edge-syncer-florin-5c4r0a44"
Creating cluster role "kcp-edge-syncer-florin-5c4r0a44" to give service account "kcp-edge-syncer-florin-5c4r0a44"

1. write and sync access to the synctarget "kcp-edge-syncer-florin-5c4r0a44"
2. write access to apiresourceimports.

Creating or updating cluster role binding "kcp-edge-syncer-florin-5c4r0a44" to bind service account "kcp-edge-syncer-florin-5c4r0a44" to cluster role "kcp-edge-syncer-florin-5c4r0a44".

Wrote physical cluster manifest to florin-syncer.yaml for namespace "kcp-edge-syncer-florin-5c4r0a44". Use

  KUBECONFIG=<edge-cluster-config> kubectl apply -f "florin-syncer.yaml"

to apply it. Use

  KUBECONFIG=<edge-cluster-config> kubectl get deployment -n "kcp-edge-syncer-florin-5c4r0a44" kcp-edge-syncer-florin-5c4r0a44

to verify the syncer pod is running.
```

An edge syncer manifest yaml file is created in your current director: `florin-syncer.yaml`. The default for the output file is the name of the SyncTarget object with “-syncer.yaml” appended. Now deploy the edge syncer to florin edge cluster:

  
```console
$ KUBECONFIG=$florin_kubeconfig kubectl apply -f florin-syncer.yaml

namespace/kcp-edge-syncer-florin-5c4r0a44 created
serviceaccount/kcp-edge-syncer-florin-5c4r0a44 created
secret/kcp-edge-syncer-florin-5c4r0a44-token created
clusterrole.rbac.authorization.k8s.io/kcp-edge-syncer-florin-5c4r0a44 created
clusterrolebinding.rbac.authorization.k8s.io/kcp-edge-syncer-florin-5c4r0a44 created
role.rbac.authorization.k8s.io/kcp-edge-dns-florin-5c4r0a44 created
rolebinding.rbac.authorization.k8s.io/kcp-edge-dns-florin-5c4r0a44 created
secret/kcp-edge-syncer-florin-5c4r0a44 created
deployment.apps/kcp-edge-syncer-florin-5c4r0a44 created
```

Check that the edge syncer pod is running:

```console
$ KUBECONFIG=$florin_kubeconfig kubectl get pods -A
NAMESPACE                         NAME                                              READY   STATUS    RESTARTS   AGE
kcp-edge-syncer-florin-5c4r0a44   kcp-edge-syncer-florin-5c4r0a44-bb8c8db4b-ng8sz   1/1     Running   0          30s
kube-system                       coredns-565d847f94-kr2pw                          1/1     Running   0          85s
kube-system                       coredns-565d847f94-rj4s8                          1/1     Running   0          85s
kube-system                       etcd-florin-control-plane                         1/1     Running   0          99s
kube-system                       kindnet-l26qt                                     1/1     Running   0          85s
kube-system                       kube-apiserver-florin-control-plane               1/1     Running   0          100s
kube-system                       kube-controller-manager-florin-control-plane      1/1     Running   0          100s
kube-system                       kube-proxy-qzhx6                                  1/1     Running   0          85s
kube-system                       kube-scheduler-florin-control-plane               1/1     Running   0          99s
local-path-storage                local-path-provisioner-684f458cdd-75wv8           1/1     Running   0          85s
``` 



c. Create the nginx workload and edge placement to deploy it to the florin cluster
...

## 5. Cleanup the environment
