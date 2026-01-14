# Usage Guide

This guide provides comprehensive examples and usage patterns for kubectl-multi.

## Basic Commands

### Get Resources

```bash
# Get nodes from all managed clusters
kubectl multi get nodes

# Get pods from all clusters in all namespaces
kubectl multi get pods -A

# Get services in specific namespace
kubectl multi get services -n kube-system

# Get deployments with label selector
kubectl multi get deployments -l app=nginx -A

# Show labels
kubectl multi get pods --show-labels -n kube-system
```

### Global Flags

- `--kubeconfig string`: Path to kubeconfig file
- `--remote-context string`: Remote hosting context (default: "its1")
- `--all-clusters`: Operate on all managed clusters (default: true)
- `-n, --namespace string`: Target namespace
- `-A, --all-namespaces`: List resources across all namespaces

## Output Examples

### Sample Input and Output

#### Getting Nodes
```bash
kubectl multi get nodes
```

**Output:**
```
CONTEXT  CLUSTER       NAME                    STATUS  ROLES          AGE    VERSION
its1     cluster1      cluster1-control-plane  Ready   control-plane  6d23h  v1.33.1
its1     cluster2      cluster2-control-plane  Ready   control-plane  6d23h  v1.33.1
its1     its1-cluster  kubeflex-control-plane  Ready   <none>         6d23h  v1.27.2+k3s1
```

#### Getting Pods with Namespace
```bash
kubectl multi get pods -n kube-system
```

**Output:**
```
CONTEXT  CLUSTER       NAME                                            READY  STATUS   RESTARTS  AGE
its1     cluster1      coredns-674b8bbfcf-6k7vc                        1/1    Running  2         6d23h
its1     cluster1      etcd-cluster1-control-plane                     1/1    Running  2         6d23h
its1     cluster1      kube-apiserver-cluster1-control-plane           1/1    Running  2         6d23h
its1     cluster2      coredns-674b8bbfcf-5c46s                        1/1    Running  2         6d23h
its1     cluster2      etcd-cluster2-control-plane                     1/1    Running  2         6d23h
its1     its1-cluster  coredns-68559449b6-g8kpn                        1/1    Running  14        6d23h
```

#### Getting Services with All Namespaces
```bash
kubectl multi get services -A
```

**Output:**
```
CONTEXT  CLUSTER       NAMESPACE    NAME          TYPE       CLUSTER-IP    EXTERNAL-IP  PORT(S)                 AGE
its1     cluster1      default      kubernetes    ClusterIP  10.96.0.1     <none>       443/TCP                 6d23h
its1     cluster1      kube-system  kube-dns      ClusterIP  10.96.0.10    <none>       53/UDP,53/TCP,9153/TCP  6d23h
its1     cluster2      default      kubernetes    ClusterIP  10.96.0.1     <none>       443/TCP                 6d23h
its1     cluster2      kube-system  kube-dns      ClusterIP  10.96.0.10    <none>       53/UDP,53/TCP,9153/TCP  6d23h
```

#### Using Label Selectors
```bash
kubectl multi get pods -l k8s-app=kube-dns -A
```

**Output:**
```
CONTEXT  CLUSTER       NAMESPACE    NAME                      READY  STATUS   RESTARTS  AGE
its1     cluster1      kube-system  coredns-674b8bbfcf-6k7vc  1/1    Running  2         6d23h
its1     cluster1      kube-system  coredns-674b8bbfcf-vhh9g  1/1    Running  2         6d23h
its1     cluster2      kube-system  coredns-674b8bbfcf-5c46s  1/1    Running  2         6d23h
its1     cluster2      kube-system  coredns-674b8bbfcf-7gft4  1/1    Running  2         6d23h
```

## Supported Resource Types

### Cluster-Scoped Resources
- `nodes` (no, node) - Kubernetes nodes
- `namespaces` (ns) - Kubernetes namespaces  
- `persistentvolumes` (pv) - Persistent volumes
- `storageclasses` (sc) - Storage classes
- `clusterroles` - RBAC cluster roles

### Namespace-Scoped Resources  
- `pods` (po) - Kubernetes pods
- `services` (svc) - Kubernetes services
- `deployments` (deploy) - Kubernetes deployments
- `configmaps` (cm) - Configuration maps
- `secrets` - Kubernetes secrets
- `persistentvolumeclaims` (pvc) - PV claims
- `ingresses` (ing) - Ingress resources

### Custom Resources
- Any CRD installed in clusters (auto-discovered)
- KubeStellar resources (managedclusters, etc.)

## Advanced Usage

### Working with Specific Clusters

By default, kubectl-multi operates on all managed clusters, but you can customize this behavior:

```bash
# Use a different ITS context
kubectl multi --remote-context my-its get nodes

# Use a different kubeconfig file
kubectl multi --kubeconfig /path/to/kubeconfig get pods
```

### Output Formatting

```bash
# Show additional labels
kubectl multi get pods --show-labels

# Use wide output (if supported by the resource)
kubectl multi get pods -o wide

# Get resource in YAML format
kubectl multi get pod mypod -o yaml
```

### Complex Selectors

```bash
# Multiple label selectors
kubectl multi get pods -l app=nginx,version=v1.0

# Field selectors (where supported)
kubectl multi get pods --field-selector status.phase=Running

# Combine selectors and namespaces
kubectl multi get pods -l tier=frontend -n production
```

## Common Workflows

### Monitoring Cluster Health

```bash
# Check node status across all clusters
kubectl multi get nodes

# Check critical system pods
kubectl multi get pods -n kube-system

# Monitor specific applications
kubectl multi get pods -l app=myapp -A
```

### Resource Discovery

```bash
# Find all services
kubectl multi get services -A

# Locate specific deployments
kubectl multi get deployments -l app=web-server

# Check persistent volumes
kubectl multi get pv
```

### Troubleshooting

```bash
# Find failing pods
kubectl multi get pods --field-selector status.phase!=Running -A

# Check resource usage
kubectl multi get pods --show-labels -A

# Monitor specific namespaces
kubectl multi get all -n problematic-namespace
```

## Best Practices

### Performance Tips

1. **Use specific namespaces** when possible to reduce output:
   ```bash
   kubectl multi get pods -n kube-system  # Better than -A
   ```

2. **Use label selectors** to filter results:
   ```bash
   kubectl multi get pods -l app=nginx  # More efficient
   ```

3. **Combine flags** effectively:
   ```bash
   kubectl multi get deployments -l tier=frontend -n production
   ```

### Error Handling

kubectl-multi gracefully handles errors from individual clusters:

- If one cluster is unavailable, others will still be queried
- Warning messages are displayed for failed clusters
- Partial results are still returned

### Output Management

For large outputs:

```bash
# Pipe to less for pagination
kubectl multi get pods -A | less

# Save output to file
kubectl multi get nodes > cluster-nodes.txt

# Filter with grep
kubectl multi get pods -A | grep nginx
```

## Troubleshooting Usage

### Common Issues

#### No Resources Found
```bash
No resources found
```
This is normal if the resource type doesn't exist in any cluster.

#### Cluster Connection Errors
```bash
Warning: failed to list pods in cluster cluster1: connection refused
```
This indicates a specific cluster is unreachable, but others will continue to work.

#### Permission Errors
```bash
Error: pods is forbidden: User "user" cannot list resource "pods"
```
Check your RBAC permissions on the managed clusters.

### Getting Help

```bash
# Get help for the main command
kubectl multi --help

# Get help for specific subcommands
kubectl multi get --help
```

## Next Steps

- Learn about the internal [Architecture](architecture_guide.md)
- Contribute to development with the [Development Guide](development_guide.md)
- Check the [API Reference](api_reference.md) for technical details