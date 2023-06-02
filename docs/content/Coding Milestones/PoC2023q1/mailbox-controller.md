---
title: "Mailbox Controller"
date: 2023-02-02
weight: 4
description: >
---


## Linking SyncTarget with Mailbox Workspace

For a given SyncTarget T, the mailbox controller currently chooses the
name of the corresponding workspace to be the concatenation of the
following.

- the ID of the logical cluster containing T
- the string "-mb-"
- T's UID

The mailbox workspace gets an annotation whose key is
`edge.kcp.io/sync-target-name` and whose value is the name of the
workspace object (as seen in its parent workspace, the edge service
provider workspace).

## Usage

The mailbox controller needs three Kubernetes client configurations.
One --- concerned with reading inventory --- is to access the
APIExport view of the `workload.kcp.io` API group, to read the
`SyncTarget` objects.  This must be a client config that is pointed at
the workspace (which is always `root`, as far as I know) that has this
APIExport and is authorized to read its view.  Another client config
is needed to give read/write access to all the mailbox workspaces, so
that the controller can create `APIBinding` objects to the edge
APIExport in those workspaces; this should be a client config that is
able to read/write in all clusters.  For example, that is in the
kubeconfig context named `base` in the kubeconfig created by `kcp
start`.  Finally, the controller also needs a kube client config that
is pointed at the edge service provider workspace and is authorized to
consume the `Workspace` objects from there.

The command line flags, beyond the basics, are as follows.

```console
      --concurrency int                  number of syncs to run in parallel (default 4)
      --espw-path string                 the pathname of the edge service provider workspace (default "root:espw")

      --inventory-cluster string         The name of the kubeconfig cluster to use for access to APIExport view of SyncTarget objects
      --inventory-context string         The name of the kubeconfig context to use for access to APIExport view of SyncTarget objects (default "root")
      --inventory-kubeconfig string      Path to the kubeconfig file to use for access to APIExport view of SyncTarget objects
      --inventory-user string            The name of the kubeconfig user to use for access to APIExport view of SyncTarget objects

      --mbws-cluster string              The name of the kubeconfig cluster to use for access to mailbox workspaces (really all clusters)
      --mbws-context string              The name of the kubeconfig context to use for access to mailbox workspaces (really all clusters) (default "base")
      --mbws-kubeconfig string           Path to the kubeconfig file to use for access to mailbox workspaces (really all clusters)
      --mbws-user string                 The name of the kubeconfig user to use for access to mailbox workspaces (really all clusters)

      --server-bind-address ipport       The IP address with port at which to serve /metrics and /debug/pprof/ (default :10203)

      --workload-cluster string          The name of the kubeconfig cluster to use for access to edge service provider workspace
      --workload-context string          The name of the kubeconfig context to use for access to edge service provider workspace
      --workload-kubeconfig string       Path to the kubeconfig file to use for access to edge service provider workspace
      --workload-user string             The name of the kubeconfig user to use for access to edge service provider workspace
```

## Try It

To exercise it, do the following steps.

Clone this repo, install kcp (the version for `github.com/kcp-dev/kcp` in `go.mod` is required) 
and start a kcp server as described [here](kubestellar-scheduler.md#steps-to-try-the-scheduler).

Do the remaining steps in a separate shell, with
`$KUBECONFIG` set to the admin config for that kcp server.  

Next, create the edge service provider workspace:

```shell
kubectl ws root
kubectl ws create espw --enter
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
kubectl ws create imw --enter
```

Then in that workspace, run the following command to create a `SyncTarget`
object.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: workload.kcp.io/v1alpha1
kind: SyncTarget
metadata:
  name: stest1
spec:
  cells:
    foo: bar
EOF
```

That should provoke logging like the following from the mailbox controller.

```console
I0305 18:07:20.490417   85556 main.go:369] "Created missing workspace" worker=0 mbwsName="niqdko2g2pwoadfb-mb-f99e773f-3db2-439e-8054-827c4ac55368"
```

And you can verify that as follows:

```console
$ kubectl ws root:espw
Current workspace is "root:espw".

$ kubectl get workspaces
NAME                                                       TYPE        REGION   PHASE   URL                                                     AGE
niqdko2g2pwoadfb-mb-f99e773f-3db2-439e-8054-827c4ac55368   universal            Ready   https://192.168.58.123:6443/clusters/0ay27fcwuo2sv6ht   22s
```

FYI, if you look inside that workspace you will see an `APIBinding`
named `bind-edge` that binds to the `APIExport` named `edge.kcp.io`
from the edge service provider workspace (and this is why the
controller needs to know the pathname of that workspace), so that the
edge API is available in the mailbox workspace.

Next, `kubectl delete` that workspace, and watch the mailbox
controller wait for it to be gone and then re-create it.

```console
I0305 18:08:15.428884   85556 main.go:369] "Created missing workspace" worker=2 mbwsName="niqdko2g2pwoadfb-mb-f99e773f-3db2-439e-8054-827c4ac55368"
```

Finally, go back to your inventory workspace to delete the `SyncTarget`:

```shell
kubectl ws \~
kubectl ws imw
kubectl delete SyncTarget stest1
```
and watch the mailbox controller react as follows.

```console
I0305 18:08:44.380421   85556 main.go:352] "Deleted unwanted workspace" worker=0 mbwsName="niqdko2g2pwoadfb-mb-f99e773f-3db2-439e-8054-827c4ac55368"
```
