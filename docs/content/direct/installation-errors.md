# Installation Errors and Solutions

## Error: `sysctl fs.inotify.max_user_watches is only 155693 but must be at least 524288`

### Cause
This error occurs because the current value of `fs.inotify.max_user_watches` is too low. This setting controls the maximum number of file watches that a user can create.

### Solution
To resolve this error, you need to increase the value of `fs.inotify.max_user_watches`. Follow the steps below:

#### For Rancher Desktop
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
       echo "fs.inotify.max_user_watches=1048576" > /etc/sysctl.d/fs.inotify.conf
       echo "fs.inotify.max_user_instances=1024" >> /etc/sysctl.d/fs.inotify.conf
       sysctl -p /etc/sysctl.d/fs.inotify.conf
   ```

3. Restart Rancher Desktop.

#### For Docker Runtimes
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

More details can be found [here](https://kind.sigs.k8s.io/docs/user/known-issues#pod-errors-due-to-too-many-open-files).
