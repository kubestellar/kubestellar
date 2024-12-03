# Some known problems

Here are some user and/or environment problems that we have seen.

For bugs, see [the issues on GitHub](https://github.com/kubestellar/kubestellar/issues) and the [release notes](release-notes.md).

## Wrong value stuck in hidden kflex state in kubeconfig

The symptom is `kflex ctx ...` commands failing. See [Confusion due to hidden state in your kubeconfig](knownissue-kflex-extension.md).

## Kind clusters failing to work

The symptom is `kind` cluster(s) that get created but fail to get their job done. See [Potential Error with Kubestellar Installation related to Issues with Kind backed by Rancher Desktop](knownissue-kind-config.md).

## Authorization fail for WSL fetching Helm chart from ghcr

The symptom is that attempting to instantiate the core Helm chart gets an authorization failure, for a user running a Linux distro using WSL. See [Authorization failure in WSL while fetching Helm chart from ghcr.io](knownissue-wsl-ghcr-helm.md).

## Missing results in a CombinedStatus object

The symptom is a missing entry in the `results` of a `CombinedStatus` object. See [Missing results in a CombinedStatus object](knownissue-collector-miss.md).
