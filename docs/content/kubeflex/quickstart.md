# Quick Start Guide

This guide will help you get started with KubeFlex quickly. Choose the scenario that best fits your needs.

## Prerequisites

- [kind](https://kind.sigs.k8s.io/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [KubeFlex CLI](../README.md#installation)

## Basic Multi-Tenant Setup

Create the hosting kind cluster with ingress controller and install the kubeflex operator:

```shell
kflex init --create-kind
```

Create a control plane:

```shell
kflex create cp1
```

Interact with the new control plane, for example get namespaces and create a new one:

```shell
kflex ctx cp1
kubectl get ns
kubectl create ns myns
```

Create a second control plane and check that the namespace created in the first control plane is not present:

```shell
kflex create cp2
kflex ctx cp2
kubectl get ns
```

To go back to the hosting cluster context, use the `ctx` command:

```shell
kflex ctx
```

To switch back to a control plane context, use the `ctx <control plane name>` command, e.g:

```shell
kflex ctx cp1
```

To delete a control plane, use the `delete <control plane name>` command, e.g:

```shell
kflex delete cp1
```

## Advanced Multi-Tenant Scenario

For a realistic development team scenario with complete isolation:

1. **Initialize the hosting cluster**:
   ```shell
   kflex init --create-kind
   ```

2. **Create Team Alpha's control plane**:
   ```shell
   kflex create team-alpha --type k8s
   ```

3. **Switch to Team Alpha's isolated environment**:
   ```shell
   kflex ctx team-alpha
   kubectl create namespace frontend
   kubectl create namespace backend
   kubectl create deployment web --image=nginx -n frontend
   ```

4. **Create Team Beta's control plane**:
   ```shell
   kflex create team-beta --type k8s
   ```

5. **Switch to Team Beta's environment**:
   ```shell
   kflex ctx team-beta
   kubectl get namespaces  # Notice: team-alpha's namespaces are not visible
   kubectl create namespace api
   kubectl create deployment api-server --image=httpd -n api
   ```

6. **Verify complete isolation**:
   ```shell
   # Team Beta cannot see Team Alpha's resources
   kubectl get deployments --all-namespaces
   # Only shows Team Beta's deployments
   
   # Switch back to Team Alpha
   kflex ctx team-alpha
   kubectl get deployments --all-namespaces
   # Only shows Team Alpha's deployments
   ```

7. **Return to host cluster management**:
   ```shell
   kflex ctx
   kubectl get controlplanes
   # Shows both team-alpha and team-beta control planes
   ```

8. **Cleanup**:
   ```shell
   kflex delete team-alpha
   kflex delete team-beta
   ```

**Result**: Each team operates with complete isolation - they cannot see or interfere with each other's resources, yet they share the underlying infrastructure efficiently.

## Next Steps

- Read the [User's Guide](users.md) for detailed usage instructions
- Explore [Architecture](architecture.md) to understand how KubeFlex works
- Check out [Multi-Tenancy Guide](multi-tenancy.md) for advanced scenarios
