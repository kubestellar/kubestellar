---
title: "2023q1 PoC commands"
linkTitle: "2023q1 PoC commands"
weight: 100
---

This PoC includes two sorts of commands for users to use.  Most are
executables delivered in the `bin` directory.  The other sort of
command for users is a `bash` script that is designed to be fetched
from github and fed directly into `bash`.

# Executables

The command lines exhibited below presume that you have added the
`bin` directory to your `$PATH`.  Alternatively: these executables can
be invoked directly using any pathname (not in your `$PATH`).

## Platform control

The `kubestellar` command has two subcommands, one for setup and one
for teardown.

### Kubestellar start

This subcommand is used after installation.
This subcommand does two things:
1. It creates and populates the edge service provider workspace (ESPW) if it does not already exist; and
2. It stops any running kubesteallar controllers and then starts them all.

```shell
kubestellar start [flags]
```

The flags can appear before and/or after the subcommand.

The available flags are as follows.

- `-V` or `--verbose`: calls for more verbose output.  This is a
  binary choice, not a matter of degree.
- `--log-folder $pathname`: says where to put the logs from the
  controllers.  Will be `mkdir -p` if absent.  Defaults to
  `${PWD}/kubestellar-logs`.
- `-h` or `--help`: print a brief usage message and terminate.

### Kubestellar stop

This subcommand undoes `kubestellar start`.  It stops any running
controllers and deletes the ESPW.

This command accepts all the same flags as `kubestellar start` but
ignores the `--log-folder`.

## Creating SyncTarget/Location pairs

In this PoC, the interface between infrastructure and workload
management is inventory API objects.  Specifically, for each edge
cluster there is a unique pair of SyncTarget and Location objects in a
so-called inventory management workspace.  The following command helps
with making that pair of objects.

This commad accepts all the global command-line options of `kubectl`
excepting `--context`.

Invoke this command when your current workspace is your chosen
inventory management workspace, or specify that workspace with the
`--imw` command line flag.  Upon completion, the current workspace
will be your chosen inventory management workspace.

This command does not depend on the action of any of the edge-mc
(Kubestellar) controllers.

```console
$ kubectl kubestellar ensure location -h
Usage: kubectl kubestellar ensure location ($kubectl_flag | --imw ws_path)* objname labelname=labelvalue ...

$ kubectl kubestellar ensure location --imw root:imw-1 demo1 foo=bar the-word=the-bird
Current workspace is "root:imw-1".
synctarget.workload.kcp.io/demo1 created
location.scheduling.kcp.io/demo1 created
synctarget.workload.kcp.io/demo1 labeled
location.scheduling.kcp.io/demo1 labeled
synctarget.workload.kcp.io/demo1 labeled
location.scheduling.kcp.io/demo1 labeled
```

The above example shows using this script to create a SyncTarget and a
Location named `demo1` with labels `foo=bar` and `the-word=the-bird`.
This was equivalent to the following commands.

```shell
kubectl ws root:imw-1
kubectl create -f -<<EOF
apiVersion: workload.kcp.io/v1alpha1
kind: SyncTarget
metadata:
  name: demo1
  labels:
    id: demo1
    foo: bar
    the-word: the-bird
---
apiVersion: scheduling.kcp.io/v1alpha1
kind: Location
metadata:
  name: demo1
  labels:
    foo: bar
    the-word: the-bird
spec:
  resource: {group: workload.kcp.io, version: v1alpha1, resource: synctargets}
  instanceSelector:
    matchLabels: {"id":"demo1"}
EOF
```

This command operates in idempotent style, making whatever changes (if
any) are needed to move from the current state to the desired state.
Current limitation: it does not cast a skeptical eye on the spec of a
pre-existing Location.

## Removing SyncTarget/Location pairs

The following script undoes whatever remains from a corresponding
usage of `kubectl kubestellar ensure location`.

This commad accepts all the global command-line options of `kubectl`
excepting `--context`.

Invoke this command when your current workspace is your chosen
inventory management workspace, or specify that workspace with the
`--imw` command line flag.  Upon completion, the current workspace
will be your chosen inventory management workspace.

This command does not depend on the action of any of the edge-mc
(Kubestellar) controllers.

```console
$ kubectl kubestellar remove location -h
Usage: kubectl kubestellar remove location ($kubectl_flag | --imw ws_path)* objname

$ kubectl ws root:imw-1
Current workspace is "root:imw-1".

$ kubectl kubestellar remove location demo1
synctarget.workload.kcp.io "demo1" deleted
location.scheduling.kcp.io "demo1" deleted

$ kubectl kubestellar remove location demo1

$ 
```

## Syncer preparation and installation

The syncer runs in each edge cluster and also talks to the
corresponding mailbox workspace.  In order for it to be able to do
that, there is some work to do in the mailbox workspace to create a
ServiceAccount for the syncer to authenticate as and create RBAC
objects to give the syncer the privileges that it needs.  The
following script does those things and also outputs YAML to be used to
install the syncer in the edge cluster.

This commad accepts all the global command-line options of `kubectl`
excepting `--context`.

Invoke this command when your current workspace is your chosen
inventory management workspace, or specify that workspace with the
`--imw` command line flag.  Upon completion, the current workspace
will be what it was when the command started.

This command will only succeed if the mailbox controller has created
and conditioned the mailbox workspace for the given SyncTarget.  This
command waits for 10 to 70 seconds for that to happen.

```console
$ kubectl kubestellar prep-for-syncer -h                     
Usage: kubectl kubestellar prep-for-syncer ($kubectl_flag | --imw ws_path | --espw ws_path | --syncer-image image_ref | -o filename)* synctarget_name

$ kubectl kubestellar prep-for-syncer --imw root:imw-1 demo1
Current workspace is "root:imw-1".
Current workspace is "root:espw"
Current workspace is "root:espw:4yqm57kx0m6mn76c-mb-406c54d1-64ce-4fdc-99b3-cef9c4fc5010" (type root:universal).
Creating service account "kcp-edge-syncer-demo1-28at01r3"
Creating cluster role "kcp-edge-syncer-demo1-28at01r3" to give service account "kcp-edge-syncer-demo1-28at01r3"

 1. write and sync access to the synctarget "kcp-edge-syncer-demo1-28at01r3"
 2. write access to apiresourceimports.

Creating or updating cluster role binding "kcp-edge-syncer-demo1-28at01r3" to bind service account "kcp-edge-syncer-demo1-28at01r3" to cluster role "kcp-edge-syncer-demo1-28at01r3".

Wrote physical cluster manifest to demo1-syncer.yaml for namespace "kcp-edge-syncer-demo1-28at01r3". Use

  KUBECONFIG=<pcluster-config> kubectl apply -f "demo1-syncer.yaml"

to apply it. Use

  KUBECONFIG=<pcluster-config> kubectl get deployment -n "kcp-edge-syncer-demo1-28at01r3" kcp-edge-syncer-demo1-28at01r3

to verify the syncer pod is running.
Current workspace is "root:espw".
```

Once that script has run, the YAML for the objects to create in the
edge cluster is in your chosen output file.  The default for the
output file is the name of the SyncTarget object with "-syncer.yaml"
appended.

Create those objects with a command like the following; adjust as
needed to configure `kubectl` to modify the edge cluster and read your
chosen output file.

```shell
KUBECONFIG=$demo1_kubeconfig kubectl apply -f demo1-syncer.yaml
```

## Edge cluster on-boarding

The following command is a combination of `kubectl kubestellar
ensure-location` and `kubectl kubestellar prep-for-syncer`.

```console
$ kubectl kubestellar prep-for-cluster -h                              
Usage: kubectl kubestellar prep-for-cluster ($kubectl_flag | --imw ws_path | --espw ws_path | --syncer-image image_ref | -o filename)* synctarget_name labelname=labelvalue...

$ kubectl kubestellar prep-for-cluster --imw root:imw-1 demo2 key1=val1
Current workspace is "root:imw-1".
synctarget.workload.kcp.io/demo2 created
location.scheduling.kcp.io/demo2 created
synctarget.workload.kcp.io/demo2 labeled
location.scheduling.kcp.io/demo2 labeled
Current workspace is "root:imw-1".
Current workspace is "root:espw".
Current workspace is "root:espw:1cpf1cd4ydy13vo1-mb-3c354acd-ed86-45bb-a60d-cee8e59973f7" (type root:universal).
Creating service account "kcp-edge-syncer-demo2-15nq4e94"
Creating cluster role "kcp-edge-syncer-demo2-15nq4e94" to give service account "kcp-edge-syncer-demo2-15nq4e94"

 1. write and sync access to the synctarget "kcp-edge-syncer-demo2-15nq4e94"
 2. write access to apiresourceimports.

Creating or updating cluster role binding "kcp-edge-syncer-demo2-15nq4e94" to bind service account "kcp-edge-syncer-demo2-15nq4e94" to cluster role "kcp-edge-syncer-demo2-15nq4e94".

Wrote physical cluster manifest to demo2-syncer.yaml for namespace "kcp-edge-syncer-demo2-15nq4e94". Use

  KUBECONFIG=<pcluster-config> kubectl apply -f "demo2-syncer.yaml"

to apply it. Use

  KUBECONFIG=<pcluster-config> kubectl get deployment -n "kcp-edge-syncer-demo2-15nq4e94" kcp-edge-syncer-demo2-15nq4e94

to verify the syncer pod is running.
```

## Creating a Workload Management Workspace

Such a workspace needs not only to be created but also populated with
an `APIBinding` to the edge API and, if desired, an `APIBinding` to
the Kubernetes containerized workload API.

Invoke this script when the current kcp workspace is the parent of the
desired workload management workspace (WMW).

The default behavior is to include an `APIBinding` to the Kubernetes
containerized workload API, and there are optional command line flags
to control this behavior.

This script works in idempotent style, doing whatever work remains to
be done.

```console
$ kubectl kubestellar ensure wmw -h
Usage: kubectl ws parent; kubectl kubestellar ensure wmw ($kubectl_flag | --with-kube boolean)* wm_workspace_name

$ kubectl ws .
Current workspace is "root:my-org".

$ kubectl kubestellar ensure wmw example-wmw
Current workspace is "root".
Current workspace is "root:my-org".
Workspace "example-wmw" (type root:universal) created. Waiting for it to be ready...
Workspace "example-wmw" (type root:universal) is ready to use.
Current workspace is "root:my-org:example-wmw" (type root:universal).
apibinding.apis.kcp.io/bind-espw created
apibinding.apis.kcp.io/bind-kube created

$ kubectl ws ..
Current workspace is "root:my-org".

$ kubectl kubestellar ensure wmw example-wmw
Current workspace is "root".
Current workspace is "root:my-org".
Current workspace is "root:my-org:example-wmw" (type root:universal).

$ kubectl ws ..
Current workspace is "root:my-org".

$ kubectl kubestellar ensure wmw example-wmw --with-kube false
Current workspace is "root".
Current workspace is "root:my-org".
Current workspace is "root:my-org:example-wmw" (type root:universal).
apibinding.apis.kcp.io "bind-kube" deleted

$ 
```

## Removing a Workload Management Workspace

Deleting a WMW can be done by simply deleting its `Workspace` object from
the parent.

```console
$ kubectl ws .
Current workspace is "root:my-org:example-wmw".

$ kubectl ws ..
Current workspace is "root:my-org".

$ kubectl delete Workspace example-wmw
workspace.tenancy.kcp.io "example-wmw" deleted

$ 
```

Alternatively, you can use the following command line whose design
completes the square here.  Invoke it when the current workspace is
the parent of the workload management workspace to delete.

```console
$ kubectl kubestellar remove wmw -h
Usage: kubectl ws parent; kubectl kubestellar remove wmw kubectl_flag... wm_workspace_name

$ kubectl ws root:my-org
Current workspace is "root:my-org".

$ kubectl kubestellar remove wmw demo1
workspace.tenancy.kcp.io "demo1" deleted

$ kubectl ws .
Current workspace is "root:my-org".

$ kubectl kubestellar remove wmw demo1

$ 
```

# Web-to-bash

## Quick Setup

This is a combination of some installation and setup steps, for use in
[the
QuickStart](https://github.com/kcp-dev/edge-mc/blob/main/QUICKSTART.md).

The script can be read directly from
https://raw.githubusercontent.com/kcp-dev/edge-mc/main/bootstrap/bootstrap-kubestellar.sh
and does the following things.

1. Downloads and installs kcp if it is not already evident on `$PATH`
   (using [the script below](#install-kcp-and-its-kubectl-plugins).
2. Starts a kcp server if one is not already running.
3. Downloads and installs kubestellar if it is not already evident on
   `$PATH` (using [the script below](#install-kubestellar).
4. `kubestellar start` if the KubeStellar controllers are not already running.

This script accepts the following command line flags; all are optional.

- `--kubestellar-version $version`: specifies the release of
  KubeStellar to use.  The default is the latest regular release.
- `--kcp-version $version`: specifies the kcp release to use.  The
  default is the one that works with the chosen release of
  KubeStellar.
- `--os $OS`: specifies the operating system to use in selecting the
  executables to download and install.  Choices are `linux` and
  `darwin`.  Autodetected if omitted.
- `--arch $IAS`: specifies the instruction set architecture to use in
  selecting the executables to download and install.  Choices are
  `amd64` and `arm64`.  Autodetected if omitted.
- `--bind-address $IPADDR`: directs that the kcp server (a) write that
  address for itself in the kubeconfig file that it constructs and (b)
  listens only at that address.  The default is to pick one of the
  host's non-loopback addresses to write into the kubeconfig file and
  not bind a listening address.
- `--ensure-folder $install_parent_dir`: specifies the parent folder
  for downloads.  Will be `mkdir -p`.  The default is the current
  working directory.  The download of kcp, if any, will go in
  `$install_parent_dir/kcp`.  The download of KubeStellar will go in
  `$install_parent_dir/kubestellar`.
- `-V` or `--verbose`: incrases the verbosity of output.  This is a
  binary thing, not a matter of degree.
- `-X`: makes the script `set -x` internally, for debugging.
- `-h` or `--help`: print brief usage message and exit.

Here "install" means only to (a) unpack the distribution archives into
the relevant places under `$install_parent_dir` and (b) enhance the
`PATH`, and `KUBECONFIG` in the case of kcp, environment variables in
the shell running the script.  Of course, if you run the script in a
subshell then those environment effects terminate with that subshell;
this script also prints out messages showing how to update the
environment in another shell.

## Install kcp and its kubectl plugins

This script is directly available at
https://raw.githubusercontent.com/kcp-dev/edge-mc/main/bootstrap/install-kcp-with-plugins.sh
and does the following things.

- Fetch and install the `kcp` server executable.
- Fetch and install the kubectl plugins of kcp.

This script accepts the following command line flags; all are
optional.

- `--version $version`: specifies the kcp release to use.  The default
  is the latest.
- `--OS $OS`: specifies the operating system to use in selecting the
  executables to fetch and install.  Choices are `darwin` and `linux`.
  Auto-detected if omitted.
- `--arch $ARCH`: specifies the instruction set architecture to use in
  selecting the executables to fetch and install.  Choices are `arm64`
  and `amd64`.  Auto-detected if omitted.
- `--ensure-folder $install_parent_dir`: specifies where to install
  to.  This will be `mkdir -p`.  The default is `./kcp`.
- `-V` or `--verbose`: incrases the verbosity of output.  This is a
  binary thing, not a matter of degree.
- `-X`: makes the script `set -x` internally, for debugging.
- `-h` or `--help`: print brief usage message and exit.

Here install means only to unpack the downloaded archives, creating
`$install_parent_dir/bin`.  If `$install_parent_dir/bin` is not
already on your `$PATH` then this script will print out a message
telling you to add it.

## Install KubeStellar

This script is direclty available at
https://raw.githubusercontent.com/kcp-dev/edge-mc/main/bootstrap/install-kubestellar.sh
and will download and install KubeStellar.

This script accepts the following command line arguments; all are
optional.

- `--version $version`: specifies the release of KubeStellar to use.
  Defaults to the latest regular release.
- `--OS $OS`: specifies the operating system to use in selecting the
  executables to fetch and install.  Choices are `darwin` and `linux`.
  Auto-detected if omitted.
- `--arch $ARCH`: specifies the instruction set architecture to use in
  selecting the executables to fetch and install.  Choices are `arm64`
  and `amd64`.  Auto-detected if omitted.
- `--ensure-folder $install_parent_dir`: specifies where to install
  to.  This will be `mkdir -p`.  The default is `./kubestellar`.
- `-V` or `--verbose`: incrases the verbosity of output.  This is a
  binary thing, not a matter of degree.
- `-X`: makes the script `set -x` internally, for debugging.
- `-h` or `--help`: print brief usage message and exit.

Here install means only to unpack the downloaded archive, creating
`$install_parent_dir/bin`.  If `$install_parent_dir/bin` is not
already on your `$PATH` then this script will print out a message
telling you to add it.
