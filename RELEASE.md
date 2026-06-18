# Release Process

This document describes the release process for the KubeStellar project, satisfying LFX Insights OSPS-BR-01.01.

## Overview
KubeStellar follows a semantic versioning strategy. Releases are automated using our CI/CD pipeline managed by Prow and GitHub Actions.

## Release Cadence
* **Major/Minor Releases:** Scheduled based on feature completion and stability milestones.
* **Patch Releases:** Issued as needed to address critical bugs or security vulnerabilities.

## Automation
The release process is fully automated. When a new tag is pushed to the repository (e.g., `v0.25.0`), the release pipeline triggers automatically to:
1.  Build the necessary binaries and container images.
2.  Run the test suite to ensure stability.
3.  Publish the artifacts to the container registry.
4.  Create a GitHub Release with the generated changelog.

## Artifacts
All release artifacts (source code, binaries, checksums) are available on the [GitHub Releases Page](https://github.com/kubestellar/kubestellar/releases).
