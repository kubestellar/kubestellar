# Feature Development Tracking

## Overview

This document tracks major feature additions across repository branches in the KubeStellar project. It provides a historical record of when key features were introduced, who authored them, and on which branch they were developed.

## Feature Commit History

| Feature Name | Commit SHA | Author | Branch | Date | Files Changed | Commit Message |
|-------------|------------|---------|---------|------|---------------|----------------|
| Shell Completion Scripts (CLI Bash-to-Go Conversion) | 0e4602431107a39e2b5a29bd9f2188c86c578179 | namasl | i414 | 2023-11-21 | 19 | Preserve REST config for each client |
| CHANGELOG/RELEASE Documentation (Version 1.0.65) | f85dd90977b54df02a9b94360667a94b26dd1448 | shivansh-gohem | fix-lfx-release-docs | 2026-01-12 | 1 | docs: add RELEASE.md to resolve LFX Insight OSPS-BR-01.01 |
| Resource Extraction Improvements (GroupResource Qualification) | f217cb6bac8fa1d2465c46560ccbb4e35b70e72a | xonas1101 | qualify-excluded-resources-by-group | 2026-01-26 | 2 | Minor tweaks, added another apiGroup for events |

## Notes

- **Shell Completion Scripts**: This feature represents the initial work on converting KubeStellar's Bash-based CLI commands (`kubectl-kubestellar-*`) to Go. The PR (#1294) introduced a Go-based CLI framework using Cobra that would later support shell completion functionality. The work was done on the `i414` branch by namasl (Nick Masluk).

- **CHANGELOG/RELEASE Documentation**: This commit added the `RELEASE.md` file documenting the project's release process, versioning strategy, and artifact locations. It addressed the LFX Insights compliance requirement (OSPS-BR-01.01) for build and release documentation. The work was done on the `fix-lfx-release-docs` branch by shivansh-gohem (Shivansh Sahu).

- **Resource Extraction Improvements**: This feature improved the resource exclusion mechanism by qualifying excluded resources using `GroupResource` instead of name-only matching. This prevents accidental exclusion of CRDs that reuse common resource names, making the resource extraction/filtering more precise. The work was done on the `qualify-excluded-resources-by-group` branch by xonas1101.
