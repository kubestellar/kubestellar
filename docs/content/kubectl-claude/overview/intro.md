---
title: Introduction
description: AI-powered kubectl plugin for multi-cluster Kubernetes management
---

# kubectl-claude

AI-powered kubectl plugin for multi-cluster Kubernetes management, with built-in diagnostic tools.

## Overview

kubectl-claude provides two main interfaces:

1. **Claude Code Plugin** - Use natural language to query and manage Kubernetes clusters directly in Claude Code
2. **CLI Tool** - Traditional kubectl plugin with AI-powered natural language queries

## Features

- **Multi-cluster management** - Discover and manage clusters from your kubeconfig
- **Health monitoring** - Check cluster and node health status
- **Workload inspection** - View pods, deployments, services, and events
- **RBAC analysis** - Analyze permissions for users, groups, and service accounts
- **Diagnostic tools** - Find pod issues, deployment problems, security misconfigurations
- **OPA Gatekeeper integration** - Manage and monitor policy violations

## Quick Start

### Installation

#### Homebrew (Recommended)

```bash
brew tap kubestellar/tap
brew install kubectl-claude
```

#### From Releases

Download from [GitHub Releases](https://github.com/kubestellar/kubectl-claude/releases).

#### From Source

```bash
git clone https://github.com/kubestellar/kubectl-claude.git
cd kubectl-claude
go build -o kubectl-claude ./cmd/kubectl-claude
sudo mv kubectl-claude /usr/local/bin/
```

### Claude Code Plugin Setup

1. Add the KubeStellar marketplace:
   ```
   /plugin marketplace add kubestellar/claude-plugins
   ```
2. Go to `/plugin` → **Discover** tab
3. Install **kubectl-claude**

#### Verify Installation

Run `/mcp` in Claude Code - you should see:
```
plugin:kubectl-claude:kubectl-claude · ✓ connected
```

#### Allow Tools Without Prompts

To avoid permission prompts for each tool call, add to `~/.claude/settings.json`:

```json
{
  "permissions": {
    "allow": [
      "mcp__plugin_kubectl-claude_kubectl-claude__*"
    ]
  }
}
```

Or run in Claude Code:
```
/allowed-tools add mcp__plugin_kubectl-claude_kubectl-claude__*
```

### Usage in Claude Code

Once installed, ask questions like:

- "List my Kubernetes clusters"
- "Find pods with issues in the production namespace"
- "Check for security misconfigurations in my cluster"
- "What permissions does the admin service account have?"
- "Show me warning events in kube-system"
- "Analyze the default namespace"

## Available Tools

### Cluster Management
| Tool | Description |
|------|-------------|
| `list_clusters` | Discover clusters from kubeconfig |
| `get_cluster_health` | Check cluster health status |
| `get_nodes` | List cluster nodes with status |
| `audit_kubeconfig` | Audit all clusters for connectivity and recommend cleanup |

### Workload Tools
| Tool | Description |
|------|-------------|
| `get_pods` | List pods with filtering options |
| `get_deployments` | List deployments |
| `get_services` | List services |
| `get_events` | Get recent events |
| `describe_pod` | Get detailed pod information |
| `get_pod_logs` | Retrieve pod logs |

### RBAC Analysis
| Tool | Description |
|------|-------------|
| `get_roles` | List Roles in a namespace |
| `get_cluster_roles` | List ClusterRoles |
| `get_role_bindings` | List RoleBindings |
| `get_cluster_role_bindings` | List ClusterRoleBindings |
| `can_i` | Check if you can perform an action |
| `analyze_subject_permissions` | Full RBAC analysis for any subject |
| `describe_role` | Detailed view of Role/ClusterRole rules |

### Diagnostic Tools
| Tool | Description |
|------|-------------|
| `find_pod_issues` | Find CrashLoopBackOff, ImagePullBackOff, OOMKilled, pending pods |
| `find_deployment_issues` | Find stuck rollouts, unavailable replicas, ReplicaSet errors |
| `check_resource_limits` | Find pods without CPU/memory limits |
| `check_security_issues` | Find privileged containers, root users, host network |
| `analyze_namespace` | Comprehensive namespace analysis |
| `get_warning_events` | Get only Warning events |
| `find_resource_owners` | Find who owns/manages resources via managedFields, labels, annotations |

### OPA Gatekeeper Policy Tools
| Tool | Description |
|------|-------------|
| `check_gatekeeper` | Check if OPA Gatekeeper is installed and healthy |
| `get_ownership_policy_status` | Get ownership policy configuration and violation count |
| `list_ownership_violations` | List resources missing required ownership labels |
| `install_ownership_policy` | Install ownership labels policy (dryrun/warn/enforce modes) |
| `set_ownership_policy_mode` | Change policy enforcement mode |
| `uninstall_ownership_policy` | Remove the ownership policy |

## CLI Usage

### As kubectl plugin

```bash
# List all clusters
kubectl claude clusters list

# Check cluster health
kubectl claude clusters health

# Natural language queries (requires ANTHROPIC_API_KEY)
kubectl claude "show me failing pods"
```

### As MCP Server

```bash
# Start MCP server (used by Claude Code)
kubectl-claude --mcp-server
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `KUBECONFIG` | Path to kubeconfig file |
| `ANTHROPIC_API_KEY` | API key for Claude AI (for natural language queries) |

## Contributing

Contributions are welcome! Please read our [contributing guidelines](https://github.com/kubestellar/kubectl-claude/blob/main/CONTRIBUTING.md).

## License

Apache License 2.0 - see [LICENSE](https://github.com/kubestellar/kubectl-claude/blob/main/LICENSE) for details.
