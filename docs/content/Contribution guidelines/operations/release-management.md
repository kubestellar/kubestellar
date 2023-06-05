### Publishing a new KubeStellar release

#### Note:
You currently need write access to the [{{ config.repo_url }}]({{ config.repo_url }}) repository to perform these tasks.

You also need an available team member with approval permissions from [https://github.com/openshift/release/blob/master/ci-operator/config/{{ config.repo_short_name }}/OWNERS](https://github.com/openshift/release/blob/master/ci-operator/config/{{ config.repo_short_name }}/OWNERS).

### Create git tags

#### Prerequisite - make sure you have a GPG signing key

https://docs.github.com/en/authentication/managing-commit-signature-verification/generating-a-new-gpg-key
https://docs.github.com/en/authentication/managing-commit-signature-verification/adding-a-gpg-key-to-your-github-account
https://docs.github.com/en/authentication/managing-commit-signature-verification/telling-git-about-your-signing-key
Create the tags


- `git fetch` from the main KubeStellar repository [{{ config.repo_url }}]({{ config.repo_url }}) to ensure you have the latest commits
- `git tag` the main module
- If your git remote for [{{ config.repo_url }}]({{ config.repo_url }}) is named something other than upstream, change REF accordingly
- If you are creating a release from a release branch, change main in REF accordingly, or you can make REF a commit hash.

```
REF=upstream/main
TAG=v1.2.3
git tag --sign --message "$TAG" "$TAG" "$REF"
Tag the pkg/sdk module, following the same logic as above for REF and TAG
```

```
REF=upstream/main
TAG=v1.2.3
git tag --sign --message "pkg/sdk/$TAG" "pkg/sdk/$TAG" "$REF"
```
Push the tags

```
REMOTE=upstream
TAG=v1.2.3
git push "$REMOTE" "$TAG" "pkg/sdk/$TAG"
```
If it's a new minor version

If this is the first release of a new minor version (e.g. the last release was v0.7.x, and you are releasing the first 0.8.x version), follow the following steps.

Otherwise, you can skip to Review/edit/publish the release in GitHub

### Create a release branch

- Set REMOTE, REF, and VERSION as appropriate.
```
REMOTE=upstream
REF="$REMOTE/main"
VERSION=1.2
git checkout -b "release-$VERSION" "$REF"
git push "$REMOTE" "release-$VERSION"
```

- Configure prow for the new release branch

    - Make sure you have openshift/release cloned
    - Create a new branch
    - Copy ci-operator/config/kcp-dev/kcp/kcp-dev-kcp-main.yaml to ci-operator/config/kcp-dev/kcp/kcp-dev-kcp-release-<version>.yaml
    - Edit the new file
    - Change main to the name of the release branch, such as release-0.8

```
zz_generated_metadata:
  branch: main
```
Change latest to the name of the release branch

```
promotion:
  namespace: kcp
  tag: latest
  tag_by_commit: true
```
    - Edit core-services/prow/02_config/kcp-dev/kcp/_prowconfig.yaml
    - Copy the main branch configuration to a new release-x.y entry
    - Run make update
    - Add the new/updated files and commit your changes
    - Push your branch to your fork
    - Open a pull request
    - Wait for it to be reviewed and merged
    - Update testgrid

- Make sure you have a clone of kubernetes/test-infra
- Edit config/testgrids/kcp/kcp.yaml
- In the test_groups section:
- Copy all the entries under # main to the bottom of the map
- Rename -main- to -release-<version>-
- In the dashboard_groups section:
    - Add a new entry under dashboard_names for kcp-release-<version>
    - In the dashboards section:
        - Copy the kcp-main entry, including dashboard_tab and all its entries, to a new entry called kcp-release-<version>
        - Rename main to release-<version> in the new entry
        - Commit your changes
        - Push your branch to your fork
        - Open a pull request
        - Wait for it to be reviewed and merged
        - Review/edit/publish the release in GitHub

The goreleaser workflow automatically creates a draft GitHub release for each tag.

Navigate to the draft release for the tag you just pushed. You'll be able to find it under the releases page.
If the release notes have been pre-populated, delete them.
For the "previous tag," select the most recent, appropriate tag as the starting point
If this is a new minor release (e.g. v0.8.0), select the initial version of the previous minor release (e.g. v0.7.0)
If this is a patch release (e.g. v0.8.7), select the previous patch release (e.g. v0.8.6)
Click "Generate release notes"
Publish the release
Notify

- Create an email addressed to [kcp-dev@googlegroups.com](https://kubestellar.io/joinus) and [kcp-users@googlegroups.com](https://groups.google.com/g/kubestellar-users)

```
Subject: [release] <version> e.g. [release] v0.8.0
```
    - In the body, include noteworthy changes
    - Provide a link to the release in GitHub for the full release notes
    - Post a message in the [#kubestellar](https://kubestellar.io/slack) Slack channel