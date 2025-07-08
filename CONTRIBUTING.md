# Contributing to KubeStellar

_Note to contributors: the contents of this file are incorporated in sections into the KubeStellar website via the file [contributing-inc](./docs/content/contribution-guidelines/contributing-inc.md). **If you add or modify links in this file you will also need to update that file**_

---

<!--guidelines-start-->
Greetings! We are grateful for your interest in joining the KubeStellar community and making a positive impact. Whether you're raising issues, enhancing documentation, fixing bugs, or developing new features, your contributions are essential to our success.

To get started, kindly read through this document and familiarize yourself with our code of conduct. If you have any inquiries, please feel free to reach out to us on [Slack](https://kubernetes.slack.com/archives/C058SUSL5AA).

We can't wait to collaborate with you!


This document describes our policies, procedures and best practices for working on KubeStellar via the project and repository on GitHub. Much of this interaction (issues, pull requests, discussions) is meant to be viewed directly at the [KubeStellar repository webpage on GitHub](https://github.com/kubestellar/kubestellar/). Other community discussions and questions are available via our slack channel. If you have any inquiries, please feel free to reach out to us on the [KubeStellar-dev Slack channel](https://kubernetes.slack.com/archives/C058SUSL5AA/).

Please read the following guidelines if you're interested in contributing to KubeStellar.

## General practices in the KubeStellar GitHub Project

### Contributing Code -- Prerequisites
<!--end-first-include-->
Please make sure that your environment has all the necessary versions as spelled out in the prerequisites section of our [user guide](./docs/content/direct/pre-reqs.md)
<!--start-second-include-->

### Issues
[View active issues on GitHub](https://github.com/kubestellar/kubestellar/issues)

Prioritization for pull requests is given to those that address and resolve existing GitHub issues. Utilize the available issue labels to identify meaningful and relevant issues to work on.

If you believe that there is a need for a fix and no existing issue covers it, feel free to create a new one.

As a new contributor, we encourage you to start with issues labeled as **[good first issue.](https://github.com/kubestellar/kubestellar/issues?q=is%3Aissue%20state%3Aopen%20label%3A%22good%20first%20issue%22)**

We also have a subset of issues we've labeled **[help wanted!](https://github.com/kubestellar/kubestellar/labels/help%20wanted)**

Your assistance in improving documentation is highly valued, regardless of your level of experience with the project.

To claim an issue that you are interested in, kindly leave a comment on the issue and request the maintainers to assign it to you.

### Committing
We encourage all contributors to adopt [best practices in git commit management](https://hackmd.io/q22nrXjERBeIGb-fqwrUSg) to facilitate efficient reviews and retrospective analysis. Note: that document was written for projects where some of the contributors are doing merges into the main branch, but in KubeStellar we have GitHub doing that for us. For the kubestellar repository, this is controlled by [Prow](https://docs.prow.k8s.io/); for the other repositories in the kubestellar organization we use the GitHub mechanisms directly.

Your git commits should provide ample context for reviewers and future codebase readers.

A recommended format for final commit messages is as follows:

```
{Short Title}: {Problem this commit is solving and any important contextual information} {issue number if applicable}
```
In conformance with CNCF expectations, we will only merge commits that indicate your agreement with the [Developer Certificate of Origin](#certificate-of-origin). The CNCF defines how to do this, and there are two cases: one for developers working for an organization that is a CNCF member, and one for contributors acting as individuals. For the latter, assent is indicated by doing a Git "sign-off" on the commit. 

<!--end-second-include-->
See [Git Commit Signoff and Signing](./docs/content/direct/pr-signoff.md)) for more information on how to do that.
<!--start-third-include-->

### Pull Requests
[View active Pull Requests on GitHub](https://github.com/kubestellar/kubestellar/pulls)

When submitting a pull request, clear communication is appreciated. This can be achieved by providing the following information:

- Detailed description of the problem you are trying to solve, along with links to related GitHub issues
- Explanation of your solution, including links to any design documentation and discussions
- Information on how you tested and validated your solution
- Updates to relevant documentation and examples, if applicable

Following are a few more things to keep in mind when making a pull request.

- Smaller pull requests are typically easier to review and merge than larger ones. If your pull request is big, it is always recommended to collaborate with the maintainers to find the best way to divide it.
- Do not make a PR from your `main` branch. Your life will be much easier if the `main` branch in your fork tracks the `main` branch in the shared repository.
- Learn to use `git rebase`. It is your friend. It is one of your most helpful friends. It is how you can cope when other changes merge while you are in the midst of working on your PR.
- There are, broadly speaking, two styles of using Git history: keeping an accurate record of your development process, or producing a simple explanation of the end result. We aim for the latter. Squash out uninteresting intermediate commits.
- Do not merge from `main` into your PR's branch. That makes a tangled Git history, and we prefer to keep it simple. Instead, rebase your PR's branch onto the latest edition of `main`.
- When adding/updating a GitHub Actions workflow, be aware of the [action reference discipline](#github-action-reference-discipline).
- For a PR that modifies the website, include a preview. That gets much easier if you follow the documentation about setting up for that (i.e., properly create your `gh-pages` branch, enabling its use in your fork's settings) and make the name of your PR's branch start with "doc-". If you already have a PR with a different sort of name, you can explicitly invoke the rendering workflow --- unless your branch name has a slash or other exotic character in it; stick to alphanumerics plus dash and dot. You can not change the name of the branch in a PR, but you can close a PR and open an equivalent one using a branch with a good name.
- For a PR that modifies the website, remember that the doc source files are viewed two ways (see the website documentation); make them work in both views.
- If you mix pervasive changes to whitespace with substantial changes, you risk GitHub's display of the diff becoming confused. DO check that. If the diff display is confused, it makes reviewing much harder. Have mercy on your reviewers; skip the pervasive whitespace changes if they confuse GitHub's diff. BTW, did you really intend to make all those whitespace changes, or are they an unintended gift from your IDE? Don't make changes that you do not really intend.

#### Titling Pull Requests
We require that the title of each pull request start with a special nickname character (emoji) that classifies the request into one of the following categories. 

The nickname characters to use for different PRs are as follows

- ✨ (nickname `:sparkles:`) feature
- 🐛 (nickname `:bug:`) bug fix
- 📖 (nickname `:book:`) docs
- 📝 (nickname `:memo:`)  proposal
- ⚠️ (nickname `:warning:`) breaking change
- 🌱 (nickname `:seedling:`) other/misc
- ❓ (nickname `:question:`) requires manual review/categorization

---

_Note: The GitHub web interface will assist you with adding the character; while editing the title of your pull request:_

- _type a colon (':')_
- _begin typing the character nickname (_e.g._ sparkles)_
- _the web interface should offer you a pick-list of corresponding characters._
- _Just click on the correct one to insert it in the title_
- _Add at least one space after the special character._

#### Continuous Integration

Pull requests are subjected to checking by a collection of [GitHub
Actions](https://docs.github.com/en/actions) workflows and
[Prow](https://docs.prow.k8s.io/docs/overview/) jobs. The [infra
repo](https://github.com/kubestellar/infra/) defines the Prow instance
used for KubeStellar. The GitHub Actions workflows are found in [the
.github/workflows
directory](https://github.com/kubestellar/kubestellar/tree/main/.github/workflows).

##### GitHub Action reference discipline

For the sake of supply chain security, every reference from a workflow
to an action identifies the action's version by a commit hash. In
particular, there is [a
file](https://github.com/kubestellar/kubestellar/blob/main/.gha-reversemap.yml)
that lists the approved commit hash for each action. The file should
be updated/extended only when you have confidence in the new/added
version. There is [a
script](https://github.com/kubestellar/kubestellar/blob/main/hack/gha-reversemap.sh)
for updating and checking this stuff. There is a workflow that checks
that every workflow follows the discipline here.

#### Review and Approval Process

Reviewers will review your PR within a business day. A PR requires both an `/lgtm` and then an `/approve` in order to get merged. These are commands to Prow, each appearing alone on a line in a comment of the PR. You may `/approve` your own PR but you may not `/lgtm` it. Once both forms of assent have been given and the other gating checks have passed, the PR will go into the Prow merge queue and eventually be merged. Once that happens, you will be notified:

_Congratulations! Your pull request has been successfully merged!_ 👏

If you have any questions about contributing, don't hesitate to reach out to us on the KubeStellar-dev [Slack channel](https://kubernetes.slack.com/archives/C058SUSL5AA/).



## Testing Locally
<!--end-third-include-->
Our [Getting Started](./docs/content/direct/get-started.md) guide shows a user how to install a simple "kick the tires" instance of KubeStellar using a helm chart and kind.

To set up and test a development system, please refer to the [_test/e2e/README.md_](./test/e2e/README.md) file in the GitHub repository. After running any of those e2e (end to end) tests you will be left with a running system that can be exercised further.

<!--start-fourth-include-->

### Testing changes to the helm chart

<!--end-fourth-include-->
If you are interested in modifying the Helm chart itself, look at the User Guide page on the [Core Helm chart](./docs/content/direct/core-chart.md) for more information on its many options before you begin, notably on how to specify using a local version of the script.

<!--start-fifth-include-->


### Testing the script against an upcoming release

Prior to making a new release, there needs to be testing that the
current Helm chart works with the executable behavior that will
appear in the new release.  

## Licensing
<!--end-fifth-include-->
KubeStellar is [Apache 2.0 licensed](./LICENSE) and we accept contributions via GitHub pull requests.

<!--start-sixth-include-->


## Certificate of Origin

By contributing to this project you agree to the Developer Certificate of
Origin (DCO). This document was created by the Linux Kernel community and is a
simple statement that you, as a contributor, have the legal right to make the
contribution. See the [DCO]({{ config.repo_url }}/blob/{{ config.ks_branch }}/DCO)</a> file for details.
