# Authorization failure while fetching Helm chart from ghcr.io

## Description of the Issue

When following the
[Getting Started recipe](get-started.md) you might get a failure from
the command to instantiate KubeStellar's core Helm chart. The error
message is as follows.

> Error: failed to authorize: failed to fetch oauth token: unexpected status from GET request to https://ghcr.io/token?scope=repository%3Akubestellar%2Fkubestellar%2Fcore-chart%3Apull&service=ghcr.io: 403 Fobidden

This is [Issue 2544](https://github.com/kubestellar/kubestellar/issues/2544).

## Root Cause

Following is one root cause that is partly understood. There may be others.

The cause is the user having a broken configuration for Docker. Even
though `helm` does not itself use containers, `helm` will consult the
user's Docker configuration file (`~/.docker/config.json`) for
registry credentials if that file exists.

Fetching a Helm chart from an OCI registry can involve getting a
temporary token. For a private Helm chart, registry credentials are
required in order to get that temporary token; for a public Helm
chart, registry credentials are not needed.  Even though fetching a
public Helm chart does not require registry credentials, `helm` tries
to get and use credentials for the `ghcr.io` registry if that Docker
configuration file exists.  When that file exists but specifies
something that does not work, that can lead to an error message about
an authorization failure in the request to get the temporary token.

This pathology is discussed in [an Issue in the Helm repository on
GitHub](https://github.com/helm/helm/issues/13179).

For an example, consider the case of someone using Rancher Desktop on
Linux. The installation instructions for Rancher Desktop, in the Linux
case, [recommend installing and initializing a package named
"pass"](https://docs.docker.com/desktop/setup/install/linux/#general-system-requirements). This
is explained in more detail in a [linked
document](https://docs.docker.com/desktop/setup/sign-in/#credentials-management-for-linux-users). If
the user does _not_ install _and_ initialize "pass" then Docker's
handling of registry credentials will be messed up.

### Testing whether Helm can fetch public charts

To test whether the problem is breakage in helm/docker, try the
command `helm show chart
oci://ghcr.io/kubestellar/kubestellar/core-chart`. If that fails all
by itself, the problem is in Helm or something that it uses.

### Resolution for lack of "pass"

Install _and initialize_ the package named "pass".

## Workarounds

If the resolution above does not work then you can try doing the
KubeStellar setup as a different user --- an ordinary user or root
(but remember that unnecessary use of root is a security risk). When
the problem is caused by the user's Docker config file, a different
user's Docker config file might not have the problem. Also, as noted
in [the GitHub Issue](https://github.com/helm/helm/issues/13179),
`helm` will succeed at fetching public charts if the user does NOT
_have_ a Docker config file.

Another way to work around a broken Docker config file is to
temporarily remove or rename it while doing the KubeStellar setup. The
KubeStellar setup does not require credentials for any registry ---
except for pull rate limit considerations. The KubeStellar setup
_does_ involve using some images from DockerHub, and DockerHub imposes
a strict rate limit on non-logged-in users.
