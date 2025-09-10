# Potential Error with Kubestellar Installation related to Issues with Kind backed by Rancher Desktop

## Description of the Issue

Kubestellar installation may fail for some users during the setup of the second cluster (cluster2) when running Kind with Docker provided by Rancher Desktop. The failure occurs while initializing the cluster, with an error related to `kubeadm`. Insufficient system parameter settings (`sysctl`) within the Rancher Desktop virtual machine may be causing this issue.

## Error Message Example

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

## Root Cause

This is caused by a [known issue with kind](https://kind.sigs.k8s.io/docs/user/known-issues/#pod-errors-due-to-too-many-open-files).

When using Rancher Desktop, the Linux machine that needs to be reconfigured is a virtual machine that Rancher Desktop is managing.

Kind requires the following minimum settings in the Linux machine:

```
fs.inotify.max_user_watches = 524288
fs.inotify.max_user_instances = 512
```

If these parameters are set lower than the suggested values, the second cluster initialization may fail.

## Steps to Reproduce the Issue

1. Install Rancher Desktop
   - Download and install Rancher Desktop from the official website.
   - Configure it to use Docker as the container runtime (`dockerd`).
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

## Expected Behavior

Cluster 2 should create successfully, and the installation should complete without errors.

## Steps to Fix

1. Check Current `sysctl` Parameter Values
   - Use the command `rdctl shell` to log in to the Rancher Desktop VM.
     Run:
     ```
     sysctl fs.inotify.max_user_watches
     sysctl fs.inotify.max_user_instances
     ```
   - Confirm if these values are below the recommended settings (524288 for max_user_watches and 512 for max_user_instances).
2. Modify the Parameter Settings
   - Setting these parameters temporarily with `sysctl` will revert after restarting Rancher Desktop. To persist the changes, you need to modify the configuration using an overlay file.
3. Create an Override Configuration File
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

4. Restart Rancher Desktop
   - Restart Rancher Desktop for the changes to take effect and ensure the new `sysctl` parameter values persist.
5. Delete Existing Kind Clusters
   - Before re-running the Kubestellar Getting Started guide, delete all previously created clusters:

     ```
     kind delete cluster --name <cluster-name>
     ```

   - Repeat for each cluster (e.g., kubeflex, cluster1, cluster2).

6. Re-run the Kubestellar Setup
   - With the updated configuration, run the Kubestellar Getting Started guide or the automated demo environment script again.
   - Verify that both clusters are created successfully without errors.

## Additional Note: Ensuring a Clean Environment for Reinstallation

Deleting all existing Kind clusters before re-running the installation ensures no leftover configurations interfere with the new setup.
