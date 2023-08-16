# Contributing to KubeStellar
Greetings! We are grateful for your interest in joining the KubeStellar community and making a positive impact. Whether you're raising issues, enhancing documentation, fixing bugs, or developing new features, your contributions are essential to our success.

To get started, kindly read through this document and familiarize yourself with our code of conduct. If you have any inquiries, please feel free to reach out to us on the KubeStellar-dev [Slack channel](https://kubernetes.slack.com/archives/C058SUSL5AA/).

We can't wait to collaborate with you!

## Contributing Code

### Prerequisites

#### Go

[Install Go](https://golang.org/doc/install/) 1.19+.  See [this gist](https://gist.github.com/jniltinho/8758e15a9ef80a189fce) for another way to install Go.
  Please note that the go language version numbers in these files must exactly agree:
  
    Your local go/go.mod file, kcp/.ci-operator.yaml, and in all the kcp/.github/workflows yaml files that specify go-version.
    
    - In ./ci-operator.yaml the go version is indicated by the "tag" attribute.
    - In go.mod it is indicated by the "go" directive.
    - In the .github/workflows yaml files it is indicated by "go-version"
    
Check out our [QuickStart Guide](../../Getting-Started/quickstart/)

#### Other packages

- GNU make
- [__ko__](https://github.com/ko-build/ko) (required for compiling KubeStellar Syncer)
  - [__slsa-verifier__](https://github.com/slsa-framework) needed in Ubuntu for ko signing

### Issues
Prioritization for pull requests is given to those that address and resolve existing GitHub issues. Utilize the available issue labels to identify meaningful and relevant issues to work on.

If you believe that there is a need for a fix and no existing issue covers it, feel free to create a new one.

As a new contributor, we encourage you to start with issues labeled as good first issues.

Your assistance in improving documentation is highly valued, regardless of your level of experience with the project.

To claim an issue that you are interested in, kindly leave a comment on the issue and request the maintainers to assign it to you.

### Committing
We encourage all contributors to adopt [best practices in git commit management](https://gist.github.com/luismts/495d982e8c5b1a0ced4a57cf3d93cf60) to facilitate efficient reviews and retrospective analysis. Your git commits should provide ample context for reviewers and future codebase readers.

A recommended format for final commit messages is as follows:

```
{Short Title}: {Problem this commit is solving and any important contextual information} {issue number if applicable}
```
### Pull Requests
When submitting a pull request, clear communication is appreciated. This can be achieved by providing the following information:

- Detailed description of the problem you are trying to solve, along with links to related GitHub issues
- Explanation of your solution, including links to any design documentation and discussions
- Information on how you tested and validated your solution
- Updates to relevant documentation and examples, if applicable

The pull request template has been designed to assist you in communicating this information effectively.

Smaller pull requests are typically easier to review and merge than larger ones. If your pull request is big, it is always recommended to collaborate with the maintainers to find the best way to divide it.

Approvers will review your PR within a business day. A PR requires both an /lgtm and then an /approve in order to get merged. You may /approve your own PR but you may not /lgtm it. Automation will add the PR it to the OpenShift PR merge queue. The OpenShift Tide bot will automatically merge your work when it is available.

Congratulations! Your pull request has been successfully merged! 👏

If you have any questions about contributing, don't hesitate to reach out to us on the KubeStellar-dev [Slack channel](https://kubernetes.slack.com/archives/C058SUSL5AA/).

## Testing Locally

Our [QuickStart](../../Getting-Started/quickstart/)
 guide shows a user how to install a
local KCP server and install the KubeStellar components and run an
example.  As a contributor you will want a different setup flow,
including `git clone` of this repo instead of fetching and unpacking a
release archive.  The same example usage should work for you, and
there is a larger example at [this link](../../Coding%20Milestones/PoC2023q1/example1/).

### Testing changes to the KubeStellar central container image

If you make a change that affects the container image holding the
central components then you will need to build a new image; perhaps
surprisingly, this is not included in `make build`.  The regular way
to build this image is with the following command.  It builds a
multi-platform image, for all the platforms that KubeStellar can run
its central components on, and pushes it to quay.io.

```bash
make kubestellar-image
```

However, that will only succeed if you have done `docker login` to
quay.io with credentials authorized to write to the
`kubestellar/kubestellar` repository.  Look on quay.io to find the
image you just pushed, you will need its tag.

For a less pushy alternative you can build a single-platform image and
not push it, using the following command.

```bash
make kubestellar-image-local
```

Follow that with `docker images` to find the tag of the image you just
built.  Get that image:tag known where you are going to run the
central container; for example, if that will be in a local `kind`
cluster then you can use [kind
load](https://kind.sigs.k8s.io/docs/user/quick-start#loading-an-image-into-your-cluster).

To get the image you just built used in your testing, edit
`scripts/kubectl-kubestellar-deploy` and update the line that defines
`$image_tag`; follow this with your `make build`.  For the sake of
future users of a merged change, your last edit like this should refer
to a tag that you pushed to quay.io/kubestellar/kubestellar.

### Testing changes to the bootstrap script

The quickstart says to fetch the [bootstrap script]({{ config.repo_url }}/blob/{{ config.ks_branch }}/bootstrap/bootstrap-kubestellar.sh) from the {{ config.ks_branch }} branch of
the KubeStellar repo; if you want to contribute a change to that script then
you will need to test your changed version.  Just run your local copy
(perhaps in a special testing directory, just to be safe) and be sure
to add the downloaded `bin` at the _front_ of your `$PATH` (contrary
to [what the scripting currently tells
you]({{ config.repo_url }}/blob/{{ config.ks_branch }}/bootstrap/bootstrap-kubestellar.sh)) so that your `git clone`'s `bin` does not shadow the one being tested.

Note that changes to the bootstrap script start being used by users as
soon as your PR merges.  Since this script can only fetch a released
version of the executables, changes to this script can not rely on any
behavior of those executables that is not in the currently latest
release.  Also, a change that restricts the range of usable releases
needs to add checking for use of incompatible releases.

### Testing the bootstrap script against an upcoming release

Prior to making a new release, there needs to be testing that the
current bootstrap script works with the executable behavior that will
appear in the new release.  To support this we will add an option to
the bootstrap script that enables it to use a local release archive
instead of fetching an archive of an actual release from github.

## Licensing
KubeStellar is [Apache 2.0 licensed](LICENSE.md) and we accept contributions via
GitHub pull requests.

Please read the following guide if you're interested in contributing to KubeStellar.

## Certificate of Origin

By contributing to this project you agree to the Developer Certificate of
Origin (DCO). This document was created by the Linux Kernel community and is a
simple statement that you, as a contributor, have the legal right to make the
contribution. See the [DCO]({{ config.repo_url }}/blob/{{ config.ks_branch }}/DCO)</a> file for details.
