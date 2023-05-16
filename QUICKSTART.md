# **KubeStellar** Quickstart

## Required Packages:
   - [docker](https://docs.docker.com/engine/install/)
   - [kind](https://kind.sigs.k8s.io/)
   - [kubectl](https://kubernetes.io/docs/tasks/tools/) (version range expected: 1.23-1.25)

## Setup Instructions

Table of contents:

- [1. Install and run **KubeStellar**](#1-install-and-run-kubestellar)
- [2. Example deployment of Apache HTTP Server workload into two kind local clusters](#2-example-deployment-of-apache-http-server-workload-into-two-kind-local-clusters)
  - [a. Stand up two kind clusters: florin and guilder](#a-stand-up-two-kind-clusters-florin-and-guilder)
  - [b. Create a KubeStellar Inventory Management Workspace (IMW) and Workload Management Workspace (WMW)](#b-create-a-kubestellar-inventory-management-workspace-imw-and-workload-management-workspace-wmw)
  - [c. Onboarding the clusters](#c-onboarding-the-clusters)
  - [d. Create and deploy the Apache Server workload into florin and guilder clusters](#e-create-and-deploy-the-apache-server-workload-into-florin-and-guilder-clusters)
- [3. Teardown the environment](#3-teardown-the-environment)


This guide is intended to show how to quickly bring up a **KubeStellar** environment with its dependencies from a binary release.

## 1. Install and run **KubeStellar**

KubeStellar works in the context of kcp, so to use KubeStellar you also need kcp. Download the kcp and **KubeStellar** binaries and scripts into a `kubestellar` subfolder in your current working directory using the following command:

```shell
bash <(curl -s https://raw.githubusercontent.com/francostellari/edge-mc/main/hack/kubestellar-bootstrap.sh) --kcp-version v0.11.0 --kubestellar-version v0.1.0 --folder . --create-folder
export PATH="$PATH:$(pwd)/kcp/bin:$(pwd)/kubestellar/bin"
export KUBECONFIG="$(pwd)/.kcp/admin.kubeconfig"
```

Check that `KubeStellar` is running:

First, check that controllers are running with the following command:

```shell
ps aux | grep -e mailbox-controller -e placement-translator -e scheduler
```

which should yield something like:

```console
user     1892  0.0  0.3 747644 29628 pts/1    Sl   10:51   0:00 mailbox-controller --inventory-context=root --mbws-context=base -v=2
user     1902  0.3  0.3 743652 27504 pts/1    Sl   10:51   0:02 scheduler -v 2 --root-user kcp-admin --root-cluster root --sysadm-context system:admin --sysadm-user shard-admin
user     1912  0.3  0.5 760428 41660 pts/1    Sl   10:51   0:02 placement-translator --allclusters-context system:admin -v=2
```

Second, check that the Edge Service Provider Workspace (`espw`) is created with the following command:

```shell
kubectl ws tree
```

which should yield:

```console
kubectl ws tree
.
└── root
    ├── compute
    └── espw
```

## 2. Example deployment of Apache HTTP Server workload into two kind local clusters

### a. Stand up a local florin and guilder kind clusters

Create the first edge cluster:

```shell
kind create cluster --name florin --config examples/florin-config.yaml
```  

Create the second edge cluster:

```shell
kind create cluster --name guilder --config examples/guilder-config.yaml
```  

### b. Create a KubeStellar Inventory Management Workspace (IMW) and Workload Management Workspace (WMW)

IMW are used by KubeStellar to store inventory objects (sync targets and placement). Create an IMW named `example-imw` with the following command:

```shell
kubectl config use-context root
kubectl ws root
kubectl ws create "example-imw"
```

WMW are used by KubeStellar to store workloads and edge placement objects. Create an WMW named `example-wmw` in a `my-org` workspace with the following command:

```shell
kubectl ws root
kubectl ws create "my-org" --enter
kubectl kubestellar ensure wmw example-wmw
```

### c. Onboarding the clusters

Let's begin by onboarding the `florin` cluster:

```shell
kubectl kubestellar prep-for-cluster --imw root:example-imw florin  env=prod
```

which should yield something like:

```console
Current workspace is "root:example-imw".
synctarget.workload.kcp.io/florin created
location.scheduling.kcp.io/florin created
synctarget.workload.kcp.io/florin labeled
location.scheduling.kcp.io/florin labeled
Current workspace is "root:example-imw".
Current workspace is "root:espw".
Current workspace is "root:espw:9nemli4rpx83ahnz-mb-c44d04db-ae85-422c-9e12-c5e7865bf37a" (type root:universal).
Creating service account "kcp-edge-syncer-florin-1yi5q9c4"
Creating cluster role "kcp-edge-syncer-florin-1yi5q9c4" to give service account "kcp-edge-syncer-florin-1yi5q9c4"

 1. write and sync access to the synctarget "kcp-edge-syncer-florin-1yi5q9c4"
 2. write access to apiresourceimports.

Creating or updating cluster role binding "kcp-edge-syncer-florin-1yi5q9c4" to bind service account "kcp-edge-syncer-florin-1yi5q9c4" to cluster role "kcp-edge-syncer-florin-1yi5q9c4".

Wrote physical cluster manifest to florin-syncer.yaml for namespace "kcp-edge-syncer-florin-1yi5q9c4". Use

  KUBECONFIG=<pcluster-config> kubectl apply -f "florin-syncer.yaml"

to apply it. Use

  KUBECONFIG=<pcluster-config> kubectl get deployment -n "kcp-edge-syncer-florin-1yi5q9c4" kcp-edge-syncer-florin-1yi5q9c4

to verify the syncer pod is running.
Current workspace is "root:example-imw".
Current workspace is "root".
```

An edge syncer manifest yaml file is created in your current director: `florin-syncer.yaml`. The default for the output file is the name of the SyncTarget object with “-syncer.yaml” appended.

Now le's deploy the edge syncer to the `florin` edge cluster:

  
```shell
KUBECONFIG=$florin_kubeconfig kubectl apply -f florin-syncer.yaml
```

which should yield something like:

```console
namespace/kcp-edge-syncer-florin-1yi5q9c4 created
serviceaccount/kcp-edge-syncer-florin-1yi5q9c4 created
secret/kcp-edge-syncer-florin-1yi5q9c4-token created
clusterrole.rbac.authorization.k8s.io/kcp-edge-syncer-florin-1yi5q9c4 created
clusterrolebinding.rbac.authorization.k8s.io/kcp-edge-syncer-florin-1yi5q9c4 created
secret/kcp-edge-syncer-florin-1yi5q9c4 created
deployment.apps/kcp-edge-syncer-florin-1yi5q9c4 created
```

Optionally, check that the edge syncer pod is running:

```console
KUBECONFIG=$florin_kubeconfig kubectl get pods -A
```

which should yield something like:

```console
NAMESPACE                         NAME                                               READY   STATUS    RESTARTS   AGE
kcp-edge-syncer-florin-1yi5q9c4   kcp-edge-syncer-florin-1yi5q9c4-77cb588c89-xc5qr   1/1     Running   0          12m
kube-system                       coredns-565d847f94-92f4k                           1/1     Running   0          58m
kube-system                       coredns-565d847f94-kgddm                           1/1     Running   0          58m
kube-system                       etcd-florin-control-plane                          1/1     Running   0          58m
kube-system                       kindnet-p8vkv                                      1/1     Running   0          58m
kube-system                       kube-apiserver-florin-control-plane                1/1     Running   0          58m
kube-system                       kube-controller-manager-florin-control-plane       1/1     Running   0          58m
kube-system                       kube-proxy-jmxsg                                   1/1     Running   0          58m
kube-system                       kube-scheduler-florin-control-plane                1/1     Running   0          58m
local-path-storage                local-path-provisioner-684f458cdd-kw2xz            1/1     Running   0          58m

``` 

Now, let's onboard the `guilder` cluster:

```shell
kubectl kcp-edge prep-for-cluster --imw root:example-imw guilder env=prod
```

Apply the created edge syncer manifest:

```shell
KUBECONFIG=$guilder_kubeconfig kubectl apply -f guilder-syncer.yaml
```


### e. Create and deploy the Apache Server workload into florin and guilder clusters

Create the `EdgePlacement` object for your workload. Its “where predicate” (the locationSelectors array) has one label selector that matches the Location objects (`florin` and `guilder`) created earlier, thus directing the workload to both edge clusters.

In the `wmw-1` workspace create the following `EdgePlacement` object: 
  
```shell
kubectl ws root:my-org:example-wmw
kubectl apply -f examples/common-placement.yaml
```

Deploy the Apache HTTP Server workload. Note the namespace label matches the label in the namespaceSelector for the EdgePlacement (`edge-placement-c`) object created above. 

```shell
kubectl apply -f examples/common-workload.yaml
```

Now, let's check that the deployment was created in the `florin` edge cluster:

```shell
KUBECONFIG=$florin_kubeconfig kubectl get deployments -A
```

which should yield something like:

```console
NAMESPACE     NAME      READY   UP-TO-DATE   AVAILABLE   AGE
commonstuff   commond   0/0     0            0           6m44s
```

Also, let's check that the deployment was created in the `guilder` edge cluster:

```shell
KUBECONFIG=$guilder_kubeconfig kubectl get deployments -A
```

which should yield something like:

```console
NAMESPACE     NAME      READY   UP-TO-DATE   AVAILABLE   AGE
commonstuff   commond   0/0     0            0           6m44s
```

Lastly, let's check that the workload is working in both clusters:

For `florin`:

```console
$ curl http://localhost:8081
<!DOCTYPE html>
<html>
  <body>
    This is a common web site.
  </body>
</html>
```

For `guilder`:

```console
$ curl http://localhost:8082
<!DOCTYPE html>
<html>
  <body>
    This is a common web site.
  </body>
</html>
```

Congratulations, you’ve just deployed a workload to two edge clusters using kubestellar! To learn more about kubestellar please visit our [User Guide](<place-holder>)

## 3. Teardown the environment

To uninstall kubestellar run the following command:

```bash
kubestellar stop
```

To uninstall kcp, kubestellar and delete all the generated files (e.g., edge syncer manifests and logs files) run the following command:

```shell
kubestellar cleanup
```
