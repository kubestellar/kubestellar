# Workload Description Spaces

- [What is a WDS?](#what-is-a-wds)
- [Creating a WDS](#creating-a-wds)
    - [Creating a kubeconfig context for accessing the WDS](#creating-a-kubeconfig-context-for-accessing-the-wds)
- [Working with a WDS](#working-with-a-wds)
- [WDS vs. ControlPlane Registration](#wds-vs-controlplane-registration)
- [Controllers that work with a WDS](#controllers-that-work-with-a-wds)



## What is a WDS?

A Workload Description Space (WDS) is a space in the [KubeStellar architecture](architecture.md) that serves as the primary interface for users to define and manage workloads for multi-cluster deployment. The WDS consists of a Kubernetes API server with storage that:

- Stores the definitions of workloads in their native Kubernetes format
- Stores the control objects (`BindingPolicy`, `Binding`, `Status Collector`, `CombinedStatus` and `CustomTransform`) that define how workloads are distributed
- Stores status information about deployed workloads
- Acts as the main user interface to the KubeStellar system


## Creating a WDS

Currently the only documented way to create a WDS is by using the [core Helm chart](core-chart.md). See [the step-by-step instructions for getting started](get-started.md#use-core-helm-chart-to-initialize-kubeflex-and-create-its-and-wds) for an example.

The adventurous user could --- after using the core Helm chart to get this `PostCreateHook` object created --- create a WDS directly using the KubeFlex CLI or API to create a suitable `ControlPlane` object that uses the same `PostCreateHook` as the core Helm chart does for creating WDSes.

### Creating a kubeconfig context for accessing the WDS

After creating your WDS, you will need access to it. To do this you will want your kubeconfig file to have a context for accessing your WDS. The aforementioned step-by-step instructions include doing this.

The following command will (1) ensure that your kubeconfig file has a context that has (a) the same name as the WDS and (b) the right contents for accessing the WDS and then (2) make that context be your current one. You only need to to this once, after creating the WDS. **BEWARE:** This command must only be launched either (a) when the current kubeconfig context is for accessing the KubeFlex hosting cluster or (b) after the KubeFlex CLI has added its [hosting cluster extension](https://github.com/kubestellar/kubeflex/blob/main/docs/users.md#hosting-context) into your kubeconfig file (see the KubeFlex documentation about that, and, for example, the `kflex ctx --set-current-for-hosting` command in the KubeStellar step-by-step setup instructions).

```shell
# Create the WDS context if it does not already exist; ensure it has the right contents; make it current.
kflex ctx --overwrite-existing-context <wds-name>
```

Any time later that you want to switch your current kubeconfig context back to the one for this WDS, you can do it with either of the following two commands.

1. `kflex ctx <wds-name>`
2. `kubectl config use-context <wds-name>`

## Working with a WDS

With a suitable context in a kubeconfig file, you can use any Kubernetes client to manipulate objects in that WDS. See the "User Guide > Usage" section of this website for more information on using a WDS.

## WDS vs. ControlPlane Registration

It's important to distinguish between:

1. **Creating a space that can serve as a WDS**: This involves setting up a Kubernetes-like API server.
2. **Registering it with KubeFlex as a ControlPlane and deploying KubeStellar components**: This is the step that makes the space function as a WDS in the KubeStellar ecosystem.

When using KubeFlex ControlPlane types `host` or `external` for your WDS, step 1 has already been done before creating the `ControlPlane` object for the WDS.

## Controllers that work with a WDS

The following two Pods run the KubeStellar controllers for a WDS.

1. The [KubeStellar Controller Manager](architecture.md#kubestellar-controller-manager).
2. The [Transport Controller](architecture.md#pluggable-transport-controller).

These controllers are managed as `Deployment` objects in the KubeFlex hosting cluster. These `Deployment` objects are created by the setup procedures discussed above.


