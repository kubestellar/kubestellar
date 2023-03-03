# Mailbox Controller

The mailbox controller runs in the edge service provider workspace and
maintains a child workspace per SyncTarget.

## Linking SyncTarget with Mailbox Workspace

For a given SyncTarget T, the mailbox controller currently chooses the
name of the corresponding workspace to be the concatenation of the
following.

- the ID of the logical cluster containing T
- the string "-w-"
- T's UID

The mailbox workspace gets an annotation whose key is
`edge.kcp.io/sync-target-name` and whose value is the name of the
Workspace object (as seen in its parent workspace, the edge service
provider workspace).

## Usage

The command line flags, beyond the basics, are as follows.

```
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

```shell
(base) mspreitz@mjs12 edge-mc % go run ./cmd/mailbox-controller -v=2
I0127 00:21:48.876022   24503 main.go:206] "Found APIExport view" exportName="workload.kcp.io" serverURL="https://192.168.58.123:6443/services/apiexport/root/workload.kcp.io"
I0127 00:21:48.877965   24503 shared_informer.go:255] Waiting for caches to sync for mailbox-controller
I0127 00:21:48.978352   24503 shared_informer.go:262] Caches are synced for mailbox-controller
I0127 00:21:48.978414   24503 main.go:169] "Informers synced"
I0303 16:45:52.528677   62289 main.go:113] "Command line flag" add_dir_header="false"
I0303 16:45:52.528859   62289 main.go:113] "Command line flag" alsologtostderr="false"
I0303 16:45:52.528865   62289 main.go:113] "Command line flag" concurrency="4"
I0303 16:45:52.528869   62289 main.go:113] "Command line flag" inventory-context="root"
I0303 16:45:52.528872   62289 main.go:113] "Command line flag" inventory-kubeconfig=""
I0303 16:45:52.528876   62289 main.go:113] "Command line flag" log_backtrace_at=":0"
I0303 16:45:52.528880   62289 main.go:113] "Command line flag" log_dir=""
I0303 16:45:52.528883   62289 main.go:113] "Command line flag" log_file=""
I0303 16:45:52.528886   62289 main.go:113] "Command line flag" log_file_max_size="1800"
I0303 16:45:52.528890   62289 main.go:113] "Command line flag" logtostderr="true"
I0303 16:45:52.528893   62289 main.go:113] "Command line flag" one_output="false"
I0303 16:45:52.528897   62289 main.go:113] "Command line flag" server-bind-address=":10203"
I0303 16:45:52.528900   62289 main.go:113] "Command line flag" skip_headers="false"
I0303 16:45:52.528904   62289 main.go:113] "Command line flag" skip_log_headers="false"
I0303 16:45:52.528907   62289 main.go:113] "Command line flag" stderrthreshold="2"
I0303 16:45:52.528911   62289 main.go:113] "Command line flag" v="2"
I0303 16:45:52.528914   62289 main.go:113] "Command line flag" vmodule=""
I0303 16:45:52.528918   62289 main.go:113] "Command line flag" workload-context=""
I0303 16:45:52.528921   62289 main.go:113] "Command line flag" workload-kubeconfig=""
I0303 16:45:52.552450   62289 main.go:248] "Found APIExport view" exportName="workload.kcp.io" serverURL="https://192.168.58.123:6443/services/apiexport/root/workload.kcp.io"
I0303 16:45:52.554741   62289 shared_informer.go:282] Waiting for caches to sync for mailbox-controller
I0303 16:45:52.654988   62289 shared_informer.go:289] Caches are synced for mailbox-controller
I0303 16:45:52.655018   62289 main.go:211] "Informers synced"
```

In a separate shell, make a workload management workspace as follows.

```
kubectl ws \~
kubectl ws create work1 --enter
```

Then in that workspace, `kubectl create` the following object.

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

```
I0303 16:47:45.921037   62289 main.go:379] "Created missing workspace" worker=1 ref={cluster:2sqxqu9zxhpsgtm4 name:stest1 uid:05c38f36-3c03-4a21-a67f-6056bfca5b05}
```

And you can verify that like so.

```shell
(base) mspreitz@mjs12 ~ % kubectl ws root:edge
Current workspace is "root:edge".

(base) mspreitz@mjs12 ~ % kubectl get Workspace
NAME                                                       TYPE        REGION   PHASE   URL                                                     AGE
2sqxqu9zxhpsgtm4-mb-05c38f36-3c03-4a21-a67f-6056bfca5b05   universal            Ready   https://192.168.58.123:6443/clusters/30l2suw35h3kwg2z   55s
```

Next, `kubectl delete` that Workspace, and watch the mailbox
controller wait for it to be gone and then re-create it.

```
I0303 16:50:04.477565   62289 main.go:379] "Created missing workspace" worker=0 ref={cluster:2sqxqu9zxhpsgtm4 name:stest1 uid:05c38f36-3c03-4a21-a67f-6056bfca5b05}
```

Finally, go back to your workload workspace and `kubectl delete
SyncTarget stest1` and watch the mailbox controller react as follows.

```
I0303 16:51:09.136632   62289 main.go:362] "Deleted unwanted workspace" worker=2 ref={cluster:2sqxqu9zxhpsgtm4 name:stest1 uid:05c38f36-3c03-4a21-a67f-6056bfca5b05}
```
