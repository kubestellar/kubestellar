---
title: "Details"
---

Want to get involved? Check out our [good-first-issue list]({{ config.repo_url }}/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22).

TODO: draw new picture

## Status of this memo

This document outlines near-term plans for building on what was
produced in the work on [the PoC for 2023q1](../../PoC2023q1/outline/).

## Introduction

PoC2023q1 was defined with the over-arching goal of supporting edge
computing scenarios.  Since then we have realized that the technical
problems that we took on are not that specific, they appear in other
multi-cluster scenarios as well.

The goals of this PoC are as follows.  Ones that are substantially
different from what has been accomplished for PoC2023q1 are
highlighted.

- Separation of infrastructure and workload management.
- The focus here is on workload management, and that strictly reads
  an inventory of infrastructure.
- What passes from inventory to workload management is kcp TMC
  Location and SyncTarget objects.
- **Compared to PoC2023q1, decoupling from kcp TMC by making our own
  copy of the definitions of SyncTarget and Location.**
- **Potentially: switch from using SyncTarget and Location to some
  other representation of inventory.**
- **Compared to PoC2023q1, decoupling from kcp core by (1) introducing
  an abstraction layer that delivers the essential functionality of
  kcp's logical clusters based on a variety of implementations and (2)
  using [kube-bind](https://github.com/kube-bind/kube-bind) instead of
  kcp's APIExport/APIBinding.  Where PoC2023q1 used the concept of a
  kcp workspace, PoC2023q4 uses the abstract concept that we call a
  "space".**
- Use of a space as the container for the central spec of a workload.
- Propagation of desired state from center outward, as directed by
  EdgePlacement objects and the referenced inventory objects.
- Interfaces designed for a large number of workload execution clusters.
- Interfaces designed with the intention that workload execution
  clusters operate independently of each other and the center (e.g.,
  can tolerate only occasional connectivity) and thus any "service
  providers" (in the technical sense from kcp) in the center or
  elsewhere.
- Rule-based customization of desired state.
- Propagation of reported state from workload execution clusters to center.
- Summarization of reported state in the center.
- **Exact, not summarized, reported state returned to workload
  description space in the case of placement on exactly 1 workload
  execution cluster.**
- Return and/or summarization of state from associated objects (e.g.,
  ReplicaSet or Pod objects associated with a given Deployment
  object).
- The TCP connections are opened in the inward direction, not outward.
- A platform "product" that can be deployed (as opposed to a service
  that is used).
- **Codified support for scenarios where some KubeStellar clients and
  the syncers in some of the workload execution clusters have to go
  through load balancers and/or other proxies to reach the central
  server(s).**
- **Compared to PoC2023q1, codification of closer to production grade
  deployment technique(s).**
- **A hierarchy with more than two levels.**

Some important things that are not attempted in this PoC include the following.

- An implementation that supports a very large volume of reported
  state (which could come from either a large number of workload
  execution clusters and/or a large amount of reported state in each
  one of those).
- User control over ordering of propagation from center outward,
  either among destinations or kinds of objects.
- More than baseline security (baseline being, e.g., HTTPS, Secret
  objects, non-rotating bearer token based service authentication).
- A good design for bootstrapping the workload management in the
  workload execution clusters.
- Very strong isolation between tenants of this platform.
