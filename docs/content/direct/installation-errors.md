# Kind host not configured for more than two clusters

[Kind](https://kind.sigs.k8s.io/) uses a docker-in-docker technique to
create multiple Kubernetes clusters on your host. But, in order for
this to work for three or more clusters, the host running Docker
typically needs an expanded configuration. This is mostly described in
[a known issue of
kind](https://kind.sigs.k8s.io/docs/user/known-issues/#pod-errors-due-to-too-many-open-files). However,
that document does not mention the additional complexity that arises
when the OS running the containers is the guest OS inside a virtual
machine on your host (e.g., a Mac, which does not natively run
containers and so uses a virtual machine with a Linux guest OS).

## Symptoms

Many KubeStellar setup paths check for the needed configuration. When
the check fails, you get an error message like the following.

```
sysctl fs.inotify.max_user_watches is only 155693 but must be at least 524288
```

If you avoid the check but the configuration is not expanded then the
symptom will most likely be setup ceasing to make progress at some
point. Or maybe other errors about things not happening or things not
existing.

## Solution
To resolve this error, you need to increase the value of `fs.inotify.max_user_watches` and/or `fs.inotify.max_user_instances`. Follow the steps below:

### For Rancher Desktop
1. Open the configuration file:
   ```sh
   vi "~/Library/Application Support/rancher-desktop/lima/_config/override.yaml"
   ```

2. Add the following script to the `provision` section:
   ```yaml
   provision:
   - mode: system
     script: |
       #!/bin/sh
       sysctl fs.inotify.max_user_watches=524288
       sysctl fs.inotify.max_user_instances=512
   ```

3. Restart Rancher Desktop.

### Docker on Linux

The resolution in the [kind known issue](https://kind.sigs.k8s.io/docs/user/known-issues#pod-errors-due-to-too-many-open-files) can be used directly.

1. Create a new configuration file:
   ```sh
   sudo vi /etc/sysctl.d/99-sysctl.conf
   ```

2. Add the following lines:
   ```sh
   fs.inotify.max_user_watches=1048576
   fs.inotify.max_user_instances=1024
   ```

3. Apply the changes:
   ```sh
   sudo sysctl -p /etc/sysctl.d/99-sysctl.conf
   ```

# Go version mismatch during image build

## Symptoms

Container image builds fail with errors indicating an unsupported or invalid Go version,
even though the project builds successfully in other environments.

## Cause

KubeStellar requires **Go 1.24 or newer** as specified in `go.mod`.
Earlier versions of the `core.Dockerfile` used Go 1.19, which caused
image build failures when the Go version requirement increased.

## Solution

Ensure that container image builds use Go 1.24 or newer.
This is handled by the updated `core.Dockerfile`, which now uses
a Go 1.24 base image.
