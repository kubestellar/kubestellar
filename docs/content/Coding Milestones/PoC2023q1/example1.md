---
short_name: example1
manifest_name: 'docs/content/Coding Milestones/PoC2023q1/example1.md'
pre_req_name: 'docs/content/common-subs/pre-req.md'
---
[![docs-ecutable - example1]({{config.repo_url}}/actions/workflows/docs-ecutable-example1.yml/badge.svg?branch={{config.ks_branch}})]({{config.repo_url}}/actions/workflows/docs-ecutable-example1.yml)
{%
   include-markdown "../../common-subs/required-packages.md"
   start="<!--required-packages-start-->"
   end="<!--required-packages-end-->"
%}
{%
   include-markdown "../../common-subs/save-some-time.md"
   start="<!--save-some-time-start-->"
   end="<!--save-some-time-end-->"
%}

This doc shows a detailed example usage of the KubeStellar components.

This example involves two edge clusters and two workloads.  One
workload goes on both edge clusters and one workload goes on only one
edge cluster.  Nothing changes after the initial activity.

This example is presented in stages.  The controllers involved are
always maintaining relationships.  This document focuses on changes as
they appear in this example.

## Stage 1

{%
   include-markdown "example1-subs/example1-pre-kcp.md"
   start="<!--example1-pre-kcp-start-->"
   end="<!--example1-pre-kcp-end-->"
%}

{%
   include-markdown "example1-subs/example1-start-kcp.md"
   start="<!--example1-start-kcp-start-->"
   end="<!--example1-start-kcp-end-->"
%}

{%
   include-markdown "example1-subs/example1-post-kcp.md"
   start="<!--example1-post-kcp-start-->"
   end="<!--example1-post-kcp-end-->"
%}

{%
   include-markdown "example1-subs/example1-post-espw.md"
   start="<!--example1-post-espw-start-->"
   end="<!--example1-post-espw-end-->"
%}

## Stage 5

### Singleton reported state return

The workload `ReplicaSet` and `Deployment` objects above request
return of reported state to the WDS when the number of executing
copies is exactly 1.

For the common workload, the number of executing copies should be 2.
Check that this is the number reported.

```shell
kubectl ws root:my-org:wmw-c
kubectl get rs -n commonstuff commond -o yaml | grep 'kubestellar.io/executing-count: "2"' || { kubectl get rs -n commonstuff commond -o yaml; false; }
```

For the special workload, the number of executing copies should be 1.  Check it.

```shell
kubectl ws root:my-org:wmw-s
kubectl get deploy -n specialstuff speciald -o yaml | grep 'kubestellar.io/executing-count: "1"' || { kubectl get deploy -n specialstuff speciald -o yaml; false; }
```

Look at the status section of the "speciald" `Deployment` and see that
it has been filled in with the information from the guilder cluster.

```shell
kubectl get deploy -n specialstuff speciald -o yaml
kubectl get deploy -n specialstuff speciald -o yaml | grep 'readyReplicas: 1'
```

### Status Summarization (aspirational)

![Summarization for special](Edge-PoC-2023q1-Scenario-1-stage-5s.svg "Status summarization for special")

The status summarizer, driven by the EdgePlacement and
SinglePlacementSlice for the special workload, creates a status
summary object in the specialstuff namespace in the special workload
workspace holding a summary of the corresponding Deployment objects.
In this case there is just one such object, in the mailbox workspace
for the guilder cluster.

![Summarization for common](Edge-PoC-2023q1-Scenario-1-stage-5c.svg "Status summarization for common")

The status summarizer, driven by the EdgePlacement and
SinglePlacementSlice for the common workload, creates a status summary
object in the commonstuff namespace in the common workload workspace
holding a summary of the corresponding Deployment objects.  Those are
the `commond` Deployment objects in the two mailbox workspaces.

## Teardown the environment

{%
   include-markdown "../../common-subs/teardown-the-environment.md"
   start="<!--teardown-the-environment-start-->"
   end="<!--teardown-the-environment-end-->"
%}
