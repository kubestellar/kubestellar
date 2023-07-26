<!--example1-stage-1a-start-->
### The mailbox controller

Running the mailbox controller will be conveniently automated.
Eventually.  In the meantime, you can use the KubeStellar command shown
here.

```shell
kubectl ws root:espw
go run ./cmd/mailbox-controller -v=2 &
sleep 60
```
``` { .bash .no-copy }
...
I0721 17:37:10.186848  189094 main.go:206] "Found APIExport view" exportName="e
dge.kcp.io" serverURL="https://10.0.2.15:6443/services/apiexport/cseslli1ddit3s
a5/edge.kcp.io"
...
I0721 19:17:21.906984  189094 controller.go:300] "Created APIBinding" worker=1
mbwsName="1d55jhazpo3d3va6-mb-551bebfd-b75e-47b1-b2e0-ff0a4cb7e006" mbwsCluster
="32x6b03ixc49cj48" bindingName="bind-edge" resourceVersion="1247"
...
I0721 19:18:56.203057  189094 controller.go:300] "Created APIBinding" worker=0
mbwsName="1d55jhazpo3d3va6-mb-732cf72a-1ca9-4def-a5e7-78fd0e36e61c" mbwsCluster
="q31lsrpgur3eg9qk" bindingName="bind-edge" resourceVersion="1329"
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
echo The guilder mailbox workspace name is $GUILDER_WS
```
``` { .bash .no-copy }
The guilder mailbox workspace name is 1t82bk54r6gjnzsp-mb-f0a82ab1-63f4-49ea-954d-3a41a35a9f1c
```

```shell
FLORIN_WS=$(kubectl get Workspace -o json | jq -r '.items | .[] | .metadata | select(.annotations ["edge.kcp.io/sync-target-name"] == "florin") | .name')
echo The florin mailbox workspace name is $FLORIN_WS
```
``` { .bash .no-copy }
The florin mailbox workspace name is 1t82bk54r6gjnzsp-mb-1a045336-8178-4026-8a56-5cd5609c0ec1
```
<!--example1-stage-1a-end-->
