# KCP-Edge Example Scenarios:

<img align="center" src="https://github.com/kcp-dev/edge-mc/blob/main/docs/content/en/docs/Coding%20Milestones/PoC2023q1/Edge-PoC-2023q1-Scenario-1-stage-4.svg" width="600" height="600">

In this example scenario we will deploy two [kind](https://kind.sigs.k8s.io/) edge clusters. We will call them “florin” and “guilder”. We will also deploy two workloads (`special & common`). The common workload goes on both edge clusters and special workload goes on only into the `guilder` edge cluster. This example is described in more details [here](https://docs.kcp-edge.io/docs/coding-milestones/poc2023q1/example1/). 

Run the following command to deploy 

```shell
./install_edge-mc.sh --stage 4
```

You can also explore others stages (e.g, 1, 2 or 3) describe [here](https://docs.kcp-edge.io/docs/coding-milestones/poc2023q1/example1/). 

It creates the following components:

-  The infrastructure and the edge service provider workspace and lets that react to the inventory
-  Two workloads, called “common” and “special” and in response to each EdgePlacement, the edge scheduler creates the corresponding SinglePlacementSlice object.
-  The placement translator reacts to the EdgePlacement objects in the workload management workspaces
- 


NB: if you're using a macOS, you may see pop-us messages similar to the one below while deploying kcp-edge: 

```shell
  Do you want the application “kcp” to accept incoming network connections?
```

You can accept it or configure your firewall to suppress them by adding our kcp-edge components to the list of permitted apps.


You should see an ouput similar to the one below:

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
    │   ├── limgjykhmrjeiwc6-mb-1c6d6132-4ef9-482e-bff5-ee7a70fb601e
    │   └── limgjykhmrjeiwc6-mb-a1d8f1cd-6493-4480-8c5e-c7a3dd53600a
    ├── imw-1
    └── my-org
        ├── wmw-c
        └── wmw-s
```

#### Two synctargets and locations objects are created, one for each cluster

```shell
kubectl ws root:imw-1
kubectl get locations
NAME         RESOURCE      AVAILABLE   INSTANCES   LABELS   AGE
location-f   synctargets   0           1                    2m21s
location-g   synctargets   0           1                    2m21s

kubectl get synctargets
NAME            AGE
sync-target-f   3m6s
sync-target-g   3m5s
```

#### Two workload management workspaces are created:

1. For workload common:

```bash
kubectl ws root:my-org:wmw-c
Current workspace is "root:my-org:wmw-c".

kubectl get ns
NAME          STATUS   AGE
commonstuff   Active   99s
default       Active   104s

kubectl -n commonstuff get deploy
NAME      READY   UP-TO-DATE   AVAILABLE   AGE
commond   0/0     0            0           111s

kubectl -n commonstuff get configmaps
NAME               DATA   AGE
httpd-htdocs       1      117s
kube-root-ca.crt   1      117s

kubectl get SinglePlacementSlice
NAME               AGE
edge-placement-c   111s
```

2. For workload special:

```bash
kubectl ws root:my-org:wmw-s
Current workspace is "root:my-org:wmw-s".

kubectl get ns
NAME           STATUS   AGE
default        Active   5m1s
specialstuff   Active   4m57s

kubectl -n specialstuff  get deploy
NAME       READY   UP-TO-DATE   AVAILABLE   AGE
speciald   0/0     0            0           5m29s

kubectl -n specialstuff  get configmaps
NAME               DATA   AGE
httpd-htdocs       1      5m35s
kube-root-ca.crt   1      5m35s

kubectl get SinglePlacementSlice
NAME               AGE
edge-placement-s   5m26s
```

#### Talk about syncer and workloads deployed at pclusters

#### Clean up kcp-edge environment

```bash
./clean_up.sh
```