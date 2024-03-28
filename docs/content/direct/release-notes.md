# Release notes

The following sections list the known issues for each release. The issue list is not differential (i.e., compared to previous releases) but a full list representing the overall state of the specific release. 

## 0.21.2 and its release candidates

Changes since 0.21.1 are mainly bug fixes and doc work; see changelog for full details.

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
* Objects on two different WDSs shouldn't have the exact same identifier (same group, version, kind, name and namespace). Such a conflict is currently not identified.

## 0.20.0 and its release candidates

* Dynamic changes to WECs are not supported. Existing ManifestWorks will not be updated when new WECs are added or when labels are added/deleted on existing WECs
* Removing of WorkStatus objects (on the transport namespace) is not supported and may not result in recreation of that object
* Singleton status return: It is the user responsibility to make sure that if a BindingPolicy requesting singleton status return matches a given workload object then no other BindingPolicy matches the same object. Currently there is no enforcement of that.
* Objects on two different WDSs shouldn't have the exact same identifier (same group, version, kind, name and namespace). Such a conflict is currently not identified.
* An update to a workload object that removes some BindingPolicies from the matching set is not handled correctly.
* Some operations are not handled correctly while the controller is down:
   * If a workload object is deleted, or changed to remove some BindingPolicies from the matching set, it will not be handled correctly.
   * A BindingPolicy update that removes workload objects or clusters from their respective matching sets is not handled correctly.
