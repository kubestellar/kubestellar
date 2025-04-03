# Contributing to KubeStellar
<!--guidelines-start-->
Greetings! We are grateful for your interest in joining the KubeStellar community and making a positive impact. Whether you're raising issues, enhancing documentation, fixing bugs, or developing new features, your contributions are essential to our success.

To get started, kindly read through this document and familiarize yourself with our code of conduct. If you have any inquiries, please feel free to reach out to us on [Slack](https://kubernetes.slack.com/archives/C058SUSL5AA).

We can't wait to collaborate with you!


This document describes our policies, procedures and best practices for working on KubeStellar via the project and repository on GitHub. Much of this interaction (issues, pull requests, discussions) is meant to be viewed directly at the [KubeStellar repository webpage on GitHub](https://github.com/kubestellar/kubestellar/). Other community discussions and questions are available via our slack channel. If you have any inquiries, please feel free to reach out to us on the [KubeStellar-dev Slack channel](https://kubernetes.slack.com/archives/C058SUSL5AA/).

Please read the following guidelines if you're interested in contributing to KubeStellar.

## General practices in the KubeStellar GitHub Project

### Contributing Code -- Prerequisites

Please make sure that your environment has all the necessary versions as spelled out in the prerequisites section of our [user guide](../direct/pre-reqs.md)
<!--end-first-include-->
(If you are viewing this file in the repository, the [pre-req listing is in the docs subfolder](./docs/content/direct/pre-reqs.md))
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

See [Git Commit Signoff and Signing](../direct/pr-signoff.md) for more information on how to do that.

<!--end-second-include-->
(If you are viewing this file in the repository, the [Git Signoff information is the docs subfolder](./docs/content/direct/pr-signoff.md))
<!--start-third-include-->


### Pull Requests
[View active Pull Requests on GitHub](https://github.com/kubestellar/kubestellar/pulls)

When submitting a pull request, clear communication is appreciated. This can be achieved by providing the following information:

- Detailed description of the problem you are trying to solve, along with links to related GitHub issues
- Explanation of your solution, including links to any design documentation and discussions
- Information on how you tested and validated your solution
- Updates to relevant documentation and examples, if applicable

The pull request template has been designed to assist you in communicating this information effectively.

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

#### Pull Request Process
Smaller pull requests are typically easier to review and merge than larger ones. If your pull request is big, it is always recommended to collaborate with the maintainers to find the best way to divide it.

Approvers will review your PR within a business day. A PR requires both an /lgtm and then an /approve in order to get merged. You may /approve your own PR but you may not /lgtm it. Automation will add the PR it to the OpenShift PR merge queue. The OpenShift Tide bot will automatically merge your work when it is available, and you will be notified:

_Congratulations! Your pull request has been successfully merged!_ 👏

If you have any questions about contributing, don't hesitate to reach out to us on the KubeStellar-dev [Slack channel](https://kubernetes.slack.com/archives/C058SUSL5AA/).



## Testing Locally

Our [Getting Started](./docs/content/direct/get-started.md) guide shows a user how to install a simple "kick the tires" instance of KubeStellar using a helm chart and kind.

To set up and test a development system, please refer to the _test/e2e/README.md_ file in the GitHub repository.
After running any of those e2e (end to end) tests you will be left with a running system that can be exercised further.

<!--end-third-include-->
(If you are viewing this file in the repository, the [Getting Started guide is here](./docs/content/direct/get-started.md) and the [End to End (E2E) testing section is here"](./test/e2e/README.md) )
<!--start-fourth-include-->

### Testing changes to the helm chart

If you are interested in modifying the Helm chart itself, look at the User Guide page on the [Core Helm chart](../direct/core-chart.md) for more information on its many options before you begin, notably on how to specify using a local version of the script.

<!--end-fourth-include-->
If you are viewing this page directly in the repository the helm chart documentation is [here in the documentation tree](./docs/content/direct/core-chart.md)
<!--start-fifth-include-->


### Testing the script against an upcoming release

Prior to making a new release, there needs to be testing that the
current Helm chart works with the executable behavior that will
appear in the new release.  

## Licensing
KubeStellar is [Apache 2.0 licensed](./license-inc.md) and we accept contributions via GitHub pull requests.

<!--end-fifth-include-->
The license is also accessible in the [root of the repository](./LICENSE)
<!--start-sixth-include-->


## Certificate of Origin

By contributing to this project you agree to the Developer Certificate of
Origin (DCO). This document was created by the Linux Kernel community and is a
simple statement that you, as a contributor, have the legal right to make the
contribution. See the [DCO]({{ config.repo_url }}/blob/{{ config.ks_branch }}/DCO)</a> file for details.
