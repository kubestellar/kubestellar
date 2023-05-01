#### 3. Experiment with the kcp-edge 2023q1 PoC example scenarios:

In this quickstart example we will deploy `stage 3` described in more details [here](https://docs.kcp-edge.io/docs/coding-milestones/poc2023q1/example1/). It creates the following components:

-  The infrastructure and the edge service provider workspace and lets that react to the inventory
-  Two workloads, called “common” and “special” and in response to each EdgePlacement, the edge scheduler creates the corresponding SinglePlacementSlice object.
-  The placement translator reacts to the EdgePlacement objects in the workload management workspaces

```bash
./install_edge-mc.sh --stage 3
```

NB: if you're using a macOS, you may see pop-us messages similar to the one below while deploying kcp-edge: 

```bash
  Do you want the application “kcp” to accept incoming network connections?
```

You can accept it or configure your firewall to suppress them by adding our kcp-edge components to the list of permitted apps.


You should see an ouput similar to the one below:

```bash
kind get clusters
florin
guilder
```

```bash
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


```bash
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

For workload common:

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

For workload special:

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

For placement translator:
```bash

kubectl ws root:my-org:wmw-c
Current workspace is "root:my-org:wmw-c".

kubectl get EdgePlacement
NAME               AGE
edge-placement-c   91s

kubectl delete EdgePlacement edge-placement-c
edgeplacement.edge.kcp.io "edge-placement-c" deleted
```
Placement translator logs:

```bash
:WorkspaceScheduled Status:True Severity: LastTransitionTime:2023-03-30 17:46:42 -0400 EDT Reason: Message:}] Initializers:[]}}
I0330 17:47:01.732064   64918 main.go:119] "Receive" key="2vh6tnanyw60negt:edge-placement-c" val=map[{APIGroup: Resource:namespaces Name:commonstuff}:{APIVersion:v1 IncludeNamespaceObject:false}]
I0330 17:47:01.732364   64918 main.go:119] "Receive" key="211ieqpc4xyydw2w:edge-placement-s" val=map[{APIGroup: Resource:namespaces Name:specialstuff}:{APIVersion:v1 IncludeNamespaceObject:false}]
I0330 17:48:08.042551   64918 main.go:119] "Receive" key="2vh6tnanyw60negt:edge-placement-c" val=map[]
```

#### 4. Delete a kcp-edge Poc2023q1 example stage:

```bash
./delete_edge-mc.sh
```