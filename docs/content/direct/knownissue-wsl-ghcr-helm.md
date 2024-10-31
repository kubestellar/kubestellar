# Authorization failure in WSL while fetching Helm chart from ghcr.io

## Description of the Issue

When using WSL (Windows Subsystem for Linxu) and following the
[Getting Started recipe](get-started.md) you might get a failure from
the command to instantiate KubeStellar's core Helm chart. The error
message is as follows.

> Error: failed to authorize: failed to fetch oauth token: unexpected status from GET request to https://ghcr.io/token?scope=repository%3Akubestellar%2Fkubestellar%2Fcore-chart%3Apull&service=ghcr.io: 403 Fobidden

This is [Issue 2544](https://github.com/kubestellar/kubestellar/issues/2544).

## Root Cause

Unknown.

## Workaround

Run the Getting Started recipe as root user in Linux. For example, use `sudo su -` to get a shell as root.
