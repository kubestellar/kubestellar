This doc attempts to show a simple example usage of the 2023q1 PoC.
This doc is a work in progress.

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

