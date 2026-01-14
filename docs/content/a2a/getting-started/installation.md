---
sidebar_position: 2
---

# Installation

This guide will help you install KubeStellar A2A on your system with all required dependencies.

## System Requirements

### Minimum Requirements
- **Python**: 3.11 or higher
- **Memory**: 512MB available RAM
- **Storage**: 100MB free disk space
- **Network**: Internet access for package downloads

### Recommended Requirements
- **Python**: 3.12 or higher
- **Memory**: 2GB available RAM
- **Storage**: 1GB free disk space
- **kubectl**: Configured with at least one Kubernetes cluster
- **Helm**: Version 3.x for advanced deployment features

## Installation Methods

### Method 1: Using uv (Recommended)

[uv](https://github.com/astral-sh/uv) is the fastest Python package installer and project manager.

```bash
# Install uv if you haven't already
curl -LsSf https://astral.sh/uv/install.sh | sh

# Clone the repository
git clone https://github.com/kubestellar/a2a.git
cd a2a

# Install KubeStellar A2A with all dependencies
uv pip install -e ".[dev]"

# Install kubectl plugin alias (a2a) and optional kubestellar name
uv tool install .
# Verify alias
which kubectl-a2a
kubectl a2a --help
# Optional: create kubestellar plugin name via symlink to the Python entrypoint
ln -sf "$(command -v kubestellar)" ~/.local/bin/kubectl-kubestellar
kubectl kubestellar --help
```

### Method 2: Using pip

```bash
# Clone the repository
git clone https://github.com/kubestellar/a2a.git
cd a2a

# Create and activate virtual environment
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install KubeStellar A2A
pip install -e .

# Optionally install as a kubectl plugin (alias a2a)
python -m pip install .
# Ensure your Python scripts dir is on PATH (e.g., ~/.local/bin)
kubectl a2a --help
```

### Method 3: Development Installation

For contributors and developers:

```bash
# Clone the repository
git clone https://github.com/kubestellar/a2a.git
cd a2a

# Install with development dependencies
uv pip install -e ".[dev,test]"

# Install pre-commit hooks (optional)
pre-commit install
```

## Verify Installation

Test your installation to ensure everything is working correctly:

```bash
# Check CLI installation
uv run kubestellar --help

# List available functions
uv run kubestellar list-functions

# Test basic functionality
uv run kubestellar execute get_kubeconfig

# Verify kubectl plugin
kubectl a2a --help
kubectl kubestellar --help || true
```

## CLI Commands Overview

### Main Commands

```bash
# Show help
uv run kubestellar --help

# List all available functions
uv run kubestellar list-functions

# Execute a function with parameters
uv run kubestellar execute <function_name> [parameters]

# Get detailed function description and schema
uv run kubestellar describe <function_name>

# Start interactive agent mode
uv run kubestellar agent
```

### Function Execution Examples

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

### Available Functions

```
- kubestellar_management
  Description: Advanced KubeStellar multi-cluster resource management with deep search capabilities
  
- get_kubeconfig
  Description: Get details from kubeconfig file including contexts, clusters, and users
  
- helm_deploy
  Description: Deploy Helm charts across clusters with KubeStellar binding policy integration
  
- namespace_utils  
  Description: List and count pods, services, deployments and other resources across namespaces
  
- gvrc_discovery
  Description: Discover and inventory all available Kubernetes API resources across clusters
  
- multicluster_create
  Description: Create Kubernetes resources across multiple clusters
  
- multicluster_logs
  Description: Aggregate and stream logs from multiple clusters
  
- deploy_to
  Description: Deploy resources to specific clusters with advanced targeting
```

## Optional Components

### AI Features Setup

For AI-powered automation and natural language interfaces:

#### OpenAI Integration
```bash
# Set your OpenAI API key
export OPENAI_API_KEY="your-openai-api-key"

# Test agent mode
uv run kubestellar agent
```

##### Agent Mode Interface

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

Provider: openai
Model: gpt-4o

Type 'help' for available commands
Type 'exit' or Ctrl+D to quit

[openai] ▶ 
```

##### Agent Commands

```bash
# Natural language queries
[openai] ▶ how many pods are running?
[openai] ▶ show me kubestellar topology
[openai] ▶ deploy nginx using helm to production clusters
[openai] ▶ check binding policy status

# Built-in commands
help          # Show available commands
clear         # Clear conversation history
provider <name>  # Switch AI provider
exit          # Exit the agent
```

#### Claude MCP Integration
Add to your Claude Desktop configuration (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "kubestellar": {
      "command": "uv",
      "args": ["run", "kubestellar-mcp"],
      "cwd": "/path/to/a2a"
    }
  }
}
```

### KubeStellar Integration

For full KubeStellar 2024 architecture support:

```bash
# Ensure you have KubeStellar installed
# Follow the official KubeStellar installation guide
# https://docs.kubestellar.io/

# Verify KubeStellar is accessible
kubectl get wds --all-namespaces
kubectl get its --all-namespaces
```

## Configuration

### Environment Variables

Set up common environment variables for seamless operation:

```bash
# Add to your shell profile (.bashrc, .zshrc, etc.)

# Kubernetes configuration
export KUBECONFIG="$HOME/.kube/config"

# AI provider - OpenAI (currently only supported provider)
export OPENAI_API_KEY="your-openai-key"

# Logging level
export LOG_LEVEL="INFO"
```

### Configuration File

Create a configuration file at `~/.kube/a2a-config.yaml`:

```yaml
# AI Provider Configuration (OpenAI only currently supported)
providers:
  openai:
    api_key: "${OPENAI_API_KEY}"
    model: "gpt-4o"
    temperature: 0.7
    
default_provider: "openai"

# UI Configuration
ui:
  show_thinking: true
  show_token_usage: true
  
# Cluster Configuration
clusters:
  default_timeout: "5m"
  auto_discovery: true
```

**Note:** Additional AI providers (Claude, Gemini, etc.) will be added in future releases.

## Troubleshooting Installation

### Common Issues

#### Python Version Issues
```bash
# Check Python version
python --version

# If using pyenv, set local version
pyenv local 3.12
```

#### Permission Issues
```bash
# On macOS/Linux, you might need to use sudo for system-wide installation
sudo uv pip install -e .

# Or install to user directory
uv pip install --user -e .
```

#### Network Issues
```bash
# If behind corporate proxy
export HTTP_PROXY="http://proxy.company.com:8080"
export HTTPS_PROXY="http://proxy.company.com:8080"

# Install with proxy settings
uv pip install -e . --proxy http://proxy.company.com:8080
```

#### Kubernetes Configuration Issues
```bash
# Verify kubectl is working
kubectl cluster-info

# List available contexts
kubectl config get-contexts

# Test kubeconfig access
uv run kubestellar execute get_kubeconfig --param detail_level=full
```

### Getting Help

If you encounter issues:

1. **Check the logs**: Set `LOG_LEVEL=DEBUG` for detailed output
2. **Verify prerequisites**: Ensure all requirements are met
3. **Update dependencies**: Run `uv pip install --upgrade -e .`
4. **Report issues**: [GitHub Issues](https://github.com/kubestellar/a2a/issues)

## Next Steps

Once installation is complete:

1. **[Quick Start Guide →](./quick-start)** - Get up and running in 5 minutes
2. **[Troubleshooting Guide →](../troubleshooting)** - Resolve common issues
3. **[GitHub Repository →](https://github.com/kubestellar/a2a)** - Source code and issues

---

*Installation complete! Ready to revolutionize your Kubernetes management? 🚀*
