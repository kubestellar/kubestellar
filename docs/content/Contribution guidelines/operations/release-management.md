### Publishing a new KubeStellar release

#### Prerequisite - make sure you have a GPG signing key

1. [https://docs.github.com/en/authentication/managing-commit-signature-verification/generating-a-new-gpg-key](https://docs.github.com/en/authentication/managing-commit-signature-verification/generating-a-new-gpg-key)
2. [https://docs.github.com/en/authentication/managing-commit-signature-verification/adding-a-gpg-key-to-your-github-account](https://docs.github.com/en/authentication/managing-commit-signature-verification/adding-a-gpg-key-to-your-github-account)
3. [https://docs.github.com/en/authentication/managing-commit-signature-verification/telling-git-about-your-signing-key](https://docs.github.com/en/authentication/managing-commit-signature-verification/telling-git-about-your-signing-key)

### Create the tags

#### Note:
You currently need write access to the [{{ config.repo_url }}]({{ config.repo_url }}) repository to perform these tasks.

<!-- You also need an available team member with approval permissions from [https://github.com/openshift/release/blob/master/ci-operator/config/{{ config.repo_short_name }}/OWNERS](https://github.com/openshift/release/blob/master/ci-operator/config/{{ config.repo_short_name }}/OWNERS). -->

### Checkout the main branch
```shell
git clone git@github.com:{{ config.repo_short_name }}.git
cd {{ config.repo_default_file_path }}
git checkout main
```

### Update the 'kubectl-kubestellar-prep_for_syncer' file with a reference to the new version of the kubestellar syncer version
```shell
vi outer-scripts/kubectl-kubestellar-prep_for_syncer
```

change the version in the following line:
```shell
syncer_image="quay.io/kubestellar/syncer:{{ config.ks_next_tag }}"
```

### Update the core-helm-chart 'Chart.yaml' and 'values.yaml' files with a reference to the new version of the KubeStellar Helm chart version
```shell
vi core-helm-chart/Chart.yaml
```

change the versions in the 'Chart.yaml' file in the following lines:
```shell hl_lines="1 2"
version: {{ config.ks_next_helm_version }}
appVersion: {{ config.ks_next_tag }}
```

then in 'values.yaml'
```shell 
vi core-helm-chart/values.yaml
```

change the version in the 'values.yaml' file in the following line:
```shell hl_lines="1"
tag: {{ config.ks_next_branch }}
```

### Update the VERSION file
The VERSION file points to the 'latest' and 'stable' release tags associated with the most recent release (latest) and the most stable release (stable).  Update the 'stable' and 'latest tags accordingly

```shell
vi VERSION
```

<b>before:</b>
```shell title="VERSION" hl_lines="2 3"
...
stable={{ config.ks_tag }}
latest={{ config.ks_tag }}
...
```

<b>after:</b>
```shell title="VERSION" hl_lines="2 3" 
...
stable={{ config.ks_tag }}
latest={{ config.ks_next_tag }}
...
```

### Push the main branch
```shell
git add .
git commit -m "updates to main to support new release"
git push -u origin main
```

### Create a release-major.minor branch
To create a release branch, identify the current 'release' branches' name (e.g. {{ config.ks_next_branch }}).  Increment the <major> or <minor> segment as part of the 'release' branches' name.  For instance, the 'release' branch is '{{ config.ks_branch }}', you might name the new release branch '{{ config.ks_next_branch }}'.
```shell
git checkout -b {{ config.ks_next_branch }}
```

### Update the mkdocs.yml file
The mkdocs.yml file points to the branch and tag associated with the branch you have checked out.  Update the ks_branch and ks_tag key/value pairs at the top of the file

```shell
vi docs/mkdocs.yml
```

<b>before:</b>
```shell title="mkdocs.yml" hl_lines="2 3 4 5 6 7 8"
...
edit_uri: edit/main/docs/content
ks_branch: 'main'
ks_tag: '{{ config.ks_tag }}'
ks_current_helm_version: {{ config.ks_current_helm_version }}
ks_next_branch: '{{ config.ks_next_branch }}'
ks_next_tag: '{{ config.ks_next_tag }}'
ks_next_helm_version: {{ config.ks_next_helm_version }}
...
```

<b>after:</b>
```shell title="mkdocs.yml" hl_lines="2 3 4 5 6 7 8" 
...
edit_uri: edit/{{ config.ks_next_branch }}/docs/content
ks_branch: '{{ config.ks_next_branch }}'
ks_tag: '{{ config.ks_next_tag }}'
ks_current_helm_version: {{ config.ks_next_helm_version }}
ks_next_branch:    # put the branch name of the next numerical branch that will come in the future
ks_next_tag:       # put the tag name of the next numerical tag that will come in the future
ks_next_helm_version: # put the number of the next logical helm version
...
```

### Update the branch name in kubestellar/docs/content/readme.md
There are about 6 instances of these in the readme.md.  They connect the GitHub Actions for the specific branch to the readme.md page.
<b>before:</b>
```shell
https://github.com/kubestellar/kubestellar/actions/workflows/docs-gen-and-push.yml/badge.svg?branch=main
```

<b>after:</b>
```shell
https://github.com/kubestellar/kubestellar/actions/workflows/docs-gen-and-push.yml/badge.svg?branch={{config.ks_next_branch}}
```

### Remove the current 'stable' alias using 'mike' (DANGER!)
Be careful, this will cause links to the 'stable' docs, which is the default for our community, to become unavailable.  For now, point 'stable' at 'main'
```shell
cd docs
mike delete stable # remove the 'stable' alias from the current '{{ config.ks_branch }}' branches' doc set
mike deploy --push --rebase --update-aliases main stable # this generates the 'main' branches' docs set and points 'stable' at it temporarily
cd ..
```


### Push the new release branch
```shell
git add .
git commit -m "new release version {{ config.ks_next_branch }}"
git push -u origin {{ config.ks_next_branch }} # replace <major>.<minor> with your incremented <major>.<minor> pair
```

### Update the 'stable' alias using 'mike'
```shell
cd docs
mike delete stable # remove the 'stable' alias from the 'main' branches' doc set
git pull
mike deploy --push --rebase --update-aliases {{ config.ks_next_branch }} stable  # this generates the new '{{ config.ks_next_branch }}' branches' doc set and points 'stable' at it
cd ..
```

### Test your doc site
Open a Chrome Incognito browser to [{{ config.docs_url }}]({{ config.docs_url }}) and look for the version drop down to be updated to the new release you just pushed with 'git' and deployed with 'mike'

### Create a tagged release
View the existing tags you have for the repo

```shell
git fetch --tags
git tag
```

create a tag that follows <major>.<minor>.<patch>.  For this example we will increment tag '{{ config.ks_tag }}' to '{{ config.ks_next_tag }}'

```shell
TAG={{ config.ks_next_tag }}
REF={{ config.ks_next_branch }}
git tag --sign --message "$TAG" "$TAG" "$REF"
git push origin --tags
```

### Create a build
```shell
./hack/make-release-full.sh {{ config.ks_next_tag }}
```

### Create a release in GH UI
- Navigate to the KubeStellar GitHub Source Repository Releases section at {{ config.repo_url }}/releases
- Click 'Draft a new release' and select the tag ('{{ config.ks_next_tag }}')
    - Select the release branch you created above ({{ config.ks_next_branch }})
    - Add a release title ({{ config.ks_next_tag }})
    - Add some release notes ('generate release notes' if you like)
    - select 'pre-release' as a the first step.  Once validated the release is working properly, come back and mark as 'release'
    - Attach the binaries that were created in the 'make-release-full' process above
        - You add the KubeStellar-specific '*.tar.gz' and the 'checksum256.txt' files
        - GitHub will automatically add the 'Source Code (zip)' and 'Source Code (tar.gz)'

    ![Release Example](gh-draft-new-release.png)

### Create the KubeStellar Core container image
First, login to quay.io with a user that has credentials to 'write' to the kubestellar quay repo
```
docker login quay.io
```

then, remove the 'buildx' container image from your local docker images
```
REPOSITORY      TAG               IMAGE ID       CREATED        SIZE
moby/buildkit   buildx-stable-1   16fc6c95ddff   10 days ago    168MB

docker rmi 16fc6c95ddff
```

finally, make the KubeStellar image from within the local copy of the release branch '{{config.ks_next_branch}}'
```
make kubestellar-image
```

### Update KubeStellar Core Helm repository
First, make sure you have a version of 'tar' that supports '--transform'
```
brew install gnu-tar
```

then, from root of local copy of https://github.com/kubestellar/kubestellar repo:
```
gtar -zcf kubestellar-core-{{config.ks_next_helm_version}}.tar.gz core-helm-chart/ --transform s/core-helm-chart/kubestellar-core/
mv kubestellar-core-{{config.ks_next_helm_version}}.tar.gz ~
shasum -a 256 ~/kubestellar-core-{{config.ks_next_helm_version}}.tar.gz
```

then, from root of local copy of https://github.com/kubestellar/helm repo
```
mv ~/kubestellar-core-{{config.ks_next_helm_version}}.tar.gz charts
```

next, update 'index.yaml' in root of local copy of helm repo (only update the data, not time, on lines 6 and 15):  
```shell title="index.yaml" hl_lines="5 6 8 13 14 15"
apiVersion: v1
entries:
  kubestellar-core:
  - apiVersion: v2
    appVersion: v0.10.0
    created: "2023-10-30T12:00:00.727185806-04:00"
    description: A Helm chart for KubeStellar Core deployment as a service
    digest: 6f42d9e850308f8852842cd23d1b03ae5be068440c60b488597e4122809dec1e
    icon: https://raw.githubusercontent.com/kubestellar/kubestellar/main/docs/favicons/favicon.ico
    name: kubestellar
    type: application
    urls:
    - https://helm.kubestellar.io/charts/kubestellar-core-{{config.ks_new_helm_version}}.tar.gz
    version: "3"
generated: "2023-10-30T12:00:00.727185806-04:00"
```



### Update the KubeStellar Core container image just build and uploaded to quay.io

Head up to quay.io and look for the image of KubeStellar Core container just uploaded.
Make this image 'stable' so that helm and other install methods pickup this image.

### Check that GH Workflows for docs are working
Check to make sure the GitHub workflows for doc generation, doc push, and broken links is working and passing
[{{ config.repo_url }}/actions/workflows/docs-gen-and-push.yml]({{ config.repo_url }}/actions/workflows/docs-gen-and-push.yml)
[{{ config.repo_url }}/actions/workflows/broken-links-crawler.yml]({{ config.repo_url }}/actions/workflows/broken-links-crawler.yml)




<!-- ### Note sure if any of this PROW Stuff is necessary - we will see the next time we do a release..
- Configure prow for the new release branch

    - Make sure you have openshift/release cloned
    - Create a new branch
    - Copy ci-operator/config/kcp-dev/edge-md/kcp-dev-kcp-main.yaml to ci-operator/config/kubestellar/kubestellar/kcp-dev-kcp-release-<version>.yaml
    - Edit the new file
    - Change main to the name of the release branch, such as {{ config.ks_next_branch }}

```
zz_generated_metadata:
  branch: main
```
Change latest to the name of the release branch

```
promotion:
  namespace: kubestellar
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
Notify -->


### Create an email addressed to [kubestellar-dev@googlegroups.com and kubestellar-users@googlegroups.com](mailto:kubestellar-dev@googlegroups.com,kubestellar-users@googlegroups.com) 

```
Subject: KubeStellar release {{ config.ks_next_tag }}

Body:

Dear KubeStellar Community,
	Release {{ config.ks_next_tag }} is now available at https://github.com/kubestellar/kubestellar/releases/tag/{{ config.ks_next_tag }}
 
What's Changed

🐛 Fix display of initial spaces after deploy in kube by @MikeSpreitzer in #1143
✨ Generalize bootstrap wrt namespace in hosting cluster by @MikeSpreitzer in #1144
✨ Generalize bootstrap wrt namespace in hosting cluster by @MikeSpreitzer in #1145
✨ Switch to use k8s code generators by @ezrasilvera in #1139
✨ Bump actions/checkout from 4.1.0 to 4.1.1 by @dependabot in #1151
🌱 Align default core image ref in chart with coming release by @MikeSpreitzer in #1146
📖Update dev-env.md by @francostellari in #1157
📖Update Chart.yaml appVersion by @francostellari in #1158
🐛 Use realpath to see through symlinks by @MikeSpreitzer in #1156
✨ Increase kind version to v0.20 for ubuntu by @fab7 in #1155
📖 Document syncer removal by @MikeSpreitzer in #1164
🌱 Rename urmeta to ksmeta by @MikeSpreitzer in #1166
✨ Make get-internal-kubeconfig fetch mid-level kubeconfig by @MikeSpreitzer in #1161
✨ Make ensure/remove wmw insensitive to current workspace by @MikeSpreitzer in #1160
New Contributors

@fab7 made their first contribution in #1155
Full Changelog: v0.8.0…v0.9.0

Thank you for your continued support,

Andy
```

### Post the same message in the [#kubestellar](https://kubestellar.io/slack) Slack channel
