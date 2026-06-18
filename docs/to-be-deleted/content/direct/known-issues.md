# Some known problems

Here are some user and/or environment problems that we have seen.

For bugs, see [the issues on GitHub](https://github.com/kubestellar/kubestellar/issues) and the [release notes](release-notes.md).

## Wrong value stuck in hidden kflex state in kubeconfig

The symptom is `kflex ctx ...` commands failing. See [Confusion due to hidden state in your kubeconfig](knownissue-kflex-extension.md).

## Kind clusters failing to work

The symptom is `kind` cluster(s) that get created but fail to get their job done. See [Potential Error with Kubestellar Installation related to Issues with Kind backed by Rancher Desktop](knownissue-kind-config.md).

## Authorization fail for Helm fetching chart from ghcr

The symptom is that attempting to instantiate the core Helm chart gets an authorization failure. See [Authorization failure while fetching Helm chart from ghcr.io](knownissue-helm-ghcr.md).

## Missing results in a CombinedStatus object

The symptom is a missing entry in the `results` of a `CombinedStatus` object. See [Missing results in a CombinedStatus object](knownissue-collector-miss.md).

## Kind host not configured for more than two clusters

This can arise when using `kind` inside a virtual machine (e.g., when using Docker on a Mac). The symptom is either a complaint from KubeStellar setup that `sysctl fs.inotify.max_user_watches is only 155693 but must be at least 524288` or setup grinding to a halt. See [Kind host not configured for more than two clusters](installation-errors.md).

## Insufficient CPU for your clusters

This can happen when you are using a docker-in-docker technique. The symptom is that setup stops making progress at some point. See [Insufficient CPU for your clusters](knownissue-cpu-insufficient-for-its1.md)
