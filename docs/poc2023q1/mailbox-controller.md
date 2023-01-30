# Mailbox Controller

The mailbox controller runs in the edge service provider workspace and
maintains a child workspace per SyncTarget.

## Temporary design detail

For a given SyncTarget T, the mailbox controller currently chooses the
name of the corresponding workspace to be the concatenation of the
following.

- the name (_not_ ID) of the workspace containing T
- "-w-"
- T's name

This is ambiguous and subject to name length overflows.  Later
revisions will do better.

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
`$KUBECONFIG` set to the admin config for that kcp server.

`kubectl ws root`

`kubectl ws create edge --enter`

After that, a run of the controller should look like the following.

```shell
(base) mspreitz@mjs12 edge-mc % go run ./cmd/mailbox-controller -v=2
I0127 00:21:48.876022   24503 main.go:206] "Found APIExport view" exportName="workload.kcp.io" serverURL="https://192.168.58.123:6443/services/apiexport/root/workload.kcp.io"
I0127 00:21:48.877965   24503 shared_informer.go:255] Waiting for caches to sync for mailbox-controller
I0127 00:21:48.978352   24503 shared_informer.go:262] Caches are synced for mailbox-controller
I0127 00:21:48.978414   24503 main.go:169] "Informers synced"
```

In a separate shell, `kubectl create` the following object.

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
I0127 00:23:52.545266   24503 main.go:330] "Created missing workspace" worker=1 wsName="269p7excqtb3xen8-w-stest1"
```

And you can verify that like so.

```shell
(base) mspreitz@mjs12 kcp % kubectl get Workspace
NAME                        TYPE        PHASE   URL                                                     AGE
269p7excqtb3xen8-w-stest1   universal           https://192.168.58.123:6443/clusters/2ktljajura89bwf2   29s
```

Next, `kubectl delete` that Workspace, and watch the mailbox
controller wait for it to be gone and then re-create it.

```
I0127 00:26:40.251001   24503 main.go:330] "Created missing workspace" worker=2 wsName="269p7excqtb3xen8-w-stest1"
```

Finally, `kubectl delete SyncTarget stest1` and watch the mailbox
controller react as follows.

```
I0127 00:26:59.369990   24503 main.go:311] "Deleted unwanted workspace" worker=1 wsName="269p7excqtb3xen8-w-stest1"
```
