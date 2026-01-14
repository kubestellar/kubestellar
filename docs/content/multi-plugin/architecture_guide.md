# Architecture Guide

This document explains how kubectl-multi works internally, its architecture, and technical implementation details.

## High-Level Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────────┐
│   kubectl CLI   │    │  kubectl-multi   │──▶│  KubeStellar ITS    │
│                 │──▶│     Plugin       │──▶│    (Discovery)      │
│ kubectl multi   │    │                  │    │                     │
│   get pods      │    │  Cluster Disco.  │    │ ManagedCluster CRDs │
└─────────────────┘    │  Command Exec.   │    └─────────────────────┘
                       │  Output Format   │
                       └──────────────────┘
                              │
                              ▼
                    ┌─────────────────────┐
                    │  Managed Clusters   │
                    │                     │
                    │ ┌─────────────────┐ │
                    │ │    cluster1     │ │
                    │ │    cluster2     │ │
                    │ │       ...       │ │
                    │ └─────────────────┘ │
                    └─────────────────────┘
```

## Plugin Architecture

```
kubectl-multi/
├── main.go                 # Entry point - delegates to cmd package
├── pkg/
│   ├── cmd/               # Command structure (Cobra-based)
│   │   ├── root.go        # Root command & global flags
│   │   ├── get.go         # Get command implementation
│   │   ├── describe.go    # Describe command
│   │   └── ...           # Other kubectl commands
│   ├── cluster/           # Cluster discovery & management
│   │   └── discovery.go   # KubeStellar cluster discovery
│   └── util/              # Utility functions
│       └── formatting.go  # Resource formatting & helpers
```

## How It Works

### 1. Cluster Discovery Process

The plugin discovers clusters through a multi-step process:

```go
func DiscoverClusters(kubeconfig, remoteCtx string) ([]ClusterInfo, error) {
	// 1. Connect to ITS cluster (e.g., "its1")
	// 2. List ManagedCluster CRDs using dynamic client
	// 3. Filter out WDS clusters (wds1, wds2, etc.)
	// 4. Build clients for each workload cluster
	// 5. Return slice of ClusterInfo with all clients
}
```

**ManagedCluster Discovery:**
```go
// Uses KubeStellar's ManagedCluster CRDs
gvr := schema.GroupVersionResource{
	Group:    "cluster.open-cluster-management.io",
	Version:  "v1", 
	Resource: "managedclusters",
}
```

**WDS Filtering:**
```go
func isWDSCluster(clusterName string) bool {
	lowerName := strings.ToLower(clusterName)
	return strings.HasPrefix(lowerName, "wds") || 
		   strings.Contains(lowerName, "-wds-") || 
		   strings.Contains(lowerName, "_wds_")
}
```

### 2. Command Processing Flow

```
User Input: kubectl multi get pods -n kube-system
     │
     ▼
┌──────────────────────────────────────────────────────────┐
│  1. Parse Command & Flags (Cobra)                       │
│     - Resource type: "pods"                             │
│     - Namespace: "kube-system"                          │
│     - Other flags: selector, output format, etc.       │
└──────────────────────────────────────────────────────────┘
     │
     ▼
┌──────────────────────────────────────────────────────────┐
│  2. Discover Clusters                                    │
│     - Connect to ITS cluster                            │
│     - List ManagedCluster CRDs                          │
│     - Filter out WDS clusters                           │
│     - Build clients for each cluster                    │
└──────────────────────────────────────────────────────────┘
     │
     ▼
┌──────────────────────────────────────────────────────────┐
│  3. Route to Resource Handler                            │
│     - handlePodsGet() for pods                          │
│     - handleNodesGet() for nodes                        │
│     - handleGenericGet() for other resources            │
└──────────────────────────────────────────────────────────┘
     │
     ▼
┌──────────────────────────────────────────────────────────┐
│  4. Execute Across All Clusters                         │
│     - Print header once                                 │
│     - For each cluster:                                 │
│       * List resources using appropriate client         │
│       * Format and append to output                     │
└──────────────────────────────────────────────────────────┘
     │
     ▼
┌──────────────────────────────────────────────────────────┐
│  5. Unified Output                                       │
│     - Single table with CONTEXT and CLUSTER columns     │
│     - Resources from all clusters combined              │
└──────────────────────────────────────────────────────────┘
```

### 3. Resource Type Handling

The plugin handles different resource types through a sophisticated routing system:

#### Built-in Resource Handlers
```go
switch strings.ToLower(resourceType) {
case "nodes", "node", "no":
	return handleNodesGet(...)
case "pods", "pod", "po":
	return handlePodsGet(...)
case "services", "service", "svc":
	return handleServicesGet(...)
case "deployments", "deployment", "deploy":
	return handleDeploymentsGet(...)
// ... more specific handlers
default:
	return handleGenericGet(...) // Uses dynamic client for discovery
}
```

#### Dynamic Resource Discovery
For unknown resource types, the plugin uses Kubernetes API discovery:

```go
func DiscoverGVR(discoveryClient discovery.DiscoveryInterface, resourceType string) (schema.GroupVersionResource, bool, error) {
	// 1. Get all API resources from the cluster
	_, apiResourceLists, err := discoveryClient.ServerGroupsAndResources()
	
	// 2. Normalize resource type (handle aliases like "po" -> "pods")
	normalizedType := normalizeResourceType(resourceType)
	
	// 3. Search through all API resources for matches
	// 4. Return GroupVersionResource + whether it's namespaced
}
```

### 4. Output Formatting

The plugin generates unified tabular output with cluster context:

#### Single Header Strategy
```go
// Print header only once at the top
fmt.Fprintf(tw, "CONTEXT\tCLUSTER\tNAME\tSTATUS\tROLES\tAGE\tVERSION\n")

// Then iterate through all clusters and resources
for _, clusterInfo := range clusters {
	for _, resource := range resources {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			clusterInfo.Context, clusterInfo.Name, ...)
	}
}
```

#### Namespace Handling
The plugin intelligently handles namespace-scoped vs cluster-scoped resources:

```go
// For namespace-scoped resources with -A flag
if allNamespaces {
	fmt.Fprintf(tw, "CONTEXT\tCLUSTER\tNAMESPACE\tNAME\t...\n")
} else {
	fmt.Fprintf(tw, "CONTEXT\tCLUSTER\tNAME\t...\n")  // No namespace column
}
```

## Technical Implementation

### Client Management

The plugin maintains multiple Kubernetes clients for each cluster:

```go
type ClusterInfo struct {
	Name            string                    // Cluster name
	Context         string                    // kubectl context
	Client          *kubernetes.Clientset    // Typed client
	DynamicClient   dynamic.Interface         // Dynamic client
	DiscoveryClient discovery.DiscoveryInterface // API discovery
	RestConfig      *rest.Config             // REST configuration
}
```

### Error Handling Strategy

The plugin uses graceful error handling to ensure partial failures don't break the entire operation:

```go
for _, clusterInfo := range clusters {
	resources, err := clusterInfo.Client.CoreV1().Pods(ns).List(...)
	if err != nil {
		fmt.Printf("Warning: failed to list pods in cluster %s: %v\n", clusterInfo.Name, err)
		continue  // Continue with other clusters
	}
	// Process resources...
}
```

### Resource Type Discovery

For unknown resource types, the plugin uses Kubernetes API discovery:

1. **Normalization**: Converts aliases (`po` → `pods`, `svc` → `services`)
2. **API Discovery**: Queries the cluster for available resources
3. **Matching**: Finds resources by name, singular name, or short names
4. **Fallback**: Uses sensible defaults for common resources

### Namespace Scope Detection

The plugin automatically detects whether resources are namespace-scoped:

```go
gvr, isNamespaced, err := util.DiscoverGVR(clusterInfo.DiscoveryClient, resourceType)

if isNamespaced && !allNamespaces && targetNS != "" {
	// List in specific namespace
	list, err = clusterInfo.DynamicClient.Resource(gvr).Namespace(targetNS).List(...)
} else {
	// List cluster-wide or all namespaces
	list, err = clusterInfo.DynamicClient.Resource(gvr).List(...)
}
```

## Key Dependencies

### Languages & Frameworks
- **Go 1.21+**: Primary language for the plugin
- **Cobra**: CLI framework for command structure and parsing
- **Kubernetes client-go**: Official Kubernetes Go client library
- **KubeStellar APIs**: For managed cluster discovery

### Key Dependencies
```go
require (
	github.com/spf13/cobra v1.8.0           // CLI framework
	k8s.io/api v0.29.0                      // Kubernetes API types
	k8s.io/apimachinery v0.29.0             // Kubernetes API machinery
	k8s.io/client-go v0.29.0                // Kubernetes Go client
	k8s.io/kubectl v0.29.0                  // kubectl utilities
)
```

## Performance Considerations

### Parallel Operations
The plugin executes operations across clusters sequentially to maintain consistent output order, but this could be optimized for parallel execution in the future.

### Caching
Currently, cluster discovery happens on each command execution. Future versions could implement caching for better performance.

### Resource Filtering
WDS cluster filtering happens early in the discovery process to avoid unnecessary API calls.

## Security Model

### Authentication
- Uses existing kubectl configuration and credentials
- Inherits all authentication mechanisms supported by kubectl
- No additional authentication required

### Authorization
- Requires appropriate RBAC permissions on each managed cluster
- Uses the same permissions model as kubectl
- Gracefully handles permission errors per cluster

### Network Security
- All communications use existing Kubernetes API security
- No additional network ports or protocols
- Respects existing network policies and firewalls

## Future Architecture Improvements

### Planned Enhancements
1. **Parallel cluster operations** for better performance
2. **Cluster discovery caching** to reduce API calls
3. **Plugin-based resource handlers** for extensibility
4. **Async operations** for long-running commands
5. **Better error aggregation** and reporting

### Extensibility Points
- Resource handlers can be easily added
- Output formatters can be customized
- Cluster discovery can be extended for other platforms
- Command structure allows easy addition of new kubectl commands