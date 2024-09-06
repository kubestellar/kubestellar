# Release notes

The following sections list the known issues for each release. The issue list is not differential (i.e., compared to previous releases) but a full list representing the overall state of the specific release. 

## 0.24.0 and its candidates and their precursors

The main functional change from 0.23.X is the completion of the status combination and introduction of the create-only feature. There is also further work on the organization of the website. There is also a major change in the GitHub repository structure: the kubestellar/ocm-transport-plugin repo's contents have been merged into the kubestellar/kubestellar repo (after `0.24.0-alpha.2`).

### Remaining limitations in 0.24.0

* Removing of WorkStatus objects (in the transport namespace) is not supported and may not result in recreation of that object
* Singleton status return: It is the user responsibility to make sure that if a BindingPolicy requesting singleton status return matches a given workload object then no other BindingPolicy matches the same object. Currently there is no enforcement of that.
* Objects on two different WDSes shouldn't have the exact same identifier (same group, version, kind, name and namespace). Such a conflict is currently not identified.

## 0.23.1

The main change from 0.23.0 is a re-organization of the website, which is still a work in progress, and archival of all website content that is outdated.

### Remaining limitations in 0.23.1

* Removing of WorkStatus objects (in the transport namespace) is not supported and may not result in recreation of that object
* Singleton status return: It is the user responsibility to make sure that if a BindingPolicy requesting singleton status return matches a given workload object then no other BindingPolicy matches the same object. Currently there is no enforcement of that.
* Objects on two different WDSes shouldn't have the exact same identifier (same group, version, kind, name and namespace). Such a conflict is currently not identified.

## 0.23.0 and its release candidates

The main change is introduction of the an all-in-one chart, called the core chart, for installing KubeStellar in a given hosting cluster and creating an initial set of WDSes and ITSes.

This release also introduces a preliminary API for combining workload object reported state from the WECs --- **BUT THE IMPLEMENTATION IS NOT DONE***. The control objects can be created but the designed response is not there. The design of the control objects is likely to change in the future too (without change in the Kubernetes API group's version string). In short, stay away from this feature in this release.

This release also features better observability (`/metrics` and `/debug/pprof`) and control over client-side self-restraint (request QPS and burst).

### Remaining limitations in 0.23.0 and its release candidates

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

* Removing of WorkStatus objects (on the transport namespace) is not supported and may not result in recreation of that object
* Singleton status return: It is the user responsibility to make sure that if a BindingPolicy requesting singleton status return matches a given workload object then no other BindingPolicy matches the same object. Currently there is no enforcement of that.
* Objects on two different WDSes shouldn't have the exact same identifier (same group, version, kind, name and namespace). Such a conflict is currently not identified.

## 0.20.0 and its release candidates

* Dynamic changes to WECs are not supported. Existing ManifestWorks will not be updated when new WECs are added or when labels are added/deleted on existing WECs
* Removing of WorkStatus objects (on the transport namespace) is not supported and may not result in recreation of that object
* Singleton status return: It is the user responsibility to make sure that if a BindingPolicy requesting singleton status return matches a given workload object then no other BindingPolicy matches the same object. Currently there is no enforcement of that.
* Objects on two different WDSes shouldn't have the exact same identifier (same group, version, kind, name and namespace). Such a conflict is currently not identified.
* An update to a workload object that removes some BindingPolicies from the matching set is not handled correctly.
* Some operations are not handled correctly while the controller is down:
   * If a workload object is deleted, or changed to remove some BindingPolicies from the matching set, it will not be handled correctly.
   * A BindingPolicy update that removes workload objects or clusters from their respective matching sets is not handled correctly.
