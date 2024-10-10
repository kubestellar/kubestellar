# Troubleshooting

This guide is a work in progress.

## Debug log levels

The KubeStellar controllers take an optional command line flag that
sets the level of debug logging to emit. Each debug log message is
associated with a _log level_, which is a non-negative integer. Higher
numbers correspond to messages that appear more frequently and/or give
more details. The flag's name is `-v` and its value sets the highest
log level that gets emitted; higher level messages are suppressed.

The KubeStellar debug log messages are assigned to log levels roughly
according to the following rules. Note that the various Kubernetes
libraries used in these controllers also emit leveled debug log
messages, according to their own numbering conventions. The
KubeStellar rules are designed to not be very inconsistent with the
Kubernetes practice.

- **0**: messages that appear O(1) times per run.
- **1**: more detailed messages that appear O(1) times per run.
- **2**: messages that appear O(1) times per lifecycle event of an API object or important conjunction of them (e.g., when a Binding associates a workload object with a WEC).
- **3**: more detailed messages that appear O(1) times per lifecycle event of an API object or important conjunction of them.
- **4**: messages that appear O(1) times per sync. A sync is when a controller reads the current state of one API object and reacts to that.
- **5**: more detailed messages that appear O(1) times per sync.

The [core Helm chart](core-chart.md) has "values" that set the
verbosity (`-v`) of various controllers.

## Things to look at

- Remember that for each of your BindingPolicy objects, there is a corresponding Binding object that reports what is matching the policy object.
- Although not part of the interface, when debugging you can look at the ManifestWork and WorkStatus objects in the ITS.
- Look at logs of controllers. If they have had container restarts that look relevant, look also at the previous logs. Do not forget OCM controllers. Do not forget that some Pods have more than one interesting container.
- If a controller's `-v` is not at least 5, increase it.
- Remember that Kubernetes controllers tend to report transient problems as errors without making it clear that the problem is transient and tend to not make it clear if/when the problem has been resolved (sigh).

## Making a good trouble report

Basic configuration information.

- Include the versions of all the relevant software; do not forget the OCM pieces.
- Report on each Kubernetes/OCP cluster involved. What sort of cluster is it (kind, k3d, OCP, ...)? What version of that?
- For each WDS and ITS involved, report on what sort of thing is playing that role (remember that a Space is a role) --- a new KubeFlex control plane (report type) or an existing cluster (report which one).

Do a simple clean demonstration of the problem, if possible.

Show the particulars of something going wrong.

- Report timestamps of when salient changes happened. Make it clear which timezone is involved in each one. Particularly interesting times are when KubeStellar did the wrong thing or failed to do anything at all in response to something.
- Look at the Binding and ManifestWork and WorkStatus objects and the controller logs. Include both in a problem report. Show the relevant workload objects, from WDS and from WEC. When the problem is behavior over time, show the objects contents from before and after the misbehavior.
- When reporting kube API object contents, include the `meta.managedFields`. For example, when using `kubectl get`, include `--show-managed-fields`.

## Potential Error with Kubestellar Installation related to Issues with Kind backed by Rancher Desktop
### Description of the Issue

Kubestellar installation may fail for some users during the setup of the second cluster (cluster2) when running Kind with Docker provided by Rancher Desktop. The failure occurs while initializing the cluster, with an error related to kubeadm and insufficient system parameter settings (sysctl) within the Rancher Desktop virtual machine.

### Error Message Example

```
Error: hub oriented command should not running against non-hub cluster
Creating cluster "cluster2" ...
...
ERROR: failed to create cluster: failed to init node with kubeadm: command "docker exec --privileged cluster2-control-plane kubeadm init --skip-phases=preflight --config=/kind/kubeadm.conf --skip-token-print --v=6" failed with error: exit status 1
Command Output: I1008 16:11:20.743111 134 initconfiguration.go:255] loading configuration from "/kind/kubeadm.conf"
...
[config] WARNING: Ignored YAML document with GroupVersionKind kubeadm.k8s.io/v1beta3, Kind=JoinConfiguration
...
```

### Root Cause

The issue arises due to low default settings for certain sysctl parameters in the Rancher Desktop virtual machine. Kind requires the following minimum settings:

```
fs.inotify.max_user_watches = 524288
fs.inotify.max_user_instances = 512
```

If these parameters are set lower than the required values, the second cluster initialization will fail.

### Steps to Reproduce the Issue
1. Install Rancher Desktop
   - Download and install Rancher Desktop from the official website.
   - Configure it to use Docker as the container runtime (dockerd).
2. Install Kind
   - Follow the installation instructions provided in the Kind documentation.
3. Install Kubestellar Prerequisites
   - Ensure that all required dependencies for Kubestellar are installed on your system. Refer to the Kubestellar documentation for a complete list.
4. Run the Kubestellar Getting Started Guide or Demo Environment Setup Script
   - Follow the steps in the Kubestellar Getting Started guide or run the automated demo environment setup script.
5. Monitor the Installation Process
   - Confirm the successful installation of kubeflex.
   - Ensure that ITS1 (Information Transformation Service 1) and WDS1 (Workload Distribution Service 1) are created.
   - Verify the creation of the first cluster (cluster1).
6. Wait for the Creation of Cluster2
   - Allow the script to attempt the creation of the second remote cluster (cluster2).
   - The error should occur during this step if the issue is present.

# Expected Behavior

Cluster 2 should create successfully, and the installation should complete without errors.

# Steps to Fix
1. Check Current sysctl Parameter Values
   - Use the command rdctl shell to log in to the Rancher Desktop VM.
Run:
```
sysctl fs.inotify.max_user_watches
sysctl fs.inotify.max_user_instances
```
   - Confirm if these values are below the recommended settings (524288 for max_user_watches and 512 for max_user_instances).
2. Modify the Parameter Settings
   - Setting these parameters temporarily with sysctl will revert after restarting Rancher Desktop. To persist the changes, you need to modify the configuration using an overlay file.
Create an Override Configuration File
   - On a Mac:
     - Open a terminal and create a new file:
```
vi ~/Library/Application\ Support/rancher-desktop/lima/_config/override.yaml
```
     - Add the following content:
```
provision:
- mode: system
  script: |
    #!/bin/sh
    echo "fs.inotify.max_user_watches=524288" > /etc/sysctl.d/fs.inotify.conf
    echo "fs.inotify.max_user_instances=512" >> /etc/sysctl.d/fs.inotify.conf
    sysctl -p /etc/sysctl.d/fs.inotify.conf
```
     - Save the file.
3. Restart Rancher Desktop
   - Restart Rancher Desktop for the changes to take effect and ensure the new sysctl parameter values persist.
4. Delete Existing Kind Clusters
   - Before re-running the Kubestellar Getting Started guide, delete all previously created clusters:

```
kind delete cluster --name <cluster-name>
```
Repeat for each cluster (e.g., kubeflex, cluster1, cluster2).
   - Re-run the Kubestellar Setup
     - With the updated configuration, run the Kubestellar Getting Started guide or the automated demo environment script again.
     - Verify that both clusters are created successfully without errors.

# Additional Note: Ensuring a Clean Environment for Reinstallation
Deleting all existing Kind clusters before re-running the installation ensures no leftover configurations interfere with the new setup.
