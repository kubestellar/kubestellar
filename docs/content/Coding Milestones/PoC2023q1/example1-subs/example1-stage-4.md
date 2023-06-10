<!--example1-stage-4-start-->
## Stage 4

![Syncer effects](../Edge-PoC-2023q1-Scenario-1-stage-4.svg "Stage 4 summary")

In Stage 4, the edge syncer does its thing.  Actually, it should have
done it as soon as the relevant inputs became available in stage 3.
Now we examine what happened.

You can check that the workloads are running in the edge clusters as
they should be.

The syncer does its thing between the florin cluster and its mailbox
workspace.  This is driven by the `SyncerConfig` object named
`the-one` in that mailbox workspace.

The syncer does its thing between the guilder cluster and its mailbox
workspace.  This is driven by the `SyncerConfig` object named
`the-one` in that mailbox workspace.

Using the kubeconfig that `kind` modified, examine the florin cluster.
Find just the `commonstuff` namespace and the `commond` Deployment.

```shell
KUBECONFIG=~/.kube/config kubectl --context kind-florin get ns
```
``` { .bash .no-copy }
NAME                                 STATUS   AGE
commonstuff                          Active   6m51s
default                              Active   57m
kubestellar-syncer-florin-1t9zgidy   Active   17m
kube-node-lease                      Active   57m
kube-public                          Active   57m
kube-system                          Active   57m
local-path-storage                   Active   57m
```

```shell
KUBECONFIG=~/.kube/config kubectl --context kind-florin get deploy -A | egrep 'NAME|stuff'
```
``` { .bash .no-copy }
NAMESPACE                         NAME                              READY   UP-TO-DATE   AVAILABLE   AGE
commonstuff                       commond                           1/1     1            1           7m59s
```

Examine the guilder cluster.  Find both workload namespaces and both
Deployments.

```shell
KUBECONFIG=~/.kube/config kubectl --context kind-guilder get ns | egrep NAME\|stuff
```
``` { .bash .no-copy }
NAME                               STATUS   AGE
commonstuff                        Active   8m33s
specialstuff                       Active   8m33s
```

```shell
KUBECONFIG=~/.kube/config kubectl --context kind-guilder get deploy -A | egrep NAME\|stuff
```
``` { .bash .no-copy }
NAMESPACE                          NAME                               READY   UP-TO-DATE   AVAILABLE   AGE
commonstuff                        commond                            1/1     1            1           8m37s
specialstuff                       speciald                           1/1     1            1           8m55s
```

Examining the common workload in the guilder cluster, for example,
will show that the replacement-style customization happened.

```shell
KUBECONFIG=~/.kube/config kubectl --context kind-guilder get deploy -n commonstuff commond -o yaml
```
``` { .bash .no-copy }
...
      containers:
      - env:
        - name: EXAMPLE_VAR
          value: env is prod
        image: library/httpd:2.4
        imagePullPolicy: IfNotPresent
        name: httpd
...
```

Check that the common workload on the florin cluster is working.

``` {.bash .hide-me}
sleep 10
```
```shell
curl http://localhost:8094
```
``` { .bash .no-copy }
<!DOCTYPE html>
<html>
  <body>
    This is a common web site.
    Running in florin.
  </body>
</html>
```

Check that the special workload on the guilder cluster is working.
``` {.bash .hide-me}
sleep 10
```
```shell
curl http://localhost:8097
```
``` { .bash .no-copy }
<!DOCTYPE html>
<html>
  <body>
    This is a special web site.
    Running in guilder.
  </body>
</html>
```

Check that the common workload on the guilder cluster is working.

```shell
curl http://localhost:8096
```
``` { .bash .no-copy }
<!DOCTYPE html>
<html>
  <body>
    This is a common web site.
    Running in guilder.
  </body>
</html>
```
<!--example1-stage-4-stop-->