# KCP-Edge Example Scenarios:

<p align="center">
<img src="https://github.com/kcp-dev/edge-mc/blob/main/docs/content/en/docs/Coding%20Milestones/PoC2023q1/Edge-PoC-2023q1-Scenario-1-stage-4.svg" width="600" height="600">
</p>

In this example scenario we deploy two [kind](https://kind.sigs.k8s.io/) edge clusters. We call them “florin” and “guilder”. We also deploy two workloads (`special & common`). The common workload goes on both edge clusters and special workload goes on only into the `guilder` edge cluster. This example is described in more details [here](https://docs.kcp-edge.io/docs/coding-milestones/poc2023q1/example1/). 

Run the following command to deploy our example scenario:

```shell
./install_edge-mc.sh --stage 4
```

NB: you can also explore others stages (e.g., 1, 2 or 3) described [here](https://docs.kcp-edge.io/docs/coding-milestones/poc2023q1/example1/). 

It creates/deploys the following components:
-  kcp with kcp-playground and edge-syncer plugins
-  kcp-edge controllers: edge-scheduler, mailbox-controller and placement-translator
-  two kind clusters: “florin” and “guilder”, each running a edge-syncer
-  five workspaces: one edge server provider workspace, two inventory management workspaces, two workload management workspaces

NB: if you're using a macOS, you may see pop-us messages similar to the one below while deploying kcp-edge: 

```shell
  Do you want the application “kcp” to accept incoming network connections?
```

You can accept it or configure your firewall to suppress them by adding our kcp-edge components to the list of permitted apps.

After the completion of the `install_edge-mc.sh` script, you should see an ouput similar to the one below:


#### Two kind clusters:

```shell
kind get clusters
florin
guilder
```

#### kcp-edge infra deployed:

```shell
kubectl ws tree
.
└── root
    ├── compute
    ├── espw
    │   ├── 2r8mzyucyiogekve-mb-18bf4a12-e019-4520-954e-a2565fe991b5
    │   └── 2r8mzyucyiogekve-mb-f366f9ba-a111-4c80-b418-1a7b3ce61ab9
    ├── imw-1
    └── my-org
        ├── wmw-c
        └── wmw-s
```

The mailbox-controller created two mailbox workspace (`limgjykhmrjeiwc6-mb-1c6d6132-4ef9-482e-bff5-ee7a70fb601e` and `limgjykhmrjeiwc6-mb-a1d8f1cd-6493-4480-8c5e-c7a3dd53600a`) for the newly created SyncTargets: sync-target-f and sync-target-g

#### Two synctargets and locations objects are created, one for each cluster:

```shell
kubectl get locations,synctargets
NAME                                    RESOURCE      AVAILABLE   INSTANCES   LABELS   AGE
location.scheduling.kcp.io/location-f   synctargets   0           1                    3m2s
location.scheduling.kcp.io/location-g   synctargets   0           1                    3m2s

NAME                                       AGE
synctarget.workload.kcp.io/sync-target-f   3m1s
synctarget.workload.kcp.io/sync-target-g   3m1s
```

#### Two workload management workspaces are created:

1. For workload common:

```shell
kubectl ws root:my-org:wmw-c
Current workspace is "root:my-org:wmw-c".

kubectl get ns
NAME          STATUS   AGE
commonstuff   Active   4m57s
default       Active   5m2s

kubectl -n commonstuff get deploy
NAME      READY   UP-TO-DATE   AVAILABLE   AGE
commond   0/0     0            0           5m24s

kubectl -n commonstuff get configmaps
NAME               DATA   AGE
httpd-htdocs       1      5m42s
kube-root-ca.crt   1      5m42s
```

An `EdgePlacement` object is created for the workload common. Its “where predicate” (the locationSelectors array) has one label selector that matches the Location objects (`location-g` and `location-f` ) created earlier, thus directing the workload to both “florin” and “guilder” edge clusters.

``` shell
kubectl get edgeplacement
NAME               AGE
edge-placement-c   4m21s
```

In response to the created EdgePlacement, the edge scheduler created a corresponding SinglePlacementSlice object:

```shell
kubectl get SinglePlacementSlice
NAME               AGE
edge-placement-c   8m11s
```

2. For workload special:

```shell
kubectl ws root:my-org:wmw-s
Current workspace is "root:my-org:wmw-s".

kubectl get ns
NAME           STATUS   AGE
default        Active   6m28s
specialstuff   Active   6m25s

kubectl -n specialstuff  get deploy
NAME       READY   UP-TO-DATE   AVAILABLE   AGE
speciald   0/0     0            0           6m38s

kubectl -n specialstuff  get configmaps
NAME               DATA   AGE
httpd-htdocs       1      6m54s
kube-root-ca.crt   1      6m55s
```

An `EdgePlacement` object is created for the workload common. Its “where predicate” (the locationSelectors array) has one label selector that matches the Location object (`location-g`) created earlier, thus directing the workload to the “guilder” edge cluster.

```shell
kubectl get edgeplacement
NAME               AGE
edge-placement-s   7m13s
```

Again, in response to the created EdgePlacement, the edge scheduler created a corresponding SinglePlacementSlice object:

```shell
kubectl get SinglePlacementSlice
NAME               AGE
edge-placement-s   7m41s
```

#### Special and Common workloads are copied to their respective mailbox workspaces

In response to the created `EdgePlacement` and `SinglePlacementSlice` objects, the placement translator copied the workload prescriptions into the mailbox workspaces and created SyncerConfig objects there.

1. Mailbox workspace associated with Synctarget `sync-target-f`:

```shell
kubectl ws root:espw:2r8mzyucyiogekve-mb-18bf4a12-e019-4520-954e-a2565fe991b5
Current workspace is "root:espw:2r8mzyucyiogekve-mb-18bf4a12-e019-4520-954e-a2565fe991b5".

kubectl get ns
NAME          STATUS   AGE
commonstuff   Active   24m
default       Active   24m

kubectl get SyncerConfig
NAME      AGE
the-one   25m
```


2. Mailbox workspace associated with Synctarget `sync-target-g`: 

```shell
kubectl get ns
NAME           STATUS   AGE
commonstuff    Active   29m
specialstuff   Active   29m
default        Active   29m

kubectl get SyncerConfig
NAME      AGE
the-one   29m
```

#### The workloads are synced into edge clusters via the edge-syncer:

1. Florin edge cluster: receives the `workload common`

```shell
cd kcp
kubectl kcp playground use pcluster florin
 🔸  Current context in kubeconfig '.kcp-playground/playground.kubeconfig' is has been set to 'pcluster-florin-admin' 🏓

kubectl get pods -A
NAMESPACE                          NAME                                                READY   STATUS    RESTARTS   AGE
commonstuff                        commond-7b5d7ddd77-bj47g                            1/1     Running   0          30m
kcp-edge-syncer-the-one-20hyh1ri   kcp-edge-syncer-the-one-20hyh1ri-858d5d49b4-pblph   1/1     Running   0          31m
kube-system                        coredns-565d847f94-tg7j2                            1/1     Running   0          32m
kube-system                        coredns-565d847f94-wwsk7                            1/1     Running   0          32m
kube-system                        etcd-florin-control-plane                           1/1     Running   0          32m
kube-system                        kindnet-gl9wc                                       1/1     Running   0          32m
kube-system                        kube-apiserver-florin-control-plane                 1/1     Running   0          32m
kube-system                        kube-controller-manager-florin-control-plane        1/1     Running   0          32m
kube-system                        kube-proxy-x9nr9                                    1/1     Running   0          32m
kube-system                        kube-scheduler-florin-control-plane                 1/1     Running   0          32m
local-path-storage                 local-path-provisioner-684f458cdd-8whx8             1/1     Running   0          32m


kubectl get deploy -A
NAMESPACE                          NAME                               READY   UP-TO-DATE   AVAILABLE   AGE
commonstuff                        commond                            1/1     1            1           30m
kcp-edge-syncer-the-one-20hyh1ri   kcp-edge-syncer-the-one-20hyh1ri   1/1     1            1           31m
kube-system                        coredns                            2/2     2            2           33m
local-path-storage                 local-path-provisioner             1/1     1            1           33m
```


1. Guilder edge cluster: receives the `workload common` and the `workload special`

```shell
kubectl kcp playground use pcluster guilder 
 🔸  Current context in kubeconfig '.kcp-playground/playground.kubeconfig' is has been set to 'pcluster-guilder-admin' 🏓

kubectl get pods -A
NAMESPACE                          NAME                                                READY   STATUS    RESTARTS   AGE
commonstuff                        commond-7b5d7ddd77-ltrtc                            1/1     Running   0          34m
kcp-edge-syncer-the-one-20dlc7rw   kcp-edge-syncer-the-one-20dlc7rw-76df8f5459-mwmk6   1/1     Running   0          35m
kube-system                        coredns-565d847f94-cnngq                            1/1     Running   0          35m
kube-system                        coredns-565d847f94-tdrgx                            1/1     Running   0          35m
kube-system                        etcd-guilder-control-plane                          1/1     Running   0          36m
kube-system                        kindnet-crbrp                                       1/1     Running   0          35m
kube-system                        kube-apiserver-guilder-control-plane                1/1     Running   0          36m
kube-system                        kube-controller-manager-guilder-control-plane       1/1     Running   0          36m
kube-system                        kube-proxy-clr8m                                    1/1     Running   0          35m
kube-system                        kube-scheduler-guilder-control-plane                1/1     Running   0          36m
local-path-storage                 local-path-provisioner-684f458cdd-7md8s             1/1     Running   0          35m
specialstuff                       speciald-7d5ff89f68-w84lw                           1/1     Running   0          34m
 

kubectl get deploy -A
NAMESPACE                          NAME                               READY   UP-TO-DATE   AVAILABLE   AGE
commonstuff                        commond                            1/1     1            1           34m
kcp-edge-syncer-the-one-20dlc7rw   kcp-edge-syncer-the-one-20dlc7rw   1/1     1            1           35m
kube-system                        coredns                            2/2     2            2           36m
local-path-storage                 local-path-provisioner             1/1     1            1           36m
specialstuff                       speciald                           1/1     1            1           34m
```

#### Clean up kcp-edge environment

```shell
./clean_up.sh
```