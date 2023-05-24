---
title: "KubeStellar Quickstart Guide"
linkTitle: "KubeStellar Quickstart Guide"
---

<img width="500px" src="../../KubeStellar with Logo.png" title="KubeStellar">

## Estimated Time: 
   ~3 minutes
   
## Required Packages:
   - [docker](https://docs.docker.com/engine/install/)
   - [kind](https://kind.sigs.k8s.io/)
   - [kubectl](https://kubernetes.io/docs/tasks/tools/) (version range expected: 1.23-1.25)

## Setup Instructions

Table of contents:

- [1. Install and run **KubeStellar**](#1-install-and-run-kubestellar)
- [2. Example deployment of Apache HTTP Server workload into two local kind clusters](#2-example-deployment-of-apache-http-server-workload-into-two-local-kind-clusters)
  - [a. Stand up two kind clusters: florin and guilder](#a-stand-up-two-kind-clusters-florin-and-guilder)
  - [b. Create a KubeStellar Inventory Management Workspace (IMW) and Workload Management Workspace (WMW)](#b-create-a-kubestellar-inventory-management-workspace-imw-and-workload-management-workspace-wmw)
  - [c. Onboarding the clusters](#c-onboarding-the-clusters)
  - [d. Create and deploy the Apache Server workload into florin and guilder clusters](#d-create-and-deploy-the-apache-server-workload-into-florin-and-guilder-clusters)
  - [e. Carrying on](#e-carrying-on)
- [3. Teardown the environment](#3-teardown-the-environment)


This guide is intended to show how to (1) quickly bring up a **KubeStellar** environment with its dependencies from a binary release and then (2) run through a simple example usage.

## 1. Install and run **KubeStellar**

KubeStellar works in the context of kcp, so to use KubeStellar you also need kcp. Download the kcp and **KubeStellar** binaries and scripts into a `kubestellar` subfolder in your current working directory using the following command:

```shell
bash <(curl -s https://raw.githubusercontent.com/kcp-dev/edge-mc/main/bootstrap/bootstrap-kubestellar.sh) --kubestellar-version latest
export PATH="$PATH:$(pwd)/kcp/bin:$(pwd)/kubestellar/bin"
export KUBECONFIG="$(pwd)/.kcp/admin.kubeconfig"
```

Check that `KubeStellar` is running:

First, check that controllers are running with the following command:

```shell
ps aux | grep -e mailbox-controller -e placement-translator -e kubestellar-scheduler
```

which should yield something like:

```console
user     1892  0.0  0.3 747644 29628 pts/1    Sl   10:51   0:00 mailbox-controller -v=2
user     1902  0.3  0.3 743652 27504 pts/1    Sl   10:51   0:02 scheduler -v 2 
user     1912  0.3  0.5 760428 41660 pts/1    Sl   10:51   0:02 placement-translator -v=2
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

## 2. Example deployment of Apache HTTP Server workload into two local kind clusters

In this example you will create two edge clusters and define one
workload that will be distributed from the center to those edge
clusters.  This example is similar to the one described more
expansively [on the
website](https://docs.kubestellar.io/release-v0.2.2/Coding%20Milestones/PoC2023q1/example1/),
but with the some steps reorganized and combined and the special
workload and summarization aspirations removed.

### a. Stand up two kind clusters: florin and guilder

Create the first edge cluster:

```shell
kind create cluster --name florin --config  kubestellar/examples/florin-config.yaml
```  

Note: if you already have a cluster named 'florin' from a previous exercise of KubeStellar, please delete the florin cluster ('kind delete cluster --name florin') and create it using the instruction above.

Create the second edge cluster:

```shell
kind create cluster --name guilder --config  kubestellar/examples/guilder-config.yaml
```  

Note: if you already have a cluster named 'guilder' from a previous exercise of KubeStellar, please delete the guilder cluster ('kind delete cluster --name guilder') and create it using the instruction above.

### b. Create a KubeStellar Inventory Management Workspace (IMW) and Workload Management Workspace (WMW)

IMWs are used by KubeStellar to store inventory objects (`SyncTargets` and `Locations`). Create an IMW named `example-imw` with the following command:

```shell
kubectl config use-context root
kubectl ws root
kubectl ws create example-imw
```

WMWs are used by KubeStellar to store workload descriptions and `EdgePlacement` objects. Create an WMW named `example-wmw` in a `my-org` workspace with the following commands:

```shell
kubectl ws root
kubectl ws create my-org --enter
kubectl kubestellar ensure wmw example-wmw
```

A WMW does not have to be created before the edge cluster is on-boarded; the WMW only needs to be created before content is put in it.

### c. Onboarding the clusters

Let's begin by onboarding the `florin` cluster:

```shell
kubectl ws root
kubectl kubestellar prep-for-cluster --imw root:example-imw florin env=prod
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

An edge syncer manifest yaml file was created in your current director: `florin-syncer.yaml`. The default for the output file is the name of the SyncTarget object with “-syncer.yaml” appended.

Now let's deploy the edge syncer to the `florin` edge cluster:

  
```shell
kubectl --context kind-florin apply -f florin-syncer.yaml
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

```shell
kubectl --context kind-florin get pods -A
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
kubectl ws root
kubectl kubestellar prep-for-cluster --imw root:example-imw guilder env=prod extended=si
```

Apply the created edge syncer manifest:

```shell
kubectl --context kind-guilder apply -f guilder-syncer.yaml
```


### d. Create and deploy the Apache Server workload into florin and guilder clusters

Create the `EdgePlacement` object for your workload. Its “where predicate” (the locationSelectors array) has one label selector that matches the Location objects (`florin` and `guilder`) created earlier, thus directing the workload to both edge clusters.

In the `example-wmw` workspace create the following `EdgePlacement` object: 
  
```shell linenums="1"
kubectl ws root:my-org:example-wmw

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

Put the prescription of the HTTP server workload into the WMW. Note the namespace label matches the label in the namespaceSelector for the EdgePlacement (`edge-placement-c`) object created above. 


```shell linenums="1"
kubectl apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: commonstuff
  labels: {common: "si"}
---
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: commonstuff
  name: httpd-htdocs
data:
  index.html: |
    <!DOCTYPE html>
    <html>
      <body>
        This is a common web site.
      </body>
    </html>
---
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: commonstuff
  name: commond
spec:
  selector: {matchLabels: {app: common} }
  template:
    metadata:
      labels: {app: common}
    spec:
      containers:
      - name: httpd
        image: library/httpd:2.4
        ports:
        - name: http
          containerPort: 80
          hostPort: 8094
          protocol: TCP
        volumeMounts:
        - name: htdocs
          readOnly: true
          mountPath: /usr/local/apache2/htdocs
      volumes:
      - name: htdocs
        configMap:
          name: httpd-htdocs
          optional: false
EOF
```

Now, let's check that the deployment was created in the `florin` edge cluster - it may take a few 10s of seconds to appear:

```shell
kubectl --context kind-florin get deployments -A
```

which should yield something like:

```console
NAMESPACE                         NAME                              READY   UP-TO-DATE   AVAILABLE   AGE
commonstuff                       commond                           1/1     1            1           6m48s
kcp-edge-syncer-florin-2upj1awn   kcp-edge-syncer-florin-2upj1awn   1/1     1            1           16m
kube-system                       coredns                           2/2     2            2           28m
local-path-storage                local-path-provisioner            1/1     1            1           28m
```

Also, let's check that the deployment was created in the `guilder` edge cluster:

```shell
kubectl --context kind-guilder get deployments -A
```

which should yield something like:

```console
NAMESPACE                          NAME                               READY   UP-TO-DATE   AVAILABLE   AGE
commonstuff                        commond                            1/1     1            1           7m54s
kcp-edge-syncer-guilder-6tuay5d6   kcp-edge-syncer-guilder-6tuay5d6   1/1     1            1           12m
kube-system                        coredns                            2/2     2            2           27m
local-path-storage                 local-path-provisioner             1/1     1            1           27m
```

Lastly, let's check that the workload is working in both clusters:

For `florin`:

```shell
curl http://localhost:8094
```
which should yield:

```html
<!DOCTYPE html>
<html>
  <body>
    This is a common web site.
  </body>
</html>
```
NOTE: if you receive the error: 'curl: (52) Empty reply from server', wait 2 minutes and attempt curl again.  It takes a minute for the Apache HTTP Server to synchronize and start.

For `guilder`:

```shell
curl http://localhost:8083
```
which should yield:

```html
<!DOCTYPE html>
<html>
  <body>
    This is a common web site.
  </body>
</html>
```

NOTE: if you receive the error: 'curl: (52) Empty reply from server', wait and attempt curl again.  It takes some time for the Apache HTTP Server to synchronize and start.

Congratulations, you’ve just deployed a workload to two edge clusters using kubestellar! To learn more about kubestellar please visit our [User Guide](user-guide.md)

### e. Carrying on

What you just did is part of the example [on the
website](https://docs.kubestellar.io/release-v0.2.2/Coding%20Milestones/PoC2023q1/example1/),
but with the some steps reorganized and combined and the special
workload and summarization aspiration removed.  You could continue
from here, doing the steps for the special workload.

## 3. Teardown the environment

To remove the example usage, delete the IMW and WMW and kind clusters run the following commands:

```shell
rm florin-syncer.yaml guilder-syncer.yaml
kubectl ws root
kubectl delete workspace example-imw
kubectl ws root:my-org
kubectl kubestellar remove wmw demo1
kubectl ws root
kubectl delete workspace my-org
kind delete cluster --name florin
kind delete cluster --name guilder
```

Stop and uninstall KubeStellar use the following command:

```shell
kubestellar stop
```

Stop and uninstall KubeStellar and kcp with the following command:

```shell
remove-kubestellar
```

## Demo Video

<a href="https://www.youtube.com/watch?v=NMGH-bwsh7s" target="_blank">
 <img src="https://img.youtube.com/vi/NMGH-bwsh7s/0.jpg" alt="KubeStellar Demo" width="700" height="500" border="10" />
</a>

