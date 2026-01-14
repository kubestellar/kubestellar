# Installation Guide

This guide covers how to install and set up kubectl-multi on your system.

## Prerequisites

- Go 1.21 or later
- kubectl installed and configured
- Access to KubeStellar managed clusters

## Installation Methods

### Method 1: Build and Install (Recommended)

```bash
# Clone the repository
git clone <repository-url>
cd kubectl-multi

# Build and install as kubectl plugin
make install

# Or install system-wide
make install-system
```

### Method 2: Manual Installation

```bash
# Build binary
make build

# Copy to PATH
cp bin/kubectl-multi ~/.local/bin/
chmod +x ~/.local/bin/kubectl-multi

# Verify installation
kubectl plugin list | grep multi
```

### Method 3: Go Install (if available)

```bash
# Install directly with Go
go install <repository-url>@latest

# Make sure your GOPATH/bin is in your PATH
export PATH=$PATH:$(go env GOPATH)/bin
```

## Verification

After installation, verify that the plugin is working correctly:

```bash
# Check if kubectl recognizes the plugin
kubectl plugin list | grep multi

# Test basic functionality
kubectl multi get nodes

# Check version (if implemented)
kubectl multi version
```

## Configuration

### Kubeconfig Setup

kubectl-multi uses your existing kubectl configuration. Make sure you have:

1. Valid kubeconfig file (usually `~/.kube/config`)
2. Access to your KubeStellar ITS cluster
3. Proper RBAC permissions for managed clusters

### Required Permissions

The plugin needs the following permissions on your ITS cluster:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubectl-multi-reader
rules:
- apiGroups: ["cluster.open-cluster-management.io"]
  resources: ["managedclusters"]
  verbs: ["get", "list"]
```

And on managed clusters:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubectl-multi-reader
rules:
- apiGroups: [""]
  resources: ["*"]
  verbs: ["get", "list"]
- apiGroups: ["apps"]
  resources: ["*"]
  verbs: ["get", "list"]
# Add other API groups as needed
```

## Troubleshooting Installation

### Common Issues

#### Plugin Not Found
```bash
Error: unknown command "multi" for "kubectl"
```

**Solution:** 
- Ensure the binary is named `kubectl-multi` and is in your PATH
- Run `kubectl plugin list` to verify the plugin is detected

#### Permission Denied
```bash
permission denied: kubectl-multi
```

**Solution:**
```bash
chmod +x /path/to/kubectl-multi
```

#### Go Build Issues
```bash
go: module not found
```

**Solution:**
```bash
# Ensure Go modules are initialized
go mod tidy
go mod download
```

#### Connection Issues
```bash
Error: failed to connect to ITS cluster
```

**Solution:**
- Verify your kubeconfig is correct
- Check if you can connect with regular kubectl: `kubectl get nodes`
- Verify the ITS cluster context exists

### Build Dependencies

If you're building from source, you may need additional tools:

```bash
# Install make (if not available)
# On macOS
brew install make

# On Ubuntu/Debian
apt-get install build-essential

# On CentOS/RHEL
yum groupinstall "Development Tools"
```

## Updating

To update kubectl-multi to the latest version:

```bash
# Pull latest changes
git pull origin main

# Rebuild and install
make install
```

## Uninstallation

To remove kubectl-multi:

```bash
# Find the plugin location
which kubectl-multi

# Remove the binary
rm /path/to/kubectl-multi

# Or if installed via make
make uninstall
```

## Next Steps

After successful installation:
1. Read the [Usage Guide](usage_guide.md) for detailed examples
2. Check the [Architecture Guide](architecture_guide.md) to understand how it works
3. See [Development Guide](development_guide.md) if you want to contribute