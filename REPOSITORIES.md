# KubeStellar Repository Ecosystem

KubeStellar is composed of multiple repositories that together enable robust multi-cluster configuration management. Below is a summary of each codebase and its primary purpose.

## Core Platform
- **kubestellar/kubestellar** — The heart of KubeStellar: a Go-based controller and API server that reconciles multi-cluster resource bindings across any number of Kubernetes clusters.
- **kubestellar/kubeflex** — A pluggable control-plane implementation, written in Go, that lets you run Kubernetes APIs on demand for any cluster, edge node, or environment.

## Transport & Integration
- **kubestellar/ocm-status-addon** — A Go add-on that integrates with the Open Cluster Management (OCM) transport layer to surface multi-cluster status in OCM dashboards.
- **kubestellar/ocm-transport-plugin** — A plugin component (Go) for OCM that handles the low-level transport of API requests between KubeStellar and downstream clusters.
- **kubestellar/galaxy** — A collection of integration modules, helper libraries, and tooling for third-party systems (e.g. GitOps, monitoring) to work seamlessly with KubeStellar.

## User Interfaces
- **kubestellar/ui** — A React/TypeScript-based web UI that provides a graphical overview of multi-cluster placements, bindings, and statuses for developers and operators.
- **kubestellar/ui-plugins** — Extension framework for the UI: reusable widgets and plugin points to display additional metrics or custom panels within the KubeStellar dashboard.

## Tools & CLI
- **kubestellar/kubectl-plugin** — A Go-based `kubectl multi` plugin that aggregates resources across all bound clusters, enabling single-command visibility of multi-cluster workloads.
- **kubestellar/a2a** — Python-based “application-to-application” server component that facilitates direct data replication and event hooks between clusters managed by KubeStellar.
- **kubestellar/infra** — Shell scripts and Terraform manifests that define the CI/CD pipelines, GitHub Actions workflows, and cloud infrastructure used to build, test, and release KubeStellar.

## Documentation & Config
- **kubestellar/docs** — The MkDocs/Hugo source for kubestellar.io: contains all user guides, tutorials, API references, and architecture overviews.
- **kubestellar/presentations** — Slides, diagrams, and markdown for conference talks, webinars, and community meetups that explain KubeStellar concepts and use cases.
- **kubestellar/.github** — Organization-wide GitHub configuration: issue and pull-request templates, default workflows, CODEOWNERS, and security policy settings.
