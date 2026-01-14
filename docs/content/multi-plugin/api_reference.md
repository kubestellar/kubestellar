# API Reference

This document provides detailed information about the kubectl-multi codebase, including package structure, key types, and functions.

## Package Structure

### main.go

Entry point that delegates to the cmd package:

```go
package main

import "kubectl-multi/pkg/cmd"

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

## pkg/cmd Package

Contains all CLI command implementations using the Cobra framework.

### root.go

Defines the root command and global configuration.

#### Global Variables

```go
var (
	kubeconfig    string // Path to kubeconfig file
	remoteContext string // Remote hosting context (default: "its1")
	allClusters   bool   // Operate on all managed clusters
	namespace     string // Target namespace
	allNamespaces bool   // List resources across all namespaces
)
```

#### Key Functions

```go
// Execute runs the root command
func Execute() error

// initConfig initializes configuration from flags and environment
func initConfig()
```

### get.go

Implements the `kubectl multi get` command with support for all Kubernetes resources.

#### Key Functions

```go
// newGetCommand creates the get command
func newGetCommand() *cobra.Command

// handleGetCommand processes get requests across clusters
func handleGetCommand(args []string) error

// Resource-specific handlers
func handleNodesGet(tw *tabwriter.Writer, clusters []cluster.ClusterInfo, ...) error
func handlePodsGet(tw *tabwriter.Writer, clusters []cluster.ClusterInfo, ...) error  
func handleServicesGet(tw *tabwriter.Writer, clusters []cluster.ClusterInfo, ...) error
func handleDeploymentsGet(tw *tabwriter.Writer, clusters []cluster.ClusterInfo, ...) error
func handleGenericGet(tw *tabwriter.Writer, clusters []cluster.ClusterInfo, ...) error
```

#### Resource Handler Pattern

All resource handlers follow this pattern:

```go
func handleResourceGet(tw *tabwriter.Writer, clusters []cluster.ClusterInfo, 
                      targetNS string, allNamespaces bool, labelSelector string, 
                      showLabels bool) error {
	
	// 1. Print header once
	printHeader(tw, showLabels, allNamespaces)
	
	// 2. Iterate through all clusters
	for _, clusterInfo := range clusters {
		// 3. List resources in current cluster
		resources, err := listResources(clusterInfo, targetNS, allNamespaces, labelSelector)
		if err != nil {
			fmt.Printf("Warning: failed to list in cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}
		
		// 4. Format and output each resource
		for _, resource := range resources {
			outputResourceRow(tw, clusterInfo, resource, showLabels, allNamespaces)
		}
	}
	
	return nil
}
```

## pkg/cluster Package

Handles cluster discovery and client management.

### discovery.go

Core cluster discovery and client management functionality.

#### ClusterInfo Type

```go
type ClusterInfo struct {
	Name            string                              // Cluster name from ManagedCluster CRD
	Context         string                              // kubectl context name
	Client          kubernetes.Interface                // Typed Kubernetes client
	DynamicClient   dynamic.Interface                   // Dynamic client for CRDs
	DiscoveryClient discovery.DiscoveryInterface        // API resource discovery
	RestConfig      *rest.Config                        // REST configuration
}
```

#### Key Functions

```go
// DiscoverClusters discovers all managed clusters from KubeStellar ITS
func DiscoverClusters(kubeconfig, remoteCtx string) ([]ClusterInfo, error)

// buildClusterClient creates a Kubernetes client for a specific cluster context  
func buildClusterClient(kubeconfig, contextOverride string) (kubernetes.Interface, 
                       dynamic.Interface, discovery.DiscoveryInterface, *rest.Config, error)

// listManagedClusters retrieves cluster names from ManagedCluster CRDs
func listManagedClusters(kubeconfig, remoteCtx string) ([]string, error)

// isWDSCluster filters out Workload Description Space clusters
func isWDSCluster(clusterName string) bool
```

#### Discovery Process

```go
func DiscoverClusters(kubeconfig, remoteCtx string) ([]ClusterInfo, error) {
	// 1. Get list of managed cluster names from ITS
	clusterNames, err := listManagedClusters(kubeconfig, remoteCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to list managed clusters: %v", err)
	}

	var clusters []ClusterInfo
	
	// 2. Build clients for each cluster
	for _, clusterName := range clusterNames {
		// Skip WDS clusters
		if isWDSCluster(clusterName) {
			continue
		}
		
		// Build client for this cluster
		client, dynamicClient, discoveryClient, restConfig, err := buildClusterClient(kubeconfig, clusterName)
		if err != nil {
			fmt.Printf("Warning: failed to build client for cluster %s: %v\n", clusterName, err)
			continue
		}
		
		// Add to cluster list
		clusters = append(clusters, ClusterInfo{
			Name:            clusterName,
			Context:         clusterName,
			Client:          client,
			DynamicClient:   dynamicClient,
			DiscoveryClient: discoveryClient,
			RestConfig:      restConfig,
		})
	}
	
	return clusters, nil
}
```

## pkg/util Package

Utility functions for formatting and resource discovery.

### formatting.go

Resource formatting and discovery utilities.

#### Resource Status Functions

```go
// GetNodeStatus returns the status of a node
func GetNodeStatus(node corev1.Node) string

// GetPodReadyContainers returns number of ready containers in a pod
func GetPodReadyContainers(pod *corev1.Pod) int32

// GetPodTotalContainers returns total number of containers in a pod  
func GetPodTotalContainers(pod *corev1.Pod) int32

// GetServiceExternalIP returns the external IP of a service
func GetServiceExternalIP(svc *corev1.Service) string

// GetServicePorts returns formatted port list for a service
func GetServicePorts(svc *corev1.Service) string
```

#### Formatting Utilities

```go
// FormatLabels formats a label map into a display string
func FormatLabels(labels map[string]string) string

// FormatAge calculates and formats the age of a resource
func FormatAge(t metav1.Time) string

// TranslateTimestampSince returns human readable time since timestamp
func TranslateTimestampSince(timestamp metav1.Time) string
```

#### Resource Discovery

```go
// DiscoverGVR discovers GroupVersionResource for a given resource type
func DiscoverGVR(discoveryClient discovery.DiscoveryInterface, resourceType string) (schema.GroupVersionResource, bool, error)

// normalizeResourceType converts resource aliases to canonical forms
func normalizeResourceType(resourceType string) string
```

#### Resource Discovery Implementation

```go
func DiscoverGVR(discoveryClient discovery.DiscoveryInterface, resourceType string) (schema.GroupVersionResource, bool, error) {
	// 1. Normalize the resource type
	normalizedType := normalizeResourceType(resourceType)
	
	// 2. Get all API resources
	_, apiResourceLists, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		return schema.GroupVersionResource{}, false, err
	}
	
	// 3. Search through API resources
	for _, apiResourceList := range apiResourceLists {
		gv, err := schema.ParseGroupVersion(apiResourceList.GroupVersion)
		if err != nil {
			continue
		}
		
		for _, apiResource := range apiResourceList.APIResources {
			// Check if resource matches by name, singular, or short names
			if strings.EqualFold(apiResource.Name, normalizedType) ||
			   strings.EqualFold(apiResource.SingularName, normalizedType) {
				return schema.GroupVersionResource{
					Group:    gv.Group,
					Version:  gv.Version,
					Resource: apiResource.Name,
				}, apiResource.Namespaced, nil
			}
			
			// Check short names
			for _, shortName := range apiResource.ShortNames {
				if strings.EqualFold(shortName, normalizedType) {
					return schema.GroupVersionResource{
						Group:    gv.Group,
						Version:  gv.Version,
						Resource: apiResource.Name,
					}, apiResource.Namespaced, nil
				}
			}
		}
	}
	
	return schema.GroupVersionResource{}, false, fmt.Errorf("resource type %s not found", resourceType)
}
```

## Build System (Makefile)

### Key Targets

```makefile
# Build the binary
build:
	go build -o bin/kubectl-multi main.go

# Install as kubectl plugin  
install: build
	cp bin/kubectl-multi ~/.local/bin/
	chmod +x ~/.local/bin/kubectl-multi

# Install system-wide
install-system: build
	sudo cp bin/kubectl-multi /usr/local/bin/
	sudo chmod +x /usr/local/bin/kubectl-multi

# Run tests
test:
	go test -v ./pkg/...

# Format code
fmt:
	go fmt ./...
	go vet ./...

# Download dependencies
deps:
	go mod tidy
	go mod download

# Clean build artifacts
clean:
	rm -rf bin/

# Run all checks
check: fmt test build
```

## Key Dependencies

### External Libraries

```go
require (
	github.com/spf13/cobra v1.8.0           // CLI framework
	k8s.io/api v0.29.0                      // Kubernetes API types
	k8s.io/apimachinery v0.29.0             // Kubernetes API machinery  
	k8s.io/client-go v0.29.0                // Kubernetes Go client
	k8s.io/kubectl v0.29.0                  // kubectl utilities
)
```

### Standard Library Usage

- `fmt`: Formatted I/O operations
- `os`: Operating system interface
- `strings`: String manipulation utilities
- `text/tabwriter`: Aligned text output
- `time`: Time and duration handling

## Error Handling Patterns

### Graceful Degradation

```go
// Continue processing other clusters if one fails
for _, clusterInfo := range clusters {
	resources, err := listResources(clusterInfo)
	if err != nil {
		fmt.Printf("Warning: failed to list resources in cluster %s: %v\n", 
		          clusterInfo.Name, err)
		continue // Continue with next cluster
	}
	processResources(resources)
}
```

### Context Propagation

```go
// Use context for cancellation and timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

list, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
```

### Error Wrapping

```go
// Wrap errors with additional context
clusters, err := cluster.DiscoverClusters(kubeconfig, remoteContext)
if err != nil {
	return fmt.Errorf("failed to discover clusters: %v", err)
}
```

## Configuration Management

### Global Configuration

Configuration is managed through global variables and command-line flags:

```go
// Global configuration variables
var (
	kubeconfig    string // --kubeconfig
	remoteContext string // --remote-context  
	namespace     string // --namespace, -n
	allNamespaces bool   // --all-namespaces, -A
	labelSelector string // --selector, -l
	showLabels    bool   // --show-labels
)
```

### Environment Variables

The plugin respects standard kubectl environment variables:

- `KUBECONFIG`: Path to kubeconfig file
- `KUBECTL_CONTEXT`: Default kubectl context

## Output Formatting

### Tabular Output Structure

All commands use consistent tabular output with these columns:

1. `CONTEXT`: kubectl context name
2. `CLUSTER`: KubeStellar cluster name  
3. Resource-specific columns (NAME, STATUS, etc.)

### Column Management

```go
// Dynamic column headers based on resource type and flags
func buildHeader(resourceType string, allNamespaces, showLabels bool) string {
	header := "CONTEXT\tCLUSTER\t"
	
	if allNamespaces {
		header += "NAMESPACE\t"
	}
	
	header += getResourceColumns(resourceType)
	
	if showLabels {
		header += "\tLABELS"
	}
	
	return header
}
```

This API reference provides a comprehensive overview of the kubectl-multi codebase. For usage examples, see the [Usage Guide](usage_guide.md), and for architectural details, see the [Architecture Guide](architecture_guide.md).