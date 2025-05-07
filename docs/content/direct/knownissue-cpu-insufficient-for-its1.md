## Description of the issue

 When following [Getting Started(https://docs.kubestellar.io/release-0.25.1/direct/get-started/), the step at "Use Core Helm chart to initialize KubeFlex, recognize ITS, and create WDS step." may produce the following error

`no kubeconfig context for its1 was found: context its1 not found for control plane its1`

## Root cause

There is insufficiant CPU resources allocated.

### MacOS

On Mac computers equipped with M1 and onwards, docker needs to make use of either **Docker Desktop** or **colima** when only its agent is installed.

In the latter case, colima default VM is configured to use `--cpu 2 --memory 4`, which is insufficient for Kubestellar components on KinD clusters. In fact, KinD inherit **colima** resources when created.

To solve this issue, increase colima resource capacity to increase KinD clusters resource capcity:

1. Stop colima VM

```bash
colima stop
```

 2. Increase colima cpu and memory capacity

 ```bash
 colima start --cpu 4 --memory 8
 ```

3. Delete kind cluster created by the tutorial, and start over again on a clean state.

