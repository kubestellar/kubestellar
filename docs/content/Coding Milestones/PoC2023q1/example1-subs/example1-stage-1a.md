<!--example1-stage-1a-start-->
### The mailbox controller

Running the mailbox controller will be conveniently automated.
Eventually.  In the meantime, you can use the KubeStellar command shown
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
<!--example1-stage-1a-end-->
