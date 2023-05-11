# **KCP-Edge** Quickstart

## Required Packages:
   - [docker](https://docs.docker.com/get-docker/)
   - [kind](https://kind.sigs.k8s.io/)
   - kubectl (version range expected: 1.23-1.25)

## Setup Instructions

Table of contents:

- [1. Install and run **KCP-Edge**](#1-install-kcp-edge-pre-requisites)
- [2. Example deployment of nginx workload into two kind local clusters](#4-Example-deployment-of-nginx-workload-into-a-kind-local-cluster)
  - [a. Stand up two kind clusters: florin and guilder](#a-Stand-up-a-local-florin-kind-cluster)
  - [b. Onboarding the florin cluster](#b-Create-a-sync-target-placement-and-edge-syncer-for-onboarding-the-created-florin-edge-cluster)
  - [c. Onboarding the guilder cluster](#b-Create-a-sync-target-placement-and-edge-syncer-for-onboarding-the-created-florin-edge-cluster)
  - [d. Create the nginx workload and deploy it to the florin and guilder clusters](#c-Create-the-nginx-workload-and-deploy-it-to-the-florin-cluster)
- [3. Cleanup the environment](#5-Cleanup-the-environment)


This guide is intended to show how to quickly bring up a **KCP-Edge** environment with its dependencies from a binary release.

## 1. Install and run **KCP-Edge**

Download the kcp **KCP-Edge** binaries and scripts into a `kcp-edge` subfolder in your current working directory using the following command:

```shell
bash <(curl -s https://raw.githubusercontent.com/francostellari/edge-mc/main/hack/kcp-edge-bootstrap.sh) --kcp-version v0.11.0 --kcp-edge-version v0.1.0 --folder . --create-folder
export PATH="$PATH:$(pwd)/kcp/bin:$(pwd)/kcp-edge/bin"
export KUBECONFIG="$(pwd)/.kcp/admin.kubeconfig"
```

Check that `KCP-Edge` is running:

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
    │   
    ├── imw-1
    └── my-org
        ├── wmw-c
        └── wmw-s
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

### b. Onboarding the florin cluster

Create a syncTarget and location inventory objects to represent the edge cluster (`florin`):

```shell
kcp-edge --create_inv_item florin  env=prod    # replaces ensure-location.sh florin  env=prod
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
kcp-edge --syncer florin  # replaces: mailbox-prep.sh florin
```


which should yield something like:

```console
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

An edge syncer manifest yaml file is created in your current director: `florin-syncer.yaml`. The default for the output file is the name of the SyncTarget object with “-syncer.yaml” appended.

Now deploy the edge syncer to the `florin` edge cluster:

  
```shell
KUBECONFIG=$florin_kubeconfig kubectl apply -f florin-syncer.yaml
```

which should yield something like:

```console
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
KUBECONFIG=$florin_kubeconfig kubectl get pods -A
```

which should yield something like:

```console
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

### c. Onboarding the guilder cluster

Similary, repeat the steps in for the guilder cluster:


```shell
kcp-edge --create_inv_item florin  env=prod    # replaces ensure-location.sh florin  env=prod
```

Generate and apply the edge syncer manifest:

```shell
kubectl ws root:espw
kcp-edge --syncer florin  # replaces: mailbox-prep.sh florin
```

```shell
KUBECONFIG=$florin_kubeconfig kubectl apply -f florin-syncer.yaml
```


### d. Create the nginx workload and deploy it to the florin cluster

Create the `EdgePlacement` object for your workload. Its “where predicate” (the locationSelectors array) has one label selector that matches the Location object (`florin`) created earlier, thus directing the workload to `florin` edge cluster.

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

Check that the deployment was created in the florin edge cluster:

```console
KUBECONFIG=$florin_kubeconfig kubectl -n commonstuff get deployment
```

which should yield something like:

```console
NAME               READY   UP-TO-DATE   AVAILABLE   AGE
nginx-deployment   3/3     3            3           8m37s
```

Also, check that the pods are running in the florin edge cluster:

```console
KUBECONFIG=$florin_kubeconfig kubectl -n commonstuff get pods
```

which should yield something like:

```console
NAME                                READY   STATUS    RESTARTS   AGE
nginx-deployment-7fb96c846b-2hkwt   1/1     Running   0          8m57s
nginx-deployment-7fb96c846b-9lxtc   1/1     Running   0          8m57s
nginx-deployment-7fb96c846b-k8pp7   1/1     Running   0          8m57s
```

## 3. Cleanup the environment

To uninstall kcp-edge run the following command:

```bash
kcp-edge stop
```

To uninstall kcp, kcp-edge and delete all the generated files (e.g., edge syncer manifests and logs files) run the following command:

```shell
kcp-edge cleanup
```
