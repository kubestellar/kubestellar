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
We encourage all contributors to adopt [best practices in git commit management](https://gist.github.com/luismts/495d982e8c5b1a0ced4a57cf3d93cf60) to facilitate efficient reviews and retrospective analysis. Your git commits should provide ample context for reviewers and future codebase readers.

A recommended format for final commit messages is as follows:

```
{Short Title}: {Problem this commit is solving and any important contextual information} {issue number if applicable}
```
### Pull Requests
[View active Pull Requests on GitHub](https://github.com/kubestellar/kubestellar/pulls)

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

## Licensing
KubeStellar is [Apache 2.0 licensed](license.md) and we accept contributions via GitHub pull requests.


## Certificate of Origin

By contributing to this project you agree to the Developer Certificate of
Origin (DCO). This document was created by the Linux Kernel community and is a
simple statement that you, as a contributor, have the legal right to make the
contribution. See the [DCO]({{ config.repo_url }}/blob/{{ config.ks_branch }}/DCO)</a> file for details.
