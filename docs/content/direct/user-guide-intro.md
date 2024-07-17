# KubeStellar User Guide

This document is an overview of the User Guide.
See the KubeStellar [overview](../readme.md) for architecture and other information.

This user guide is an ongoing project. If you find errors, please point them out in our [Slack channel](https://kubernetes.slack.com/archives/C058SUSL5AA/) or open an issue in our [github repository](https://github.com/kubestellar/kubestellar)!

## Simple Example

If you want to try a simple installation process and example then you can try out [Getting Started](get-started.md), which uses [kind](https://kind.sigs.k8s.io/) and a helm chart. The [helm chart](core-chart.md) supports many options; the instructions on the Getting Started page show only the chart's usage in that recipe.

## In Brief

If you want a simple rough grouping, you can divide the concepts here into:

- "setup" (steps 1--7 below), exemplified in [the Setup section of Getting Started](get-started.md#setup), and
- "usage" (the remaining steps), illustrated by [the example scenarios document](example-scenarios.md).

However, you do not need to follow that dichotomy. As noted below, the relevant components can be organized more flexibly.

## The Full Story

Installing and using KubeStellar progresses through the following steps.

1. Install software prerequisites. See [prerequisites](pre-reqs.md).
1. Acquire the ability to use a Kubernetes cluster to serve as the [KubeFlex](https://github.com/kubestellar/kubeflex/) hosting cluster. See [Acquire cluster for KubeFlex hosting](acquire-hosting-cluster.md).
1. [Initialize that cluster as a KubeFlex hosting cluster](init-hosting-cluster.md).
1. Create an [Inventory and Transport Space](its.md) (ITS).
1. Create a [Workload Description Space](wds.md) (WDS).
1. Create a [Workload Execution Cluster](wec.md) (WEC).
1. [Register the WEC in the ITS](wec-registration.md).
1. Maintain workload desired state in the WDS.
1. Maintain [control objects](control.md) in the WDS to bind workload with WEC and modulate the state propagation back and forth. The [API reference](https://pkg.go.dev/github.com/kubestellar/kubestellar/api/control/v1alpha1) documents all of them. There are control objects for the following topics.
    1. [Binding workload with WEC(s)](binding.md).
    1. [Transforming desired state](transforming.md) as it travels from WDS to WEC.
    1. [Summarizing reported state](combined-status.md) from WECs into WDS.
1. Enjoy the effects of workloads being propagated to the WEC.
1. Consume reported state from WDS.

By "maintain" we mean create, read, update, delete, list, and watch as you like, over time. KubeStellar is eventually consistent: you can change your inputs as you like over time, and KubeStellar continually strives to achieve what you are currently asking it to do.

There is some flexibility in the ordering of those steps. The following flowchart shows the key ordering constraints. 

![Ordering among installation and usage actions](images/usage-outline.svg)

You can have multiple ITSes, WDSes, and WECs, created and deleted over time as you like.

Besides "Start", the other green items in that graph are entry points for extending usage at any later time. You could also see them as distinct user roles or authorities.

KubeStellar's [Core Helm chart](core-chart.md) combines initializing the KubeFlex hosting cluster, creating some ITSes, and creating some WDSes.

You can find an example run through of steps 2--7 in [Getting Started](get-started.md). This dovetails with [the example scenarios document](example-scenarios.md), which shows examples of the later steps.

## Troubleshooting

See [the Troubleshooting guide](troubleshooting.md).
