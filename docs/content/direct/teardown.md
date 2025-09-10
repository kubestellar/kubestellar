# Teardown

This document describes a way to tear down a KubeStellar system, back
to the point where the WECs and KubeFlex hosting cluster exist but
have nothing from OCM, KubeFlex, and KubeStellar installed in
them. You may find it useful to refer to [the setup and usage
flowchart](user-guide-intro.md#the-full-story).

There is a documented procedure for detaching an OCM "managed cluster"
from an OCM "hub cluster" [on the OCM
website](https://open-cluster-management.io/docs/getting-started/installation/register-a-cluster/#detach-the-cluster-from-hub). In
my experience this hangs if invoked while the managed cluster has any
workload objects being managed by the klusterlet. So teardown has to
proceed in mincing steps, first removing workload from the WECs before
detaching them from their ITS.

## Preparation

You need to be clear on what is the KubeFlex hosting cluster and have
a kubeconfig context that refers to that cluster. Put the name of that
context in the shell variable `host_context`.

Next, get a list of the KubeFlex ControlPlanes; a command like the
following will do that.

```shell
kubectl --context $host_context get controlplanes
```

Following is an example of the output from such a command. Note that
the output that you should expect to see depends on the ITSes and
WDSes that you have defined.

```console
NAME   SYNCED   READY   TYPE       AGE
its1   True     True    vcluster   196d
wds0   True     True    host       186d
```

## Teardown procedure depends on Helm charts used

If the [KubeStellar core Helm chart](core-chart.md) was used then you
will be deleting the instances of that, and this relieves you of
having to do some deletions directly.

## Deleting the WDSes

In the KubeFlex hosting cluster delete each `ControlPlane` object for
a WDS that was not created by an instance of the KubeStellar core Helm
chart.

## Deleting the OCM workload and addon

For each ITS (remember, a KubeStellar ITS is an OCM hub cluster) you
need to make sure that all the OCM workload and addons are gone.

Get a kubeconfig context for accessing the ITS, either using the
KubeFlex CLI `kflex` or by directly reading the relevant
`ControlPlane` and `Secret` objects. See [the KubeFlex user
guide](https://github.com/kubestellar/kubeflex/blob/main/docs/users.md)
for details.

Get a listing of `ManifestWork` objects in the ITS. The following command will do that,
supposing that the current kubeconfig file has a context for accessing
the ITS and the context's name is in the shell variable `its_context`.

```shell
kubectl --context $its_context get manifestworks -A
```

Following is an example of the output to expect; it shows that
KubeStellar's OCM Status Addon is still installed.

```console
NAMESPACE                NAME                          AGE
edgeplatform-test-wec1   addon-addon-status-deploy-0   196d
edgeplatform-test-wec2   addon-addon-status-deploy-0   196d
```

Delete any `ManifestWork` objects that are _NOT_ from KubeStellar's
OCM Status Addon. The ones from that addon are maintained by that
addon, so it is not effective for you to simply delete them.

KubeStellar's OCM Status Addon was installed by a Helm chart from
[ks/OSA](https://github.com/kubestellar/ocm-status-addon). When using
KubeStellar's [core Helm chart](core-chart.md), the OSA chart is
instantiated by a command run inside a container of a `Job`, so merely
deleting the core chart instance does not remove the OSA chart
instance; you must do it yourself.

Following is an example of a command that lists the Helm chart
instances ("releases", in Helm jargon) in an ITS.

```shell
helm --kube-context $its_context list -A
```

Following is an example of output from that command.

```console
NAME        	NAMESPACE              	REVISION	UPDATED                                	STATUS  	CHART                            	APP VERSION
status-addon	open-cluster-management	1       	2024-06-04 02:37:37.505803572 +0000 UTC	deployed	ocm-status-addon-chart-v0.2.0-rc9	v0.2.0-rc9
```

Following is an example of a command to delete such a chart instance.

```shell
helm --kube-context $its_context delete status-addon -n open-cluster-management
```

Check whether that got the `ManifestWork` objects deleted, using a
command like the following.

```shell
kubectl --context $its_context get manifestworks -A
```

You will probably find that they are still there. So delete them
explicitly, with a command like the following.

```shell
kubectl --context $its_context delete manifestworks -A --all
```

## Removing OCM from the WECs

The OCM website has instructions for [disconnecting a managed cluster
from its
hub](https://open-cluster-management.io/docs/getting-started/installation/register-a-cluster/#detach-the-cluster-from-hub). Do this.

Check whether all traces of OCM are gone; you can use a command like
the following, supposing that your current kubeconfig file has a
context for accessing the WEC and the context's name is in the shell
variable `wec_context`.

```shell
kubectl --context $wec_context get ns  | grep open-cluster
```

You will probably find that OCM is not all gone; following is example output.

```console
open-cluster-management                            Active   196d
```

If you find that Namespace remaining, delete it. You can use a command
like the following; it may take a few tens of seconds to complete.

```shell
kubectl --context $wec_context delete ns open-cluster-management
```

### Deleting CRDs and their instances in the WECs

The steps above still leave some `CustomResourceDefinition` (CRD)
objects from OCM in the WECs. You should remove those, and their
instances. See [Deleting CRDs and their instances in the KubeFlex
hosting
cluster](#deleting-crds-and-their-instances-in-the-kubeflex-hosting-cluster),
but do it for the WECs instead of the KubeFlex hosting cluster and for
CRDs from OCM instead of from KubeStellar (grep for
"open-cluster-management" instead of "kubestellar").

## Deleting the ITSes

Now that there is no trace of OCM left in the WECs, it is safe to
delete the ITSes. In the KubeFlex hosting cluster, delete each ITS
`ControlPlane` object that was not created due to an instance of the
KubeStellar core Helm chart.

## Deleting Remaining Stuff in the KubeFlex hosting cluster

### Delete Helm chart instances in the KubeFlex hosting cluter

Look at the Helm chart instances in the KubeFlex hosting cluster. You
might use a command like the following.

```shell
helm --kube-context $host_context list -A
```

The output may look something like the following.

```console
NAME                          	NAMESPACE            	REVISION	UPDATED                                	STATUS  	CHART                                    	APP VERSION
...
kubeflex-operator             	kubeflex-system      	1       	2024-06-03 22:35:24.360137 -0400 EDT   	deployed	kubeflex-operator-v0.6.2                 	v0.6.2
postgres                      	kubeflex-system      	1       	2024-06-03 22:33:33.210041 -0400 EDT   	deployed	postgresql-13.1.5                        	16.0.0
...
```

If you used the [KubeStellar core chart](core-chart.md) then one or
more instances of it will be in that list. In this case you can simply
delete those Helm chart instances. If the core chart was not used to
install KubeFlex in its hosting cluster then you will need to delete
the `kubeflex-system` Helm chart instance. That will probably not
delete the postgress chart instance. If that remains, delete it too.

Following is an example of a command that deletes a Helm chart
instance.

```shell
helm --kube-context $host_context delete kubeflex-operator -n kubeflex-system
```

Next, delete the `kubeflex-system` Namespace. That should be the only
one that you can find related to KubeFlex or KubeStellar.

### Deleting CRDs and their instances in the KubeFlex hosting cluster

Deleting those Helm chart instances does not remove the
`CustomResourceDefinition` (CRD) objects created by containers in
those charts. You have to delete those by hand. If I recall correctly:
before deleting the _definition_ of a custom resource, you should (for
the sake of not leaving junk in the underlying object storage, which
is bad enough in itself and could be a problem if you later introduce
a different definition for a resource of the same name) also delete
instances (objects) of that resource. The first step is to get a
listing of all the KubeStellar CRDs. You could do that with a command
like the following.

```shell
kubectl --context $host_context get crds | grep kubestellar
```

Following is an example of output from such a command.

```console
bindingpolicies.control.kubestellar.io           2024-03-01T22:03:47Z
bindings.control.kubestellar.io                  2024-03-01T22:03:48Z
campaigns.stacker.kubestellar.io                 2024-02-26T21:10:25Z
clustermetrics.galaxy.kubestellar.io             2024-06-05T15:07:08Z
controlplanes.tenancy.kflex.kubestellar.org      2024-06-03T13:58:14Z
customtransforms.control.kubestellar.io          2024-06-04T01:29:16Z
galaxies.stacker.kubestellar.io                  2024-02-26T21:10:25Z
missions.stacker.kubestellar.io                  2024-02-26T21:10:26Z
placements.control.kubestellar.io                2024-02-14T16:26:00Z
placements.edge.kubestellar.io                   2024-02-14T13:23:44Z
postcreatehooks.tenancy.kflex.kubestellar.org    2024-06-03T13:58:14Z
stars.stacker.kubestellar.io                     2024-02-26T21:10:26Z
universes.stacker.kubestellar.io                 2024-02-26T21:10:27Z
```

Next, for each one of those CRD objects, find and delete the instances
of the resource that it defines. Following is an example of a command
that gets a list for a given custom resource; this is not strictly
necessary (see the delete command below), but you may want to do it
for your own information.

```shell
kubectl --context $host_context get -A postcreatehooks.tenancy.kflex.kubestellar.org
```

Following is an example of output from such a command.

```console
NAME             SYNCED   READY   TYPE   AGE
kubestellar                              196d
ocm                                      196d
openshift-crds                           196d
```

You can delete all of the instances of a given resource with a command
like the following (which works for both cluster-scoped and namespaced
resources).

```shell
kubectl --context $host_context delete -A --all postcreatehooks.tenancy.kflex.kubestellar.org
```

Following is an example of output from such a command.

```console
postcreatehook.tenancy.kflex.kubestellar.org "kubestellar" deleted
postcreatehook.tenancy.kflex.kubestellar.org "ocm" deleted
postcreatehook.tenancy.kflex.kubestellar.org "openshift-crds" deleted
```
