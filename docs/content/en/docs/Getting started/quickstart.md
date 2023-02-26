---
categories: ["Quickstart"]
tags: ["start"]
title: "KCP-Edge Quickstart Guide"
linkTitle: "KCP-Edge Quickstart Guide"
date: 2023-02-25
description: >
---

<!-- {{% pageinfo %}}
This document provides instructions on how to build and run KCP-Edge locally.
{{% /pageinfo %}} -->

To use components from KCP-Edge you must:
1. Install and configure KCP to create a working KCP environment
2. Build KCP-Edge

## 1. Install and Configure KCP
### Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
- A Kubernetes cluster (for local testing, consider [kind](http://kind.sigs.k8s.io))

### Download kcp

Visit our [latest release page](https://github.com/kcp-dev/kcp/releases/latest) and download `kcp`
and `kubectl-kcp-plugin` that match your operating system and architecture.

Extract `kcp` and `kubectl-kcp-plugin` and place all the files in the `bin` directories somewhere in your `$PATH`.

### Start kcp

You can start kcp using this command:

```shell
kcp start
```

This launches kcp in the foreground. You can press `ctrl-c` to stop it.

To see a complete list of server options, run `kcp start options`.

### Set your KUBECONFIG

During its startup, kcp generates a kubeconfig in `.kcp/admin.kubeconfig`. Use this to connect to kcp and display the
version to confirm it's working:

```shell
$ export KUBECONFIG=.kcp/admin.kubeconfig
$ kubectl version
WARNING: This version information is deprecated and will be replaced with the output from kubectl version --short.  Use --output=yaml|json to get the full version.
Client Version: version.Info{Major:"1", Minor:"24", GitVersion:"v1.24.4", GitCommit:"95ee5ab382d64cfe6c28967f36b53970b8374491", GitTreeState:"clean", BuildDate:"2022-08-17T18:46:11Z", GoVersion:"go1.19", Compiler:"gc", Platform:"darwin/amd64"}
Kustomize Version: v4.5.4
Server Version: version.Info{Major:"1", Minor:"24", GitVersion:"v1.24.3+kcp-v0.8.0", GitCommit:"41863897", GitTreeState:"clean", BuildDate:"2022-09-02T18:10:37Z", GoVersion:"go1.18.5", Compiler:"gc", Platform:"darwin/amd64"}
```

### Configure kcp to sync to your cluster

kcp can't run pods by itself - it needs at least one physical cluster for that. For this example, we'll be using a
local `kind` cluster.  It does not have to exist yet.

In this recipe we use the root workspace to hold the description of the workload and where it goes.  These usually would go elsewhere, but we use the root workspace here for simplicity.

Run the following command to tell kcp about the `kind` cluster (replace the syncer image tag as needed; CI now puts built images in https://github.com/orgs/kcp-dev/packages):

```shell
$ kubectl kcp workload sync kind --syncer-image ghcr.io/kcp-dev/kcp/syncer:v0.10.0 -o syncer-kind-main.yaml
Creating synctarget "kind"
Creating service account "kcp-syncer-kind-25coemaz"
Creating cluster role "kcp-syncer-kind-25coemaz" to give service account "kcp-syncer-kind-25coemaz"

 1. write and sync access to the synctarget "kcp-syncer-kind-25coemaz"
 2. write access to apiresourceimports.

Creating or updating cluster role binding "kcp-syncer-kind-25coemaz" to bind service account "kcp-syncer-kind-25coemaz" to cluster role "kcp-syncer-kind-25coemaz".

Wrote physical cluster manifest to syncer-kind-main.yaml for namespace "kcp-syncer-kind-25coemaz". Use

  KUBECONFIG=<pcluster-config> kubectl apply -f "syncer-kind-main.yaml"

to apply it. Use

  KUBECONFIG=<pcluster-config> kubectl get deployment -n "kcp-syncer-kind-25coemaz" kcp-syncer-kind-25coemaz

to verify the syncer pod is running.
```

Next, we need to install the syncer pod on our `kind` cluster - this is what actually syncs content from kcp to the
physical cluster. The kind cluster needs to be running by now. Run the following command:

```shell
$ KUBECONFIG=</path/to/kind/kubeconfig> kubectl apply -f "syncer-kind-main.yaml"
namespace/kcp-syncer-kind-25coemaz created
serviceaccount/kcp-syncer-kind-25coemaz created
secret/kcp-syncer-kind-25coemaz-token created
clusterrole.rbac.authorization.k8s.io/kcp-syncer-kind-25coemaz created
clusterrolebinding.rbac.authorization.k8s.io/kcp-syncer-kind-25coemaz created
secret/kcp-syncer-kind-25coemaz created
deployment.apps/kcp-syncer-kind-25coemaz created
```

### Bind to workload APIs and create default placement

If you are running kcp version v0.10.0 or higher, you will need to run the following commmand (continuing in the `root` workspace)
to create a binding to the workload APIs export and a default placement for your physical cluster:

```shell
$ kubectl kcp bind compute root
Binding APIExport "root:compute:kubernetes".
placement placement-1pfxsevk created.
Placement "placement-1pfxsevk" is ready.
```

### Create a deployment in kcp

Let's create a deployment in our kcp workspace and see it get synced to our cluster:

```shell
$ kubectl create deployment --image=gcr.io/kuar-demo/kuard-amd64:blue --port=8080 kuard
deployment.apps/kuard created
```

Once your cluster has pulled the image and started the pod, you should be able to verify the deployment is running in
kcp:

```shell
$ kubectl get deployments
NAME    READY   UP-TO-DATE   AVAILABLE   AGE
kuard   1/1     1            1           3s
```

We are still working on adding support for `kubectl logs`, `kubectl exec`, and `kubectl port-forward` to kcp. For the
time being, you can check directly in your cluster.

kcp translates the names of namespaces in workspaces to unique names in a physical cluster. We first must get this
translated name; if you're following along, your translated name might be different.

```shell
$ KUBECONFIG=</path/to/kind/kubeconfig> kubectl get pods --all-namespaces --selector app=kuard
NAMESPACE          NAME                     READY   STATUS    RESTARTS   AGE
kcp-26zq2mc2yajx   kuard-7d49c786c5-wfpcc   1/1     Running   0          4m28s
```

Now we can e.g. check the pod logs:

```shell
$ KUBECONFIG=</path/to/kind/kubeconfig> kubectl --namespace kcp-26zq2mc2yajx logs deployment/kuard | head
2022/09/07 14:04:35 Starting kuard version: v0.10.0-blue
2022/09/07 14:04:35 **********************************************************************
2022/09/07 14:04:35 * WARNING: This server may expose sensitive
2022/09/07 14:04:35 * and secret information. Be careful.
2022/09/07 14:04:35 **********************************************************************
2022/09/07 14:04:35 Config:
{
  "address": ":8080",
  "debug": false,
  "debug-sitedata-dir": "./sitedata",
```

## 2. Build and Install KCP-Edge
To build and install KCP-Edge components:
- clone the latest KCP-Edge [repository](https://github.com/kcp-dev/edge-mc)
- make tools
- make build
- make install

The [Makefile](Makefile) provides a set of targets to help simplify the build
tasks.


```sh
make build
```

The following targets can be used to lint, test and build the KCP-Edge controllers:
```sh
make golint

make test

make install
```



## Next steps

Thanks for checking out our quickstart!

If you're interested in learning more about all the features KCP and KCP-Edge have to offer, please check out our additional
documentation:

### KCP
- [Concepts](https://github.com/kcp-dev/kcp/tree/main/docs/content/en/main/concepts/concepts.md) - a high level overview of kcp concepts
- [Workspaces](https://github.com/kcp-dev/kcp/tree/main/docs/content/en/main/concepts/workspaces.md) - a more thorough introduction on kcp's workspaces
- [Locations & scheduling](https://github.com/kcp-dev/kcp/tree/main/docs/content/en/main/concepts/locations-and-scheduling.md) - details on kcp's primitives that abstract over clusters
- [Syncer](https://github.com/kcp-dev/kcp/tree/main/docs/content/en/main/concepts/syncer.md) - information on running the kcp agent that syncs content between kcp and a physical cluster
- [kubectl plugin](https://github.com/kcp-dev/kcp/tree/main/docs/content/en/main/concepts/kubectl-kcp-plugin.md)
- [Authorization](https://github.com/kcp-dev/kcp/tree/main/docs/content/en/main/concepts/authorization.md) - how kcp manages access control to workspaces and content
- [Virtual workspaces](https://github.com/kcp-dev/kcp/tree/main/docs/content/en/main/concepts/virtual-workspaces.md) - details on kcp's mechanism for virtual views of workspace content

### KCP-Edge
TBD

## Contributing

We ❤️ our contributors! If you're interested in helping us out, please head over to our [Contributing](CONTRIBUTING.md).

## Getting in touch

There are several ways to communicate with us:

- The [`#kcp-dev` channel](https://app.slack.com/client/T09NY5SBT/C021U8WSAFK) in the [Kubernetes Slack workspace](https://slack.k8s.io)
- Our mailing lists:
    - [kcp-dev](https://groups.google.com/g/kcp-dev) for development discussions

## Additional references

- [Let's Learn kcp - A minimal Kubernetes API server with Saiyam Pathak - July 7, 2021](https://www.youtube.com/watch?v=M4mn_LlCyzk)
