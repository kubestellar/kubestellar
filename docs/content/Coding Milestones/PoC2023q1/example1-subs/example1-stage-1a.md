<!--example1-stage-1a-start-->
### The mailbox controller

The mailbox controller is one of the central controllers of
KubeStellar.  If you have deployed the KubeStellar core as Kubernetes
workload then this controller is already running in a pod in your
hosting cluster. If instead you are running these controllers as bare
processes then launch this controller as follows.

```shell
KUBECONFIG=$SM_CONFIG mailbox-controller --external-access -v=4 &> /tmp/mailbox-controller.log &
sleep 20
```

This controller is in charge of maintaining the collection of mailbox
spaces, which are an implementation detail not intended for user
consumption. You can use the following command to wait for the
appearance of the mailbox spaces implied by the florin and guilder
`SyncTarget` objects that you made earlier.

```shell
while [ $(KUBECONFIG=$SM_CONFIG kubectl get spaces -A | grep "\-mb\-" | wc -l) -ne 2 ]; do
  sleep 10
done
```

If it is working correctly, lines like the following will appear in
the controller's log (which is being written into /tmp/mailbox-controller.log if you ran the controller as a bare process above, otherwise you can fetch [as directed](../../commands/#fetch-a-log-from-a-kubestellar-runtime-container)).

``` { .bash .no-copy }
...
I1218 16:36:14.743434  239027 controller.go:277] "Created missing space" worker=0 mbsName="imw1-mb-ace9fd79-1c35-410d-879d-d73ec6f9fe6b"
...
I1218 16:36:14.759426  239027 controller.go:277] "Created missing space" worker=1 mbsName="imw1-mb-11c4c16a-82f9-48ad-b036-879aba135959"
```

You need a `-v` setting of 2 or numerically higher to get log messages
about individual mailbox spaces.

A mailbox space name is distinguished by `-mb-` separator.
You can get a listing of those mailbox spaces as follows.

```shell
KUBECONFIG=$SM_CONFIG kubectl get spaces -A
```
``` { .bash .no-copy }
NAMESPACE               NAME                                           AGE
spaceprovider-default   espw                                           84s
spaceprovider-default   imw1                                           76s
spaceprovider-default   imw1-mb-11c4c16a-82f9-48ad-b036-879aba135959   20s
spaceprovider-default   imw1-mb-ace9fd79-1c35-410d-879d-d73ec6f9fe6b   20s
spaceprovider-default   wmw1                                           53s
```

More usefully, using custom columns you can get a listing that shows
the _name_ of the associated SyncTarget.

```shell
KUBECONFIG=$SM_CONFIG kubectl get spaces -n spaceprovider-default -o "custom-columns=NAME:.metadata.name,SYNCTARGET:.metadata.annotations['edge\.kubestellar\.io/sync-target-name'],CLUSTER:.spec.cluster"
```
``` { .bash .no-copy }
NAME                                           SYNCTARGET   CLUSTER
espw                                           <none>       <none>
imw1                                           <none>       <none>
imw1-mb-11c4c16a-82f9-48ad-b036-879aba135959   florin       <none>
imw1-mb-ace9fd79-1c35-410d-879d-d73ec6f9fe6b   guilder      <none>
wmw1                                           <none>       <none>
```

Also: if you ever need to look up just one mailbox space by
SyncTarget name, you could do it as follows.

```shell
GUILDER_SPACE=$(KUBECONFIG=$SM_CONFIG kubectl get space -n spaceprovider-default -o json | jq -r '.items | .[] | .metadata | select(.annotations ["edge.kubestellar.io/sync-target-name"] == "guilder") | .name')
echo The guilder mailbox space name is $GUILDER_SPACE
```
``` { .bash .no-copy }
The guilder mailbox space name is imw1-mb-ace9fd79-1c35-410d-879d-d73ec6f9fe6b
```

```shell
FLORIN_SPACE=$(KUBECONFIG=$SM_CONFIG kubectl get space -n spaceprovider-default -o json | jq -r '.items | .[] | .metadata | select(.annotations ["edge.kubestellar.io/sync-target-name"] == "florin") | .name')
echo The florin mailbox space name is $FLORIN_SPACE
```
``` { .bash .no-copy }
The florin mailbox space name is imw1-mb-11c4c16a-82f9-48ad-b036-879aba135959
```
<!--example1-stage-1a-end-->
