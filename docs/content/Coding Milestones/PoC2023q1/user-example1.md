---
short_name: user-example1
manifest_name: 'docs/content/Coding Milestones/PoC2023q1/user-example1.md'
pre_req_name: 'docs/content/common-subs/pre-req.md'
---
[![docs-ecutable - user-example1]({{config.repo_url}}/actions/workflows/docs-ecutable-user-example1.yml/badge.svg?branch={{config.ks_branch}})]({{config.repo_url}}/actions/workflows/docs-ecutable-user-example1.yml)
{%
   include-markdown "../../common-subs/required-packages-b.md"
   start="<!--required-packages-b-start-->"
   end="<!--required-packages-b-end-->"
%}

Ready to try out KubeStellar.  Here are a few quick steps to get started.

This example involves two edge clusters and two workloads.  One
workload goes on both edge clusters and one workload goes on only one
edge cluster.

{%
   include-markdown "user-example1-subs/user-example1-pre.md"
   start="<!--user-example1-pre-start-->"
   end="<!--user-example1-pre-end-->"
%}

{%
   include-markdown "user-example1-subs/user-example1-install-all.md"
   start="<!--user-example1-install-all-start-->"
   end="<!--user-example1-install-all-end-->"
%}

{%
   include-markdown "user-example1-subs/user-example1-post-install.md"
   start="<!--user-example1-post-install-start-->"
   end="<!--user-example1-post-install-end-->"
%}

{%
   include-markdown "user-example1-subs/user-example1-post-espw.md"
   start="<!--user-example1-post-espw-start-->"
   end="<!--user-example1-post-espw-end-->"
%}

## Stage 5

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
   include-markdown "../../common-subs/brew-teardown-the-environment.md"
   start="<!--brew-teardown-the-environment-start-->"
   end="<!--brew-teardown-the-environment-end-->"
%}
