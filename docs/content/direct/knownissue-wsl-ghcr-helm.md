# Authorization failure in WSL/Ubuntu while fetching Helm chart from ghcr.io

## Description of the Issue

When working with WSL (Windows Subsystem for Linux)/Ubuntu and trying to instantiate KubeStellar's core Helm chart, you might encounter the following error:

> Error: failed to authorize: failed to fetch oauth token: unexpected status from GET request to https://ghcr.io/token?scope=repository%3Akubestellar%2Fkubestellar%2Fcore-chart%3Apull&service=ghcr.io: 403 Forbidden

This error indicates that there was an issue fetching the OAuth token from GitHub's container registry (`ghcr.io`), which is preventing you from pulling the Helm chart.

## Root Cause

The issue is caused by missing or misconfigured credentials for `ghcr.io` in the Docker configuration. Docker fails to authenticate & retrieve the necessary OAuth token from GitHub's container registry.
## Workaround

To bypass this issue, follow these steps:

1. Run the Getting Started recipe as the root user in Linux. This can be done using the `sudo su -` command to get a shell as root.

```bash
sudo su -
```

This will allow you to execute the Helm chart instantiation command with the necessary privileges.

## 2. Authenticate Docker with GitHub Container Registry

Docker requires authentication to access `ghcr.io` (GitHub's container registry). To authenticate, you need to log into the registry using your GitHub token.

Run the following command to log in to GitHub’s container registry:

```bash
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin
```

Where:
- `$GITHUB_TOKEN` is a GitHub token with the necessary scope for accessing the registry.
- `USERNAME` is your GitHub username.

Make sure the token has the proper permissions (e.g., read access to the repository).

## 3. Install `pass` and Dependencies(optional if not authenticating with GitHub Container Registry)

First, update your package list and install `pass` along with `gnupg2` (used for GPG key management):

```bash
sudo apt update && sudo apt install pass gnupg2 -y

gpg --full-generate-key
pass init "<your_gpg_key_id>"
jq '.credsStore="pass"' ~/.docker/config.json > tmp.json && mv tmp.json ~/.docker/config.json
systemctl restart docker
docker login
```
## Run 

```bash
  helm show chart oci://ghcr.io/kubestellar/kubestellar/core-chart --version 0.25.1
```
## Conclusion

The issue occurs when Docker is not authenticated with GitHub’s container registry. By running the recipe as root and properly authenticating Docker, you should be able to resolve the authorization failure and successfully fetch Helm charts from `ghcr.io
