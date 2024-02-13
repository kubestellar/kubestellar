# Release notes

The following sections list the known issues for each release. The issue list is not differential (i.e., compared to previous releases) but a full list representing the overall state of the specific release. 

## Known constraints and limitations release: v0.20-alpha.2
* Dynamic changes to WECs are not supported. Existing placements will not be updated when new WECs are added or when labels are added/deleted on existing WECs
* Removing of WorkStatus objects (on the transport namespace) is not supported and may not result in recreation of that object
* Singleton: It's the user responsibility to make sure there are no shared objects in two different (singleton) placements that target two different WECs. Currently there is no enforcement on on that. 
* Objects on two different WDSs shouldn't have the exact same identifier (same group, version, kind, name and namespace). Such a conflict is currently not identified.
* An update to a workload object that removes some Placements from the matching set is not handled correctly.
* Some operations are not handled correctly while the controller is down:
   * If a workload object is deleted, or changed to remove some Placements from the matching set, it will not be handled correctly.
   * A Placement update that (a) removes workload objects or clusters from their respective matching sets is not handled correctly.

## Known constraints and limitations release: v0.20-alpha.3
