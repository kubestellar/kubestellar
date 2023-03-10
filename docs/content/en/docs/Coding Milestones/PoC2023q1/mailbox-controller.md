---
title: "Mailbox Controller"
date: 2023-02-02
weight: 4
description: >
---

{{% pageinfo %}}
The mailbox controller runs in the edge service provider workspace and maintains a child workspace per SyncTarget.
{{% /pageinfo %}}

## Linking SyncTarget with Mailbox Workspace

For a given SyncTarget T, the mailbox controller currently chooses the
name of the corresponding workspace to be the concatenation of the
following.

- the ID of the logical cluster containing T
- the string "-mb-"
- T's UID

The mailbox workspace gets an annotation whose key is
`edge.kcp.io/sync-target-name` and whose value is the name of the
Workspace object (as seen in its parent workspace, the edge service
provider workspace).

## Usage

The command line flags, beyond the basics, are as follows.

```console
      --concurrency int                  number of syncs to run in parallel (default 4)
      --inventory-context string         current-context override for inventory-kubeconfig (default "root")
      --inventory-kubeconfig string      pathname of kubeconfig file for inventory service provider workspace
      --server-bind-address ipport       The IP address with port at which to serve /metrics and /debug/pprof/ (default :10203)
      --workload-context string          current-context override for workload-kubeconfig
      --workload-kubeconfig string       pathname of kubeconfig file for edge workload service provider workspace
```

## Try It

To exercise it, do the following steps.

Start a kcp server.  Do the remaining steps in a separate shell, with
`$KUBECONFIG` set to the admin config for that kcp server.  This will
create the edge service provider workspace.

```shell
kubectl ws root
kubectl ws create edge --enter
```

After that, a run of the controller should look like the following.

```console
$ go run ./cmd/mailbox-controller -v=2
I0305 18:06:20.046741   85556 main.go:110] "Command line flag" add_dir_header="false"
I0305 18:06:20.046954   85556 main.go:110] "Command line flag" alsologtostderr="false"
I0305 18:06:20.046960   85556 main.go:110] "Command line flag" concurrency="4"
I0305 18:06:20.046965   85556 main.go:110] "Command line flag" inventory-context="root"
I0305 18:06:20.046971   85556 main.go:110] "Command line flag" inventory-kubeconfig=""
I0305 18:06:20.046976   85556 main.go:110] "Command line flag" log_backtrace_at=":0"
I0305 18:06:20.046980   85556 main.go:110] "Command line flag" log_dir=""
I0305 18:06:20.046985   85556 main.go:110] "Command line flag" log_file=""
I0305 18:06:20.046989   85556 main.go:110] "Command line flag" log_file_max_size="1800"
I0305 18:06:20.046993   85556 main.go:110] "Command line flag" logtostderr="true"
I0305 18:06:20.046997   85556 main.go:110] "Command line flag" one_output="false"
I0305 18:06:20.047002   85556 main.go:110] "Command line flag" server-bind-address=":10203"
I0305 18:06:20.047006   85556 main.go:110] "Command line flag" skip_headers="false"
I0305 18:06:20.047011   85556 main.go:110] "Command line flag" skip_log_headers="false"
I0305 18:06:20.047015   85556 main.go:110] "Command line flag" stderrthreshold="2"
I0305 18:06:20.047019   85556 main.go:110] "Command line flag" v="2"
I0305 18:06:20.047023   85556 main.go:110] "Command line flag" vmodule=""
I0305 18:06:20.047027   85556 main.go:110] "Command line flag" workload-context=""
I0305 18:06:20.047031   85556 main.go:110] "Command line flag" workload-kubeconfig=""
I0305 18:06:20.070071   85556 main.go:247] "Found APIExport view" exportName="workload.kcp.io" serverURL="https://192.168.58.123:6443/services/apiexport/root/workload.kcp.io"
I0305 18:06:20.072088   85556 shared_informer.go:282] Waiting for caches to sync for mailbox-controller
I0305 18:06:20.172169   85556 shared_informer.go:289] Caches are synced for mailbox-controller
I0305 18:06:20.172196   85556 main.go:210] "Informers synced"
```

In a separate shell, make a inventory management workspace as follows.

```shell
kubectl ws \~
kubectl ws create inv1 --enter
```

Then in that workspace, `kubectl create` the following `SyncTarget`
object.

```yaml
apiVersion: workload.kcp.io/v1alpha1
kind: SyncTarget
metadata:
  name: stest1
spec:
  cells:
    foo: bar
```

That should provoke logging like the following from the mailbox controller.

```console
I0305 18:07:20.490417   85556 main.go:369] "Created missing workspace" worker=0 mbwsName="niqdko2g2pwoadfb-mb-f99e773f-3db2-439e-8054-827c4ac55368"
```

And you can verify that like so.

```console
$ kubectl ws root:edge
Current workspace is "root:edge".

$ kubectl get Workspace
NAME                                                       TYPE        REGION   PHASE   URL                                                     AGE
niqdko2g2pwoadfb-mb-f99e773f-3db2-439e-8054-827c4ac55368   universal            Ready   https://192.168.58.123:6443/clusters/0ay27fcwuo2sv6ht   22s
```

Next, `kubectl delete` that Workspace, and watch the mailbox
controller wait for it to be gone and then re-create it.

```console
I0305 18:08:15.428884   85556 main.go:369] "Created missing workspace" worker=2 mbwsName="niqdko2g2pwoadfb-mb-f99e773f-3db2-439e-8054-827c4ac55368"
```

Finally, go back to your inventory workspace and `kubectl delete
SyncTarget stest1` and watch the mailbox controller react as follows.

```console
I0305 18:08:44.380421   85556 main.go:352] "Deleted unwanted workspace" worker=0 mbwsName="niqdko2g2pwoadfb-mb-f99e773f-3db2-439e-8054-827c4ac55368"
```
