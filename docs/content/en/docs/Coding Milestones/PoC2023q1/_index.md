---
categories: ["Coding", "Sprints", "Milesones", "PoC", "q1"]
tags: ["code","milestone","poc", "q1"] 
title: "Milestone PoC2023q1"
linkTitle: "Milestone PoC2023q1"
weight: 2
description: >
  Consider collaborating on a coding milestone
---

<!-- {{% pageinfo %}}
This is a placeholder page that shows you how to use this template site.
{{% /pageinfo %}} -->

PoC 2023q1 will be presented on March 30th, 2023.  The PoC will include:
- Separation of inventory and workload management.
- The focus here is on workload management, and that strictly reads inventory.
- What passes from inventory to workload management is kcp TMC Location and SyncTarget objects.
- Use of a kcp workspace as the container for the central spec of a workload.
- Propagation of desired state from center to edge, as directed by EdgePlacement objects and the Location and SyncTarget objects they reference.
- Interfaces designed for a large number of edge clusters.
- Interfaces designed with the intention that edge clusters operate independently of each other and the center (e.g., can tolerate only occasional connectivity) and thus any “service providers” (in the technical sense from kcp) in the center or elsewhere.
- Rule-based customization of desired state.
- Propagation of reported state from edge to center.
- Summarization of reported state in the center.
- Return and/or summarization of reported state from associated objects (e.g., ReplicaSet or Pod objects associated with a given Deployment object).
- The edge opens connections to the center, not vice-versa.
- An edge computing platform “product” that can be deployed (as opposed to a service that is used).

