# KubeStellar Roadmap

This document outlines the planned direction for KubeStellar — the multi-cluster workload orchestration engine. It is a living document updated as priorities evolve based on community feedback, user needs, and ecosystem changes.

KubeStellar is part of [CNCF Sandbox](https://www.cncf.io/projects/) and follows the [CNCF Code of Conduct](CODE_OF_CONDUCT.md).

## Completed Milestones

### v0.28–v0.30 (2025–2026)
- Pluggable transport controller with ITS (Internal Transport System) abstraction
- CustomTransform support for workload-object mutation before propagation
- BindingPolicy with cluster and object selectors
- Multi-WDS (Workload Description Space) support
- Status return from WECs to WDS objects
- GoReleaser-based automated release pipeline with signed binaries
- Prow-based CI with `ok-to-test` gate for external contributors
- OpenSSF Best Practices badge (passing level)
- Comprehensive threat model (`THREAT-MODEL.md`)
- Security contacts and vulnerability disclosure process (`SECURITY.md`)
- LFX mentorship program — active cohorts for 2025 and 2026

## Near-Term (Q2–Q3 2026)

These items are actively tracked or in-progress:

### Core Engine
- **Kubernetes 1.32 migration** (#3543, #3546) — advance core dependencies, toolchain, and CI to Kubernetes 1.32; unblock users on current clusters
- **Aggregation for remaining resource Kinds** (#3514) — extend multi-WEC status aggregation beyond Jobs and StatefulSets to remaining applicable Kinds
- **Workload propagation status for BindingPolicy** (#2161) — surface per-BindingPolicy propagation health as a first-class status condition
- **CustomTransform correctness** (#3723, PR#3764) — ensure CustomTransform is correctly applied to already-propagated objects and tracked on wrapped objects

### Reliability & Observability
- **Transport controller stability** — fix ctNameToSpec leak when no bindings remain for a GroupResource (#3759); fix unhashable-type panic (#3719)
- **Reconciliation metrics** (PR#3785) — expose binding controller reconciliation latency and error rate metrics for Prometheus
- **ITS initialization reliability epic** (#3151) — improve debuggability and reliability of ITS bootstrap sequence

### Testing
- **Automated testing policy** (PR#3775) — formalize and document automated testing requirements for OpenSSF compliance
- **Unit test expansion** (#3727, PR#3739) — focused unit tests for deterministic logic in `pkg/binding` and `pkg/status`
- **Regular E2E test expansion** (#2506) — broaden scenario coverage, add k3d support, harden against flakes

### Developer Experience
- **Modernize demo environment scripts** (#3722, PR#3786) — replace circular dependency in `create-kubestellar-demo-env.sh`; align with modern tooling

### OpenSSF Badge Improvement
- **Raise OpenSSF Best Practices score** (#3223) — target Silver-level criteria including automated test coverage, security review responses, and vulnerability tracking
- **Vulnerability response SLA** (#3257) — ensure all vulnerability reports receive a response within 14 days
- **CVE documentation** (#3253) — document CVE fixes in release notes going forward

## Mid-Term (Q3–Q4 2026)

### Core Engine
- **Create-only workload objects** (#2336) — allow users to mark objects as "create-only" so KubeStellar does not overwrite manual changes on WECs
- **Multiple ITS support** (#2025) — enable a single WDS to propagate workloads via more than one ITS
- **Workload conflict detection** (#1791) — report conflicts when multiple BindingPolicies attempt to apply the singleton label to the same object
- **Conflict handling for multiple WDSs** (#1538) — policy and user experience for resolving workload conflicts across WDS boundaries

### Ecosystem
- **CLI for KubeStellar** (#2508) — a dedicated `kubestellar` CLI for common operations (cluster registration, BindingPolicy lifecycle, status inspection) as an alternative to raw `kubectl` workflows
- **ArgoCD declarative setup** (#3147, PR#3493) — register WDSes and create applications from the core chart via ArgoCD PostCreateHook; unify GitOps and KubeStellar workflows
- **Gateway API migration** (#3553) — migrate from NGINX Ingress to Kubernetes Gateway API for improved portability

### Observability
- **ServiceMonitor → PodMonitor migration** (#2371) — move Prometheus configuration to PodMonitor for broader compatibility
- **Additional controller metrics** (#2534) — expose workqueue depth, requeue rates, and sync durations for all KubeStellar controllers

### Community
- **Multi-language documentation** (#3386) — localization framework for docs, starting with most-requested languages
- **Adopter program** — formalize adopter tiers (Evaluating / Workload Integration / Production); grow `ADOPTERS.md` to ≥5 documented entries
- **Competitive positioning docs** — publish `docs/competitive-positioning.md` comparing KubeStellar with Karmada, Liqo, and OCM to reduce evaluation friction

## Long-Term (2027+)

- **CNCF incubation application** — complete adopter documentation, security audit, and governance requirements for incubation submission
- **Multi-cluster network policy** — native NetworkPolicy propagation with WEC-side enforcement awareness
- **BindingPolicy simulation/dry-run** — allow users to preview which objects would be propagated before applying a BindingPolicy
- **Web UI integration** — first-class connection between KubeStellar Console and core orchestration engine status APIs
- **Pluggable conflict resolution** — policy-driven resolution strategies for workload conflicts (last-write-wins, priority-based, merge)
- **Federated status aggregation** — aggregate status across WECs into a single, queryable status object per workload

## Non-Goals

KubeStellar intentionally does **not** aim to:

- **Replace Karmada or OCM** — KubeStellar builds on top of OCM and addresses a complementary policy-driven propagation use case; it is not a fork or replacement of either
- **Be a full GitOps engine** — KubeStellar handles runtime workload propagation policy; GitOps delivery (ArgoCD, Flux) is complementary and can work alongside it
- **Provide its own cluster provisioner** — KubeStellar observes and manages existing clusters; infrastructure provisioning is out of scope
- **Replace kubectl or Helm** — KubeStellar extends Kubernetes-native tooling; it does not replace CLI or package management workflows

## How to Influence the Roadmap

- **GitHub Issues** — Open an issue with the `kind/feature` label and describe the use case
- **Discussions** — Join [#kubestellar-dev on CNCF Slack](https://cloud-native.slack.com/archives/C097094RZ3M/)
- **LFX Mentorship** — See open LFX project slots for structured contribution opportunities
- **Community meetings** — Attend the [KubeStellar community call](https://github.com/kubestellar/kubestellar/blob/main/CONTRIBUTING.md) (schedule in CONTRIBUTING.md)

Items on this roadmap are subject to change. For the most current status, check the linked GitHub issues and the project's [milestone page](https://github.com/kubestellar/kubestellar/milestones).
