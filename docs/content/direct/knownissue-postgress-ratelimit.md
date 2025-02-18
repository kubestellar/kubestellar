# KubeFlex install fails due to DockerHub rate limit on pulling postgresql

## Symptom

Instantiation of the KubeStellar Helm chart fails, because the Job
`ks-core-install-postgresql` fails. Following is an example log tail.

```console
Starting the process to install KubeStellar core: kind-kubeflex...
Release "ks-core" does not exist. Installing it now.
Pulled: ghcr.io/kubestellar/kubestellar/core-chart:0.26.0-rc.1
Digest: sha256:b7f4eb868ee23fb456aaef5f2f490f519534273e295f4e9bd911fda3f2e730f9
Error: failed post-install: 1 error occurred:
	* job ks-core-install-postgresql failed: BackoffLimitExceeded
```

More broadly, this problem can strike anything that installs KubeFlex
or anything else that uses the postgresql container image from
DockerHub.

This is KubeStellar [Issue
2786](https://github.com/kubestellar/kubestellar/issues/2786).

## Root Cause

The KubeStellar Helm chart includes a copy of the KubeFlex Helm chart,
which calls for creating a Pod using the postgresql container image
from Bitnami, which is on DockerHub, which limits the rate at which
clients can pull images.

## Resolutions

You could wait a while and try again.

Alternatively, if you have a DockerHub account, it might be possible
to configure `kind` to use your credentials for that; see [the kind
documentation](https://kind.sigs.k8s.io/docs/user/private-registries/).
