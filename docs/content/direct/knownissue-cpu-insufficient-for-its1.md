# Insufficient CPU for your clusters

When following [Getting Started](get-started.md), you may find that it hangs at some point --- simply stops making progress. For example: after instantiating [the core Helm chart](core-chart.md), a `kflex ctx` command may grind to a halt with the following output and no more.

```console
$ kflex ctx --overwrite-existing-context its1
no kubeconfig context for its1 was found: context its1 not found for control plane its1
✔ Overwriting existing context for control plane
trying to load new context its1 from server...
```

## Root cause

You are using `kind`, `k3d`, GitHub Codespaces, or any other docker-in-docker based technique and your host does not have enough CPU for all of your clusters.

## Resolution

### General

Stop any irrelevant containers.

### MacOS

On Mac computers, Docker runs all of your containers in a virtual machine. Examples of things that do this include Docker Desktop, Rancher desktop, and colima. You may need to increase the CPU allocated to this virtual machine.

For example, the colima default VM is configured to use `--cpu 2 --memory 4` --- which is insufficient for Kubestellar components on KinD clusters. In fact, KinD inherit **colima** resources when created.

To solve this issue, increase colima resource capacity to increase KinD clusters resource capcity:

1. Stop colima VM

```bash
colima stop
```

2.  Increase colima cpu and memory capacity

```bash
colima start --cpu 4 --memory 8
```

3. Delete kind cluster created by the tutorial, and start over again on a clean state.
