# KubeStellar Ecosystem

KubeStellar is not a single tool — it's a platform ecosystem of coordinated projects. This document maps the full ecosystem so evaluators, contributors, and adopters understand the complete picture.

## Platform Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                    KubeStellar Platform                          │
│                                                                  │
│  ┌─────────────────┐   BindingPolicy   ┌──────────────────────┐ │
│  │  kubestellar    │ ─────────────────▶│  Workload Execution  │ │
│  │  (core engine)  │◀─── status sync ──│  Clusters (WEC)      │ │
│  │  WDS + ITS      │                   └──────────────────────┘ │
│  └────────┬────────┘                                             │
│           │ REST/Watch                                           │
│  ┌────────▼────────────────────────────────────────────────────┐ │
│  │  kubestellar/console   (AI-powered dashboard)               │ │
│  │  · AI Missions (natural-language cluster operations)        │ │
│  │  · Multi-cluster policy visualization                       │ │
│  │  · 32/32 nightly tests · v0.3.30-weekly cadence            │ │
│  └────────┬──────────────────┬──────────────────────────────┘ │
│           │                  │                                   │
│  ┌────────▼──────┐  ┌────────▼──────────────────────────────┐  │
│  │   console-    │  │   console-kb                           │  │
│  │   marketplace │  │   AI Mission Knowledge Base            │  │
│  │               │  │   · 2MB+ community fixes               │  │
│  │  153 card     │  │   · 12 categories (cncf-generated,     │  │
│  │  types across │  │     llm-d, orbit, multi-cluster, ...)  │  │
│  │  19 categories│  │   · Operational runbooks               │  │
│  └───────────────┘  └────────────────────────────────────────┘  │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  kubestellar-mcp   (AI Agent Integration)                │   │
│  │  · MCP server for Claude, Cursor, Windsurf, VS Code      │   │
│  │  · RBAC audit, drift detection, multi-cluster operations  │   │
│  │  · Homebrew: kubectl-claude, kubectl-claude-deploy        │   │
│  └──────────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────────┘
```

## Project Summary

| Repository | Role | Release Cadence | Latest |
|------------|------|----------------|--------|
| [kubestellar/kubestellar](https://github.com/kubestellar/kubestellar) | Core multi-cluster orchestration engine | On milestone | v0.30.0 |
| [kubestellar/console](https://github.com/kubestellar/console) | AI-powered multi-cluster dashboard | Weekly | v0.3.30 |
| [kubestellar/console-marketplace](https://github.com/kubestellar/console-marketplace) | Community dashboards, card presets, themes | Continuous | — |
| [kubestellar/console-kb](https://github.com/kubestellar/console-kb) | AI mission knowledge base (fixes + runbooks) | Continuous | — |
| [kubestellar/kubestellar-mcp](https://github.com/kubestellar/kubestellar-mcp) | MCP server for AI agent tooling | On milestone | v0.8.21 |
| [kubestellar/docs](https://github.com/kubestellar/docs) | Documentation site (kubestellar.io) | Continuous | — |
| [kubestellar/homebrew-tap](https://github.com/kubestellar/homebrew-tap) | Homebrew distribution for CLI tools | On release | — |

## Core Engine — kubestellar/kubestellar

The core orchestration engine provides:
- **BindingPolicy** — declarative workload placement rules across multiple clusters
- **Workload Description Space (WDS)** — where you describe what to deploy
- **ITS (In-cluster Transport Shim)** — pluggable transport layer (OCM, direct)
- **Workload Execution Cluster (WEC)** — target clusters receiving workloads

**Who should use the core engine**: platform teams managing multiple Kubernetes clusters with policy-driven workload placement.

## Console — kubestellar/console

An AI-powered multi-cluster Kubernetes dashboard featuring:
- **AI Missions** — describe cluster operations in natural language; the AI executes them
- **Multi-cluster visualization** — unified view of all WECs and BindingPolicy status
- **153 installable card types** via the marketplace
- **Weekly releases** — v0.3.x with a nightly test suite (32/32 passing as of 2026-06-08)

## Console Marketplace — kubestellar/console-marketplace

A community registry of installable content for the Console:

### Card Categories

| Category | Examples |
|----------|--------|
| Core Monitoring | cluster health, metrics, costs, network |
| Pods & Workloads | pod issues, deployment progress, workload monitor |
| GPU & ML | GPU fleet, utilization, LLM inference, ML jobs |
| ArgoCD / GitOps | sync status, health, Kustomization, drift detection |
| Security | OPA/Gatekeeper, Kyverno, Falco, Trivy, Kubescape |
| Cost | OpenCost, Kubecost |
| AI Inference (LLM-d) | request flow, KV cache, EPP routing, prefill/decode |
| AI Agents (Kagenti) | agent fleet, build pipelines, tool registry, topology |
| Games & Utilities | kubectl terminal, Doom, Chess, Weather, RSS |

All content is **config-only** — JSON files with no code execution.

### LLM-d Integration

KubeStellar Console has native monitoring for [LLM-d](https://github.com/llm-d/llm-d), a CNCF project for LLM inference disaggregation. The 12 LLM-d cards provide visibility into:
- Prefill/Decode disaggregation
- KV cache utilization
- EPP (Efficient Prefill Pooling) routing
- Inference benchmarks and AI insights

**Positioning**: KubeStellar is the only multi-cluster platform with native LLM inference monitoring — ideal for teams deploying AI workloads across GPU clusters.

### Kagenti AI Agent Platform Integration

KubeStellar Console includes [Kagenti](https://kagenti.io) platform monitoring cards, providing fleet-level visibility into AI agent deployments:
- Agent fleet overview and health
- Agent build pipelines
- Tool registry and discovery
- Agent security posture and topology

**Positioning**: KubeStellar + Kagenti = multi-cluster infrastructure for AI agent deployments.

## Console Knowledge Base — kubestellar/console-kb

A community hub for sharing AI mission fixes and operational runbooks:

### Fix Categories

| Category | Description |
|----------|-------------|
| `cncf-generated/` | Auto-generated fixes from CNCF project issue data |
| `cncf-install/` | Installation missions for CNCF projects (cert-manager, Cilium, etc.) |
| `llm-d/` | LLM-d stack operation and debugging missions |
| `multi-cluster/` | Cross-cluster deployment and federation patterns |
| `orbit/` | KubeStellar Orbit operational tasks |
| `security/` | RBAC, secrets, policy, CVE remediation |
| `troubleshooting/` | General diagnostic and recovery missions |
| `workloads/` | Deployment, rollout, storage patterns |

### Operational Runbooks

Pre-built runbooks for day-2 operations:
- KubeStellar controller install, upgrade, and rollback
- Certificate rotation, cluster upgrade, node drain
- Disaster recovery (etcd backup/restore, Velero)
- RBAC audit and remediation

## MCP Server — kubestellar/kubestellar-mcp

An [MCP (Model Context Protocol)](https://modelcontextprotocol.io) server that connects AI agents to KubeStellar multi-cluster operations:

**Supported AI clients**: Claude (Anthropic), Cursor, Windsurf, VS Code, and any MCP-compatible host.

**Install**:
```bash
brew install kubestellar/tap/kubectl-claude        # ops server
brew install kubestellar/tap/kubectl-claude-deploy # deploy server
```

**Capabilities**:
- Multi-cluster health checks and RBAC audit
- GitOps drift detection across clusters
- Gatekeeper/OPA policy inspection
- BindingPolicy status queries via natural language

**Why this matters**: `kubestellar-mcp` is the only MCP server bridging AI agents to a production multi-cluster Kubernetes orchestration layer. With MCP becoming the standard for AI tool integration (Anthropic, Google, OpenAI all adopting it), this is a strategic position.

## Relationship to Competing Projects

See [docs/competitive-positioning.md](competitive-positioning.md) for a detailed comparison. In brief:

| Project | Relationship to KubeStellar |
|---------|-----------------------------|
| OCM (Open Cluster Management) | **Complementary** — KubeStellar can use OCM as its transport layer |
| Karmada | **Alternative** — different architecture (no spoke agent, different policy model) |
| Liqo | **Different scope** — transparent compute offloading vs policy propagation |
| ArgoCD | **Complementary** — GitOps for delivery, KubeStellar for runtime multi-cluster policy |

## Community Channels

- **Slack**: [#kubestellar](https://cloud-native.slack.com/archives/C097094RZ3M) in CNCF Slack
- **Community calls**: see [kubestellar.io](https://kubestellar.io) for schedule
- **LFX Mentorship**: active slots available (see open issues tagged `LFX`)
- **ADOPTERS.md**: [Tell us you're using KubeStellar](../ADOPTERS.md) — organizations at any stage welcome

## CNCF Status

KubeStellar is a [CNCF Sandbox](https://www.cncf.io/projects/kubestellar/) project.

The KubeStellar ecosystem demonstrates CNCF community health across multiple dimensions:
- Multi-repository activity (7 repos, all actively maintained)
- Community contributions (console-kb has 2MB+ of community-contributed AI fixes)
- AI-native tooling (kubestellar-mcp, AI missions)
- Integration partnerships (LLM-d, Kagenti, OCM, ArgoCD)
