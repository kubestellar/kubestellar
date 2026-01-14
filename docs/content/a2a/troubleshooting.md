---
sidebar_position: 4
---

# Troubleshooting

This comprehensive guide will help you resolve common issues when using KubeStellar A2A. Follow the troubleshooting steps to quickly identify and fix problems.

## Quick Diagnostic Commands

Start with these commands to gather basic system information:

```bash
# Check KubeStellar A2A installation
uv run kubestellar --version

# Verify function registry
uv run kubestellar list-functions

# Test basic connectivity
uv run kubestellar execute get_kubeconfig

# Check logs with debug level
LOG_LEVEL=DEBUG uv run kubestellar execute get_kubeconfig
```

## Installation Issues

### Python Version Problems

**Problem**: `Python version too old` or `SyntaxError` during installation

**Solution**:
```bash
# Check current Python version
python --version

# Install Python 3.11+ using pyenv (recommended)
curl https://pyenv.run | bash
pyenv install 3.12
pyenv global 3.12

# Or use your system package manager
# Ubuntu/Debian:
sudo apt update && sudo apt install python3.12 python3.12-venv

# macOS with Homebrew:
brew install python@3.12
```

### Package Installation Failures

**Problem**: `Failed building wheel` or `pip install` errors

**Solution**:
```bash
# Update pip and setuptools
pip install --upgrade pip setuptools wheel

# Clear pip cache
pip cache purge

# Install with verbose output to see detailed errors
pip install -e . -v

# For macOS compilation issues
export CFLAGS=-Wno-error=incompatible-function-pointer-types

# For Linux missing headers
sudo apt-get install python3-dev build-essential
```

### uv Installation Issues

**Problem**: `uv: command not found` or installation fails

**Solution**:
```bash
# Install uv using the official installer
curl -LsSf https://astral.sh/uv/install.sh | sh

# Reload shell configuration
source ~/.bashrc  # or ~/.zshrc

# Alternatively, install via pip
pip install uv

# For Windows
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"
```

## Kubernetes Connectivity Issues

### Kubeconfig Not Found

**Problem**: `Kubeconfig file not found` or `No clusters found`

**Solution**:
```bash
# Check if kubectl is configured
kubectl cluster-info

# Verify kubeconfig file exists
ls -la ~/.kube/config

# Set KUBECONFIG environment variable explicitly
export KUBECONFIG="$HOME/.kube/config"

# Test with specific kubeconfig path
uv run kubestellar execute get_kubeconfig \
  --param kubeconfig_path="/path/to/your/kubeconfig"
```

### Cluster Access Denied

**Problem**: `Forbidden` or `Access denied` errors

**Solution**:
```bash
# Check current context and permissions
kubectl config current-context
kubectl auth can-i "*" "*" --all-namespaces

# Switch to admin context if available
kubectl config get-contexts
kubectl config use-context admin-context

# Verify cluster-admin role
kubectl get clusterrolebindings | grep cluster-admin
```

### Multiple Cluster Configuration

**Problem**: Issues with multiple clusters or contexts

**Solution**:
```bash
# List all contexts
kubectl config get-contexts

# Test each context individually
for context in $(kubectl config get-contexts -o name); do
  echo "Testing context: $context"
  kubectl --context="$context" cluster-info
done

# Merge multiple kubeconfig files
KUBECONFIG=config1:config2:config3 kubectl config view --merge --flatten > merged-config
export KUBECONFIG="merged-config"
```

## Function Execution Issues

### Function Not Found

**Problem**: `Function 'function_name' not found`

**Solution**:
```bash
# List all available functions
uv run kubestellar list-functions

# Check exact function name (case-sensitive)
uv run kubestellar describe function_name

# Verify installation is complete
uv pip install -e . --force-reinstall
```

### Parameter Validation Errors

**Problem**: `Invalid parameter` or `Schema validation failed`

**Solution**:
```bash
# Get function schema and parameter requirements
uv run kubestellar describe function_name

# Use correct parameter format
uv run kubestellar execute function_name -P param1=value1 -P param2=value2

# For complex parameters, use JSON format
uv run kubestellar execute function_name \
  --params '{"param1": "value1", "param2": ["item1", "item2"]}'

# Debug parameter parsing
LOG_LEVEL=DEBUG uv run kubestellar execute function_name -P param=value
```

### Timeout Issues

**Problem**: Functions timeout or hang indefinitely

**Solution**:
```bash
# Check cluster connectivity
kubectl cluster-info

# Increase timeout for long-running operations
uv run kubestellar execute helm_deploy \
  -P timeout=10m \
  -P chart_name=large-application

# Use async operations for multiple clusters
uv run kubestellar execute multicluster_create \
  -P resource_type=configmap \
  -P dry_run=true  # Test without actual creation first
```

## Helm Integration Issues

### Helm Charts Not Found

**Problem**: `Chart not found` or `Repository errors`

**Solution**:
```bash
# Update Helm repositories
helm repo update

# Add missing repositories
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo add stable https://charts.helm.sh/stable

# Search for charts
helm search repo chart-name

# Test with local chart path
uv run kubestellar execute helm_deploy \
  -P chart_path=./local-chart \
  -P release_name=test-release
```

### Helm Permission Issues

**Problem**: `Insufficient permissions` for Helm operations

**Solution**:
```bash
# Create service account for Helm (if using RBAC)
kubectl create serviceaccount helm-service-account
kubectl create clusterrolebinding helm-cluster-admin \
  --clusterrole=cluster-admin \
  --serviceaccount=default:helm-service-account

# Verify Helm permissions
helm version
helm list --all-namespaces
```

## KubeStellar Integration Issues

### KubeStellar Resources Not Found

**Problem**: `WDS not found` or `ITS not accessible`

**Solution**:
```bash
# Verify KubeStellar installation
kubectl api-resources | grep kubestellar

# Check WDS and ITS spaces
kubectl get wds --all-namespaces
kubectl get its --all-namespaces

# Verify KubeStellar version compatibility
kubectl get crd | grep kubestellar
```

### Binding Policy Issues

**Problem**: `BindingPolicy creation failed` or `Policy not applied`

**Solution**:
```bash
# Check existing binding policies
kubectl get bindingpolicies --all-namespaces

# Verify policy syntax
kubectl describe bindingpolicy policy-name

# Test policy creation manually
cat << EOF | kubectl apply -f -
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: test-policy
spec:
  clusterSelectors:
  - matchLabels:
      environment: test
EOF
```

## AI Integration Issues

### OpenAI API Issues

**Problem**: `OpenAI API key invalid` or `Rate limit exceeded`

**Solution**:
```bash
# Verify API key format
echo $OPENAI_API_KEY | wc -c  # Should be ~51 characters

# Test API connectivity
curl -H "Authorization: Bearer $OPENAI_API_KEY" \
  https://api.openai.com/v1/models

# Set rate limiting and retry configuration
export A2A_OPENAI_MAX_RETRIES=3
export A2A_OPENAI_TIMEOUT=30

# Use different model if quota exceeded
uv run kubestellar agent --model gpt-3.5-turbo
```

### Claude MCP Server Issues

**Problem**: MCP server not connecting to Claude Desktop

**Solution**:
```bash
# Check Claude Desktop configuration file location
# macOS: ~/Library/Application Support/Claude/claude_desktop_config.json
# Windows: %APPDATA%/Claude/claude_desktop_config.json

# Verify configuration syntax
cat ~/Library/Application\ Support/Claude/claude_desktop_config.json | jq .

# Test MCP server manually
uv run kubestellar-mcp

# Check Claude Desktop logs
# macOS: ~/Library/Logs/Claude/claude_desktop.log
tail -f ~/Library/Logs/Claude/claude_desktop.log
```

## Performance Issues

### Slow Function Execution

**Problem**: Functions take too long to execute

**Solution**:
```bash
# Enable performance profiling
export A2A_ENABLE_PROFILING=true
LOG_LEVEL=DEBUG uv run kubestellar execute function_name

# Use targeted cluster operations
uv run kubestellar execute multicluster_create \
  -P target_clusters='["specific-cluster"]' \
  -P resource_type=configmap

# Implement caching for repeated operations
export A2A_ENABLE_CACHE=true
```

### Memory Issues

**Problem**: High memory usage or out-of-memory errors

**Solution**:
```bash
# Monitor memory usage
python -c "
import psutil
print(f'Memory usage: {psutil.virtual_memory().percent}%')
"

# Use streaming for large datasets
uv run kubestellar execute multicluster_logs \
  -P follow=false \
  -P tail=100  # Limit log lines

# Process clusters in batches
uv run kubestellar execute namespace_utils \
  -P target_clusters='["cluster1"]'  # Process one at a time
```

## Debug Mode and Logging

### Enable Debug Logging

For detailed troubleshooting information:

```bash
# Enable debug logging globally
export LOG_LEVEL=DEBUG

# Enable debug for specific components
export A2A_DEBUG_FUNCTIONS=true
export A2A_DEBUG_KUBERNETES=true
export A2A_DEBUG_HELM=true

# Save debug output to file
LOG_LEVEL=DEBUG uv run kubestellar execute function_name 2>&1 | tee debug.log
```

### Common Debug Patterns

```bash
# Test individual components
uv run kubestellar execute get_kubeconfig -P detail_level=full

# Validate configuration before execution
uv run kubestellar execute helm_deploy -P dry_run=true -P chart_name=test

# Check resource availability
uv run kubestellar execute gvrc_discovery -P output_format=detailed

# Monitor real-time operations
uv run kubestellar execute multicluster_logs -P follow=true -P tail=10
```

## Getting More Help

### Community Support

- **GitHub Issues**: [Report bugs and request features](https://github.com/kubestellar/a2a/issues)
- **GitHub Discussions**: [Ask questions and share experiences](https://github.com/kubestellar/a2a/discussions)
- **KubeStellar Community**: [Join the broader KubeStellar community](https://kubestellar.io/community)

### Providing Debug Information

When reporting issues, include:

1. **System Information**:
   ```bash
   uv run kubestellar --version
   python --version
   kubectl version --client
   ```

2. **Configuration**:
   ```bash
   kubectl config current-context
   kubectl cluster-info
   ```

3. **Debug Logs**:
   ```bash
   LOG_LEVEL=DEBUG uv run kubestellar execute function_name 2>&1 | tee issue-debug.log
   ```

4. **Function Schema**:
   ```bash
   uv run kubestellar describe function_name
   ```

### Known Issues and Workarounds

#### Issue: Functions fail with "Resource not found"
**Workaround**: Ensure all required CRDs are installed and accessible
```bash
kubectl api-resources | grep -E "(kubestellar|argo|helm)"
```

#### Issue: Multi-cluster operations partially fail
**Workaround**: Use `target_clusters` to specify working clusters explicitly
```bash
uv run kubestellar execute multicluster_create \
  -P target_clusters='["working-cluster1", "working-cluster2"]'
```

#### Issue: Large cluster environments cause timeouts
**Workaround**: Process clusters in smaller batches
```bash
# Process 2 clusters at a time
for batch in cluster1,cluster2 cluster3,cluster4; do
  uv run kubestellar execute namespace_utils \
    -P target_clusters="[\"${batch//,/\",\"}\"]"
done
```

---

*Still having issues? Don't hesitate to reach out to our community for help! 🤝*