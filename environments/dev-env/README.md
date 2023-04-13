## Required Packages:
   - ko: https://ko.build/install/ 
   - gcc
   - docker 
   - jq
   - make
   - go (version expected 1.19)
   - kind
   - kubectl  


## Supported OS Platforms 
  - Linux
  - MacOS
  - Windows WSL/Ubuntu

For MacOS only:

```bash
brew install ko gcc jq make go@1.19 kind kubectl
```

Run the following script to install the required package (Linux or MacOS ):

```bash
./install_req.sh
```

For Windows WSL/Ubuntu platform, follow the instructions [here](docs/README.md)


## Quickstart

#### 1. Clone this repo:

```bash
git clone -b dev-env-v3 https://github.com/dumb0002/edge-mc.git
```

#### 2. Change into the following directory path:

```bash
cd edge-mc/environments/dev-env
```

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

## Bring your own workload (BYOW)

#### 1. Create your own edge infrastructure (pclusters):

For example: create a kind cluster

```bash
kind create cluster --name florin
``` 

#### 2. Deploy the kcp-edge platform:

  * Step-1: Clone this repo:

    ```bash
      git clone -b dev-env-v3 https://github.com/dumb0002/edge-mc.git
    ```

  * Step-2: change into the following directory path:

    ```bash
      cd edge-mc/environments/dev-env
    ```

  * Step-3: Deploy kcp-edge

    ```bash
    ./install_edge-mc.sh --stage 0
    ```

    This will start `kcp` and create/deploy the following components:

    - 4 kcp workspaces: edge service provider workspace (`espw`), inventory management workspace (`imw`) and workload management workspace (`wmw`) under my-org workspace

    ```bash
    .
    └── root
        ├── compute
        ├── espw
        ├── imw-1
        └── my-org
            └── wmw-1
    ```
    - 3 kcp-edge controllers: [edge-scheduler](https://docs.kcp-edge.io/docs/coding-milestones/poc2023q1/edge-scheduler/), [mailbox-controller](https://docs.kcp-edge.io/docs/coding-milestones/poc2023q1/mailbox-controller/) and [placement-translator](https://docs.kcp-edge.io/docs/coding-milestones/poc2023q1/placement-translator/)

    ```bash
    ps aux | grep -e mailbox-controller -e placement-translator -e cmd/scheduler/main.go
    user      2898   0.0  0.1 34922264  45188 s004  S     3:22PM   0:01.84 go run cmd/scheduler/main.go -v 2 --root-user shard-main-kcp-admin --root-cluster shard-main-root --sysadm-context shard-main-system:admin --sysadm-user shard-main-shard-admin
    user      2872   0.0  0.2 34925136  56132 s004  S     3:22PM   0:02.44 go run ./cmd/mailbox-controller --inventory-context=shard-main-root -v=2
    user      2929   0.0  0.2 34922964  69724 s004  S     3:22PM   0:03.74 go run ./cmd/placement-translator --allclusters-context shard-main-system:admin
    ```

#### 3. Deploy your own workload: 

 * Populate the `imw`:
    * Step-1: enter the target workspace:

      ```bash
          kubectl ws root:imw-1
      ```
    
    * Step-2: create a SyncTarget object to represent your pcluster. For example:

      ```bash
      cat <<EOF | kubectl apply -f -
      apiVersion: workload.kcp.io/v1alpha1
      kind: SyncTarget
      metadata:
        name: sync-target-f
        labels:
          example: si
          extended: non
      spec:
        cells:
          foo: bar
      EOF
      ```

   * Step-3: create a Location object describing your pcluster. For example:

      ```bash
      cat <<EOF | kubectl apply -f -
      apiVersion: scheduling.kcp.io/v1alpha1
      kind: Location
      metadata:
        name: location-f
        labels:
          env: prod
      spec:
        resource: {group: workload.kcp.io, version: v1alpha1, resource: synctargets}
        instanceSelector:
          matchLabels: {"example":"si", "extended":"non"}
      EOF
     ```

      A location and synctarget objects will be created:

      ```bash
      kubectl get locations,synctargets
      NAME                                    RESOURCE      AVAILABLE   INSTANCES   LABELS   AGE
      location.scheduling.kcp.io/default      synctargets   0           1                    36s
      location.scheduling.kcp.io/location-f   synctargets   0           1                    25s

      NAME                                       AGE
      synctarget.workload.kcp.io/sync-target-f   36s
      ```

      The [mailbox-controller](https://docs.kcp-edge.io/docs/coding-milestones/poc2023q1/mailbox-controller/) creates a mailbox workspace for the newly created SyncTarget: `sync-target-f`:

      ```bash
      kubectl ws root
      Current workspace is "root".

      kubectl ws tree
      .
      └── root
          ├── compute
          ├── espw
          │   └── 1q1p9rsh18rhjuy4-mb-2c1b6ce7-bc4a-4071-887d-871ba293f303
          ├── imw-1
          └── my-org
              └── wmw-1
      ```

* Populate the `wmw`: 

  * Step-1: Enter the target workspace: `wmw-1`
 
    ```bash
      kubectl ws root:my-org:wmw-1
    ```

  * Step-2: create your workload objects. For example:

    ```bash
    cat <<EOF | kubectl apply -f -
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

  * Step-3: create the `EdgePlacement` object for your workload. Its “where predicate” (the locationSelectors array) has one label selector that matches the Location object created earlier, thus directing the workload to your pcluster.
 
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: edge.kcp.io/v1alpha1
    kind: EdgePlacement
    metadata:
      name: edge-placement-c
    spec:
      locationSelectors:
      - matchLabels: {"env":"prod"}
      namespaceSelector:
        matchLabels: {"common":"si"}
    EOF
    ```
 
    In response to the created `EdgePlacement`, the edge scheduler will create a corresponding `SinglePlacementSlice` object:

    ```bash
        kubectl get SinglePlacementSlice -o yaml edge-placement-c

        apiVersion: edge.kcp.io/v1alpha1
        destinations:
        - cluster: 2vsg1pyc1uqtyk40
          locationName: location-f
          syncTargetName: sync-target-f
          syncTargetUID: b7acd821-319b-4061-be47-084fbce36f29
        kind: SinglePlacementSlice
        metadata:
          annotations:
            kcp.io/cluster: 19nyul7j8az21nkp
          creationTimestamp: "2023-04-13T03:58:20Z"
          generation: 1
          name: edge-placement-c
          ownerReferences:
          - apiVersion: edge.kcp.io/v1alpha1
            kind: EdgePlacement
            name: edge-placement-c
            uid: c3d718b8-9a2f-4a34-9b80-b98d091cff67
          resourceVersion: "1088"
          uid: b7a8e302-1532-4e48-a69e-910a9ac2aa62
    ```

   * Step-4: delete your kcp-edge environment

   ```bash
   ./delete_edge-mc.sh
   ```
