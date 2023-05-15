# **KubeStellar** Quickstart

## Required Packages:
   - [docker](https://docs.docker.com/get-docker/)
   - [kind](https://kind.sigs.k8s.io/)
   - [kubectl](https://kubernetes.io/docs/tasks/tools/) (version range expected: 1.23-1.25)

## Setup Instructions

Table of contents:

- [1. Install and run **KubeStellar**](#1-install-kubestellar-pre-requisites)
- [2. Example deployment of nginx workload into two kind local clusters](#4-Example-deployment-of-nginx-workload-into-a-kind-local-cluster)
  - [a. Stand up two kind clusters: florin and guilder](#a-Stand-up-two-kind-clusters-florin-and-guilder)
  - [b. Create a KubeStellar Inventory Management Workspace (IMW) and Workload Management Workspace (WMW)](#b-onboarding-the-clusters)
  - [c. Onboarding the clusters](#b-onboarding-the-clusters)
  - [d. Create and deploy the nginx workload into florin and guilder clusters](#c-Create-and-deploy-the-nginx-workload-into-florin-and-guilder-clusters)
- [3. Cleanup the environment](#3-Cleanup-the-environment)


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
    ├── espw
```

## 2. Example deployment of nginx workload into two kind local clusters

### a. Stand up a local florin and guilder kind clusters

Create the first edge cluster:

```shell
kind create cluster --name florin
```  

Create the second edge cluster:

```shell
kind create cluster --name guilder
```  

### b. Create a KubeStellar Inventory Management Workspace (IMW) and Workload Management Workspace (WMW)

IMW are used by KubeStellar to store inventory objects (sync targets and placement). Create an IMW named `example-imw` with the following command:

```shell
kubectl ws root
kubectl ws create "example-imw"
```

WMW are used by KubeStellar to store workloads and edge placement objects. Create an WMW named `example-wmw` in a my-org workspace with the following command:

```shell
kubectl ws root
kubectl ws create "my-org"
ensure-wmw.sh "example-wmw"
```
 
### c. Onboarding the clusters

Create a syncTarget and location inventory objects to represent the `florin` cluster:

```shell
kubectl kcp-edge prep-for-cluster --imw root:example-imw florin  env=prod
```

The following commands list the objects that were created:

```console
$ kubectl get locations,synctargets
NAME                                RESOURCE      AVAILABLE   INSTANCES   LABELS   AGE
location.scheduling.kcp.io/florin   synctargets   0           1                    57s

NAME                                AGE
synctarget.workload.kcp.io/florin   58s
```

Generate the edge syncer manifest:

```shell
kubectl ws root:espw
kubestellar --syncer florin  # replaces: mailbox-prep.sh florin
```


which should yield something like:

```console
Current workspace is "root:espw:19igldm1mmolruzr-mb-6b0309f0-84f3-4926-9344-81df2f989f69" (type root:universal).

Creating service account "kubestellar-syncer-florin-5c4r0a44"
Creating cluster role "kubestellar-syncer-florin-5c4r0a44" to give service account "kubestellar-syncer-florin-5c4r0a44"

1. write and sync access to the synctarget "kubestellar-syncer-florin-5c4r0a44"
2. write access to apiresourceimports.

Creating or updating cluster role binding "kubestellar-syncer-florin-5c4r0a44" to bind service account "kubestellar-syncer-florin-5c4r0a44" to cluster role "kubestellar-syncer-florin-5c4r0a44".

Wrote physical cluster manifest to florin-syncer.yaml for namespace "kubestellar-syncer-florin-5c4r0a44". Use

  KUBECONFIG=<edge-cluster-config> kubectl apply -f "florin-syncer.yaml"

to apply it. Use

  KUBECONFIG=<edge-cluster-config> kubectl get deployment -n "kubestellar-syncer-florin-5c4r0a44" kubestellar-syncer-florin-5c4r0a44

to verify the syncer pod is running.
```

An edge syncer manifest yaml file is created in your current director: `florin-syncer.yaml`. The default for the output file is the name of the SyncTarget object with “-syncer.yaml” appended.

Now deploy the edge syncer to the `florin` edge cluster:

  
```shell
KUBECONFIG=$florin_kubeconfig kubectl apply -f florin-syncer.yaml
```

which should yield something like:

```console
namespace/kubestellar-syncer-florin-5c4r0a44 created
serviceaccount/kubestellar-syncer-florin-5c4r0a44 created
secret/kubestellar-syncer-florin-5c4r0a44-token created
clusterrole.rbac.authorization.k8s.io/kubestellar-syncer-florin-5c4r0a44 created
clusterrolebinding.rbac.authorization.k8s.io/kubestellar-syncer-florin-5c4r0a44 created
role.rbac.authorization.k8s.io/kubestellar-dns-florin-5c4r0a44 created
rolebinding.rbac.authorization.k8s.io/kubestellar-dns-florin-5c4r0a44 created
secret/kubestellar-syncer-florin-5c4r0a44 created
deployment.apps/kubestellar-syncer-florin-5c4r0a44 created
```

Optionally, check that the edge syncer pod is running:

```console
KUBECONFIG=$florin_kubeconfig kubectl get pods -A
```

which should yield something like:

```console
NAMESPACE                         NAME                                              READY   STATUS    RESTARTS   AGE
kubestellar-syncer-florin-5c4r0a44   kubestellar-syncer-florin-5c4r0a44-bb8c8db4b-ng8sz   1/1     Running   0          30s
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

Now, let's onboard the `guilder` cluster:

```shell
kubectl kcp-edge prep-for-cluster --imw root:example-imw guilder env=prod
```

Apply the edge syncer manifest:

```shell
KUBECONFIG=$guilder_kubeconfig kubectl apply -f guilder-syncer.yaml
```


### e. Create and deploy the nginx workload into florin and guilder clusters

Create the `EdgePlacement` object for your workload. Its “where predicate” (the locationSelectors array) has one label selector that matches the Location objects (`florin` and `guilder`) created earlier, thus directing the workload to both edge clusters.

In the `wmw-1` workspace create the following `EdgePlacement` object: 
  
```console
kubectl ws root:my-org:wmw-1
```

```console
  kubectl apply -f - <<EOF
  apiVersion: edge.kcp.io/v1alpha1
  kind: EdgePlacement
  metadata:
    name: edge-placement-c
  spec:
    locationSelectors:
    - matchLabels: {"env":"prod"}
    namespaceSelector:
      matchLabels: {"common":"si"}
    nonNamespacedObjects:
    - apiGroup: apis.kcp.io
      resources: [ "apibindings" ]
      resourceNames: [ "bind-kube" ]
    upsync:
    - apiGroup: "group1.test"
      resources: ["sprockets", "flanges"]
      namespaces: ["orbital"]
      names: ["george", "cosmo"]
    - apiGroup: "group2.test"
      resources: ["cogs"]
      names: ["william"]
  EOF
```

Deploy the nginx workload. Note the namespace label matches the label in the namespaceSelector for the EdgePlacement (`edge-placement-c`) object created above. 

```console
  kubectl apply -f - <<EOF
  apiVersion: v1
  kind: Namespace
  metadata:
    name: commonstuff
    labels: {common: "si"}
  ---
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: nginx-deployment
    namespace: commonstuff
    labels:
      app: nginx
  spec:
    replicas: 3
    selector:
      matchLabels:
        app: nginx
    template:
      metadata:
        labels:
          app: nginx
      spec:
        containers:
        - name: nginx
          image: nginx:1.14.2
          ports:
          - containerPort: 80
  EOF
  ```

Now, let's check that the deployment was created in the `florin` edge cluster:

```shell
KUBECONFIG=$florin_kubeconfig kubectl -n commonstuff get deployment
```

which should yield something like:

```console
NAME               READY   UP-TO-DATE   AVAILABLE   AGE
nginx-deployment   3/3     3            3           8m37s
```

Also, let's check that the deployment was created in the `guilder` edge cluster:

```shell
KUBECONFIG=$guilder_kubeconfig kubectl -n commonstuff get deployment
```

which should yield something like:

```console
NAME               READY   UP-TO-DATE   AVAILABLE   AGE
nginx-deployment   3/3     3            3           8m37s
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
    This is a special web site.
  </body>
</html>
```

Congratulations, you’ve just deployed a workload to two edge clusters using kubestellar! To learn more about kubestellar please visit our [User Guide](<place-holder>)

## 3. Cleanup the environment

To uninstall kubestellar run the following command:

```bash
kubestellar stop
```

To uninstall kcp, kubestellar and delete all the generated files (e.g., edge syncer manifests and logs files) run the following command:

```shell
kubestellar cleanup
```
