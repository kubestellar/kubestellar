# Release notes

The following sections list the known issues for each release. The issue list is not differential (i.e., compared to previous releases) but a full list representing the overall state of the specific release. 

## 0.28.0

Helm chart, the name of the subobject for ArgoCD has changed from
`argo-cd` to `argocd`.

There have been minor fixups, including to the website.

We have advanced the version of the kube-rbac-proxy image used, from 0.18.0 (which is based on Kubernetes 1.30) to 0.19.1 (which is based on Kubernetes 1.31). Depending on a later minor release of Kubernetes is generally risky, but expected to work OK in this case.

### 0.28.0-rc.1

There is one breaking change for users: in the "values" for the core
Helm chart, the name of the subobject for ArgoCD has changed from
`argo-cd` to `argocd`.

There have been minor fixups, including to the website.

### 0.28.0-alpha.2

The main change is advancing the version of the kube-rbac-proxy image used, from 0.18.0 (which is based on Kubernetes 1.30) to 0.19.1 (which is based on Kubernetes 1.31). Depending on a later minor release of Kubernetes is generally risky, but expected to work OK in this case.

### 0.28.0-alpha.1

The main changes are moving from Kubernetes 1.29 to 1.30, and picking
up advances in other dependencies (but staying limited to Kubernetes
1.30).

### Remaining limitations in 0.28.0

* Although the create-only feature can be used with Job objects to avoid trouble with `.spec.selector`, requesting singleton reported state return will still lead to a controller fight over `.status.replicas` while the Job is in progress.
* Removing of WorkStatus objects (in the transport namespace) is not supported and may not result in recreation of that object
* Objects on two different WDSes shouldn't have the exact same identifier (same group, version, kind, name and namespace). Such a conflict is currently not identified.
* Creation, deletion, and modification of `CustomTransform` objects does not cause corresponding updates to the workload objects in the WECs; the current state of the `CustomTransform` objects is simply read at any moment when the objects in the WECs are being updated for other reasons.
* It is not known what actually happens when two different `Binding` objects list the same workload object and either or both say "create only".
* If (a) the workload object count or volume vs the configured limits on content of a `ManifestWork` causes multiple `ManifestWork` to be created for one `Binding` (`BindingPolicy`) AND (b) the limit on number of workload objects in one `ManifestWork` is greater then 1, then there may be transients where workload objects are deleted and re-created in a WEC --- which, in addition to possibly being troubling on its own, will certainly thwart the "create-only" functionality. The default limit on the number of workload objects in one `ManifestWork` is 1, so this issue will only arise when you use a non-default value. In this case you will avoid this issue if you set that limit to be at least the highest number of workload objects that will appear in a `Binding` (do check your `Binding` objects, lest you be surprised) AND your workload is not so large that multiple `ManifestWork` are created due to the limit on their size.

## 0.27.2

Fixes `scripts/check_pre_req.sh` so that when it objects to the version of `clusteradm`, this is more obvious (restores the RED X that was inadvertently removed in the previous patch release).

Also some doc improvements, and bumps to some build-time dependencies.

## 0.27.1

Bumps version of ingrex-nginx used to 0.12.1, to avoid recently disclosed vulnerabilities in older versions.

Avoids use of release 0.11 of clusteradm, **which introduced an incompatible change in the name of a ServiceAccount**.

## 0.27.0 and its RCs

The major changes since 0.26.0 are as follows.

- Adding the ability to use a pre-existing cluster as an ITS.
- Reliability improvement: The core Helm chart now uses KubeFlex release 0.8.1, which avoids pulling from DockerHub (which is rate-limited).

## 0.26.0 and its RCs

The major changes since 0.25.1 are as follows.

- Increase the Kubernetes release that kubestellar depends on, from 1.28 to 1.29.
- The demo environment creation script is much more reliable, mainly due to no longer attempting concurrent operations. Still, external network/server hiccups can cause the script to fail.
- This release removes the thrashing of workload objects in the WEC in the case where the transport controller's `max-num-wrapped` is 1.
- This release adds reporting, in `BindingPolicy` and `Binding` status, of whether any of the referenced `StatusCollector` objects do not exist.
- This release changes the schema for a `BindingPolicy` so that the request for sigleton status return is made/not-made independently in each `DownsyncPolicyClause` rather than once on the whole `BindingPolicySpec`. The schema for `Binding` objects is changed correspondingly. **This is a breaking change in the YAML schema for Binding[Policy] objects that request singleton status return.**

### Remaining limitations in 0.26.0

* Although the create-only feature can be used with Job objects to avoid trouble with `.spec.selector`, requesting singleton reported state return will still lead to a controller fight over `.status.replicas` while the Job is in progress.
* Removing of WorkStatus objects (in the transport namespace) is not supported and may not result in recreation of that object
* Objects on two different WDSes shouldn't have the exact same identifier (same group, version, kind, name and namespace). Such a conflict is currently not identified.
* Creation, deletion, and modification of `CustomTransform` objects does not cause corresponding updates to the workload objects in the WECs; the current state of the `CustomTransform` objects is simply read at any moment when the objects in the WECs are being updated for other reasons.
* It is not known what actually happens when two different `Binding` objects list the same workload object and either or both say "create only".
* If (a) the workload object count or volume vs the configured limits on content of a `ManifestWork` causes multiple `ManifestWork` to be created for one `Binding` (`BindingPolicy`) AND (b) the limit on number of workload objects in one `ManifestWork` is greater then 1, then there may be transients where workload objects are deleted and re-created in a WEC --- which, in addition to possibly being troubling on its own, will certainly thwart the "create-only" functionality. The default limit on the number of workload objects in one `ManifestWork` is 1, so this issue will only arise when you use a non-default value. In this case you will avoid this issue if you set that limit to be at least the highest number of workload objects that will appear in a `Binding` (do check your `Binding` objects, lest you be surprised) AND your workload is not so large that multiple `ManifestWork` are created due to the limit on their size.

## 0.26.0-alpha.5

This release is intended to have the same functionality as 0.26.0-alpha.3 and 0.26.0-alpha.4 but test a change to the GitHub Actions workflow that makes a release; the change adds SBOM generation ([PR 2718](https://github.com/kubestellar/kubestellar/pull/2718)).


## 0.26.0-alpha.4

This release is intended to have the same functionality as 0.26.0-alpha.3 but test a change to the GitHub Actions workflow that makes a release; the change suppresses attachment of useless binary archives ([PR 2704](https://github.com/kubestellar/kubestellar/pull/2704)).

## 0.26.0-alpha.3

This release adds the option for the core Helm chart to not take responsibility for running `clusteradm init` on an ITS. **Somebody** has to, but not necessarily this chart.

### Remaining limitations in 0.26.0-alpha.3

* Although the create-only feature can be used with Job objects to avoid trouble with `.spec.selector`, requesting singleton reported state return will still lead to a controller fight over `.status.replicas` while the Job is in progress.
* Removing of WorkStatus objects (in the transport namespace) is not supported and may not result in recreation of that object
* Objects on two different WDSes shouldn't have the exact same identifier (same group, version, kind, name and namespace). Such a conflict is currently not identified.
* Creation, deletion, and modification of `CustomTransform` objects does not cause corresponding updates to the workload objects in the WECs; the current state of the `CustomTransform` objects is simply read at any moment when the objects in the WECs are being updated for other reasons.
* It is not known what actually happens when two different `Binding` objects list the same workload object and either or both say "create only".
* If (a) the workload object count or volume vs the configured limits on content of a `ManifestWork` causes multiple `ManifestWork` to be created for one `Binding` (`BindingPolicy`) AND (b) the limit on number of workload objects in one `ManifestWork` is greater then 1, then there may be transients where workload objects are deleted and re-created in a WEC --- which, in addition to possibly being troubling on its own, will certainly thwart the "create-only" functionality. The default limit on the number of workload objects in one `ManifestWork` is 1, so this issue will only arise when you use a non-default value. In this case you will avoid this issue if you set that limit to be at least the highest number of workload objects that will appear in a `Binding` (do check your `Binding` objects, lest you be surprised) AND your workload is not so large that multiple `ManifestWork` are created due to the limit on their size.


## 0.26.0-alpha.1, 0.26.0-alpha.2

This release removes the thrashing of workload objects in the WEC in the case where the transport controller's `max-num-wrapped` is 1.

This release changes the schema for a `BindingPolicy` so that the request for sigleton status return is made/not-made independently in each `DownsyncPolicyClause` rather than once on the whole `BindingPolicySpec`. The schema for `Binding` objects is changed correspondingly.

### Remaining limitations in 0.26.0-alpha.1 and 0.26.0-alpha.2

* Although the create-only feature can be used with Job objects to avoid trouble with `.spec.selector`, requesting singleton reported state return will still lead to a controller fight over `.status.replicas` while the Job is in progress.
* Removing of WorkStatus objects (in the transport namespace) is not supported and may not result in recreation of that object
* Objects on two different WDSes shouldn't have the exact same identifier (same group, version, kind, name and namespace). Such a conflict is currently not identified.
* Creation, deletion, and modification of `CustomTransform` objects does not cause corresponding updates to the workload objects in the WECs; the current state of the `CustomTransform` objects is simply read at any moment when the objects in the WECs are being updated for other reasons.
* It is not known what actually happens when two different `Binding` objects list the same workload object and either or both say "create only".
* If (a) the workload object count or volume vs the configured limits on content of a `ManifestWork` causes multiple `ManifestWork` to be created for one `Binding` (`BindingPolicy`) AND (b) the limit on number of workload objects in one `ManifestWork` is greater then 1, then there may be transients where workload objects are deleted and re-created in a WEC --- which, in addition to possibly being troubling on its own, will certainly thwart the "create-only" functionality. The default limit on the number of workload objects in one `ManifestWork` is 1, so this issue will only arise when you use a non-default value. In this case you will avoid this issue if you set that limit to be at least the highest number of workload objects that will appear in a `Binding` (do check your `Binding` objects, lest you be surprised) AND your workload is not so large that multiple `ManifestWork` are created due to the limit on their size.

## 0.25.1

This patch release fixes some bugs and some documentation oversights. Following are the most notable ones.

- The transport controller bugs that strike when there is more than one `ManifestWork` for a given `Binding` (`BindingPolicy`) have been fixed (we hope).
- The [Getting Started document](get-started.md) has been updated to include documentation of how to use the script that does the steps listed in that document.

### Remaining limitations in 0.25.1

* Although the create-only feature can be used with Job objects to avoid trouble with `.spec.selector`, requesting singleton reported state return will still lead to a controller fight over `.status.replicas` while the Job is in progress.
* Removing of WorkStatus objects (in the transport namespace) is not supported and may not result in recreation of that object
* Objects on two different WDSes shouldn't have the exact same identifier (same group, version, kind, name and namespace). Such a conflict is currently not identified.
* Creation, deletion, and modification of `CustomTransform` objects does not cause corresponding updates to the workload objects in the WECs; the current state of the `CustomTransform` objects is simply read at any moment when the objects in the WECs are being updated for other reasons.
* It is not known what actually happens when two different `Binding` objects list the same workload object and either or both say "create only".
* If the workload object count or volume vs the configured limits on content of a `ManifestWork` causes multiple `ManifestWork` to be created for one `Binding` (`BindingPolicy`) then there may be transients where workload objects are deleted and re-created in a WEC --- which, in addition to possibly being troubling on its own, will certainly thwart the "create-only" functionality. Unless you workload is very large, you can avoid this situation by setting the `transport_controller.max_num_wrapped` "value" of [the core Helm chart](core-chart.md) to a number that is larger than the number of your workload objects (double check your count in your `Binding` object).


## 0.25.0 and its candidates

* The main advance in this release is finishing the implementation of the create-only feature. It is now available for use.
* The default value of transport controller's `max-num-wrapped` flag is changed to 1, in the core Helm chart.

### Remaining limitations in 0.25.0 and its candidates

* Although the create-only feature can be used with Job objects to avoid trouble with `.spec.selector`, requesting singleton reported state return will still lead to a controller fight over `.status.replicas` while the Job is in progress.
* Removing of WorkStatus objects (in the transport namespace) is not supported and may not result in recreation of that object
* Objects on two different WDSes shouldn't have the exact same identifier (same group, version, kind, name and namespace). Such a conflict is currently not identified.
* Creation, deletion, and modification of `CustomTransform` objects does not cause corresponding updates to the workload objects in the WECs; the current state of the `CustomTransform` objects is simply read at any moment when the objects in the WECs are being updated for other reasons.
* If the workload object count or volume vs the configured limits on content of a `ManifestWork` causes multiple `ManifestWork` to be created for one `Binding` (`BindingPolicy`) then there are bugs in the updating of workload objects in the WECs.
* It is not known what actually happens when two different `Binding` objects list the same workload object and either or both say "create only".
* If the workload object count or volume vs the configured limits on content of a `ManifestWork` causes multiple `ManifestWork` to be created for one `Binding` (`BindingPolicy`) then there may be transients where workload objects are deleted and re-created in a WEC --- which, in addition to possibly being troubling on its own, will certainly thwart the "create-only" functionality. Unless you workload is very large, you can avoid this situation by setting the `transport_controller.max_num_wrapped` "value" of [the core Helm chart](core-chart.md) to a number that is larger than the number of your workload objects (double check your count in your `Binding` object).



## 0.25.0-alpha.1 test releases

These test the release-building functionality, which has been revised in the course of merge the controller-manager and transport-controller Helm charts into the core Helm chart.


## 0.24.0 and its candidates and their precursors

The main functional change from 0.23.X is the completion of the status combination and the partial introduction of the create-only feature (its API is there but its implementation is not --- DO NOT TRY TO USE THIS FEATURE). There is also further work on the organization of the website. There is also a major change in the GitHub repository structure: the kubestellar/ocm-transport-plugin repository's contents have been merged into the kubestellar/kubestellar repo (after `0.24.0-alpha.2`).

### Remaining limitations in 0.24.0

* Job objects are not properly supported.
* Removing of WorkStatus objects (in the transport namespace) is not supported and may not result in recreation of that object
* Singleton status return: It is the user responsibility to make sure that if a BindingPolicy requesting singleton status return matches a given workload object then no other BindingPolicy matches the same object. Currently there is no enforcement of that.
* Objects on two different WDSes shouldn't have the exact same identifier (same group, version, kind, name and namespace). Such a conflict is currently not identified.
* Creation, deletion, and modification of `CustomTransform` objects does not cause corresponding updates to the workload objects in the WECs; the current state of the `CustomTransform` objects is simply read at any moment when the objects in the WECs are being updated for other reasons.
* If the workload object count or volume vs the configured limits on content of a `ManifestWork` causes multiple `ManifestWork` to be created for one `Binding` (`BindingPolicy`) then there are bugs in the updating of workload objects in the WECs.
* It is not known what actually happens when two different `Binding` objects list the same workload object and either or both say "create only".
* If the workload object count or volume vs the configured limits on content of a `ManifestWork` causes multiple `ManifestWork` to be created for one `Binding` (`BindingPolicy`) then there may be transients where workload objects are deleted and re-created in a WEC --- which, in addition to possibly being troubling on its own, will certainly thwart the "create-only" functionality.

## 0.23.1

The main change from 0.23.0 is a re-organization of the website, which is still a work in progress, and archival of all website content that is outdated.

### Remaining limitations in 0.23.1

* Job objects are not properly supported.
* Removing of WorkStatus objects (in the transport namespace) is not supported and may not result in recreation of that object
* Singleton status return: It is the user responsibility to make sure that if a BindingPolicy requesting singleton status return matches a given workload object then no other BindingPolicy matches the same object. Currently there is no enforcement of that.
* Objects on two different WDSes shouldn't have the exact same identifier (same group, version, kind, name and namespace). Such a conflict is currently not identified.

## 0.23.0 and its release candidates

The main change is introduction of the an all-in-one chart, called the core chart, for installing KubeStellar in a given hosting cluster and creating an initial set of WDSes and ITSes.

This release also introduces a preliminary API for combining workload object reported state from the WECs --- **BUT THE IMPLEMENTATION IS NOT DONE***. The control objects can be created but the designed response is not there. The design of the control objects is likely to change in the future too (without change in the Kubernetes API group's version string). In short, stay away from this feature in this release.

This release also features better observability (`/metrics` and `/debug/pprof`) and control over client-side self-restraint (request QPS and burst).

### Remaining limitations in 0.23.0 and its release candidates

* Job objects are not properly supported.
* Removing of WorkStatus objects (in the transport namespace) is not supported and may not result in recreation of that object
* Singleton status return: It is the user responsibility to make sure that if a BindingPolicy requesting singleton status return matches a given workload object then no other BindingPolicy matches the same object. Currently there is no enforcement of that.
* Objects on two different WDSes shouldn't have the exact same identifier (same group, version, kind, name and namespace). Such a conflict is currently not identified.

## 0.22.0 and its release candidates

The changes include adding the following features.

- Custom WEC-independent transformations of workload objects on their way from WDS to WEC.
- WEC-dependent Go template expansion in the strings of a workload object on its way from WDS to WEC.
- `PriorityClass` objects (from API group `scheduling.k8s.io`) propagate now.
- Support multiple WDSes.
- Allow multiple ITSes.
- Use the new Helm chart from kubestellar/ocm-transport-plugin for deploying the transport controller.

Prominent bug fixes include more discerning cleaning of workload objects on their way from WDS to WEC. This includes keeping a "headless" `Service` headless and removing the `spec.suspend` field from a `Job`.

See [the changelogs on GitHub](https://github.com/kubestellar/kubestellar/releases) for full details.

### Remaining limitations in 0.22.0 and its release candidates

* Job objects are not properly supported.
* Removing of WorkStatus objects (in the transport namespace) is not supported and may not result in recreation of that object
* Singleton status return: It is the user responsibility to make sure that if a BindingPolicy requesting singleton status return matches a given workload object then no other BindingPolicy matches the same object. Currently there is no enforcement of that.
* Objects on two different WDSes shouldn't have the exact same identifier (same group, version, kind, name and namespace). Such a conflict is currently not identified.


## 0.21.2 and its release candidates

The changes since 0.21.1 include efficiency improvements, reducing costs of running the kubestellar-controller-manager for a WDS that is an OpenShift cluster. There are also bug fixes and documentation improvements.

## 0.21.1
This release mainly updates the documentation exposed under kubestellar.io.

## 0.21.0 and its release candidates


### Major changes for 0.21.0 and its release candidates

* This release introduces pluggable transport. Currently the only plugin is [the OCM transport plugin](https://github.com/kubestellar/ocm-transport-plugin).

### Bug fixes in 0.21.0 and its release candidates

* dynamic changes to WECs **are supported**. Existing Bindings and ManifestWorks will be updated when new WECs are added/updated/delete or when labels are added/updated/deleted on existing WECs
* An update to a workload object that removes some BindingPolicies from the matching set _is_ handled correctly.
* These changes that happen while a controller is down are handled correctly:
   * If a workload object is deleted, or changed to remove some BindingPolicies from the matching set;
   * A BindingPolicy update that removes workload objects or clusters from their respective matching sets.

### Remaining limitations in 0.21.0 and its release candidates

* Job objects are not properly supported.
* Removing of WorkStatus objects (on the transport namespace) is not supported and may not result in recreation of that object
* Singleton status return: It is the user responsibility to make sure that if a BindingPolicy requesting singleton status return matches a given workload object then no other BindingPolicy matches the same object. Currently there is no enforcement of that.
* Objects on two different WDSes shouldn't have the exact same identifier (same group, version, kind, name and namespace). Such a conflict is currently not identified.

## 0.20.0 and its release candidates

* Job objects are not properly supported.
* Dynamic changes to WECs are not supported. Existing ManifestWorks will not be updated when new WECs are added or when labels are added/deleted on existing WECs
* Removing of WorkStatus objects (on the transport namespace) is not supported and may not result in recreation of that object
* Singleton status return: It is the user responsibility to make sure that if a BindingPolicy requesting singleton status return matches a given workload object then no other BindingPolicy matches the same object. Currently there is no enforcement of that.
* Objects on two different WDSes shouldn't have the exact same identifier (same group, version, kind, name and namespace). Such a conflict is currently not identified.
* An update to a workload object that removes some BindingPolicies from the matching set is not handled correctly.
* Some operations are not handled correctly while the controller is down:
   * If a workload object is deleted, or changed to remove some BindingPolicies from the matching set, it will not be handled correctly.
   * A BindingPolicy update that removes workload objects or clusters from their respective matching sets is not handled correctly.
