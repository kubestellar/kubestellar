---
short_name: commands
---
This PoC includes two sorts of commands for users to use.  Most are
executables delivered in the `bin` directory.  The other sort of
command for users is a `bash` script that is designed to be fetched
from github and fed directly into `bash`.

# Executables

The command lines exhibited below presume that you have added the
`bin` directory to your `$PATH`.  Alternatively: these executables can
be invoked directly using any pathname (not in your `$PATH`).

**NOTE**: all of the kubectl plugin usages described here certainly or
potentially change the setting of which kcp workspace is "current" in
your chosen kubeconfig file; for this reason, they are not suitable
for executing concurrently with anything that depends on that setting
in that file.

## kcp control

KubeStellar has some commands to support situations in which some
clients of the kcp server need to connect to it by opening a
connection to a DNS domain name rather than an IP address that the
server can bind its socket to.  The key idea is using the
`--tls-sni-cert-key` command line flag of `kcp start` to configure the
server to respond with a bespoke server certificate in TLS handshakes
in which the client addresses the server by a given domain name.

These commands are designed so that they can be used multiple times if
there are multiple sets of clients that need to use a distinct domain
name.

### kubestellar-ensure-kcp-server-creds

#### kubestellar-ensure-kcp-server-creds pre-reqs

The `kubestellar-ensure-kcp-server-creds` command requires that
Easy-RSA is installed.  As outlined in
https://easy-rsa.readthedocs.io/en/latest/#obtaining-and-using-easy-rsa,
this involves selecting a release archive from [the list on
GitHub](https://github.com/OpenVPN/easy-rsa/releases), unpacking it,
and adding the EasyRSA directory to your `$PATH`; `easyrsa` is a bash
script, so you do not need to worry about building or fetching a
binary specific to your OS or computer architecture.

Easy-RSA uses [OpenSSL](https://www.openssl.org/), so you will need
that installed too.

#### kubestellar-ensure-kcp-server-creds usage

This command is given exactly one thing on the command line, a DNS
domain name.  This command creates --- or re-uses if it finds already
existing --- a private key and public X.509 certificate for the kcp
server to use.  The certificate has exactly one
SubjectAlternativeName, which is of the DNS form and specifies the
given domain name.  For example: `kubestellar-ensure-kcp-server-creds
foo.bar` creates a certificate with one SAN, commonly rendered as
`DNS:foo.bar`.

This command uses a Public Key Infrastructure (PKI) and Certificate
Authority (CA) implemented by
[easy-rsa](https://github.com/OpenVPN/easy-rsa), rooted at the
subdirectory `pki` of the current working directory.  This command
will create the PKI if it does not already exist, and will initialize
the CA if that has not already been done.  The CA's public certificate
appears at the usual place for easy-rsa: `pki/ca.crt`.

This command prints some stuff --- mostly progress remarks from
easy-rsa --- to stderr and prints one line of results to stdout.  The
`bash` shell will parse that one line as three words.  Each is an
absolute pathname of one certificate or key.  The three are as
follows.

1. The CA certificate for the client to use in verifying the server.
2. The X.509 certificate for the server to put in its ServerHello in a
   TLS handshake in which the ClientHello has a Server Name Indicator
   (SNI) that matches the given domain name.
3. The private key corresponding to that server certificate.

Following is an example of invoking this command and examining its
results.

```console
bash-5.2$ eval pieces=($(scripts/kubestellar-ensure-kcp-server-creds yep.yep))
Re-using PKI at /Users/mspreitz/go/src/github.com/kubestellar/kubestellar/pki
Re-using CA at /Users/mspreitz/go/src/github.com/kubestellar/kubestellar/pki/private/ca.key
Accepting existing credentials

bash-5.2$ echo ${pieces[0]}
/Users/mspreitz/go/src/github.com/kubestellar/kubestellar/pki/ca.crt

bash-5.2$ echo ${pieces[1]}
/Users/mspreitz/go/src/github.com/kubestellar/kubestellar/pki/issued/kcp-DNS-yep.yep.crt

bash-5.2$ echo ${pieces[2]}
/Users/mspreitz/go/src/github.com/kubestellar/kubestellar/pki/private/kcp-DNS-yep.yep.key
```

Following is an example of using those results in launching the kcp
server.  The `--tls-sni-cert-key` flag can appear multiple times on
the command line, configuring the server to respond in a different way
to each of multiple different SNIs.

```console
bash-5.2$ kcp start --tls-sni-cert-key ${pieces[1]},${pieces[2]} &> /tmp/kcp.log &
[1] 66974
```

### wait-and-switch-domain

This command is for using after the kcp server has been launched.
Since the `kcp start` command really means `kcp run`, all usage of
that server has to be done by concurrent processes.  The
`wait-and-switch-domain` command bundles two things: waiting for the
kcp server to start handling kubectl requests, and making an alternate
kubeconfig file for a set of clients to use.  This command is pointed
at an existing kubeconfig file and reads it but does not write it; the
alternate config file is written, with the same base name, into a
parallel directory (which this command `mkdir -p`s).  This command
takes exactly six command line positional arguments, as follows.

1. Pathname (absolute or relative) of the directory holding the input kubeconfig.
2. Filename of the input kubeconfig (just the base, no directory part).
3. Name of the kubeconfig "context" that identifies what to replace.
4. Domain name to put in the replacement server URLs.
5. Port number to put in the replacement server URLs.
6. Pathname (absolute or relative) of a file holding the CA
   certificate to put in the alternate kubeconfig file.

Creation of the alternate kubeconfig file starts by looking in the
input kubeconfig file for the "context" with the given name, to find
the name of a "cluster".  The server URL of that cluster is examined,
and its `protocol://host:port` prefix is extracted.  The alternate
kubeconfig will differ from the input kubeconfig in the contents of
the cluster objects whose server URLs start with that same prefix.
There will be the following two differences.

1. In the server URL's `protocol://host:port` prefix, the host will be
   replaced by the given domain name and the port will be replaced by
   the given port.
2. The cluster will be given a `certificate-authority-data` that holds
   the contents (base64 encoded, as usual) of the given CA certificate
   file.

This command does a `mkdir -p` on a directory name that is the input
kubeconfig directory name followed by a dash ("-") and the given
domain name.  This command writes the alternate kubeconfig file into
that directory.

Following is an example of using this command and examining the
results.  The context and port number chosen work for the kubeconfig
file that `kcp start` (kcp release v0.11.0) creates by default.

```console
bash-5.2$ scripts/wait-and-switch-domain .kcp admin.kubeconfig root yep.yep 6443 ${pieces[0]}

bash-5.2$ diff -w .kcp/admin.kubeconfig .kcp-yep.yep/admin.kubeconfig 
4,5c4,5
<     certificate-authority-data: LS0...LQo=
<     server: https://192.168.something.something:6443
---
>       certificate-authority-data: LS0...LQo=
>       server: https://yep.yep:6443
8,9c8,9
<     certificate-authority-data: LS0...LQo=
<     server: https://192.168.something.something:6443/clusters/root
---
>       certificate-authority-data: LS0...LQo=
>       server: https://yep.yep:6443/clusters/root
```

Following is an example of using the alternate kubeconfig file, in a
context where the domain name "yep.yep" resolves to an IP address of
the network namespace in which the kcp server is running.

```console
bash-5.2$ KUBECONFIG=.kcp-yep.yep/admin.kubeconfig kubectl ws .
Current workspace is "root".
```

Because this command reads the given kubeconfig file, it is important
to invoke this command while nothing is concurrently writing it and
while the caller reliably knows the name of a kubeconfig context that
identifies what to replace.

### switch-domain

This command is the second part of `wait-and-switch-domain`: the part
of creating the alternate kubeconfig file.  It has the same inputs and
outputs.

## KubeStellar Platform control

The `kubestellar` command has three subcommands, one to finish setup
and two for process control.

The usage synopsis is as follows.

``` { .bash .no-copy }
kubestellar [flags] subcommand [flags]
```

This command accepts the following command line flags, which can
appear before and/or after the subcommand.  The `--log-folder` flag is
only used in the `start` subcommand.

- `-V` or `--verbose`: calls for more verbose output.  This is a
  binary choice, not a matter of degree.
- `-X`: turns on echoing of script lines
- `--log-folder $pathname`: says where to put the logs from the
  controllers.  Will be `mkdir -p` if absent.  Defaults to
  `${PWD}/kubestellar-logs`.
- `-h` or `--help`: print a brief usage message and terminate.

### Kubestellar init

This subcommand is used after installation to finish setup and does
two things.  One is to ensure that the edge service provider workspace
(ESPW) exists and has the required contents.  The other is to ensure
that the `root:compute` workspace has been extended with the RBAC
objects that enable the syncer to propagate reported state for
downsynced objects defined by the APIExport from that workspace of a
subset of the Kubernetes API for managing containerized workloads.

### KubeStellar start

This subcommand is used after installation or process stops.

This subcommand stops any running kubestellar controllers and then
starts them all.  It also does the same thing as `kubestellar init`.

### KubeStellar stop

This subcommand undoes the primary function of `kubestellar start`,
stopping any running KubeStellar controllers.  It does _not_ tear down
the ESPW.

## KubeStellar-release

This command just echoes the [semantic version](https://semver.org/)
of the release used.  This command is only available in archives built
for a release.  Following is an example usage.

```shell
kubestellar-release
```
``` { .bash .no-copy }
v0.2.3-preview
```

## Kubestellar-version

This executable prints information about itself captured at build
time.  If built by `make` then this is information conveyed by the
Makefile; otherwise it is [the Kubernetes
defaults](https://github.com/kubernetes/client-go/blob/master/pkg/version/base.go).

It will either print one requested property or a JSON object
containing many.

```shell
kubestellar-version help
```
``` { .bash .no-copy }
Invalid component requested: "help"
Usage: kubestellar-version [buildDate|gitCommit|gitTreeState|platform]
```

```shell
kubestellar-version buildDate
```
``` { .bash .no-copy }
2023-05-19T02:54:01Z
```

```shell
kubestellar-version gitCommit
```
``` { .bash .no-copy }
1747254b
```

```shell
kubestellar-version          
```
``` { .bash .no-copy }
{"major":"1","minor":"24","gitVersion":"v1.24.3+kcp-v0.2.1-20-g1747254b880cb7","gitCommit":"1747254b","gitTreeState":"dirty","buildDate":"2023-05-19T02:54:01Z","goVersion":"go1.19.9","compiler":"gc","platform":"darwin/amd64"}
```

## Creating SyncTarget/Location pairs

In this PoC, the interface between infrastructure and workload
management is inventory API objects.  Specifically, for each edge
cluster there is a unique pair of SyncTarget and Location objects in a
so-called inventory management workspace.  These kinds of objects were
originally defined in kcp TMC, and now there is a copy of those
definitions in KubeStellar.  It is the definitions in KubeStellar that
should be referenced.  Those are in the Kubernetes API group
`edge.kcp.io`, and they are exported from the
[KubeStellar Core Space (KCS)](../../../../Getting-Started/user-guide/)) (the kcp workspace
named `root:espw`).

The following command helps with making that SyncTarget and Location
pair and adding the APIBinding to `root:espw:edge.kcp.io` if needed.

The usage synopsis is as follows.

```shell
kubectl kubestellar ensure location flag... objname labelname=labelvalue...
```

Here `objname` is the name for the SyncTarget object and also the name
for the Location object.  This command ensures that these objects exist
and have at least the given labels.

The flags can also appear anywhere later on the command line.

The acceptable flags include all those of `kubectl` except for
`--context`.  This command also accepts the following flags.

- `--imw workspace_path`: specifies which workspace to use as the
  inventory management workspace.  The default value is the current
  workspace.

The current workspaces does not matter if the IMW is explicitly
specified.  Upon completion, the current workspace will be your chosen
IMW.

This command does not depend on the action of any of the KubeStellar
controllers but does require that the KubeStellar Core Space (KCS) has been set up.

An example usage follows.

```shell
kubectl kubestellar ensure location --imw root:imw-1 demo1 foo=bar the-word=the-bird
```
``` { .bash .no-copy }
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

The creation of the APIBinding is equivalent to the following command
(in the same workspace).

```shell
kubectl create -f <<EOF
apiVersion: apis.kcp.io/v1alpha1
kind: APIBinding
metadata:
  name: edge.kcp.io
spec:
  reference:
    export:
      path: root:espw
      name: edge.kcp.io
EOF
```

This command operates in idempotent style, making whatever changes (if
any) are needed to move from the current state to the desired state.
Current limitation: it does not cast a skeptical eye on the spec of a
pre-existing Location or APIBinding object.

## Removing SyncTarget/Location pairs

The following script undoes whatever remains from a corresponding
usage of `kubectl kubestellar ensure location`.  It has all the same
command line syntax and semantics except that the
`labelname=labelvalue` pairs do not appear.

This command does not depend on the action of any of the KubeStellar controllers.

The following session demonstrates usage, including idempotency.

```shell
kubectl ws root:imw-1
```
``` { .bash .no-copy }
Current workspace is "root:imw-1".
```

```shell
kubectl kubestellar remove location demo1
```
``` { .bash .no-copy }
synctarget.workload.kcp.io "demo1" deleted
location.scheduling.kcp.io "demo1" deleted
```

```shell
kubectl kubestellar remove location demo1
```

## Syncer preparation and installation

The syncer runs in each edge cluster and also talks to the
corresponding mailbox workspace.  In order for it to be able to do
that, there is some work to do in the mailbox workspace to create a
ServiceAccount for the syncer to authenticate as and create RBAC
objects to give the syncer the privileges that it needs.  The
following script does those things and also outputs YAML to be used to
install the syncer in the edge cluster.

The usage synopsis is as follows.

```shell
kubectl kubestellar prep-for-syncer flag... synctarget_name
```

Here `synctarget_name` is the name of the `SyncTarget` object, in the
relevant IMW, corresponding to the relevant edge cluster.

The flags can also appear anywhere later on the command line.

The acceptable flags include all those of `kubectl` except for
`--context`.  This command also accepts the following flags.

- `--imw workspace_path`: specifies which workspace holds the relevant
  SyncTarget object.  The default value is the current workspace.
- `--espw workspace_path`: specifies where to find the edge service
  provider workspace.  The default is the standard location,
  `root:espw`.
- `--syncer-image image_ref`: specifies the container image that runs
  the syncer.  The default is `quay.io/kubestellar/syncer:{{ config.ks_tag }}`.
- `-o output_pathname`: specifies where to write the YAML definitions
  of the API objects to create in the edge cluster in order to deploy
  the syncer there.  The default is `synctarget_name +
  "-syncer.yaml"`.

The current workspaces does not matter if the IMW is explicitly
specified.  Upon completion, the current workspace will be what it was
when the command started.

This command will only succeed if the mailbox controller has created
and conditioned the mailbox workspace for the given SyncTarget.  This
command will wait for 10 to 70 seconds for that to happen.

An example usage follows.

```shell
kubectl kubestellar prep-for-syncer --imw root:imw-1 demo1
```
``` { .bash .no-copy }
Current workspace is "root:imw-1".
Current workspace is "root:espw"
Current workspace is "root:espw:4yqm57kx0m6mn76c-mb-406c54d1-64ce-4fdc-99b3-cef9c4fc5010" (type root:universal).
Creating service account "kubestellar-syncer-demo1-28at01r3"
Creating cluster role "kubestellar-syncer-demo1-28at01r3" to give service account "kubestellar-syncer-demo1-28at01r3"

 1. write and sync access to the synctarget "kubestellar-syncer-demo1-28at01r3"
 2. write access to apiresourceimports.

Creating or updating cluster role binding "kubestellar-syncer-demo1-28at01r3" to bind service account "kubestellar-syncer-demo1-28at01r3" to cluster role "kubestellar-syncer-demo1-28at01r3".

Wrote physical cluster manifest to demo1-syncer.yaml for namespace "kubestellar-syncer-demo1-28at01r3". Use

  KUBECONFIG=<workload-execution-cluster-config> kubectl apply -f "demo1-syncer.yaml"

to apply it. Use

  KUBECONFIG=<workload-execution-cluster-config> kubectl get deployment -n "kubestellar-syncer-demo1-28at01r3" kubestellar-syncer-demo1-28at01r3

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
ensure-location` and `kubectl kubestellar prep-for-syncer`, and takes
the union of their command line flags and arguments.  Upon completion,
the kcp current workspace will be what it was at the start.

An example usage follows.

```shell
kubectl kubestellar prep-for-cluster --imw root:imw-1 demo2 key1=val1
```
``` { .bash .no-copy }
Current workspace is "root:imw-1".
synctarget.workload.kcp.io/demo2 created
location.scheduling.kcp.io/demo2 created
synctarget.workload.kcp.io/demo2 labeled
location.scheduling.kcp.io/demo2 labeled
Current workspace is "root:imw-1".
Current workspace is "root:espw".
Current workspace is "root:espw:1cpf1cd4ydy13vo1-mb-3c354acd-ed86-45bb-a60d-cee8e59973f7" (type root:universal).
Creating service account "kubestellar-syncer-demo2-15nq4e94"
Creating cluster role "kubestellar-syncer-demo2-15nq4e94" to give service account "kubestellar-syncer-demo2-15nq4e94"

 1. write and sync access to the synctarget "kubestellar-syncer-demo2-15nq4e94"
 2. write access to apiresourceimports.

Creating or updating cluster role binding "kubestellar-syncer-demo2-15nq4e94" to bind service account "kubestellar-syncer-demo2-15nq4e94" to cluster role "kubestellar-syncer-demo2-15nq4e94".

Wrote physical cluster manifest to demo2-syncer.yaml for namespace "kubestellar-syncer-demo2-15nq4e94". Use

  KUBECONFIG=<workload-execution-cluster-config> kubectl apply -f "demo2-syncer.yaml"

to apply it. Use

  KUBECONFIG=<workload-execution-cluster-config> kubectl get deployment -n "kubestellar-syncer-demo2-15nq4e94" kubestellar-syncer-demo2-15nq4e94

to verify the syncer pod is running.
```

## Creating a Workload Management Workspace

Such a workspace needs not only to be created but also populated with
an `APIBinding` to the edge API and, if desired, an `APIBinding` to
the Kubernetes API for management of containerized workloads.

**NOTE**: currently, only a subset of the Kubernetes containerized
workload management API is supported.  In particular, only the
following object kinds are supported: `Deployment`, `Pod`, `Service`,
`Ingress`.  To be clear: this is in addition to the generic object
kinds that are supported; illustrative examples include RBAC objects,
`CustomResourceDefinition`, and `ConfigMap`.  For a full description,
see [the categorization in the design](../outline/#data-objects).

The usage synopsis for this command is as follows.

```shell
kubectl ws parent_pathname; kubectl kubestellar ensure wmw flag... wm_workspace_name
```

Here `parent_pathname` is the workspace pathname of the parent of the
WMW, and `wm_workspace_name` is the name (not pathname, just a bare
one-segment name) of the WMW to ensure.  Thus, the pathname of the WMW
will be `parent_pathname:wm_workspace_name`.

Upon completion, the WMW will be the current workspace.

The flags can also appear anywhere later on the command line.

The acceptable flags include all those of `kubectl` except for
`--context`.  This command also accepts the following flags.

- `--with-kube boolean`: specifies whether or not the WMW should
  include an APIBinding to the Kubernetes API for management of
  containerized workloads.

This script works in idempotent style, doing whatever work remains to
be done.

The following session shows some example usages, including
demonstration of idempotency and changing whether the kube APIBinding
is included.

```shell
kubectl ws .
```
``` { .bash .no-copy }
Current workspace is "root:my-org".
```

```shell
kubectl kubestellar ensure wmw example-wmw
```
``` { .bash .no-copy }
Current workspace is "root".
Current workspace is "root:my-org".
Workspace "example-wmw" (type root:universal) created. Waiting for it to be ready...
Workspace "example-wmw" (type root:universal) is ready to use.
Current workspace is "root:my-org:example-wmw" (type root:universal).
apibinding.apis.kcp.io/bind-espw created
apibinding.apis.kcp.io/bind-kube created
```

```shell
kubectl ws ..
```
``` { .bash .no-copy }
Current workspace is "root:my-org".
```

```shell
kubectl kubestellar ensure wmw example-wmw
```
``` { .bash .no-copy }
Current workspace is "root".
Current workspace is "root:my-org".
Current workspace is "root:my-org:example-wmw" (type root:universal).
```

```shell
kubectl ws ..
```
``` { .bash .no-copy }
Current workspace is "root:my-org".
```

```shell
kubectl kubestellar ensure wmw example-wmw --with-kube false
```
``` { .bash .no-copy }
Current workspace is "root".
Current workspace is "root:my-org".
Current workspace is "root:my-org:example-wmw" (type root:universal).
apibinding.apis.kcp.io "bind-kube" deleted 
```

## Removing a Workload Management Workspace

Deleting a WMW can be done by simply deleting its `Workspace` object from
the parent.

```shell
kubectl ws .
```
``` { .bash .no-copy }
Current workspace is "root:my-org:example-wmw".
```

```shell
kubectl ws ..
```
``` { .bash .no-copy }
Current workspace is "root:my-org".
```

```shell
kubectl delete Workspace example-wmw
```
``` { .bash .no-copy }
workspace.tenancy.kcp.io "example-wmw" deleted 
```

Alternatively, you can use the following command line whose design
completes the square here.  Invoke it when the current workspace is
the parent of the workload management workspace to delete.

```shell
kubectl kubestellar remove wmw -h
```
``` { .bash .no-copy }
Usage: kubectl ws parent; kubectl kubestellar remove wmw kubectl_flag... wm_workspace_name
```

```shell
kubectl ws root:my-org
```
``` { .bash .no-copy }
Current workspace is "root:my-org".
```

```shell
kubectl kubestellar remove wmw demo1
```
``` { .bash .no-copy }
workspace.tenancy.kcp.io "demo1" deleted
```

```shell
kubectl ws .
```
``` { .bash .no-copy }
Current workspace is "root:my-org".
```

```shell
kubectl kubestellar remove wmw demo1
```

# Web-to-bash

## Quick Setup

This is a combination of some installation and setup steps, for use in
[the
QuickStart](../../../Getting-Started/quickstart/).

The script can be read directly from
{{ config.repo_raw_url }}/{{ config.ks_branch }}/bootstrap/bootstrap-kubestellar.sh
and does the following things.

1. Downloads and installs kcp if it is not already evident on `$PATH`
   (using [the script below](#install-kcp-and-its-kubectl-plugins).
2. Starts a kcp server if one is not already running.
3. Downloads and installs kubestellar if it is not already evident on
   `$PATH` (using [the script below](#install-kubestellar).
4. `kubestellar start` if the KubeStellar controllers are not already
   running or the ESPW does not (yet) exist.

This script accepts the following command line flags; all are optional.

- `--kubestellar-version $version`: specifies the release of
  KubeStellar to use.  When using a specific version, include the
  leading "v".  The default is the latest regular release, and the
  value "latest" means the same thing.
- `--kcp-version $version`: specifies the kcp release to use.  The
  default is the one that works with the chosen release of
  KubeStellar.
- `--os $OS`: specifies the operating system to use in selecting the
  executables to download and install.  Choices are `linux` and
  `darwin`.  Auto-detected if omitted.
- `--arch $IAS`: specifies the instruction set architecture to use in
  selecting the executables to download and install.  Choices are
  `amd64` and `arm64`.  Auto-detected if omitted.
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
- `-V` or `--verbose`: increases the verbosity of output.  This is a
  binary thing, not a matter of degree.
- `-X`: makes the script `set -x` internally, for debugging.
- `-h` or `--help`: print brief usage message and exit.

Here "install" means only to (a) unpack the distribution archives into
the relevant places under `$install_parent_dir` and (b) enhance the
`PATH`, and `KUBECONFIG` in the case of kcp, environment variables in
the shell running the script.  Of course, if you run the script in a
sub-shell then those environment effects terminate with that
sub-shell; this script also prints out messages showing how to update
the environment in another shell.

## Install kcp and its kubectl plugins

This script is directly available at [{{ config.repo_url }}/blob/{{ config.ks_branch }}/bootstrap/install-kubestellar.sh]({{ config.repo_url }}/blob/{{ config.ks_branch }}/bootstrap/install-kcp-with-plugins.sh) and does the following things.

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
- `-V` or `--verbose`: increases the verbosity of output.  This is a
  binary thing, not a matter of degree.
- `-X`: makes the script `set -x` internally, for debugging.
- `-h` or `--help`: print brief usage message and exit.

Here install means only to unpack the downloaded archives, creating
`$install_parent_dir/bin`.  If `$install_parent_dir/bin` is not
already on your `$PATH` then this script will print out a message
telling you to add it.

## Install KubeStellar

This script is directly available at
[{{ config.repo_url }}/blob/{{ config.ks_branch }}/bootstrap/install-kubestellar.sh]({{ config.repo_url }}/blob/{{ config.ks_branch }}/bootstrap/install-kubestellar.sh)
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
- `-V` or `--verbose`: increases the verbosity of output.  This is a
  binary thing, not a matter of degree.
- `-X`: makes the script `set -x` internally, for debugging.
- `-h` or `--help`: print brief usage message and exit.

Here install means only to unpack the downloaded archive, creating
`$install_parent_dir/bin`.  If `$install_parent_dir/bin` is not
already on your `$PATH` then this script will print out a message
telling you to add it.
