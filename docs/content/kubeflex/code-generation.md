# Code Generation Guide

This document describes how to generate and maintain typed Kubernetes clients for the kubeflex CRDs.

## Overview

kubeflex uses the standard Kubernetes code-generator tools to generate:

- **Typed Clientsets**: Strongly-typed clients for interacting with CRDs via the Kubernetes API
- **Informers**: SharedInformerFactory for watching CRD resources with local caching
- **Listers**: Typed listers for reading CRD resources from informer caches

## Prerequisites

- Go 1.24.5+ installed
- Access to `k8s.io/code-generator` (automatically fetched during generation)

## Generating Code

To regenerate all typed clients, informers, and listers:

```bash
make generate-clients
```

This runs `hack/update-codegen.sh`, which invokes the Kubernetes code-generator tools.

## Verifying Generated Code

To verify that generated code is up-to-date:

```bash
make verify-codegen
```

This is useful in CI pipelines to ensure generated code is committed after API changes.

## When to Regenerate

Regenerate typed clients when API types in `api/v1alpha1/` are modified.

## Troubleshooting

### Code generation fails

Ensure you have the correct version of Go installed and that `k8s.io/code-generator` is accessible.

### Generated code doesn't compile

Check that:
1. API types have proper markers (`+k8s:deepcopy-gen=package`, `+groupName=...`, `+genclient`)

