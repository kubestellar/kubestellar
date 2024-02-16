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

### Update the 'kubectl-kubestellar-prep_for_syncer' file with a reference to the new version of the kubestellar syncer version IF you made a new version (and see [the syncer doc](../../../Coding Milestones/PoC2023q1/kubestellar-syncer/#how-to-build-the-image-with-multiple-architectures-and-push-it-to-docker-registry) for how to do that, being careful to not exclude the git commit and cleanliness information from all the tags on the image).
```shell
vi scripts/outer/kubectl-kubestellar-prep_for_syncer
```

change the version in the following line:
```shell hl_lines="1"
syncer_image="quay.io/kubestellar/syncer:{{ config.ks_next_tag }}"
```

### Update the core-helm-chart 'Chart.yaml' and 'values.yaml' files with a reference to the new version of the KubeStellar Helm chart version
```shell
vi core-helm-chart/Chart.yaml
```

change the versions in the 'Chart.yaml' file in the following lines:
```shell hl_lines="2 4"
...
version: {{ config.ks_next_helm_version }}
...
appVersion: {{ config.ks_next_tag }}
...
```

then in 'values.yaml'
```shell 
vi core-helm-chart/values.yaml
```

change the version in the 'values.yaml' file in the following line:
```shell hl_lines="5 11"
# KubeStellar image parameters
image:
  repository: quay.io/kubestellar/kubestellar
  pullPolicy: IfNotPresent
  tag: {{ config.ks_next_branch }}
...
# Space abstraction layer image parameters
spaceimage:
  repository: quay.io/kubestellar/space-framework
  pullPolicy: IfNotPresent
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
stable={{ config.ks_stable_tag }}
latest={{ config.ks_current_tag }}
...
```

<b>after:</b>
```shell title="VERSION" hl_lines="2 3" 
...
stable={{ config.ks_stable_tag }}
latest={{ config.ks_next_tag }}
...
```

### Update the mkdocs.yml file (pre branch)
The mkdocs.yml file points to the branch and tag associated with the branch you have checked out.  Update the ks_branch and ks_tag key/value pairs at the top of the file

```shell
vi docs/mkdocs.yml
```

<b>before:</b>
```shell title="mkdocs.yml" hl_lines="2 3 4 6 7 8"
...
ks_current_branch: '{{ config.ks_current_branch }}'
ks_current_tag: '{{ config.ks_current_tag }}'
ks_current_helm_version: {{ config.ks_current_helm_version }}

ks_next_branch: '{{ config.ks_next_branch }}'
ks_next_tag: '{{ config.ks_next_tag }}'
ks_next_helm_version: {{ config.ks_next_helm_version }}
...
```

<b>after:</b>
```shell title="mkdocs.yml" hl_lines="2 3 4 6 7 8" 
...
ks_current_branch: '{{ config.ks_next_branch }}'
ks_current_tag: '{{ config.ks_next_tag }}'
ks_current_helm_version: {{ config.ks_next_helm_version }}

ks_next_branch:    # put the branch name of the next numerical branch that will come in the future
ks_next_tag:       # put the tag name of the next numerical tag that will come in the future
ks_next_helm_version: # put the number of the next logical helm version
...
```


### Push the main branch
```shell
git add .
git commit -m "updates to main to support new release {{ config.ks_next_tag }}"
git push -u origin main
```

### Create a release-major.minor branch
To create a release branch, identify the current 'release' branches' name (e.g. {{ config.ks_next_branch }}).  Increment the <major> or <minor> segment as part of the 'release' branches' name.  For instance, the 'release' branch is '{{ config.ks_branch }}', you might name the new release branch '{{ config.ks_next_branch }}'.
```shell
git checkout -b {{ config.ks_next_branch }}
```

### Update the mkdocs.yml file (post branch)
The mkdocs.yml file points to the branch and tag associated with the branch you have checked out.  Update the ks_branch and ks_tag key/value pairs at the top of the file

```shell
vi docs/mkdocs.yml
```

<b>before:</b>
```shell title="mkdocs.yml" hl_lines="2 3 4"
...
edit_uri: edit/main/docs/content
ks_branch: 'main'
ks_tag: '{{ config.ks_tag }}'
...
```

<b>after:</b>
```shell title="mkdocs.yml" hl_lines="2 3 4" 
...
edit_uri: edit/{{ config.ks_next_branch }}/docs/content
ks_branch: '{{ config.ks_next_branch }}'
ks_tag: '{{ config.ks_next_tag }}'
...
```

### Update the branch name in /README.MD
There are quite a few references to the main branch /README.MD.  They connect the GitHub Actions for the specific branch to the README.MD page.  Since we are on the new release branch, its time to update these to point to the release itself.

```shell
vi README.MD
```

<b>before:</b>
```shell hl_lines="1"
https://github.com/kubestellar/kubestellar/actions/workflows/docs-gen-and-push.yml/badge.svg?branch=main
```

<b>after:</b>
```shell hl_lines="1"
https://github.com/kubestellar/kubestellar/actions/workflows/docs-gen-and-push.yml/badge.svg?branch={{config.ks_next_branch}}
```

### Push the new release branch
```shell
git add .
git commit -m "new release version {{ config.ks_next_branch }}"
git push -u origin {{ config.ks_next_branch }} # replace <major>.<minor> with your incremented <major>.<minor> pair
```

### Remove the current 'stable' alias using 'mike' (DANGER!)
Be careful, this will cause links to the 'stable' docs, which is the default for our community, to become unavailable.  For now, point 'stable' at 'main'
```shell
cd docs
mike delete stable # remove the 'stable' alias from the current '{{ config.ks_branch }}' branches' doc set
mike deploy --push --rebase --update-aliases main stable # this generates the 'main' branches' docs set and points 'stable' at it temporarily
cd ..
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

create a tag that follows <major>.<minor>.<patch>.  For this example we will increment tag '{{ config.ks_current_tag }}' to '{{ config.ks_next_tag }}'

```shell
TAG={{ config.ks_next_tag }}
REF={{ config.ks_next_branch }}
git tag --sign --message "$TAG" "$TAG" "$REF"
git push origin --tags
```

### Clean out previous release tar images and the checksums256.txt file from your local build environment
When you create a build, output goes to your local __/build/release__.  Make sure this path is empty before you start so there is no mixup with your current build.

### Create a KubeStellar full build
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
        - You add the KubeStellar-specific '*.tar.gz' and the 'checksums256.txt' files
        - GitHub will automatically add the 'Source Code (zip)' and 'Source Code (tar.gz)'

    ![Release Example](gh-draft-new-release.png)

### Create the KubeStellar Core container image
First, login to quay.io with a user that has credentials to 'write' to the kubestellar quay repo
```
docker login quay.io
```

then, remove any running container from moby/buildkit
```
CONTAINER ID   IMAGE                           COMMAND              
c943925fd137   moby/buildkit:buildx-stable-1   "buildkitd" 

docker rm c943925fd137 -f
```

and remove the 'buildx' container image from your local docker images
```
REPOSITORY      TAG               IMAGE ID       CREATED        SIZE
moby/buildkit   buildx-stable-1   16fc6c95ddff   10 days ago    168MB

docker rmi 16fc6c95ddff
```

finally, make the KubeStellar image from within the local copy of the release branch '{{config.ks_next_branch}}'
```
make kubestellar-image
```

### Create a Space Core build
```shell
pushd space-framework
./make spacecore-image {{ config.ks_next_tag }}
popd
```

### Update the KubeStellar and Space Core container images just build and uploaded to quay.io

Head up to quay.io and look for the image of KubeStellar Core container just uploaded.
Tag the image with: __'latest'__, __'{{ config.ks_next_branch }}'__, and __'{{ config.ks_next_tag }}'__ so that helm and other install methods pickup this image.


### Update KubeStellar Core Helm repository
First, make sure you have a version of __'tar'__ that supports the __'--transform'__ command line option
```
brew install gnu-tar
```

then, from root of local copy of [https://github.com/kubestellar/kubestellar](https://github.com/kubestellar/kubestellar) repo:
```
gtar -zcf kubestellar-core-{{config.ks_next_helm_version}}.tar.gz core-helm-chart/ --transform s/core-helm-chart/kubestellar-core/
mv kubestellar-core-{{config.ks_next_helm_version}}.tar.gz ~
shasum -a 256 ~/kubestellar-core-{{config.ks_next_helm_version}}.tar.gz
```
Clone the homebrew-kubestellar repo
```shell
git clone git@github.com:{{ config.helm_repo_short_name }}.git
cd {{ config.helm_repo_default_file_path }}
git checkout main
```

then, from root of local copy of [https://github.com/kubestellar/helm](https://github.com/kubestellar/helm) repo
```
mv ~/kubestellar-core-{{config.ks_next_helm_version}}.tar.gz charts
```

next, update 'index.yaml' in root of local copy of __helm repo__ (only update the data, not time, on lines 6 and 15):  
```shell title="index.yaml" hl_lines="5 6 8 13 14 15"
apiVersion: v1
entries:
  kubestellar-core:
  - apiVersion: v2
    appVersion: {{ config.ks_next_tag }}
    created: "2023-10-30T12:00:00.727185806-04:00"
    description: A Helm chart for KubeStellar Core deployment as a service
    digest: 6f42d9e850308f8852842cd23d1b03ae5be068440c60b488597e4122809dec1e
    icon: https://raw.githubusercontent.com/kubestellar/kubestellar/main/docs/favicons/favicon.ico
    name: kubestellar
    type: application
    urls:
    - https://helm.kubestellar.io/charts/kubestellar-core-{{config.ks_new_helm_version}}.tar.gz
    version: "{{config.ks_next_helm_version}}"
generated: "2023-10-30T12:00:00.727185806-04:00"
```

finally,
finally, push to the main branch
```shell
git add .
git commit -m "updates to main to support release {{ config.ks_next_tag }} of KubeStellar Helm component"
git push -u origin main
```


### Update KubeStellar CLI Brew repository
Clone the homebrew-kubestellar repo
```shell
git clone git@github.com:{{ config.brew_repo_short_name }}.git
cd {{ config.brew_repo_default_file_path }}
git checkout main
```

edit the kubestellar_cli.rb file
```shell
vi Formula/kubestellar_cli.rb
```

update all instances of 'url' from {{ config.ks_current_tag }} to __{{ config.ks_next_tag }}__ (should be 6 of these)
```shell hl_lines="3"
...
    when :arm64
      url "https://github.com/kubestellar/kubestellar/releases/download/{{ config.ks_next_tag }}/kubestellaruser_{{ config.ks_next_tag }}_darwin_arm64.tar.gz"
      sha256 "5be4c0b676e8a4f5985d09f2cfe6c473bd2f56ebd3ef4803ca345e6f04d83d6b" 
...
```

then, update all instances of 'sha256' with the corresponding sha256 hash values in the build/release/checksums256.txt you create during the make-full-release.sh section above. (should be 6 of these)

```shell hl_lines="4"
...
    when :arm64
      url "https://github.com/kubestellar/kubestellar/releases/download/{{ config.ks_next_tag }}/kubestellaruser_{{ config.ks_next_tag }}_darwin_arm64.tar.gz"
      sha256 "<corresponding sha256 hash from checksums256.txt>" 
...
```

finally, push to the main branch
```shell
git add .
git commit -m "updates to main to support release {{ config.ks_next_tag }} of KubeStellar Brew component"
git push -u origin main
```

and, to test
```shell
brew update
brew install kubestellar-cli
```

you should see output that indicates an update for the kubestellar brew tap and then an update to version {{ config.ks_next_tag }} of the kubestellar_cli brew formula.

### Check that GH Workflows for docs are working
Check to make sure the GitHub workflows for doc generation, doc push, and broken links is working and passing
[{{ config.repo_url }}/actions/workflows/docs-gen-and-push.yml]({{ config.repo_url }}/actions/workflows/docs-gen-and-push.yml)
[{{ config.repo_url }}/actions/workflows/broken-links-crawler.yml]({{ config.repo_url }}/actions/workflows/broken-links-crawler.yml)

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
