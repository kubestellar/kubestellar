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

## 3. Create a **KCP-Edge** Inventory Management Workspace (IMW)

**IMW** are used by **KCP-Edge** to store *sync targets* and *placement* objects.
Create an **IMW** named `imw-1` with the following command:

```bash
kubectl ws root
kubectl ws create "imw-1"
```

## 4. Create a **KCP-Edge** Workload Management Workspace (WMW)

**WMW** are used by **KCP-Edge** to store *workloads* amd *edge placement* objects.
Create an **WMW** named `wmw-1` in a `my-org` workspace with the following command:

```bash
kubectl ws root
kubectl ws create "my-org"
ensure-wmw.sh "wmw-1"
```
