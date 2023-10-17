---
short_name: commands
---

This PoC includes two sorts of commands for users to use. Most are
executables designed to be accessed in the usual ways: by appearing in
a directory on the user's `$PATH` or by the user writing a pathname
for the executable on the command line. The one exception is [the
bootstrap script](#bootstrap), which is designed to be fetched from
github and fed directly into `bash`.

There are two ways of deploying kcp and the central KubeStellar
components: (1) as processes on a machine of supported OS and ISA, and
(2) as workload in a Kubernetes (possibly OpenShift) cluster.  The
latter is newer and not yet well exercised.

This document describes both commands that a platform administrator
uses and commands that a platform user uses.

Most of these executables require that the kcp server be running and
that `kubectl` is configured to use the kubeconfig file produced by
`kcp start`. That is: either `$KUBECONFIG` holds the pathname of that
kubeconfig file, its contents appear at `~/.kube/config`, or
`--kubeconfig $pathname` appears on the command line.  The exceptions
are the bootstrap script, [the kcp control script that runs before
starting the kcp server](#kubestellar-ensure-kcp-server-creds), and
the admin commands for deploying KubeStellar in a Kubernetes cluster.

**NOTE**: most of the kubectl plugin usages described here certainly
or potentially change the setting of which kcp workspace is "current"
in your chosen kubeconfig file; for this reason, they are not suitable
for executing concurrently with anything that depends on that setting
in that file.  The exceptions are the admin commands for deploying
KubeStellar in a Kubernetes cluster.

## Bare process deployment

### kcp process control

KubeStellar has some commands to support situations in which some
clients of the kcp server need to connect to it by opening a
connection to a DNS domain name rather than an IP address that the
server can bind its socket to.  The key idea is using the
`--tls-sni-cert-key` command line flag of `kcp start` to configure the
server to respond with a bespoke server certificate in TLS handshakes
in which the client addresses the server by a given domain name.

These commands are used separately from starting the kcp server and
are designed so that they can be used multiple times if there are
multiple sets of clients that need to use a distinct domain name.
Starting the kcp server is not described here, beyond the particular
consideration needed for the `--tls-sni-cert-key` flag, because it is
otherwise ordinary.

#### kubestellar-ensure-kcp-server-creds

##### kubestellar-ensure-kcp-server-creds pre-reqs

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

##### kubestellar-ensure-kcp-server-creds usage

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
bash-5.2$ eval pieces=($(kubestellar-ensure-kcp-server-creds yep.yep))
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

#### wait-and-switch-domain

This command is for using after the kcp server has been launched.
Since the `kcp start` command really means `kcp run`, all usage of
that server has to be done by concurrent processes.  The
`wait-and-switch-domain` command bundles two things: waiting for the
kcp server to start handling kubectl requests, and making an alternate
kubeconfig file for a set of clients to use.  This command is pointed
at an existing kubeconfig file and reads it but does not write it; the
alternate config file is written (its directory must already exist and
be writable).  This command takes exactly six command line positional
arguments, as follows.

1. Pathname (absolute or relative) of the input kubeconfig.
2. Pathname (absolute or relative) of the output kubeconfig.
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

Following is an example of using this command and examining the
results.  The context and port number chosen work for the kubeconfig
file that `kcp start` (kcp release v0.11.0) creates by default.

```console
bash-5.2$ wait-and-switch-domain .kcp/admin.kubeconfig test.yaml root yep.yep 6443 ${pieces[0]}

bash-5.2$ diff -w .kcp/admin.kubeconfig test.yaml
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

#### switch-domain

This command is the second part of `wait-and-switch-domain`: the part
of creating the alternate kubeconfig file.  It has the same inputs and
outputs and concurrency considerations.

### KubeStellar process control

The `kubestellar` command has three subcommands, one to finish setup
and two for process control.

The usage synopsis is as follows.

``` { .bash .no-copy }
kubestellar [flags] subcommand [flags]
```

This command accepts the following command line flags, which can
appear before and/or after the subcommand.  The `--log-folder` flag is
only meaningful for the `start` subcommand. The `--local-kcp` flag is
not meaningful for the `stop` subcommand. The `--ensure-imw` and 
`--ensure-wmw` flags are only meaningful for the `start` or `init` subcommands.

- `-V` or `--verbose`: calls for more verbose output.  This is a
  binary choice, not a matter of degree.
- `-X`: turns on echoing of script lines
- `--log-folder $pathname`: says where to put the logs from the
  controllers.  Will be `mkdir -p` if absent.  Defaults to
  `${PWD}/kubestellar-logs`.
- `--local-kcp $bool`: says whether to expect to find a local process
  named "kcp".  Defaults to "true".
- `--ensure-imw`: provide a comma separated list of pathnames for inventory workspaces, _e.g._ "root:imw1,root:imw2". Defaults to "root:imw1". To prevent the creation of any inventory workspace, then pass "".
- `--ensure-wmw`: provide a comma separated list of pathnames for workload management workspaces, _e.g._ "root:wmw1,root:imw2". Defaults to "root:wmw1". To prevent the creation of any workload management workspace, then pass "".
- `-h` or `--help`: print a brief usage message and terminate.

#### Kubestellar init

This subcommand is used after installation to finish setup and does
the following five things.

1. Waits for the kcp server to be in service and the `root:compute`
   workspace to be usable.

2. Ensure that the edge service provider workspace (ESPW) exists and
has the required contents.

3. Ensure that the `root:compute` workspace has been extended with the
RBAC objects that enable the syncer to propagate reported state for
downsynced objects defined by the APIExport from that workspace of a
subset of the Kubernetes API for managing containerized workloads.

4. Ensure the existence of zero, one, or more inventory management workspaces
depending on the value of `--ensure-imw` flag. Default is one inventory management
workspaces at pathname "root:imw1".

5. Ensure the existence of zero, one, or more workload management workspaces 
depending on the value of `--ensure-wmw` flag. Default is one workload management
workspaces at pathname "root:wmw1". The workload management workspaceshace APIBindings
that import the namespaced Kubernetes resources (kinds of objects) for management of
containerized workloads. At the completion of `kubestellar init` the current workspace will be
"root".

#### KubeStellar start

This subcommand is used after installation or process stops.

This subcommand stops any running kubestellar controllers and then
starts them all.  It also does the same thing as `kubestellar init`.

#### KubeStellar stop

This subcommand undoes the primary function of `kubestellar start`,
stopping any running KubeStellar controllers.  It does _not_ tear down
the ESPW.

## Deployment into a Kubernetes cluster

These commands administer a deployment of the central components ---
the kcp server, PKI, and the central KubeStellar components --- in a
Kubernetes cluster that will be referred to as "the hosting cluster".
These commands are framed as "kubectl plugins" and thus need to be
explicitly or implicitly given a kubeconfig file for the hosting
cluster.

You need a Kubernetes cluster with an Ingress controller deployed and
configured in a way that does _not_ terminate TLS connections (this
abstinence is often called "SSL passthrough"). An OpenShift cluster
would be one qualifying thing. Another would be an ordinary Kubernetes
cluster with the [nginx Ingress
controller](https://docs.nginx.com/nginx-ingress-controller/) deployed
and configured appropriately. Please note that [special
considerations](https://kind.sigs.k8s.io/docs/user/ingress/) apply
when deploying an ingress controller in `kind`. See [a fully worked
example with kind and
nginx](../environments/dev-env/#hosting-kubestellar-in-a-kind-cluster). You
will need to know the port number at which the Ingress controller is
listening for HTTPS connections.

**IF** your Kubernetes cluster has any worker nodes --- real or
virtual --- with the x86_64 instruction set, they need to support the
extended instruction set known as "x64-64-v2". If using hardware
bought in the last 10 years, you can assume that is true. If using
emulation, you need to make sure that your emulator is emulating that
extended instruction set --- some emulators do not do this by
default. See [QEMU configuration
recommendations](https://www.qemu.org/docs/master/system/i386/cpu.html),
for example.

You will need a domain name that, on each of your clients, resolves to
an IP address that gets to the Ingress controller's listening socket.

### Deploy to cluster

Deployment is done with a "kubectl plugin" that is invoked as `kubectl
kubestellar deploy` and creates a [Helm
"release"](https://helm.sh/docs/intro/using_helm/#three-big-concepts)
in the hosting cluster. As such, it relies on explicit (on the command
line) or implicit (in environment variables and/or `~/.kube/config`)
configuration needed to execute Helm commands. The following flags can
appear on the command line, in any order.

- `--openshift $bool`, saying whether the hosting cluster is an
  OpenShift cluster.  If so then a Route will be created to the kcp
  server; otherwise, an Ingress object will direct incoming TLS
  connections to the kcp server.  The default is `false`.
- `--external-endpoint $domain_name:$port`, saying how the kcp server
  will be reached from outside the cluster.  The given domain name
  must be something that the external clients will resolve to an IP
  address where the cluster's Ingress controller or OpenShift router
  will be listening, and the given port must be the corresponding TCP
  port number.  For a plain Kubernetes cluster, this must be
  specified.  For an OpenShift cluster this may be omitted, in which
  case the command will (a) assume that the external port number is
  443 and (b) extract the external hostname from the Route object
  after it is created and updated by OpenShift.  FYI, that external
  hostname will start with a string derived from the Route in the
  chart (currently "kubestellar-route-kubestellar") and continue with
  "." and then the ingress domain name for the cluster.
- a command line flag for the `helm upgrade` command. This includes
  the usual control over namespace: you can set it on the command
  line, otherwise the namespace that is current in your kubeconfig
  applies.
- `-X` turns on debug echoing of all the commands in the script that
  implements this command.
- `-h` prints a brief usage message and terminates with success.

For example, to deploy to a plain Kubernetes cluster whose Ingress
controller can be reached at
`my-long-application-name.my-region.some.cloud.com:1234`, you would
issue the following command.

```shell
kubectl kubestellar deploy --external-endpoint my-long-application-name.my-region.some.cloud.com:1234
```

The Helm chart takes care of setting up the KubeStellar MCCM core,
accomplishing the same thing as the [kubestellar
start](#kubestellar-start) command above.

### Fetch kubeconfig for internal clients

To fetch a kubeconfig for use by clients inside the hosting cluster,
use the `kubectl kubestellar get-internal-kubeconfig` command.  It
takes the following on the command line.

- `-o $output_pathname`, saying where to write the kubeconfig. This
  must appear exactly once on the command line.
- a `kubectl` command line flag, for accessing the hosting
  cluster. This includes the usual control over namespace.
- `-X` turns on debug echoing of all the commands in the script that
  implements this command.
- `-h` prints a brief usage message and terminates with success.

### Fetch kubeconfig for external clients

To fetch a kubeconfig for use by clients outside of the hosting
cluster --- those that will reach the kcp server via the external
endpoint specified in the deployment command --- use the `kubectl
kubestellar get-external-kubeconfig` command.  It takes the following
on the command line.

- `-o $output_pathname`, saying where to write the kubeconfig. This
  must appear exactly once on the command line.
- a `kubectl` command line flag, for accessing the hosting cluster.
  This includes the usual control over namespace.
- `-X` turns on debug echoing of all the commands in the script that
  implements this command.
- `-h` prints a brief usage message and terminates with success.

### Fetch a log from a KubeStellar runtime container

{%
   include-markdown "../../../../core-helm-chart/README.md"
   start="<!--check-log-start-->"
   end="<!--check-log-end-->"
%}

### Remove deployment to a Kubernetes cluster

The deployment of kcp and KubeStellar as a Kubernetes workload is done
by making a "release" (this is a [technical term in
Helm](https://helm.sh/docs/intro/cheatsheet/#basic-interpretationscontext)
whose meaning might surprise you) of a Helm chart.  This release is
named "kubestellar".  To undo the deployment, just use the Helm
command to delete that release.

```shell
helm delete kubestellar
```

**NOTE**: this is a detail that might change in the future.

## KubeStellar platform user commands

The remainder of the commands in this document are for users rather
than administrators of the service that KubeStellar provides.

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
`edge.kubestellar.io`, and they are exported from the
[KubeStellar Core Space (KCS)](../../../../Getting-Started/user-guide/)) (the kcp workspace
named `root:espw`).

The following command helps with making that SyncTarget and Location
pair and adding the APIBinding to `root:espw:edge.kubestellar.io` if needed.

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
- `-X`: turn on debug echoing of the commands inside the script that
  implements this command.

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
  name: edge.kubestellar.io
spec:
  reference:
    export:
      path: root:espw
      name: edge.kubestellar.io
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
- `-s`: exceptionally low info output.
- `-X`: turn on debug echoing of commands inside the script that
  implements this command.

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
- `-X`: turn on debug echoing of the commands inside the script that
  implements this command.

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
Usage: kubectl ws parent; kubectl kubestellar remove wmw [-X] kubectl_flag... wm_workspace_name
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

## Bootstrap

This is a combination of some installation and setup steps, for use in
[the
QuickStart](../../../Getting-Started/quickstart/).

The script can be read directly from
{{ config.repo_raw_url }}/{{ config.ks_branch }}/bootstrap/bootstrap-kubestellar.sh
and does the following things.

1. Downloads and installs kcp executables if they are not already
   evident on `$PATH`.
2. Downloads and installs kubestellar executables if they are not
   already evident on `$PATH`.
3. Ensures that kcp and KubeStellar are deployed (i.e., their
   processes are running and their initial configurations have been
   established) either as bare processes or as workload in a
   pre-existing Kubernetes cluster.

This script accepts the following command line flags; all are optional.  The `--os`, `--arch`, and `--bind-address` flags are only useful when deploying as bare processes.  The deployment will be into a Kubernetes cluster if either `--external-endpoint` or `--openshift true` is given.

- `--kubestellar-version $version`: specifies the release of
  KubeStellar to use.  When using a specific version, include the
  leading "v".  The default is the latest regular release, and the
  value "latest" means the same thing.
- `--kcp-version $version`: specifies the kcp release to use.  The
  default is the one that works with the chosen release of
  KubeStellar.
- `--openshift $bool`: specifies whether to the hosting cluster is an
  OpenShift cluster. The default value is `false`.
- `--endpoint-address $domain_name:$port`: specifies where an Ingress
  controller or OpenShift router is listening for incoming TLS
  connections from external (to the hosting cluster) clients of kcp
  and KubeStellar.
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
- `--host-ns $namespace_in_hosting_cluster`: specifies the namespace
  in the hosting cluster where the core will be deployed. Defaults to
  "kubestellar".
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
