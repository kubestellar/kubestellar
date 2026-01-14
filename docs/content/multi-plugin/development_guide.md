# Development Guide

This guide covers how to contribute to kubectl-multi, set up your development environment, and understand the codebase.

## Development Setup

### Prerequisites

- Go 1.21 or later
- kubectl installed and configured
- Access to KubeStellar managed clusters for testing
- Make (for build automation)
- Git

### Building from Source

```bash
# Clone repository
git clone <repository-url>
cd kubectl-multi

# Download dependencies
make deps

# Build binary
make build

# Run tests
make test

# Format code
make fmt

# Run all checks
make check
```

### Setup Kubestellar Demo Environment
This can help developers to set up multi-cluster where they can test their **kubectl multi** commands.
```
cd kubectl-multi
# Script for creating demo environment
./scripts/create-kubestellar-demo-env.sh
```
Follow [Get-Started](https://docs.kubestellar.io/release-0.28.0/direct/get-started/) for detailed guide.

### Development Workflow

1. **Fork the repository**
2. **Create feature branch**: `git checkout -b feature/new-command`
3. **Make changes**: Follow Go best practices
4. **Add tests**: Test new functionality
5. **Run checks**: `make check`
6. **Submit PR**: With detailed description

## Project Structure

```
kubectl-multi/
├── main.go                 # Entry point
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
├── Makefile               # Build automation
├── README.md              # Main documentation
├── docs/                  # Documentation folder
│   ├── installation.md    # Installation guide
│   ├── usage.md           # Usage examples
│   ├── architecture.md    # Architecture details
│   ├── development.md     # This file
│   └── api-reference.md   # Code organization
├── pkg/
│   ├── cmd/               # Command implementations
│   │   ├── root.go        # Root command & CLI setup
│   │   ├── get.go         # Get command (fully implemented)
│   │   ├── describe.go    # Describe command (basic)
│   │   ├── apply.go       # Apply command (placeholder)
│   │   └── delete.go      # Other commands (placeholders)
│   ├── cluster/           # Cluster discovery & management
│   │   └── discovery.go   # KubeStellar cluster discovery
│   └── util/              # Utility functions
│       └── formatting.go  # Resource formatting utilities
└── bin/                   # Build output directory
    └── kubectl-multi      # Compiled binary
```

## Adding New Commands

To add a new kubectl command (e.g., `logs`):

### 1. Create Command File

Create `pkg/cmd/logs.go`:

```go
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"kubectl-multi/pkg/cluster"
)

func newLogsCommand() *cobra.Command {
	var (
		follow     bool
		previous   bool
		container  string
		since      string
		sinceTime  string
		timestamps bool
		tailLines  int64
	)

	cmd := &cobra.Command{
		Use:   "logs [-f] [-p] POD [-c CONTAINER]",
		Short: "Print logs for a container in a pod across managed clusters",
		Long: `Print the logs for a container in a pod across all managed clusters.
		
If the pod has only one container, the container name is optional.`,
		Args: cobra.ExactArgs(1), // Require exactly one argument (pod name)
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleLogsCommand(args[0], container, follow, previous, since, sinceTime, timestamps, tailLines)
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Specify if the logs should be streamed")
	cmd.Flags().BoolVarP(&previous, "previous", "p", false, "Print the logs for the previous instance of the container")
	cmd.Flags().StringVarP(&container, "container", "c", "", "Print the logs of this container")
	cmd.Flags().StringVar(&since, "since", "", "Only return logs newer than a relative duration like 5s, 2m, or 3h")
	cmd.Flags().StringVar(&sinceTime, "since-time", "", "Only return logs after a specific date (RFC3339)")
	cmd.Flags().BoolVar(&timestamps, "timestamps", false, "Include timestamps on each line in the log output")
	cmd.Flags().Int64Var(&tailLines, "tail", -1, "Lines of recent log file to display")

	return cmd
}

func handleLogsCommand(podName, container string, follow, previous bool, since, sinceTime string, timestamps bool, tailLines int64) error {
	// 1. Discover clusters
	clusters, err := cluster.DiscoverClusters(kubeconfig, remoteContext)
	if err != nil {
		return fmt.Errorf("failed to discover clusters: %v", err)
	}

	// 2. Get logs from each cluster
	for _, clusterInfo := range clusters {
		fmt.Printf("=== Cluster: %s ===\n", clusterInfo.Name)
		
		// Build log options
		logOptions := &corev1.PodLogOptions{
			Container:  container,
			Follow:     follow,
			Previous:   previous,
			Timestamps: timestamps,
		}
		
		if tailLines >= 0 {
			logOptions.TailLines = &tailLines
		}
		
		// Handle since options
		if since != "" {
			duration, err := time.ParseDuration(since)
			if err != nil {
				fmt.Printf("Warning: invalid since duration for cluster %s: %v\n", clusterInfo.Name, err)
				continue
			}
			sinceSeconds := int64(duration.Seconds())
			logOptions.SinceSeconds = &sinceSeconds
		}
		
		if sinceTime != "" {
			t, err := time.Parse(time.RFC3339, sinceTime)
			if err != nil {
				fmt.Printf("Warning: invalid since-time for cluster %s: %v\n", clusterInfo.Name, err)
				continue
			}
			sinceTime := metav1.NewTime(t)
			logOptions.SinceTime = &sinceTime
		}

		// Get logs
		req := clusterInfo.Client.CoreV1().Pods(namespace).GetLogs(podName, logOptions)
		logs, err := req.Stream(context.TODO())
		if err != nil {
			fmt.Printf("Warning: failed to get logs from cluster %s: %v\n", clusterInfo.Name, err)
			continue
		}
		defer logs.Close()

		// Stream logs to output
		scanner := bufio.NewScanner(logs)
		for scanner.Scan() {
			fmt.Printf("[%s] %s\n", clusterInfo.Name, scanner.Text())
		}
		
		if err := scanner.Err(); err != nil {
			fmt.Printf("Warning: error reading logs from cluster %s: %v\n", clusterInfo.Name, err)
		}
		
		fmt.Println() // Add spacing between clusters
	}
	
	return nil
}