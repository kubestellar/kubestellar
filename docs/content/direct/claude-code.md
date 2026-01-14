# Claude Code Integration

Use Claude Code's AI capabilities to manage multiple Kubernetes clusters simultaneously with natural language.

## Overview

kubectl-claude is an AI-powered kubectl plugin designed for **managing multiple clusters simultaneously**. It integrates with Claude Code, enabling you to:

- Query Kubernetes resources across multiple clusters using natural language
- Diagnose pod issues (CrashLoopBackOff, OOMKilled, pending pods)
- Analyze RBAC permissions for any subject
- Detect security misconfigurations
- Monitor events and cluster health

## Installation

### Option 1: Homebrew (Recommended)

```bash
brew tap kubestellar/tap
brew install kubectl-claude
```

### Option 2: Claude Code Plugin Marketplace

In Claude Code, run:

```
/plugin marketplace add kubestellar/claude-plugins
```

Then go to `/plugin` → **Discover** tab and install **kubectl-claude**.

### Verify Installation

Run `/mcp` in Claude Code to see connected MCP servers:

```
plugin:kubectl-claude:kubectl-claude · ✓ connected
```

## Configuration

### Allow Tools Without Prompts

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

## Slash Commands

The kubectl-claude plugin provides specialized slash commands for common Kubernetes operations:

### /k8s-health

Check the health of all Kubernetes clusters in your kubeconfig.

```
/k8s-health
```

This will:
1. Discover all available clusters
2. Check health status of each cluster
3. Summarize any issues with recommended actions

### /k8s-issues

Find all issues across your Kubernetes clusters.

```
/k8s-issues
```

Checks for:
- Pod issues (CrashLoopBackOff, ImagePullBackOff, OOMKilled, pending)
- Deployment issues (stuck rollouts, unavailable replicas)
- Warning events

### /k8s-analyze

Perform a comprehensive analysis of a Kubernetes namespace.

```
/k8s-analyze
```

Provides insights on:
- Workload health (pods, deployments)
- Resource usage and limits
- Potential issues and recommendations

### /k8s-rbac

Analyze RBAC permissions for any subject (user, group, or service account).

```
/k8s-rbac
```

Shows:
- All RoleBindings and ClusterRoleBindings
- Effective permissions
- Overly permissive access warnings
- Security recommendations

### /k8s-security

Perform a security audit across all clusters.

```
/k8s-security
```

Finds:
- Privileged containers
- Containers running as root
- Host network/PID/IPC usage
- Pods without resource limits
- Security misconfigurations

### /k8s-audit-kubeconfig

Audit your kubeconfig and clean up stale clusters.

```
/k8s-audit-kubeconfig
```

Shows:
- Which clusters are accessible (with versions)
- Which clusters are inaccessible
- Cleanup commands for stale configurations

## Usage Examples

Once installed, ask Claude questions like:

### Cluster Management

- "List my Kubernetes clusters"
- "Check cluster health"
- "Show me nodes in the production cluster"
- "Audit my kubeconfig for stale clusters"

### Workload Diagnostics

- "Find pods with issues in the production namespace"
- "Show me CrashLoopBackOff pods"
- "What's wrong with the frontend deployment?"
- "Get logs from the api-server pod"

### RBAC Analysis

- "What permissions does the admin service account have?"
- "Can I create deployments in the default namespace?"
- "Show me all cluster roles"
- "Analyze RBAC for user john@example.com"

### Security

- "Check for security misconfigurations in my cluster"
- "Find pods running as root"
- "Show me containers with privileged access"
- "Check resource limits across namespaces"

### Events and Monitoring

- "Show me warning events in kube-system"
- "What events happened in the last hour?"
- "Analyze the production namespace"

## Available Tools

### Cluster Tools

| Tool | Description |
|------|-------------|
| `list_clusters` | Discover clusters from kubeconfig |
| `get_cluster_health` | Check cluster health status |
| `get_nodes` | List cluster nodes with status |
| `audit_kubeconfig` | Audit all clusters for connectivity |

### Workload Tools

| Tool | Description |
|------|-------------|
| `get_pods` | List pods with filtering options |
| `get_deployments` | List deployments |
| `get_services` | List services |
| `get_events` | Get recent events |
| `describe_pod` | Get detailed pod information |
| `get_pod_logs` | Retrieve pod logs |

### RBAC Tools

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
| `find_pod_issues` | Find CrashLoopBackOff, OOMKilled, pending pods |
| `find_deployment_issues` | Find stuck rollouts, unavailable replicas |
| `check_resource_limits` | Find pods without CPU/memory limits |
| `check_security_issues` | Find privileged containers, root users |
| `analyze_namespace` | Comprehensive namespace analysis |
| `get_warning_events` | Get only Warning events |

## CLI Usage

kubectl-claude also works as a standalone kubectl plugin:

```bash
# List all clusters
kubectl claude clusters list

# Check cluster health
kubectl claude clusters health

# Natural language queries (requires ANTHROPIC_API_KEY)
kubectl claude "show me failing pods"
```

## Links

- [kubectl-claude on GitHub](https://github.com/kubestellar/kubectl-claude)
- [Claude Plugins Marketplace](https://github.com/kubestellar/claude-plugins)
- [Homebrew Tap](https://github.com/kubestellar/homebrew-tap)
