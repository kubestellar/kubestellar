> ⚠️ **Important notice**
>
> Release **v0.9.2** of Kubeflex is **broken** due to a known image tag mismatch issue.
> Please **do not use this version**.

# User's Guide

## Breaking changes

### v0.9.0

Kubeflex configuration is stored within Kubeconfig file. Prior this version, `kflex` put its configuration under

```yaml
preferences:
  extensions:
  - extension:
      data:
        kflex-initial-ctx-name: kind-kubeflex  # indicates to kflex the hosting cluster context
      metadata:
        creationTimestamp: null
        name: kflex-config-extension-name
    name: kflex-config-extension-name          # indicates to kflex that this extension belongs to it
```

Starting `v0.9.0`, configuration remains within kubeconfig file but leverages `extensions:` and context-scope extension. For instance, the previous example would be translated as follow:

```yaml
extensions:                                          # change -> no longer preferences. See https://kubernetes.io/docs/reference/config-api/kubeconfig.v1/#Config
  - extension:
      data:
        hosting-cluster-ctx-name: kind-kubeflex      # change -> key name "hosting-cluster-ctx-name"
      metadata:
        creationTimestamp: 2025-06-09T22:36:17+02:00 # creation timestamp using ISO 8601 seconds
        name: kubeflex                               # same as below
    name: kubeflex                                   # change -> new extension name "kubeflex"
# ...
# Find the context corresponding to "hosting-cluster-ctx-name" (here "kind-kubeflex")
contexts:
  - context:
      cluster: kind-kubeflex
      user: kind-kubeflex
      # add extensions below and its information
      extensions:
      - extension:
        data:
          is-hosting-cluster-ctx: true                 # change -> key name "is-hosting-cluster-ctx" with "true"
        metadata:
          creationTimestamp: 2025-06-09T22:36:17+02:00 # creation timestamp using ISO 8601 seconds
          name: kubeflex                               # same as below
      name: kubeflex                                   # change -> new extension name "kubeflex"
    name: kind-kubeflex
  - context:
      cluster: mysupercp-cluster
      extensions:
      - extension:
          data:
            controlplane-name: mysupercp # change -> control plane name is saved under extension
          metadata:
            creationTimestamp: "2025-06-27T06:07:03Z"
            name: kubeflex
        name: kubeflex
      namespace: default
      user: mysupercp-admin
    name: mysupercp
  current-context: mysupercp
```

Proceed to change the kubeconfig file to match `v0.9.0`, as follow:

1. Set new hosting cluster context name running:
```bash
kflex config set-hosting $ctx_name
```
where `$ctx_name` represents the desired hosting context name

2. Delete `preferences:` related to **kubeflex** by editing your kubeconfig file manually.

At the moment, the change must be done manually until issue [#389](https://github.com/kubestellar/kubeflex/issues/389) is implemented.

## Installation

[kind](https://kind.sigs.k8s.io) and [kubectl](https://kubernetes.io/docs/tasks/tools/) are
required. Note that we plan to add support for other Kube distros. A hosting kind cluster
is created automatically by the kubeflex CLI.

Download the latest kubeflex CLI binary release for your OS/Architecture from the
[release page](https://github.com/kubestellar/kubeflex/releases) and copy it
to `/usr/local/bin` or another location in your `$PATH`. For example, on linux amd64:

```shell
OS_ARCH=linux_amd64
LATEST_RELEASE_URL=$(curl -H "Accept: application/vnd.github.v3+json"   https://api.github.com/repos/kubestellar/kubeflex/releases/latest   | jq -r '.assets[] | select(.name | test("'${OS_ARCH}'")) | .browser_download_url')
curl -LO $LATEST_RELEASE_URL
tar xzvf $(basename $LATEST_RELEASE_URL)
sudo install -o root -g root -m 0755 bin/kflex /usr/local/bin/kflex
```

Alternatively use the the single command below that will automatically detect the host OS type and architecture:

```shell
sudo su <<EOF
bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubeflex/main/scripts/install-kubeflex.sh) --ensure-folder /usr/local/bin --strip-bin
EOF
```

If you have [Homebrew](https://brew.sh), use the following commands to install kubeflex:

```shell
brew tap kubestellar/kubeflex https://github.com/kubestellar/kubeflex
brew install kflex
```

## Starting Kubeflex

Once the CLI is installed, the following CLI command creates a kind cluster and installs
the KubeFlex operator:

```shell
kflex init --create-kind
```

**NOTE**: After (a) doing `kflex init`, with or without
  `--create-kind`, or (b) installing KubeFlex with a Helm chart, you
  **MUST NOT** change your kubeconfig current context by any other
  means before using `kflex create`.

## Install KubeFlex on an existing cluster

You can install KubeFlex on an existing cluster with nginx ingress configured for SSL passthru,
or on a OpenShift cluster. At this time, we have only tested this option with Kind, k3d and OpenShift.

### Installing on kind

To create a kind cluster with nginx ingress, follow the instructions [here](https://kind.sigs.k8s.io/docs/user/ingress/).
Once you have your ingress running, you will need to configure nginx ingress for SSL passthru. Run the command:

```shell
kubectl edit deployment ingress-nginx-controller -n ingress-nginx
```

and add `--enable-ssl-passthrough` to the list of args for the container named `controller`. Then you can
run the command to install KubeFlex:

```shell
kflex init
```

### Installing on k3d

These steps have only been tested with k3d v5.6.0. Create a k3d cluster with `traefik` disabled and nginx ingress as follows:

```shell
k3d cluster create -p "9443:443@loadbalancer" --k3s-arg "--disable=traefik@server:*" kubeflex
helm install ingress-nginx ingress-nginx --set "controller.extraArgs.enable-ssl-passthrough=true" --repo https://kubernetes.github.io/ingress-nginx --version 4.6.1 --namespace ingress-nginx --create-namespace
```


```shell
kflex init --host-container-name k3d-kubeflex-server-0
```

### Installing on OpenShift

If you are installing on an OpenShift cluster you do not need any special configuration. Just run
the command:

```shell
kflex init
```

## Installing KubeFlex with Helm

To install KubeFlex on a cluster that already has nginx ingress with SSL passthru enabled,
you can use Helm instead of the KubeFlex CLI. Install KubeFlex with the following commands:

```shell
helm upgrade --install kubeflex-operator oci://ghcr.io/kubestellar/kubeflex/chart/kubeflex-operator \
--version <latest-release-version-tag> \
--set domain=localtest.me \
--set externalPort=9443
```

The `kubeflex-system` namespace is required for installing and running KubeFlex.
If it does not already exists, the Helm chart will create one.
Do not use any other namespace for this purpose.

### Installing KubeFlex with Helm on OpenShift

If you are installing on OpenShift with the `kflex` CLI, the CLI auto-detects OpenShift and autoimatically
configure the installation of the shared DB and the operator, but if you are using directly Helm to install
you will need to add additional parameters:

To install KubeFlex on OpenShift using Helm use the following commands:

```shell
helm upgrade --install kubeflex-operator oci://ghcr.io/kubestellar/kubeflex/chart/kubeflex-operator \
--version <latest-release-version-tag> \
--set isOpenShift=true
```

## Upgrading Kubeflex

The KubeFlex CLI can be upgraded with `brew upgrade kflex` (for brew installs). For linux
systems, manually download and update the binary. To upgrade the KubeFlex controller, just
upgrade the Helm chart according to the instructions for [kubernetes](#installing-kubeflex-with-helm)
or for [OpenShift](#installing-kubeflex-with-helm-on-openshift).

Note that for a kind test/dev installation, the simplest approach to get a fresh install
after updating the 'kflex' binary is to use `kind delete --name kubeflex` and re-running
`kflex init --create-kind`.

## Use a different DNS service

To use a different domain for DNS resolution, you can specify the `--domain` option when
you run `kflex init`. This domain should point to the IP address of your ingress controller,
which handles the routing of requests to different control plane instances based on the hostname.
A wildcard DNS service is recommended, so that any subdomain of your domain (such as *.\<domain\>)
will resolve to the same IP address. The default domain in KubeFlex is localtest.me, which is a
wildcard DNS service that always resolves to 127.0.0.1.
For example, `cp1.localtest.me` and `cp2.localtest.me` will both resolve to your local machine.
Note that this option is ignored if you are installing on OpenShift.

## Creating a new control plane

You can create a new control plane using the KubeFlex CLI (`kflex`) or using any Kubernetes client or `kubectl`.

**NOTE**: A pre-condition of using `kflex` to create a new control
plane is that either (a) your kubeconfig current-context is the one
used to access the hosting cluster or (b) the name of that context has
been stored in an extension in your kubeconfig file (see
[below](#hosting-context)). When this precondition is not met, the
failure will look like the following.

    $ kflex create cp1
    ✔ Checking for saved hosting cluster context...
    ◐ Creating new control plane cp1 of type k8s ...Error creating instance: no matches for kind "ControlPlane" in version "tenancy.kflex.kubestellar.org/v1alpha1"

To create a new control plane with name `cp1` using the KubeFlex CLI:

```shell
kflex create cp1
```

The KubeFlex CLI applies a `ControlPlane` CR, then waits for the control plane to become available
and finally it retrieves the `Kubeconfig` file for the new control plane, merges it with the current
Kubeconfig and sets the current context to the new control plane context.

At this point you may interact with the new control plane using `kubectl`, for example:

```shell
kubectl get ns
kubectl create ns myns
```
to switch the context back to the hosting cluster context, you may use the `ctx` command:

```shell
kflex ctx
```

That command requires your kubeconfig file to hold an extension that `kflex init` created to hold the name of the hosting cluster context. See [below](#hosting-context) for more information.

To update or refresh outdated or corrupted context information for a control plane stored in
the kubeconfig file, you can forcefully reload and overwrite the existing context data from
the KubeFlex hosting cluster. This can be accomplished by using the `--overwrite-existing-context`
flag. Here is an example:

```shell
kflex ctx cp1 --overwrite-existing-context
```

To switch back to a control plane context, use the
`ctx <control plane name>` command, e.g:

```shell
kflex ctx cp1
```

If there is not currently a kubeconfig context named for that control plane then that command requires your kubeconfig file to hold an extension that `kflex init` created to hold the name of the hosting cluster context. See [below](#hosting-context) for more information.


The same result can be accomplished with kubectl by using the `ControlPlane`` CR, for example:


```shell
kubectl apply -f - <<EOF
apiVersion: tenancy.kflex.kubestellar.org/v1alpha1
kind: ControlPlane
metadata:
  name: cp1
spec:
  backend: shared
  type: k8s
EOF
```

After applying the CR to the hosting kind cluster, you may check the status according to the usual
kubernetes conventions:

```shell
$ kubectl get controlplanes
NAME   SYNCED   READY   AGE
cp1    True     True    5d18h
```

In the above example, `SYNCED=True` means that all resources required to run the control plane
have been succssfully applied on the hosting cluster, and `READY=True` means that the Kube APIServer
for the control plane is available. You may also use `kubectl describe` to get more info about the
control plane.

To delete a control plane, you just have to delete the CR for that control plane, for example
using `kubectl delete controlplane cp1`. However, if you created the control plane with the `kflex`
CLI it would be better to use the `kflex` CLI so that it will remove the Kubeconfig for the control plane
from the current Kubeconfig and switch the context back to the context for the hosting cluster.

To delete a control plane with the kubeflex CLI use the command:

```shell
kubectl delete <control-plane-name>
```

If you are not using the kflex CLI to create the control plane and require access to the control plane,
you may retrieve the secret containing the control plane Kubeconfig, which is hosted in the control
plane hosting namespace (by convention `\<control-plane-name\>-system`) and is named `admin-kubeconfig`.

For example, the following commands retrieves the Kubeconfig for the control plane `cp1`:

```shell
NAMESPACE=cp1-system
kubectl get secrets -n ${NAMESPACE} admin-kubeconfig -o jsonpath='{.data.kubeconfig}' | base64 -d
```

### Accessing the control plane from within a kind cluster

For control plane of type k8s, the Kube API client can only use the 127.0.0.1 address. The DNS name
`\<control-plane-name\>.localtest.me`` is convenient for local test and dev but always resolves to 127.0.0.1, that does not work in a container. For accessing the control plane from within the KubeFlex hosting
cluster, you may use the controller manager Kubeconfig, which is maintained in the secret with name
`cm-kubeconfig` in the namespace hosting the control plane, or you may use the Kubeconfig in the
`admin-kubeconfig` secret with the address for the server `https://\<control-plane-name\>.\<control-plane-namespace\>:9443`.

To access the control plane API server from another kind cluster on the same docker network, you
can find the value of the nodeport for the service exposing the control plane API service, and construct
the URL for the server as `https://kubeflex-control-plane:\<nodeport\>`


## Control Plane Types

At this time KubFlex supports the following control plane types:

- k8s: this is the stock Kube API server with a subset of controllers running in the controller manager.
- ocm: this is the [Open Cluster Management Multicluster Control Plane](https://github.com/open-cluster-management-io/multicluster-controlplane), which provides a basic set of capabilities such as
clusters registration and support for the [`ManifestWork` API](https://open-cluster-management.io/concepts/manifestwork/).
- vcluster: this is based on the [vcluster project](https://www.vcluster.com) and provides the ability to create pods in the hosting namespace of the hosting cluster.
- host: this control plane type exposes the underlying hosting cluster with the same control plane abstraction
used by the other control plane types.
- external: this control plane type represents an existing cluster that was not created by KubeFlex and is not the KubeFlex hosting cluster.

## Control Plane Backends

KubeFlex roadmap aims to provide different types of backends: shared, dedicated, and for
each type the ability to choose if etcd or sql. At this time only the following
combinations are supported based on control plane type:

- k8s: shared postgresql
- ocm: dedicated etcd
- vcluster: dedicated sqlite

## Creating with a selected control plane type

If you are using the kflex CLI, you can use the flag `--type` or `-t` to select a particular
control plane type. If this flag is not specified, the default `k8s` is used.

To create a control plane of type `vcluster` run the command:

```shell
kflex create cp2 --type vcluster
```

To create a control plane of type `ocm` run the command:

```shell
kflex create cp3 --type ocm
```

To create a control plane of type `host` run the command:

```shell
kflex create cp4 --type host
```

To create a control plane of type `external` with the required options, run the command:

```shell
kflex adopt --adopted-context <kubeconfig-context-of-external-cluster> cp5
```

*Important*: This command generates a secret containing a long-lived token for accessing
the external cluster within the namespace associated with the control plane. The secret is automatically
removed when the associated control plane is deleted.

### Creating a control plane of type `external` with the API

To create a control plane of type `external` with the API, you need to provide
first a **bootstrap secret** containing a bootstrap Kubeconfig for accessing the external cluster.
The bootstrap Kubeconfig is used by the KubeFlex controllers to generate a long-lived
token for accessing the external cluster.  The bootstrap kubeconfig is required to have only one context,
so given a Kubeconfig for the external cluster `$EXTERNAL_KUBECONFIG` with context for the external
cluster `$EXTERNAL_CONTEXT` you can generate the `$BOOTSTRAP_KUBECONFIG` with the command:

```shell
kubectl --kubeconfig=$EXTERNAL_KUBECONFIG config view --minify --flatten \
--context=$EXTERNAL_CONTEXT > $BOOTSTRAP_KUBECONFIG
```

If the Kubeconfig for your external cluster uses a loopback address for the server URL, you
need to follow these [steps](#determining-the-endpoint-for-an-external-cluster-using-loopback-address)
to determine the address to use for `cluster.server` in the Kubeconfig and set that value in
the file referenced by`$BOOTSTRAP_KUBECONFIG` created in the previous step. If the address is the value of `$INTERNAL_ADDRESS` then you can update the bootstrap Kubeconfig as follows:

```shell
# e.g. INTERNAL_ADDRESS=https://ext1-control-plane:6443
kubectl --kubeconfig=$BOOTSTRAP_KUBECONFIG config set-cluster $(kubectl --kubeconfig=$BOOTSTRAP_KUBECONFIG config current-context) --server=$INTERNAL_ADDRESS
```

At this point, you can create the bootstrap secret with the command:

```shell
CP_NAME=ext1
kubectl create secret generic ${CP_NAME}-bootstrap --from-file=kubeconfig-incluster=$BOOTSTRAP_KUBECONFIG --namespace kubeflex-system
```
where `${CP_NAME}` is the name of the control plane to create.

*Important*: once the KubeFlex controller generates a long-lived token, it removes the bootstrap secret.

Finally, you can create the new control plane of type "external" applying the following yaml:

```shell
kubectl apply -f - <<EOF
apiVersion: tenancy.kflex.kubestellar.org/v1alpha1
kind: ControlPlane
metadata:
  name: ${CP_NAME}
spec:
  type: external
  bootstrapSecretRef:
    inClusterKey: kubeconfig-incluster
    name: ${CP_NAME}-bootstrap
    namespace: kubeflex-system
EOF
```
You can verify that the control plane has been created correctly with the command:

```console
$ kubectl get cps
NAME   SYNCED   READY   TYPE       AGE
ext1   True     True    external   5s
```

and check that the secret with the long-lived token has been created in `${CP_NAME}-system`:

```console
$ kubectl get secrets -n ${CP_NAME}-system
NAME               TYPE     DATA   AGE
admin-kubeconfig   Opaque   1      4m47s
```

## Working with a vcluster control plane

Let's create a vcluster control plane:

```console
$ kflex create cp2 --type vcluster
✔ Checking for saved hosting cluster context...
✔ Creating new control plane cp2...
✔ Waiting for API server to become ready...
```

Now interact with the new control plane, for example creating a new nginx pod:

```shell
kubectl run nginx --image=nginx
```

Verify the pod is running:

```console
$ kubectl get pods
NAME    READY   STATUS    RESTARTS   AGE
nginx   1/1     Running   0          24s
```

Access the pod logs:

```console
$ kubectl logs nginx
/docker-entrypoint.sh: /docker-entrypoint.d/ is not empty, will attempt to perform configuration
/docker-entrypoint.sh: Looking for shell scripts in /docker-entrypoint.d/
...
```

Exec into the pod and run the `ls` command:

```console
$ kubectl exec -it nginx -- sh
# ls
bin   dev                  docker-entrypoint.sh  home  media  opt   product_uuid  run   srv  tmp  var
boot  docker-entrypoint.d  etc                   lib   mnt    proc  root          sbin  sys  usr
```

Switch context back to the hosting Kubernetes and check that pod is running in the `cp2-system`
namespace:

```shell
kflex ctx
```

```console
$ kubectl get pods -n cp2-system
NAME                                                READY   STATUS    RESTARTS   AGE
coredns-64c4b4d78f-2w9bx-x-kube-system-x-vcluster   1/1     Running   0          6m58s
nginx-x-default-x-vcluster                          1/1     Running   0          4m26s
vcluster-0                                          2/2     Running   0          7m15s
```

The nginx pod is the one with the name `nginx-x-default-x-vcluster`.

## Listing Control Planes

To list all control planes managed by KubeFlex, use the following command:

```shell
kflex list

## Working with an external control plane

In this section, we will show an example of creating an external control plane to adopt
a kind cluster named `ext1`. This example supposes that the external cluster `ext1`
and the KubeFlex hosting cluster are on the same docker network.

### Determining the endpoint for an external cluster using loopback address

This is a common scenario when adopting kind or k3d. For clusters using the
default `kind` docker network, execute the following command to
check the DNS name of the external cluster `ext1` on the docker network:

```shell
docker inspect ext1-control-plane | jq '.[].NetworkSettings.Networks.kind.DNSNames'
```

The output will show something similar to the following:

```shell
[
  "ext1-control-plane",
  "79540574c3c7"
]
```

The endpoint for the adopted cluster should then be set to `https://ext1-control-plane:6443`. Note that
the port `6443` is a default value used by kind.

If you're not utilizing the default `kind` network, you'll need to make sure that the external cluster `ext1`
and the KubeFlex hosting cluster are on the same docker network.

```shelll
docker inspect ext1-control-plane | jq '.[].NetworkSettings.Networks | keys[]'
docker inspect kubeflex-control-plane | jq '.[].NetworkSettings.Networks | keys[]'
```

## Adopting the external cluster

To set up the external cluster ext1 as a control plane named cpe, use the following command:

```shell
kflex adopt --adopted-context kind-ext1 --url-override https://ext1-control-plane:6443 ext1
```

Explanation of command parameters:

- `--adopted-context kind-ext1`:
    This specifies the context name, kind-ext1, for the ext1 cluster. Ensure that this context is correctly set in your current kubeconfig file.``

- `--url-override https://ext1-control-plane:6443`:
    This parameter sets the endpoint URL for the external control plane. It's crucial to use this option when the server URL in the existing kubeconfig uses a local loopback address, which is common for kind or k3d servers running on your local machine. Here, replace https://ext1-control-plane:6443 with the actual endpoint you have determined for your external control plane in the previous step.

- `ext1`:
   This is the name of the new control plane.

### External clusters with reachable network address

If the network address of the external cluster's API server in the bootstrap Kubeconfig is accessible by the controllers operating within the KubeFlex hosting cluster, there is no need to specify a `url-override`.

## Manipulate contexts

Kubeflex offers the ability to manipulate context through `kflex ctx`. The available commands are:

### `kflex ctx`

Switch to the hosting cluster context (default name `*-kubeflex`)

### `kflex ctx CONTEXT`

Switch context to the one provided `CONTEXT`

### `kflex ctx get`

Return the current context (alias command of `kubectl config current-context`)

### `kflex ctx rename OLD_CONTEXT NEW_CONTEXT`

Rename a context within your kubeconfig file. By default, when creating a control plane `mycp`, the context, user, and cluster name are named as such:

```
context: mycp
cluster: mycp-cluster
user: mycp-admin
```

Therefore, applying the context rename command `kflex ctx rename mycp mycp-renamed` will change these 3 values as follow:

```
context: mycp-renamed
cluster: mycp-renamed-cluster
user: mycp-renamed-admin
```

### `kflex ctx delete CONTEXT`

Delete a context within your kubeconfig file. If the context deleted is your current context, `kflex` automatically switch your current context to the hosting cluster.

## Post-create hooks

With post-create hooks you can automate applying kubernetes templates on the hosting cluster or on
a hosted control plane right after the creation of a control plane. Some relevant use cases are:

- Applying OpenShift CRDs on a control plane to be used as a Workload Description Space (WDS) for deplying
workloads to OpenShift clusters.

- Starting a new controller in the namespace of a control plane in the hosting cluster that interacts
with objects in the control plane.

- Installing software components on a hosted control plane of type vcluster. An example of that is installing
the Open Cluster Management Hub on a vcluster.

### Defining hooks

To use a post-create hook, first you define the templates to apply when a control plane is created in
a `PostCreateHook` custom resource. An example "hello world" hook is defined as follows:

```yaml
apiVersion: tenancy.kflex.kubestellar.org/v1alpha1
kind: PostCreateHook
metadata:
  name: hello
  labels:
    mylabelkey: mylabelvalue
spec:
  templates:
  - apiVersion: batch/v1
    kind: Job
    metadata:
      name: hello
    spec:
      template:
        spec:
          containers:
          - name: hello
            image: public.ecr.aws/docker/library/busybox:1.36
            command: ["echo",  "Hello", "World"]
          restartPolicy: Never
      backoffLimit: 1
```

This hook will launch a job in the same namespace of the control plane that will print
"Hello World" to the standard output. Typically, a hook runs a job that by default
interacts with the hosting cluster API server. To make the job interact with the hosted
control plane  API server you can mount the secret with the in-cluster kubeconfig
for that API server. For example, for a control plane of type `k8s` you can define
a volume for a secret as follows:

```yaml
volumes:
- name: kubeconfig
  secret:
    secretName: admin-kubeconfig
```

Then, you can mount the volume and define the `KUBECONFIG` env variable as follows:

```yaml
env:
- name: KUBECONFIG
  value: "/etc/kube/kubeconfig-incluster"
volumeMounts:
- name: kubeconfig
  mountPath: "/etc/kube"
  readOnly: true
```

A complete example for installing OpenShift CRDs on a control plane is available
[here](../config/samples/postcreate-hooks/openshift-crds.yaml). More examples
are available [here](../config/samples/postcreate-hooks).

### Labels propagation

There are scenarios where you may need to setup labels on control planes based on the
features that the control plane acquires after the hook runs. For example you may want
to label a control plane where the OpenShift CRDs have been applied as a control plane
with OpenShift flavor.

To propagate labels, simply set the labels on the PostCreateHook as shown in the example
*hello* hook. The labels are then automatically propagated to any newly created control plane
where the hook is applied.

### Using the hooks

Once you define a new hook, you can just apply it in the KubeFlex hosting cluster:

```shell
kflex ctx
kubectl apply -f <hook-file.yaml> # e.g. kubectl apply -f hello.yaml
```

You can then reference the hook by name when you create a new control plane.

#### Single PostCreateHook (Legacy)

With kflex CLI (you can use --postcreate-hook or -p):

```shell
kflex create cp1 --postcreate-hook \<my-hook-name\> # e.g. kflex create cp1 -p hello
```

If you are using directly a ControlPlane CRD with kubectl, you can create a control plane
with the post-create hook as in the following example:

```shell
kubectl apply -f - <<EOF
apiVersion: tenancy.kflex.kubestellar.org/v1alpha1
kind: ControlPlane
metadata:
  name: cp1
spec:
  backend: shared
  postCreateHook: hello
  type: k8s
EOF
```

**Note**: The `postCreateHook` field is deprecated. Use `postCreateHooks` instead for better functionality.

#### Multiple PostCreateHooks (Recommended)

You can now specify multiple post-create hooks for a single control plane, allowing for more complex automation workflows. Each hook can have its own variables and will be executed when the control plane is created.

**Using kubectl:**

```shell
kubectl apply -f - <<EOF
apiVersion: tenancy.kflex.kubestellar.org/v1alpha1
kind: ControlPlane
metadata:
  name: cp1
spec:
  backend: shared
  postCreateHooks:
    - hookName: openshift-crds
      vars:
        version: "4.14"
    - hookName: install-operator
      vars:
        namespace: "operators"
    - hookName: configure-rbac
  waitForPostCreateHooks: true
  type: k8s
EOF
```

**CLI support for multiple hooks:** *Coming soon - CLI enhancement to support multiple hooks is planned.*

### Hook execution and timing

The `waitForPostCreateHooks` field controls whether the control plane waits for all post-create hooks to complete before marking itself as Ready:

- `waitForPostCreateHooks: true` (recommended): The control plane will not be marked as Ready until ALL specified hooks complete successfully. This ensures your automation completes before the control plane is considered available.

- `waitForPostCreateHooks: false` (default): The control plane is marked as Ready as soon as the API server is available, without waiting for hooks to complete. Hooks run in the background.

**Example with waiting:**

```yaml
apiVersion: tenancy.kflex.kubestellar.org/v1alpha1
kind: ControlPlane
metadata:
  name: cp-wait-example
spec:
  backend: shared
  postCreateHooks:
    - hookName: setup-hook
    - hookName: config-hook
  waitForPostCreateHooks: true  # Wait for all hooks to complete
  type: k8s
```

**Status tracking:** You can monitor hook completion in the control plane status:

```shell
kubectl get controlplane cp1 -o jsonpath='{.status.postCreateHooks}'
```

This shows the completion status of each individual hook.

While `kflex create` waits for the control plane to be available, when `waitForPostCreateHooks: false` it does not guarantee the hook's completion. Use `kubectl` commands to verify the status of resources created by the hook.

### Built-in objects

You can specify built-in objects in the templates that will be replaced at run-time.
Variables are specified using Helm-like syntax:

```yaml
"{{.<Object Name>}}"
```

Note that the double quotes are required for a valid yaml.

Currently avilable built-in objects are:

- "{{.Namespace}}" - the namespace hosting the control plane
- "{{.ControlPlaneName}}" - the name of the control plane
- "{{.HookName}}" - the name of the hook.

### User-Provided objects

In addition to the built-in objects, you can specify your own objects
to inject arbitrary values in the template. These objects are specified using
Helm-like syntax as well:

```yaml
"{{.<Your Object Name>}}"
```

#### Per-hook variables (Multiple PostCreateHooks)

When using multiple PostCreateHooks, you can specify different variables for each hook:

```yaml
apiVersion: tenancy.kflex.kubestellar.org/v1alpha1
kind: ControlPlane
metadata:
  name: cp1
spec:
  backend: shared
  postCreateHooks:
    - hookName: setup-namespace
      vars:
        namespace: "my-app"
        environment: "production"
    - hookName: deploy-operator
      vars:
        version: "v1.2.3"
        replicas: "3"
  globalVars:
    cluster: "production"
    region: "us-west-2"
  type: k8s
```

Variables are resolved with the following precedence:
1. **Built-in variables** (highest priority): Namespace, ControlPlaneName, HookName
2. **Per-hook variables**: Defined in `postCreateHooks[].vars`
3. **Global variables**: Defined in `globalVars`
4. **Default variables**: Defined in `PostCreateHook.spec.defaultVars`

#### Legacy single hook variables

For backward compatibility, when using the deprecated `postCreateHook` field, you can specify values using key/value pairs under `postCreateHookVars`:

```yaml
apiVersion: tenancy.kflex.kubestellar.org/v1alpha1
kind: ControlPlane
metadata:
  name: cp1
spec:
  backend: shared
  postCreateHook: hello
  postCreateHookVars:
    version: "0.1.0"
  type: k8s
```

You can also specify these values with the `kflex` CLI with the `--set name=value`
flag. For example:

```shell
kflex create cp1 -p hello --set version=0.1.0
```

You can specify multiple key/value pairs using a `--set` flag for each pair, for
example:

```shell
kflex create cp1 -p hello --set version=0.1.0 --set message=hello
```

## Hosting Context

The KubeFlex CLI (kflex) uses an extension in the user's kubeconfig
file to store the name of the context used for accessing the hosting
cluster. This context name is _not_ stored during the operation of
`kflex init` or instantiation of a KubeFlex Helm chart; the name is
stored later when `kflex create` or `kflex ctx $cpnaem` switches to a
different context. This is why the user **MUST NOT** change the
kubeconfig current context by other means in the interim.

The KubeFlex CLI needs to know the hosting cluster context name in
order to do `kflex ctx`, or to do `kflex ctx $cpname` when the user's
kubeconfig does not already hold a context named `$cpname` and the
current context is not the hosting cluster context.

If the relevant extension is deleted or overwritten by other apps, you
need to take steps to restore it. Otherwise, kflex context switching
may not work.

You can do this in either of the two following ways.

### Restore Hosting Context Preference by kflex ctx cpname

If the relevant extension is missing then you can restore it by using
`kubectl config use-context` to set the current context to the hosting
cluster context and then using `kflex ctx --set-current-for-hosting`
to restore the needed kubeconfig extension.

### Restore Hosting Context Preference by editing kubeconfig file

The other way is manually editing the kubeconfig file. Following is an
excerpt from an example kubeconfig file when the extension is present
and saying that the name of the context used to access the hosting
cluster is `kind-kubeflex`.

```yaml
preferences:
  extensions:
  - extension:
      data:
        kflex-initial-ctx-name: kind-kubeflex
      metadata:
        creationTimestamp: null
        name: kflex-config-extension-name
    name: kflex-config-extension-name
```

## Uninstalling KubeFlex

To uninstall KubeFlex, first ensure you remove all you control planes:

```shell
kubectl delete cps --all
```

Then, uninstall KubeFlex with the commands:

```shell
helm delete -n kubeflex-system kubeflex-operator
helm delete -n kubeflex-system postgres
kubectl delete pvc data-postgres-postgresql-0
kubectl delete ns kubeflex-system
```

## Listing Available Contexts

To list all available contexts in your kubeconfig file, use the following command:

```shell
kflex ctx list
```

## PostCreateHook Template Variables

PostCreateHooks support template variables with the following precedence:

1. **System Variables** (highest priority)
  - `Namespace`: Control plane namespace
  - `ControlPlaneName`: Name of the control plane
  - `HookName`: Name of the PostCreateHook

2. **User Variables**
  Defined in `ControlPlane.spec.postCreateHookVars`

3. **Default Variables**
  Defined in `PostCreateHook.spec.defaultVars`
