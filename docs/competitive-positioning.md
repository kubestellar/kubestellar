# KubeStellar: Positioning and Comparison Guide

This document helps engineers evaluate KubeStellar in the context of the broader multi-cluster ecosystem. It is intended to reduce evaluation friction and answer the most common question we hear: *"How does KubeStellar compare to [Karmada / Liqo / OCM / ArgoCD]?"*

## TL;DR

KubeStellar is a **policy-driven workload propagation engine** for Kubernetes. It answers the question: *"Given a desired state and a set of clusters, which objects should go where, and how should I know they arrived?"*

It is **not** a cluster provisioner, a GitOps engine, or a service mesh. It is most naturally used **alongside** those tools, not instead of them.

---

## Core Concepts (2-minute primer)

| Concept | What it is |
|---------|-----------|
| **WDS** (Workload Description Space) | A Kubernetes API server where you define workloads and propagation policies |
| **ITS** (Internal Transport System) | The mechanism that carries objects from WDS to WECs; defaults to Open Cluster Management |
| **WEC** (Workload Execution Cluster) | A downstream cluster where workloads actually run |
| **BindingPolicy** | A KubeStellar CRD that says "propagate objects matching X to clusters matching Y" |
| **CustomTransform** | A KubeStellar CRD that modifies objects in-flight before they reach a WEC |
| **StatusCollector** | Returns status from WECs back to WDS objects |

---

## KubeStellar vs Open Cluster Management (OCM)

**Relationship: KubeStellar builds on top of OCM — they are complementary, not competing.**

| | KubeStellar | OCM |
|--|-------------|-----|
| Layer | Policy + propagation logic | Transport + cluster registration |
| Objects | KubeStellar `BindingPolicy` CRD | OCM `Subscription`, `Channel`, `ManifestWork` |
| Abstraction | High-level: "send workload X to clusters matching label Y" | Low-level: `ManifestWork` per cluster |
| Status | `StatusCollector` aggregates across WECs | `ManifestWork` status per cluster |
| When to use alone | You need rich propagation policies and status aggregation | You need cluster registration and basic `ManifestWork` delivery |
| When to use together | ✅ Recommended | ✅ Recommended |

**Recommendation:** Use KubeStellar when you need declarative, policy-driven propagation with status feedback. OCM provides the transport; KubeStellar provides the intelligence on top of it.

---

## KubeStellar vs Karmada

| | KubeStellar | Karmada |
|--|-------------|---------|
| Architecture | Hub runs no agent on spoke clusters (ITS is a separate API server) | Hub talks to spoke clusters; spoke requires `karmada-agent` (push mode) or API aggregation (pull mode) |
| Policy model | `BindingPolicy` — Kubernetes-native CRD on WDS | `PropagationPolicy` + `ClusterPropagationPolicy`; `ResourceBinding` is an intermediate resource |
| Transport | Pluggable (defaults to OCM); ITS is decoupled from WDS | Built-in; tightly coupled to hub |
| Status | `StatusCollector` per workload; flexible aggregation rules | `ResourceBinding` status; built-in aggregation for some resource types |
| Multi-WDS | Multiple WDS instances supported (planned) | Single hub |
| Object modification | `CustomTransform` CRD | `OverridePolicy` / `ClusterOverridePolicy` |
| CNCF status | Sandbox | Sandbox |
| Maturity | v0.30.x; production-ready for policy propagation | v1.x; more mature, wider community |

**When to choose KubeStellar:**
- You want a pluggable transport layer (e.g., swap OCM for a future transport)
- You don't want agents running inside spoke clusters
- You prefer separate control planes (WDS and ITS are different API servers)
- You are already using OCM for cluster management

**When Karmada may fit better:**
- You need a mature, stable API surface (v1.x vs v0.3x)
- You want a single-binary hub-and-spoke model
- Your organization already standardizes on Karmada

**Can you use both?** Technically yes, for different workloads, but there is no official integration. Choose one as your propagation layer.

---

## KubeStellar vs Liqo

| | KubeStellar | Liqo |
|--|-------------|------|
| Model | Policy-driven propagation | Peer-to-peer resource offloading via virtual nodes |
| Abstraction | BindingPolicy selects which objects go where | Virtual node abstraction; scheduler-transparent offloading |
| Network | Network-agnostic; you manage connectivity | Builds its own overlay network between peers |
| Object types | Any Kubernetes object | Pod-scheduling focus (though Liqo supports broader resource reflection) |
| Status | StatusCollector returns status to WDS | Reflected resource status on virtual node |
| When to choose | Policy-driven multi-cluster with fine-grained object selection | Transparent compute offloading without changing workload manifests |

**Liqo and KubeStellar solve different problems.** Liqo is excellent if you want a cluster to seamlessly overflow workloads to a peer without changing your deployment manifests. KubeStellar is the right choice when you need explicit, policy-driven control over which objects land on which clusters.

---

## KubeStellar vs ArgoCD (Multi-cluster)

**Relationship: Complementary. Use ArgoCD for GitOps delivery; use KubeStellar for runtime propagation policy.**

| | KubeStellar | ArgoCD |
|--|-------------|--------|
| Model | Runtime propagation policy | GitOps delivery from git to cluster |
| Trigger | BindingPolicy (label selectors on objects + clusters) | Git commit / sync webhook |
| Target | Multiple clusters, policy-driven | `Application` targets a single cluster (ApplicationSet for multi-cluster) |
| Object modification | CustomTransform in-flight | Helm values, Kustomize overlays |
| Status | StatusCollector aggregates across WECs | Application health per cluster |
| Drift detection | Status return shows WEC state; not ArgoCD-style sync | ArgoCD detects and optionally auto-syncs drift |

**Common pattern:** GitOps (ArgoCD) manages the WDS objects; KubeStellar propagates those objects to the right WECs at runtime. This is the "GitOps + KubeStellar" deployment model documented in issue #3147.

---

## Feature Comparison Matrix

| Feature | KubeStellar | Karmada | Liqo | OCM |
|---------|------------|---------|------|-----|
| No spoke agent required | ✅ | ❌ (karmada-agent) | ❌ (liqo-agent) | ❌ (OCM agent) |
| Label-based cluster selection | ✅ | ✅ | Partial | Partial |
| Object-level propagation policy | ✅ | ✅ | ❌ | Partial |
| In-flight object transformation | ✅ (CustomTransform) | ✅ (OverridePolicy) | ❌ | ❌ |
| Status aggregation | ✅ (StatusCollector) | ✅ | ❌ | Partial |
| Pluggable transport | ✅ | ❌ | ❌ | — |
| Multi-WDS | ✅ (planned) | ❌ | ❌ | ❌ |
| Network overlay | ❌ | ❌ | ✅ | ❌ |
| CNCF Sandbox | ✅ | ✅ | ✅ | ✅ (OCM) |

---

## Frequently Asked Questions

### "KubeStellar requires OCM — am I locked in to OCM's maturity level?"

KubeStellar's transport layer is pluggable via the ITS abstraction. Today OCM is the default ITS implementation; future alternative transports are on the roadmap. The KubeStellar policy layer (BindingPolicy, CustomTransform, StatusCollector) is completely independent of the transport.

### "Karmada has more GitHub stars — is KubeStellar less mature?"

Karmada has a larger community and is at v1.x. KubeStellar is at v0.30.x and is the right choice when you prioritize transport pluggability, no-spoke-agent architecture, and tighter OCM integration. Both are CNCF Sandbox projects.

### "Can I migrate from Karmada to KubeStellar?"

The concepts are similar (propagation policy, override, status) but the APIs differ. There is no automated migration tooling; see issue #3537 (LFX mentorship: Alternative Solution Migration Specialist) for ongoing work in this area.

### "Is KubeStellar production-ready?"

KubeStellar v0.30.0 is used in production for policy-driven workload propagation. See [ADOPTERS.md](ADOPTERS.md) for documented adopters.

---

## Further Reading

- [KubeStellar Getting Started](docs/content/direct/get-started.md)
- [BindingPolicy API reference](docs/content/direct/binding-policy.md)
- [KubeStellar Architecture](docs/content/direct/architecture.md)
- [CNCF project page](https://www.cncf.io/projects/kubestellar/)
- Community questions: [#kubestellar-dev on CNCF Slack](https://cloud-native.slack.com/archives/C097094RZ3M/)
