<!--user-example1-stage-1a-start-->
### Deploy the KubeStellar helm chart

```shell
kubectl kubestellar deploy --external-endpoint my-long-application-name.my-region.some.cloud.com:1234
```

get the names of all workspaces created with associated synctargets

```shell
kubectl get Workspace -o "custom-columns=NAME:.metadata.name,SYNCTARGET:.metadata.annotations['edge\.kcp\.io/sync-target-name'],CLUSTER:.spec.cluster"
```
``` { .bash .no-copy }
NAME                                                       SYNCTARGET   CLUSTER
1t82bk54r6gjnzsp-mb-1a045336-8178-4026-8a56-5cd5609c0ec1   ren       1najcltzt2nqax47
1t82bk54r6gjnzsp-mb-f0a82ab1-63f4-49ea-954d-3a41a35a9f1c   stimpy      1y7wll1dz806h3sb
```

get the name of the specific mailbox associated with the synctarget

```shell
REN_WS=$(kubectl get Workspace -o json | jq -r '.items | .[] | .metadata | select(.annotations ["edge.kcp.io/sync-target-name"] == "ren") | .name')
echo The ren mailbox workspace name is $REN_WS
```


``` { .bash .no-copy }
The ren mailbox workspace name is 1t82bk54r6gjnzsp-mb-1a045336-8178-4026-8a56-5cd5609c0ec1
```
<!--user-example1-stage-1a-end-->
