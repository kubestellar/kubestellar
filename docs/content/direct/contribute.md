# Contributing to KubeStellar

Welcome to the KubeStellar Contribution Guide! We are excited to have you here.

You can join the community via our [Slack channel](https://kubernetes.slack.com/archives/C058SUSL5AA/).

This section provides information on the Code of Conduct, guidelines, terms, and conditions that define the KubeStellar contribution processes. By contributing, you are enabling the success of KubeStellar users, and that goes a long way to make everyone happier, including you. We welcome individuals who are new to open-source contributions.

There are different ways you can contribute to the KubeStellar development:

- **Documentation:** Enhance the documentation by fixing typos, enabling semantic clarity, adding links, updating information on changelogs and release versions, and implementing content strategy.
- **Code:** Indicate your interest in developing new features, modifying existing features, raising concerns, or fixing bugs.

Before you start contributing, familiarize yourself with our community [Code of Conduct](../contribution-guidelines/coc-inc.md).

## Visit the GitHub repository

The KubeStellar [GitHub organization](https://github.com/kubestellar) is a collection of the different KubeStellar repositories that you can start contributing to.

### Sign off your contribution

Ensure that you comply with the rules and policy guiding the repository contribution indicated in the [Developer Certificate of Origin (DCO)](https://github.com/kubestellar/kubestellar/blob/main/DCO).

If you are contributing via the GitHub web interface, navigate to the **Settings** section of your forked repository and enable the **Require contributors to sign off on web-based commits** setting. This will allow you to automatically sign off your commits via GitHub directly, as shown below.

![signoff-via-github-ui](https://github.com/user-attachments/assets/ddfd3988-142e-4380-a738-1a767b1aaba6)

If you are contributing via the command line terminal, run the `git commit --signoff --message [commit message]` or `git commit -s -m [commit message]` command when making each commit.

## Contribution Resources

Read the resources to gain a better understanding of the contribution processes.

- **[Code of Conduct](../contribution-guidelines/coc-inc.md)** The CNCF code of conduct for the KubeStellar community
- **[Contribution Guidelines](../contribution-guidelines/contributing-inc.md)** General Guidelines for our Github processes
- **[License](../contribution-guidelines/license-inc.md)** The Apache 2.0 license under which KubeStellar is published
- **[Governance](../contribution-guidelines/governance-inc.md)** The protocols under which the KubeStellar project is run
- **[Onboarding](../contribution-guidelines/onboarding-inc.md)** The procedures for adding/removing members of our Github organization
- **Website**
  - **[Build Overview](../contribution-guidelines/operations/document-management.md)** How our website is built and how to collaboratively work on changes to it using Github staging
  - **[Testing website PRs](../contribution-guidelines/operations/testing-doc-prs.md)** how to test website changes using only your local workstation
- **Security**
  - **[Policy](../contribution-guidelines/security/security-inc.md)** Security Policies
  - **[Contacts](../contribution-guidelines/security/security_contacts-inc.md)** Who to contact with security concerns
- **[Testing](testing.md)** How to use the preconfigured tests in the repository
- **[Packaging](packaging.md)** How the components of KubeStellar are organized
- **[Release Process](release.md)** All the steps involved in creating and publishing a new release of KubeStellar
- **[Release Testing](release-testing.md)** Steps involved in testing a release or release candidate before merging it into the main branch.
- **[Sign-off and Signing Contributions](pr-signoff.md)** How to properly configure your commits so they are both signed and "signed off" (and how those terms differ)
