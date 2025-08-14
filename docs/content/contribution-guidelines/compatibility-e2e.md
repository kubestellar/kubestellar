# KubeStellar & KubeFlex Compatibility E2E Testing

## Overview
KubeStellar runs automated end-to-end (E2E) compatibility tests with KubeFlex on every main branch merge and release for either project. This ensures integration issues are detected early.

## How It Works
- **Trigger:** The test is triggered by a Prow job (`pull-kubestellar-kubeflex-compatibility-e2e`) on merges/releases to KubeStellar or KubeFlex. Postsubmit jobs for both repositories are defined in the `kubestellar/infra` Prow configuration.
- **Cluster:** Tests run in an Oracle Cloud Infrastructure (OCI) Kubernetes cluster.
- **Matrix:** The job tests multiple version pairs (main/main, main/stable, stable/main, stable/stable).
- **Periodic:** A periodic job runs nightly to ensure regular compatibility coverage even without active merges.
- **Script:** The E2E script installs both projects and validates integration scenarios:
  - **WEC registration** – registers a simulated Workload Execution Cluster and verifies it appears in KubeStellar.
  - **ITS/WDS interaction** – creates and validates resources handled by these controllers.
  - **Object sync** – creates an object in a source namespace and verifies it appears in the target namespace.
- **Reporting:** Results are posted to GitHub PRs and TestGrid. Failures may notify maintainers on Slack. In TestGrid, results can be found under the `kubestellar-compatibility` dashboard, tab `KubeStellar-KubeFlex E2E Compatibility`.

## Debugging Locally
You can run the E2E matrix script on your own cluster:

```bash
# Prerequisites: bash, kubectl, access to a test cluster
cd scripts/
bash e2e-compatibility-matrix.sh
```
Set environment variables to test specific versions:
```bash
KUBESTELLAR_VERSION=main,stable KUBEFLEX_VERSION=main,stable bash e2e-compatibility-matrix.sh
```

## Adding New Scenarios
Edit `scripts/e2e-compatibility-matrix.sh` to add new integration checks. Ensure the script returns nonzero exit on failure for correct Prow reporting.

## Troubleshooting
- Check pod logs in the test cluster for either namespace (`kubestellar`, `kubeflex`).
- Review the Prow job logs and TestGrid output (`kubestellar-compatibility` dashboard, tab `KubeStellar-KubeFlex E2E Compatibility`) for failure details.

## Questions?
Ask in the #kubestellar-dev Slack channel or open an issue in the repo.
