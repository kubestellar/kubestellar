# Rename argo-cd to argocd in documentation

This PR updates the documentation to use "argocd" instead of "argo-cd" in Helm commands and YAML examples, aligning with the changes made to the core-chart files.

## Changes

- Updated `docs/content/direct/core-chart-argocd.md` to use "argocd" instead of "argo-cd" in:
  - Helm commands (`--set argocd.install=true` instead of `--set argo-cd.install=true`)
  - YAML examples (`argocd:` instead of `argo-cd:`)
  - JSON examples (`argocd.applications` instead of `argo-cd.applications`)

## Testing

The changes have been tested by:
1. Installing KubeStellar with ArgoCD enabled
2. Creating an ITS and WDS
3. Creating an ArgoCD application
4. Verifying the application is synced and deployed to the WDS

## Notes

- This PR only updates the documentation to match the code changes already made in the core-chart files.
- The URLs to ArgoCD documentation (like `https://argo-cd.readthedocs.io/`) remain unchanged as they are external references.