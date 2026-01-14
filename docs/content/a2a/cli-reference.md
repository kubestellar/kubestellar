---
sidebar_position: 4
---

# CLI and kubectl Plugin Reference

The KubeStellar CLI provides powerful command-line access for Kubernetes multi-cluster management.

In addition to the `kubestellar` CLI, you can use KubeStellar as a kubectl plugin via executables named `kubectl-<name>` on your PATH.

- Primary plugin name: `kubestellar` (binary and Krew). Executable: `kubectl-kubestellar` → usage: `kubectl kubestellar ...`.
- Python-installed alias: `a2a`. Executable: `kubectl-a2a` → usage: `kubectl a2a ...`.

```
╭─────────────────────────────────────────────────────────────────────────────────────────────╮
│  ██╗  ██╗██╗   ██╗██████╗ ███████╗███████╗████████╗███████╗██╗     ██╗      █████╗ ██████╗  │
│  ██║ ██╔╝██║   ██║██╔══██╗██╔════╝██╔════╝╚══██╔══╝██╔════╝██║     ██║     ██╔══██╗██╔══██╗ │
│  █████╔╝ ██║   ██║██████╔╝█████╗  ███████╗   ██║   █████╗  ██║     ██║     ███████║██████╔╝ │
│  ██╔═██╗ ██║   ██║██╔══██╗██╔══╝  ╚════██║   ██║   ██╔══╝  ██║     ██║     ██╔══██║██╔══██╗ │
│  ██║  ██╗╚██████╔╝██████╔╝███████╗███████║   ██║   ███████╗███████╗███████╗██║  ██║██║  ██║ │
│  ╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚══════╝╚══════╝   ╚═╝   ╚══════╝╚══════╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝ │
│                       Multi-Cluster Kubernetes Management Agent                             │
╰─────────────────────────────────────────────────────────────────────────────────────────────╯
```

## Installation

```bash
# Install with uv
uv pip install -e ".[dev]"
```

## Commands

### Basic Commands

```bash
# Show help
uv run kubestellar --help

# List all available functions
uv run kubestellar list-functions

# Execute a specific function
uv run kubestellar execute <function_name>

# Describe a function (show parameters and schema)
uv run kubestellar describe <function_name>

# Start interactive AI agent
uv run kubestellar agent
```

### Function Execution

Execute functions with parameters using multiple syntax options:

```bash
# Using --param flag
uv run kubestellar execute get_kubeconfig --param context=production --param detail_level=full

# Using -P shorthand (recommended)
uv run kubestellar execute get_kubeconfig -P context=staging -P detail_level=contexts

# Using JSON parameters
uv run kubestellar execute get_kubeconfig --params '{"context": "production", "detail_level": "full"}'

# Complex array parameters
uv run kubestellar execute namespace_utils -P target_namespaces='["prod","staging"]' -P all_namespaces=true
```

### Interactive Agent Mode

The agent provides natural language interface for cluster management:

```bash
# Start the agent
uv run kubestellar agent

# Example queries in agent mode:
[openai] ▶ how many pods are running?
[openai] ▶ show me kubestellar topology
[openai] ▶ deploy nginx using helm to production clusters
[openai] ▶ check binding policy status
```

Agent commands:
- `help` - Show available commands
- `clear` - Clear conversation history
- `provider <name>` - Switch AI provider
- `exit` - Exit the agent

## kubectl Plugin Examples

```bash
kubectl kubestellar --help
kubectl kubestellar list-functions
kubectl kubestellar execute kubestellar_management -P operation=deep_search

# alias
kubectl a2a providers
```

Install methods are detailed in “Getting Started → Installation”. For Krew, use the `kubestellar.yaml` manifest attached to a release, or submit it to the central krew-index to enable `kubectl krew install kubestellar`.

## Available Functions

### Core Functions

- **get_kubeconfig** - Analyze kubeconfig file
- **kubestellar_management** - Multi-cluster resource management
- **helm_deploy** - Deploy Helm charts with binding policies
- **namespace_utils** - Manage namespaces across clusters
- **gvrc_discovery** - Discover API resources

### Multi-Cluster Functions

- **multicluster_create** - Create resources across clusters
- **multicluster_logs** - Aggregate logs from multiple clusters
- **deploy_to** - Deploy to specific clusters

## Examples

### Get Cluster Information
```bash
# Get current context
uv run kubestellar execute get_kubeconfig

# Get full details
uv run kubestellar execute get_kubeconfig -P detail_level=full
```

### Deploy Applications
```bash
# Deploy Helm chart
uv run kubestellar execute helm_deploy \
  -P chart_name=nginx \
  -P repository_url=https://charts.bitnami.com/bitnami \
  -P target_clusters='["prod-cluster"]'

# Create deployment across namespaces
uv run kubestellar execute multicluster_create \
  -P resource_type=deployment \
  -P resource_name=web-app \
  -P image=nginx:1.21 \
  -P all_namespaces=true
```

### Resource Discovery
```bash
# Discover all resources
uv run kubestellar execute gvrc_discovery

# List all namespaces
uv run kubestellar execute namespace_utils \
  -P operation=list \
  -P all_namespaces=true
```

## Configuration

### Agent Configuration

Configure AI provider in `~/.kube/a2a-config.yaml`:

```yaml
# OpenAI is currently the only supported provider
providers:
  openai:
    api_key: "your-openai-key"
    model: "gpt-4o"
    temperature: 0.7
    
default_provider: "openai"

ui:
  show_thinking: true
  show_token_usage: true
```

Or use environment variables:
```bash
export OPENAI_API_KEY="your-key"
```

**Note:** Additional AI providers (Claude, Gemini, etc.) will be added in future releases.
