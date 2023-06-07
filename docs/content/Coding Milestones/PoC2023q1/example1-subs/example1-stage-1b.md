<!--example1-stage-1b-start-->
### Populate the edge service provider workspace

This puts the definition and export of the edge-mc API in the edge
service provider workspace.

Use the following command.

```shell
kubectl create -f config/exports
```

### The mailbox controller

Running the mailbox controller will be conveniently automated.
Eventually.  In the meantime, you can use the edge-mc command shown
here.

```shell
go run ./cmd/mailbox-controller -v=2 &
sleep 45
```
``` { .bash .no-copy }
...
I0423 01:09:37.991080   10624 main.go:196] "Found APIExport view" exportName="workload.kcp.io" serverURL="https://192.168.58.123:6443/services/apiexport/root/workload.kcp.io"
...
I0423 01:09:38.449395   10624 controller.go:299] "Created APIBinding" worker=1 mbwsName="apmziqj9p9fqlflm-mb-bf452e1f-45a0-4d5d-b35c-ef1ece2879ba" mbwsCluster="yk9a66vjms1pi8hu" bindingName="bind-edge" resourceVersion="914"
...
I0423 01:09:38.842881   10624 controller.go:299] "Created APIBinding" worker=3 mbwsName="apmziqj9p9fqlflm-mb-b8c64c64-070c-435b-b3bd-9c0f0c040a54" mbwsCluster="12299slctppnhjnn" bindingName="bind-edge" resourceVersion="968"
^C
```

You need a `-v` setting of 2 or numerically higher to get log messages
about individual mailbox workspaces.

This controller creates a mailbox workspace for each SyncTarget and
puts an APIBinding to the edge API in each of those mailbox
workspaces.  For this simple scenario, you do not need to keep this
controller running after it does those things (hence the `^C` above);
normally it would run continuously.

You can get a listing of those mailbox workspaces as follows.

```shell
kubectl get Workspaces
```
``` { .bash .no-copy }
NAME                                                       TYPE        REGION   PHASE   URL                                                     AGE
1t82bk54r6gjnzsp-mb-1a045336-8178-4026-8a56-5cd5609c0ec1   universal            Ready   https://192.168.58.123:6443/clusters/1najcltzt2nqax47   50s
1t82bk54r6gjnzsp-mb-f0a82ab1-63f4-49ea-954d-3a41a35a9f1c   universal            Ready   https://192.168.58.123:6443/clusters/1y7wll1dz806h3sb   50s
```

More usefully, using custom columns you can get a listing that shows
the _name_ of the associated SyncTarget.

```shell
kubectl get Workspace -o "custom-columns=NAME:.metadata.name,SYNCTARGET:.metadata.annotations['edge\.kcp\.io/sync-target-name'],CLUSTER:.spec.cluster"
```
``` { .bash .no-copy }
NAME                                                       SYNCTARGET   CLUSTER
1t82bk54r6gjnzsp-mb-1a045336-8178-4026-8a56-5cd5609c0ec1   florin       1najcltzt2nqax47
1t82bk54r6gjnzsp-mb-f0a82ab1-63f4-49ea-954d-3a41a35a9f1c   guilder      1y7wll1dz806h3sb
```

Also: if you ever need to look up just one mailbox workspace by
SyncTarget name, you could do it as follows.

```shell
GUILDER_WS=$(kubectl get Workspace -o json | jq -r '.items | .[] | .metadata | select(.annotations ["edge.kcp.io/sync-target-name"] == "guilder") | .name')
```
``` { .bash .no-copy }
1t82bk54r6gjnzsp-mb-f0a82ab1-63f4-49ea-954d-3a41a35a9f1c
```

```shell
FLORIN_WS=$(kubectl get Workspace -o json | jq -r '.items | .[] | .metadata | select(.annotations ["edge.kcp.io/sync-target-name"] == "florin") | .name')
```
``` { .bash .no-copy }
1t82bk54r6gjnzsp-mb-1a045336-8178-4026-8a56-5cd5609c0ec1
```

### Connect guilder edge cluster with its mailbox workspace

The following command will (a) create, in the mailbox workspace for
guilder, an identity and authorizations for the edge syncer and (b)
write a file containing YAML for deploying the syncer in the guilder
cluster.

```shell
kubectl kubestellar prep-for-syncer --imw root:imw-1 guilder
```
``` { .bash .no-copy }
Current workspace is "root:imw-1".
Current workspace is "root:espw".
Current workspace is "root:espw:1t82bk54r6gjnzsp-mb-f0a82ab1-63f4-49ea-954d-3a41a35a9f1c" (type root:universal).
Creating service account "kcp-edge-syncer-guilder-wfeig2lv"
Creating cluster role "kcp-edge-syncer-guilder-wfeig2lv" to give service account "kcp-edge-syncer-guilder-wfeig2lv"

 1. write and sync access to the synctarget "kcp-edge-syncer-guilder-wfeig2lv"
 2. write access to apiresourceimports.

Creating or updating cluster role binding "kcp-edge-syncer-guilder-wfeig2lv" to bind service account "kcp-edge-syncer-guilder-wfeig2lv" to cluster role "kcp-edge-syncer-guilder-wfeig2lv".

Wrote physical cluster manifest to guilder-syncer.yaml for namespace "kcp-edge-syncer-guilder-wfeig2lv". Use

  KUBECONFIG=<pcluster-config> kubectl apply -f "guilder-syncer.yaml"

to apply it. Use

  KUBECONFIG=<pcluster-config> kubectl get deployment -n "kcp-edge-syncer-guilder-wfeig2lv" kcp-edge-syncer-guilder-wfeig2lv

to verify the syncer pod is running.
Current workspace is "root:espw".
```

The file written was, as mentioned in the output,
`guilder-syncer.yaml`.  Next `kubectl apply` that to the guilder
cluster.  That will look something like the following; adjust as
necessary to make kubectl manipulate **your** guilder cluster.

```shell
KUBECONFIG=~/.kube/config kubectl --context kind-guilder apply -f guilder-syncer.yaml
```
``` { .bash .no-copy }
namespace/kcp-edge-syncer-guilder-wfeig2lv created
serviceaccount/kcp-edge-syncer-guilder-wfeig2lv created
secret/kcp-edge-syncer-guilder-wfeig2lv-token created
clusterrole.rbac.authorization.k8s.io/kcp-edge-syncer-guilder-wfeig2lv created
clusterrolebinding.rbac.authorization.k8s.io/kcp-edge-syncer-guilder-wfeig2lv created
secret/kcp-edge-syncer-guilder-wfeig2lv created
deployment.apps/kcp-edge-syncer-guilder-wfeig2lv created
```

You might check that the syncer is running, as follows.

```shell
KUBECONFIG=~/.kube/config kubectl --context kind-guilder get deploy -A
```
``` { .bash .no-copy }
NAMESPACE                          NAME                               READY   UP-TO-DATE   AVAILABLE   AGE
kcp-edge-syncer-guilder-saaywsu5   kcp-edge-syncer-guilder-saaywsu5   1/1     1            1           52s
kube-system                        coredns                            2/2     2            2           35m
local-path-storage                 local-path-provisioner             1/1     1            1           35m
```

### Connect florin edge cluster with its mailbox workspace

Do the analogous stuff for the florin cluster.

```shell
kubectl kubestellar prep-for-syncer --imw root:imw-1 florin
```
``` { .bash .no-copy }
Current workspace is "root:imw-1".
Current workspace is "root:espw".
Current workspace is "root:espw:1t82bk54r6gjnzsp-mb-1a045336-8178-4026-8a56-5cd5609c0ec1" (type root:universal).
Creating service account "kcp-edge-syncer-florin-32uaph9l"
Creating cluster role "kcp-edge-syncer-florin-32uaph9l" to give service account "kcp-edge-syncer-florin-32uaph9l"

 1. write and sync access to the synctarget "kcp-edge-syncer-florin-32uaph9l"
 2. write access to apiresourceimports.

Creating or updating cluster role binding "kcp-edge-syncer-florin-32uaph9l" to bind service account "kcp-edge-syncer-florin-32uaph9l" to cluster role "kcp-edge-syncer-florin-32uaph9l".

Wrote physical cluster manifest to florin-syncer.yaml for namespace "kcp-edge-syncer-florin-32uaph9l". Use

  KUBECONFIG=<pcluster-config> kubectl apply -f "florin-syncer.yaml"

to apply it. Use

  KUBECONFIG=<pcluster-config> kubectl get deployment -n "kcp-edge-syncer-florin-32uaph9l" kcp-edge-syncer-florin-32uaph9l

to verify the syncer pod is running.
Current workspace is "root:espw".
```

And deploy the syncer in the florin cluster.

```shell
KUBECONFIG=~/.kube/config kubectl --context kind-florin apply -f florin-syncer.yaml 
```
``` { .bash .no-copy }
namespace/kcp-edge-syncer-florin-32uaph9l created
serviceaccount/kcp-edge-syncer-florin-32uaph9l created
secret/kcp-edge-syncer-florin-32uaph9l-token created
clusterrole.rbac.authorization.k8s.io/kcp-edge-syncer-florin-32uaph9l created
clusterrolebinding.rbac.authorization.k8s.io/kcp-edge-syncer-florin-32uaph9l created
secret/kcp-edge-syncer-florin-32uaph9l created
deployment.apps/kcp-edge-syncer-florin-32uaph9l created
```
<!--example1-stage-1b-end-->