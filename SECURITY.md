<!--security-start-->
# Security Announcements

Stay updated with the latest security advisories and major API changes by joining the [kubestellar-security-announce](https://groups.google.com/u/1/g/kubestellar-security-announce) mailing list.

# Report a Vulnerability

We appreciate the efforts of security researchers and users who report vulnerabilities to the KubeStellar Open Source Community. All reports are thoroughly investigated by a dedicated group of community volunteers.

### How to Report a Vulnerability

To report a security vulnerability, email the private [kubestellar-security-announce@googlegroups.com](mailto:kubestellar-security-announce@googlegroups.com) list with detailed information about the vulnerability. Ensure that you include the information required for [all KubeStellar bug reports](https://github.com/kubestellar/kubestellar/blob/main/.github/ISSUE_TEMPLATE/bug_report.yaml).

### When Should You Report a Vulnerability?

- You believe you’ve discovered a potential security vulnerability in KubeStellar.
- You are unsure of how a vulnerability might affect KubeStellar.
- You believe you've found a vulnerability in a dependency used by KubeStellar.

### When Should You NOT Report a Vulnerability?

- You need assistance in configuring KubeStellar components for security.
- You need help applying security-related updates.
- The issue you encountered is not security-related.

# Security Vulnerability Response

Each security report is acknowledged and analyzed by the KubeStellar maintainers within 3 working days.

Vulnerability information shared with the Security Response Committee will remain confidential within the KubeStellar project until a resolution is confirmed. It will not be disseminated to other projects unless required for resolving the issue.

As the security issue progresses from triage to an identified fix, and finally to release planning, we will keep the reporter updated.

# Public Disclosure Policy

KubeStellar is committed to transparency and timely disclosure. The public disclosure date for each vulnerability is negotiated between the Security Response Committee and the reporter. Our preferred approach is to disclose the vulnerability as soon as a user mitigation is available.

However, disclosure may be delayed for reasons such as:
- The vulnerability fix is not yet fully understood or tested.
- Coordination is required with vendors or other affected projects.

In cases where a vulnerability has a straightforward mitigation, we aim for a disclosure timeline of approximately 7 days from the report date. For more complex cases, the disclosure period may extend to a few weeks.

The KubeStellar maintainers hold the final say in setting the disclosure date.

# Security Best Practices

KubeStellar encourages contributors to adhere to secure coding practices:

- Always validate inputs to prevent injection attacks.
- Use strong cryptographic algorithms (e.g., AES-256, RSA).
- Ensure sensitive data is encrypted both in transit and at rest.

Contributors are required to undergo **security awareness training** to follow secure development practices and to help prevent common security pitfalls. All security-conscious changes will be validated and monitored in the project’s continuous integration pipeline.

We recommend integrating automated static code analysis tools such as Bandit and Semgrep into your development workflow to automatically catch vulnerabilities early in the development cycle.

# Reporting Process for Dependencies

If you identify a security issue in a dependency that KubeStellar uses, please report it directly to the maintainers of that project. However, if the vulnerability impacts KubeStellar, please also notify us so that we can track it and apply necessary patches to our codebase.

# Security Patches and Fixes

Security vulnerabilities will be patched and deployed within 14 days from the date they are reported, unless otherwise stated. The patch will be included in the next available release, and we will provide upgrade instructions for affected users.

## Contact

For any security-related inquiries, please contact [security@example.com](mailto:security@example.com).
<!--security-end-->
