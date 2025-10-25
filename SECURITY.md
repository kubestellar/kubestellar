## Security Announcements

Join the [kubestellar-security-announce](https://groups.google.com/u/1/g/kubestellar-security-announce) group for emails about security and major API announcements.

## Dependencies Policy

KubeStellar manages its dependencies with the following policy:

- **Dependency Detection:** We use [Dependabot](https://github.com/dependabot) to automatically check for and propose updates to dependencies in Go modules, Python requirements, Dockerfiles, Helm charts, and GitHub Actions workflows. Dependabot PRs serve as prompts but are not automatically accepted.
- **Update Process:** After Dependabot creates a PR, maintainers wait for potential issues to surface before proceeding. The handling then depends on the type of dependency and whether Dependabot's proposal is functional:
    - **GitHub Actions:** Maintainers create their own PR that follows our [GitHub Action reference discipline](https://github.com/kubestellar/kubestellar/blob/main/CONTRIBUTING.md#github-action-reference-discipline) and other established practices.
    - **Go Dependencies:** If Dependabot's proposal is functional, it may be accepted directly. If the proposal is broken, maintainers create their own PR to address the dependency update properly.
- **Review Process:** All dependency update pull requests are subject to the same [review process](https://github.com/kubestellar/kubestellar/blob/main/CONTRIBUTING.md#pull-requests) as other code changes. Maintainers verify that updates do not introduce breaking changes or known vulnerabilities before merging.
- **Vulnerability Checking:** Before merging dependency updates, maintainers perform security assessments:
    - **Security Scanning:** Given that KubeStellar imports various types of dependencies (Go packages, pre-built binaries, container images, Helm charts, and GitHub Actions), we rely on GitHub's security advisory database and Dependabot's vulnerability detection capabilities. Specific additional security scanning tools are not currently standardized across all dependency types.
    - **Security Advisories:** Review security advisories and release notes for the updated dependencies
    - **Breaking Changes:** Verify that updates do not introduce breaking changes or compatibility issues
    - **GitHub Actions:** For GitHub Actions specifically, ensure updates follow our [GitHub Action Reference Discipline](https://github.com/kubestellar/kubestellar/blob/main/CONTRIBUTING.md#github-action-reference-discipline) and use approved commit hashes. The [verify-action-hashes workflow](https://github.com/kubestellar/kubestellar/blob/main/.github/workflows/verify-action-hashes.yaml) automatically checks that each GitHub Action reference uses an approved commit hash.
    - **SBOM Generation:** Generate Software Bill of Materials (SBOM) using [Anchore's syft tool](https://github.com/kubestellar/kubestellar/blob/main/.github/workflows/goreleaser.yml) during releases to identify and track dependencies for security analysis
    - **Testing:** Run available tests to verify that updated dependencies work correctly with the codebase
- **Security Best Practices:** We avoid using unmaintained or deprecated dependencies. Monitoring for security advisories affecting our dependencies is primarily done through GitHub's security advisory database and Dependabot notifications. Vulnerabilities in dependencies are prioritized for prompt remediation.
- **Documentation:** The dependency update process is documented in the repository's README and CONTRIBUTING guidelines.

## Report a Vulnerability

We're extremely grateful for security researchers and users that report vulnerabilities to the KubeStellar Open Source Community. All reports are thoroughly investigated by a set of community volunteers.

You can also email the private [kubestellar-security-announce@googlegroups.com](mailto:kubestellar-security-announce@googlegroups.com) list with the security details and the details expected for [all KubeStellar bug reports](https://github.com/kubestellar/kubestellar/blob/main/.github/ISSUE_TEMPLATE/bug_report.yaml).

### When Should I Report a Vulnerability?

- You think you discovered a potential security vulnerability in KubeStellar
- You are unsure how a vulnerability affects KubeStellar
- You think you discovered a vulnerability in another project that KubeStellar depends on
    - For projects with their own vulnerability reporting and disclosure process, please report it directly there


### When Should I NOT Report a Vulnerability?

- You need help tuning KubeStellar components for security
- You need help applying security related updates
- Your issue is not security related

## Security Vulnerability Response

Each report is acknowledged and analyzed by the maintainers of KubeStellar within 3 working days.

Any vulnerability information shared with Security Response Committee stays within KubeStellar project and will not be disseminated to other projects unless it is necessary to get the issue fixed.

As the security issue moves from triage, to identified fix, to release planning we will keep the reporter updated.

## Public Disclosure Timing

A public disclosure date is negotiated by the KubeStellar Security Response Committee and the bug submitter. We prefer to fully disclose the bug as soon as possible once a user mitigation is available. It is reasonable to delay disclosure when the bug or the fix is not yet fully understood, the solution is not well-tested, or for vendor coordination. The timeframe for disclosure is from immediate (especially if it's already publicly known) to a few weeks. For a vulnerability with a straightforward mitigation, we expect report date to disclosure date to be on the order of 7 days. The KubeStellar maintainers hold the final say when setting a disclosure date.