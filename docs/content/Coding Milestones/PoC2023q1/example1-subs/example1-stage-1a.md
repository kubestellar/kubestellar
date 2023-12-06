<!--example1-stage-1a-start-->
### The mailbox controller

The mailbox controller is one of the central controllers of
KubeStellar.  If you have deployed the KubeStellar core as Kubernetes
workload then this controller is already running in a pod in your
hosting cluster. If instead you are running these controllers as bare
processes then launch this controller as follows.

```shell
ESPW_SPACE_CONFIG="${PWD}/temp-space-config/spaceprovider-default-espw"
kubectl-kubestellar-get-config-for-space --space-name espw --sm-core-config $SM_CONFIG --sm-context $SM_CONTEXT --output $ESPW_SPACE_CONFIG
(
  KUBECONFIG=$SM_CONFIG mailbox-controller  &
  sleep 20
)
```

This controller is in charge of maintaining the collection of mailbox
spaces, which are an implementation detail not intended for user
consumption. You can use the following command to wait for the
appearance of the mailbox spaces implied by the florin and guilder
`SyncTarget` objects that you made earlier.

```shell
while [ KUBECONFIG=$SM_CONFIG $(kubectl get spaces -A | grep "\-mb\-" | wc -l) -ne 2 ]; do
  sleep 10
done
```

If it is working correctly, lines like the following will appear in
the controller's log (which is being written into your shell if you ran the controller as a bare process above, otherwise you can fetch [as directed](../../commands/#fetch-a-log-from-a-kubestellar-runtime-container)).

``` { .bash .no-copy }
...
I0721 17:37:10.186848  189094 main.go:206] "Found APIExport view" exportName="e
dge.kubestellar.io" serverURL="https://10.0.2.15:6443/services/apiexport/cseslli1ddit3s
a5/edge.kubestellar.io"
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
about individual mailbox spaces.

A mailbox space name is distinguished by `-mb-` separator.
You can get a listing of those mailbox spaces as follows.

```shell
KUBECONFIG=$SM_CONFIG kubectl get spaces -A
```
``` { .bash .no-copy }
NAME                                                       TYPE          REGION   PHASE   URL                                                     AGE
1t82bk54r6gjnzsp-mb-1a045336-8178-4026-8a56-5cd5609c0ec1   universal              Ready   https://192.168.58.123:6443/clusters/1najcltzt2nqax47   50s
1t82bk54r6gjnzsp-mb-f0a82ab1-63f4-49ea-954d-3a41a35a9f1c   universal              Ready   https://192.168.58.123:6443/clusters/1y7wll1dz806h3sb   50s
compute                                                    universal              Ready   https://172.20.144.39:6443/clusters/root:compute        6m8s
espw                                                       organization           Ready   https://172.20.144.39:6443/clusters/root:espw           2m4s
imw1                                                       organization           Ready   https://172.20.144.39:6443/clusters/root:imw1           1m9s
```

More usefully, using custom columns you can get a listing that shows
the _name_ of the associated SyncTarget.

```shell
KUBECONFIG=$SM_CONFIG kubectl get spaces -n spaceprovider-default -o "custom-columns=NAME:.metadata.name,SYNCTARGET:.metadata.annotations['edge\.kubestellar\.io/sync-target-name'],CLUSTER:.spec.cluster"
```
``` { .bash .no-copy }
NAME                                                       SYNCTARGET   CLUSTER
1t82bk54r6gjnzsp-mb-1a045336-8178-4026-8a56-5cd5609c0ec1   florin       1najcltzt2nqax47
1t82bk54r6gjnzsp-mb-f0a82ab1-63f4-49ea-954d-3a41a35a9f1c   guilder      1y7wll1dz806h3sb
compute                                                    <none>       mqnl7r5f56hswewy
espw                                                       <none>       2n88ugkhysjbxqp5
imw1                                                       <none>       4d2r9stcyy2qq5c1
```

Also: if you ever need to look up just one mailbox space by
SyncTarget name, you could do it as follows.

```shell
GUILDER_SPACE=$(KUBECONFIG=$SM_CONFIG kubectl get space -n spaceprovider-default -o json | jq -r '.items | .[] | .metadata | select(.annotations ["edge.kubestellar.io/sync-target-name"] == "guilder") | .name')
echo The guilder mailbox space name is $GUILDER_SPACE
```
``` { .bash .no-copy }
The guilder mailbox space name is 1t82bk54r6gjnzsp-mb-f0a82ab1-63f4-49ea-954d-3a41a35a9f1c
```

```shell
FLORIN_SPACE=$(KUBECONFIG=$SM_CONFIG kubectl get space -n spaceprovider-default -o json | jq -r '.items | .[] | .metadata | select(.annotations ["edge.kubestellar.io/sync-target-name"] == "florin") | .name')
echo The florin mailbox space name is $FLORIN_SPACE
```
``` { .bash .no-copy }
The florin mailbox space name is 1t82bk54r6gjnzsp-mb-1a045336-8178-4026-8a56-5cd5609c0ec1
```
<!--example1-stage-1a-end-->
