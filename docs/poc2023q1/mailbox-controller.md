# Mailbox Controller

## Try It

To exercise it, do the following steps.

Start a kcp server.  Do the remaining steps in a separate shell, with
`$KUBECONFIG` set to the admin config for that kcp server.

`kubectl ws root`

`kubectl ws create edge --enter`

After that, a run of the controller should look like the following.

```
(base) mspreitz@mjs12 edge-mc % go run ./cmd/mailbox-controller
I0126 17:29:18.326193   44934 main.go:160] "Found APIExport view" exportName="workload.kcp.io" serverURL="https://192.168.58.123:6443/services/apiexport/root/workload.kcp.io"
```

In a separate shell, create a workspace that is a child of
`root:edge`.  You should see the controller log the creation and
several updates of that workspace.
