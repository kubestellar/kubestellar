# Making KubeStellar Releases

This document defines how releases of the KubeStellar repository are made. This document is a work-in-progress. In particular, the dependency cycle between the `kubestellar` and `ocm-tansport-plugin` repos is not well documented and it is still not clear how to test and release in the presence of this cycle.

See the associated [packaging and delivery doc](packaging.md) for some
clues about the problem.

## Goals and limitations

The release process has the following goals.

- A release is identified using [semantic versioning](https://semver.org). This means that the associated semantics are followed, in terms of what sort of changes to the repo require what sort of changes to the release identifier.
- A user can pick up and use a given existing release without being perturbed by on-going contributor work. A release is an immutable thing.
- A release with a given semver identifier is built from a commit of this Git repository tagged with a tag whose name is "v" followed by the release identifier.
- There is a concept of a release branch, with a name following the pattern `release-$major.$minor`. This allows contributors to do maintenance work on older release branches. This also allows a user to track work along such a line.
- The contents of `main` always work. This includes passing CI tests. This includes documentation being accurate. We allow point-in-time specific documentation, such as a document that says "Here is how to use release 1.2.3" --- which would refer to a release made in the past. We do not require the documentation in `main` to document all releases.
- A git tag is immutable. Once associated with a given Git commit, that association is not changed later.
- We do not put self-references into Git. For example, making release `1.2.3` does not require changing any file in Git to have the string `1.2.3` in it.

We have the following limitations.

- The only way to publish artifacts (broadly construed, not (necessarily) GitHub "release artifacts") is to make a release.
- The only way to test published artifacts is to make a release and test it.
- Thus, it is necessary to keep users clearly appraised of the quality (or status of evaluating the quality) of each release.
- Because of the lack of self references, most user instructions (e.g., examples) and tests do not have concrete release identifiers in them; instead, the user has to chose and supply the release identifier. There can also be documentation of a specific past release (e.g., the latest stable release) that uses the literal identifier for that past release.
- **PAY ATTENTION TO THIS ONE**: Because of the prohibition of self references, **Git will not contain the exact bytes of our Helm chart definitions**. Where a Helm chart states its own version or has a container image reference to an image built from the same release, the bytes in Git have a placeholder for that image's tag and the process of creating the published release artifacts fills in that placeholder. Think of this as being analogous to the linking done when building a binary executable file.
- The design below falls short of the goal of not putting self-references in files under Git control. One way is in the KubeFlex PostCreateHook that installs the kubestellar-controller-manager (KCM), where the version of the container image for the KCM appears. Another is in the examples document, which also holds references to its own release. Another is in the examples.md file, and another is in the `docs/content/direct/README.md` file.

## Technology

There is a GitHub workflow that creates the published artifacts for each Git tag whose name starts with "v". The rest of the tag name is required to be a semver release identifier. Note that this document does not (yet, anyway) specify **how** that GitHub workflow gets its job done. This workflow is confusingly named "goreleaser" and in a file named "goreleaser.yml" and has a job named "goreleaser" despite the fact that it does more than use goreleaser.

For each tag `v$version` the following published artifacts will be created.

- The container image for the kubestellar-controller-manager (KCM), at `ghcr.io/kubestellar/kubestellar/controller-manager`. Image tag will be `$version`. This GitHub "package" will be connected to the ks/ks repo (this connection is something that an admin will do once, it will stick for all versions).
- The Helm chart (for installing the KCM for a WDS), at `ghcr.io/kubestellar/kubestellar/controller-manager-chart` with version `$version` and Helm "appVersion" `$version`. This GitHub "package" will also be connected to the ks/ks repo. The chart has a reference to container image for the KCM and that reference is `ghcr.io/kubestellar/kubestellar/controller-manager:$version`. **In Git the chart has only placeholders in these places, _not_ `$version`; the `$version` is inserted into a distinct copy by the GitHub workflow, which then publishes this specialized copy.**
- Note that there is no automation (yet) concerning the KubeFlex PostCreateHook that installs the KCM.

## Website

We use `mike` and `MkDocs` to derive and publish GitHub pages. See `docs/README.md` for details.

The published GitHub pages are organized into "releases".  Each release in the GitHub pages corresponds to a git branch whose name begins with "release-" or is "main".

Our documentation is, mostly, viewable in either of two ways. The source documents can be viewed directly through GitHub's web UI for files. The other way is through the website.

## Testing and Examples

The unit tests (of which we have almost none right now), integration tests (of which we also have very few), and end-to-end (E2E) tests in this repository are run in the context of a local copy of this repository and test that version of this repository --- not using any published release artifacts. Additionally, some E2E tests have the option to test published artifacts instead of the local copy of this repo.

The end-to-end tests include ones written in `bash`, and these are the only documentation telling a user how to use the present version of this repository. Again, these tests do not use any published artifacts from a release of this repo.

We have another category of tests, _release tests_. These test a given release, using the published artifacts of that release. Currently all the release tests are a subset of the E2E tests --- those that can be told to test published artifacts. In particular, they can test the published artifacts reached through the kubestellar PostCreatHook, which contains an explicit reference to one particular release (as explained elsewhere in this document).

We have GitHub workflows that exercise the E2E tests, normally on the copy of the repo that the workflow applies to. However, these workflows are parameterized and can be told to test the released artifacts instead.

We also have a GitHub workflow, named "Test latest release" in `.github/workflows/test-latest-release.yml`, that invokes those E2E tests on the latest release. This workflow can be triggered manually, and is also configured to run after completion of the workflow ("goreleaser") that publishes release artifacts.

We will maintain a document that lists releases that pass our quality bar. The latest of those is thus the latest stable release. This document is updated in `main` as quality evaluations come in.

In [docs/content/direct/README.md](README.md) we maintain a statement of what is the latest stable release.

We maintain an [examples document](examples.md) that tells users how to exercise the release that the document appears in. This requires a self-reference that is updated as part of the release process.

## Policy

We aim for all regular releases to be working. In order to do that, we have to make test releases and test them. The widely recognized pattern for doing that is to make "release candidates" (i.e., releases for testing purposes) `1.2.3-rc0`, `1.2.3-rc1`, `1.2.3-rc2`, and so on, while trying to get to a quality release `1.2.3`. Once one of them is judged to be of passing quality, we make a release without the `-rc<N>` suffix. Due to the self-reference in the KCM PostCreateHook, this will involve making a new commit.

Right after making a release we test it thoroughly.

### Deliberately feature-incomplete releases

We plan a few deliberately feature-incomplete releases. They will be regular releases as far as the technology here is concerned. They will be announced only to selected users who acknowledge that they are getting something that is incomplete. In GitHub, these will be marked as "pre-releases". The status of these releases will be made clear in their documentation (which currently appears in [the release notes](release-notes.md).

### Website

We aim to keep the documents viewable both through the website and GitHub's web UI for viewing files. We aim for all of the documentation to be reachable on the website and in the GitHub file UI starting from the repo's README.md.

We create a release in the GitHub pages for every release. A patch release is a release. A test release is a release. Creating that GHP release is done by creating a git branch named `release-$version`.

## Step-by-Step

Making a new release requires a contributor to do the following things. Here `$version` is the semver identifier for the release (e.g., `1.2.3-rc2`).

- Edit the source for the KCM PCH (in `config/postcreate-hooks/kubestellar.yaml`) and update the tag in the reference to the KCM container image (it appears in the last object, a `Job`).

- Edit [the examples document](examples.md) to update the self-references for the coming release.

- Until we have our first stable release, edit [the README](README.md#latest-stable-release) where it wishes it could cite a stable release but instead cites the latest release, to rever to the coming release.

- Edit [the release notes](release-notes.md).

- Make a new Git commit with those changes and get it into the right branch in the shared repo (through the regular PR process if not authorized to cheat).

- Apply the Git tag `v$version` to that new commit in the shared repo.

- After that, the "goreleaser" GitHub workflow then creates and publishes the artifacts for that release (as discussed [above](#technology)) and then the "Test latest release" workflow will run the E2E tests using those artifacts.

- After the release artifacts have been published, create and push to the shared repo a branch named `release-$version`. This will also trigger the workflow that tests the latest release. Every push to a branch with such a name triggers that workflow, in case there has been a change in an E2E test for that release.

- Do additional testing: more scenarios, more platforms.

## Future Process Development

We intend to get rid of the self-reference in the KCM PCH, as follows. Define a Helm chart for installing the PCH. Update the release workflow to specialize that Helm chart, similarly to the specialization done for the KCM Helm chart.

## Open questions

What to do about the dependency cycle between ks/ks and ks/ocm-transport-plugin?

Exactly when does a new release branch diverge from `main`? What about cherry-picking between `main` and the latest (or also earlier?) release branch?

What about the clusteradm container image?
