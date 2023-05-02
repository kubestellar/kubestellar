# KCP-Edge Contributor Development Environment

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


Run the following script to install the required package (Linux or MacOS ):

```bash
./install_req.sh
```

Alternatively, run the following command for MacOS platform:

```bash
brew install ko gcc jq make go@1.19 kind kubectl
```

For Windows WSL/Ubuntu platform, follow the instructions [here](docs/wls_ubuntu_platform.md)


## Quickstart

For a single-quick automation to deploy

#### 1. Clone the kcp-edge repo:

```shell
git clone https://github.com/kcp-dev/edge-mc.git  kcp-edge
```

#### 2. Build the kcp-edge binaries:

```shell
cd kcp-edge
make build
```

#### 3. Create your own edge infrastructure (edge clusters) - Bring Your Own Cluster (BYOC)

Create your edge cluster or bring your own k8s edge cluster. In this example, we will use [kind](https://kind.sigs.k8s.io/) to create an edge cluster that we name “florin”:

```shell
    kind create cluster --name florin
```  

#### 4. Deploy the kcp-edge platform:

  * Step-1: download kcp binaries for your platform:

    ```shell
       VERSION=0.11.0 # choose the latest version (without v prefix)
       OS=darwin   # or darwin
       ARCH=amd64  # or amd64
       curl -sSfL "https://github.com/kcp-dev/kcp/releases/download/v${VERSION}/kcp_${VERSION}_${OS}_${ARCH}.tar.gz" > kcp.tar.gz
       curl -sSfL "https://github.com/kcp-dev/kcp/releases/download/v${VERSION}/kubectl-kcp-plugin_${VERSION}_${OS}_${ARCH}.tar.gz" > kubectl-kcp-plugin.tar.gz
    ```

    Extract kcp and kubectl-kcp-plugin and place all the files in the bin directories somewhere in your $PATH. For example:

    ```shell
    tar -xvf kcp.tar.gz
    tar -xvf kubectl-kcp-plugin.tar.gz
    export PATH="$PATH:$(pwd)/kcp/bin"
    ```

  * Step-2: start kcp

    ```shell
    kcp start >& kcp_log.txt &
    ```

  * Step-3: Deploy kcp-edge infra
     
    Run the following command inside the `kcp-edge/environments/dev-env` directory:
    ```bash
    ./kcp-edge.sh start
    ```

    This will create/deploy the following components:

    - 1 kcp workspace: edge service provider workspace (`espw`)

    ```bash
    kubectl ws tree
    .
    └── root
        ├── compute
        ├── espw
    ```

    - 3 kcp-edge controllers: [edge-scheduler](https://docs.kcp-edge.io/docs/coding-milestones/poc2023q1/edge-scheduler/), [mailbox-controller](https://docs.kcp-edge.io/docs/coding-milestones/poc2023q1/mailbox-controller/) and [placement-translator](https://docs.kcp-edge.io/docs/coding-milestones/poc2023q1/placement-translator/)

    ```bash
    ps aux | grep -e mailbox-controller -e placement-translator -e cmd/scheduler/main.go
    user      2898   0.0  0.1 34922264  45188 s004  S     3:22PM   0:01.84 go run cmd/scheduler/main.go -v 2 --root-user shard-main-kcp-admin --root-cluster shard-main-root --sysadm-context shard-main-system:admin --sysadm-user shard-main-shard-admin
    user      2872   0.0  0.2 34925136  56132 s004  S     3:22PM   0:02.44 go run ./cmd/mailbox-controller --inventory-context=shard-main-root -v=2
    user      2929   0.0  0.2 34922964  69724 s004  S     3:22PM   0:03.74 go run ./cmd/placement-translator --allclusters-context shard-main-system:admin

    ```

#### 5. Connect your edge pcluster to the kcp-edge platform:

  * Step-1: Populate the `imw`: enter the target workspace:

    ```bash
        kubectl ws root:imw-1
    ```
    
  * Step-2: create a SyncTarget object to represent your edge pcluster. For example:

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

  * Step-3: create a Location object describing your edge pcluster. For example:

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
  
    You can remove the default location object created:

    ```bash
    kubectl delete location default
    location.scheduling.kcp.io "default" deleted
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

  * Step-4: create the edge syncer manifest

    ```bash
       ./build-edge-syncer.sh  --syncTarget sync-target-f

       -----------------------------------------------------------
        Edge-syncer manifest created:  sync-target-f-syncer.yaml
        Current workspace: root:espw:1q1p9rsh18rhjuy4-mb-2c1b6ce7-bc4a-4071-887d-871ba293f303
       -----------------------------------------------------------
    ```
    An edge syncer manifest yaml file is created in your current director: `sync-target-f-syncer.yaml`

  * Step-5: deploy the edge syncer to your edge pcluter

    For example: switch to the context of the florin kind cluster

    ```bash
       kubectl config use-context kind-florin
       Switched to context "kind-florin".
    ```

    Apply the edge syncer manifest:

    ```bash
      kubectl apply -f sync-target-g-syncer.yaml

      namespace/kcp-edge-syncer-the-one-2p3eqojn created
      serviceaccount/kcp-edge-syncer-the-one-2p3eqojn created
      secret/kcp-edge-syncer-the-one-2p3eqojn-token created
      clusterrole.rbac.authorization.k8s.io/kcp-edge-syncer-the-one-2p3eqojn created
      clusterrolebinding.rbac.authorization.k8s.io/kcp-edge-syncer-the-one-2p3eqojn created
      role.rbac.authorization.k8s.io/kcp-edge-dns-the-one-2p3eqojn created
      rolebinding.rbac.authorization.k8s.io/kcp-edge-dns-the-one-2p3eqojn created
      secret/kcp-edge-syncer-the-one-2p3eqojn created
      deployment.apps/kcp-edge-syncer-the-one-2p3eqojn created
    ```

    Check that the edge syncer pod is running:

    ```bash
        kubectl get pods -A
        NAMESPACE                          NAME                                                READY   STATUS    RESTARTS   AGE
        kcp-edge-syncer-the-one-2p3eqojn   kcp-edge-syncer-the-one-2p3eqojn-6884cd645b-tn6s7   1/1     Running   0          2m16s
        kube-system                        coredns-565d847f94-fw7vt                            1/1     Running   0          4m40s
        kube-system                        coredns-565d847f94-kc4gk                            1/1     Running   0          4m40s
        kube-system                        etcd-florin-control-plane                           1/1     Running   0          4m56s
        kube-system                        kindnet-9vzg8                                       1/1     Running   0          4m40s
        kube-system                        kube-apiserver-florin-control-plane                 1/1     Running   0          4m55s
        kube-system                        kube-controller-manager-florin-control-plane        1/1     Running   0          4m55s
        kube-system                        kube-proxy-qhprt                                    1/1     Running   0          4m40s
        kube-system                        kube-scheduler-florin-control-plane                 1/1     Running   0          4m55s
        local-path-storage                 local-path-provisioner-684f458cdd-bc5c4             1/1     Running   0          4m40s
    ``` 


#### 5. Bring Your Own Workload (BYOW) 

  * Step-1: Populate the `wmw` with your workload objects -  enter the target workspace: `wmw-1`
  
      ```bash
        kubectl ws root:my-org:wmw-1
      ```

    N.B: if your using the `florin` kind edge pcluster created in this example, then you should switch back to the kcp edge context first before executing the command above:

      ```bash
        kubectl config use-context shard-main-root
      ```

  * Step-2: deploy your workload. For example:

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

    Check that your workload was deployed successfully: 

    ```bash
        kubectl -n commonstuff get deploy
        NAME               READY   UP-TO-DATE   AVAILABLE   AGE
        nginx-deployment   0/3     0            0           13s
    ```


#### 2. Create the `EdgePlacement` object for your workload. 

In the `wmw-1` workspace create the following `EdgePlacement` object: 
 
  ```bash
  kubectl ws root:my-org:wmw-1

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
Its “where predicate” (the locationSelectors array) has one label selector that matches the Location object created earlier, thus directing the workload to your edge pcluster.
 
In response to the created `EdgePlacement`, the edge [scheduler](https://docs.kcp-edge.io/docs/coding-milestones/poc2023q1/edge-scheduler/) will create a corresponding `SinglePlacementSlice` object:

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

#### 3. Check that the workloads objects are copied to mailbox workspace:

In response to the created EdgePlacement and SinglePlacementSlice objects, the [placement translator](https://docs.kcp-edge.io/docs/coding-milestones/poc2023q1/placement-translator/) will copy the workload prescriptions into the mailbox workspaces and create `SyncerConfig` objects there.

```bash
  kubectl ws root:espw:1q1p9rsh18rhjuy4-mb-2c1b6ce7-bc4a-4071-887d-871ba293f303
  Current workspace is "root:espw:1q1p9rsh18rhjuy4-mb-2c1b6ce7-bc4a-4071-887d-871ba293f303".

  kubectl get ns
  NAME          STATUS   AGE
  commonstuff   Active   2m16s
  default       Active   73m

  kubectl -n commonstuff get deploy
  NAME               READY   UP-TO-DATE   AVAILABLE   AGE
  nginx-deployment   0/3     0            0           2m25s

  kubectl get syncerConfig
  NAME      AGE
  the-one   21m
```

```bash
    kubectl get syncerConfig the-one -o yaml
    apiVersion: edge.kcp.io/v1alpha1
    kind: SyncerConfig
    metadata:
      annotations:
        kcp.io/cluster: 32uj0ldbsw50x5np
      creationTimestamp: "2023-04-19T16:06:03Z"
      generation: 4
      name: the-one
      resourceVersion: "1150"
      uid: 2f75a6b6-96ad-4cc6-8a53-2cfbcbf6ce84
    spec:
      clusterScope:
      - apiVersion: v1alpha1
        group: apis.kcp.io
        objects:
        - bind-kube
        resource: apibindings
      namespaceScope:
        namespaces:
        - commonstuff
        resources:
        - apiVersion: v1
          group: rbac.authorization.k8s.io
          resource: roles
        - apiVersion: v1
          group: coordination.k8s.io
          resource: leases
        - apiVersion: v1
          group: ""
          resource: limitranges
        - apiVersion: v1
          group: networking.k8s.io
          resource: ingresses
        - apiVersion: v1
          group: ""
          resource: endpoints
        - apiVersion: v1
          group: ""
          resource: services
        - apiVersion: v1
          group: apps
          resource: deployments
        - apiVersion: v1
          group: ""
          resource: serviceaccounts
        - apiVersion: v1
          group: ""
          resource: configmaps
        - apiVersion: v1
          group: ""
          resource: secrets
        - apiVersion: v1
          group: ""
          resource: resourcequotas
        - apiVersion: v1
          group: ""
          resource: pods
        - apiVersion: v1
          group: rbac.authorization.k8s.io
          resource: rolebindings
      upsync:
      - apiGroup: group1.test
        names:
        - george
        - cosmo
        namespaces:
        - orbital
        resources:
        - sprockets
        - flanges
      - apiGroup: group2.test
        names:
        - william
        resources:
        - cogs
    status: {}
```

#### 4. Check that the workloads are running in the edge pclusters:

   * Step-1: switch to your edge pcluster: `wmw-1`
    
     ```bash
        kubectl config use-context kind-florin
        Switched to context "kind-florin".
     ```

   * Step-2: check your workload:

      ```bash
          kubectl get ns
          NAME                               STATUS   AGE
          commonstuff                        Active   4m58s
          default                            Active   49m
          kcp-edge-syncer-the-one-2p3eqojn   Active   46m
          kube-node-lease                    Active   49m
          kube-public                        Active   49m
          kube-system                        Active   49m
          local-path-storage                 Active   48m

          kubectl -n commonstuff get deploy
          NAME               READY   UP-TO-DATE   AVAILABLE   AGE
          nginx-deployment   3/3     3            3           4m55s


          kubectl -n commonstuff get pods
          NAME                                READY   STATUS    RESTARTS   AGE
          nginx-deployment-7fb96c846b-6v6fq   1/1     Running   0          5m1s
          nginx-deployment-7fb96c846b-f52qc   1/1     Running   0          5m1s
          nginx-deployment-7fb96c846b-r85t6   1/1     Running   0          5m1s
      ```
#### 5. Delete your kcp-edge environment:

```bash
./delete_edge-mc.sh
```
