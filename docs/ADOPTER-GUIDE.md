# KubeStellar Adopter Guide

Are you using KubeStellar in your organization? We would love to add you to [ADOPTERS.md](../ADOPTERS.md) and welcome you to the KubeStellar community.

This guide explains what counts as an adopter, which maturity tier fits your situation, and how to submit your entry.

---

## Who should be listed?

Any organization that uses KubeStellar in one of the following ways:

| Maturity Level | Criteria |
|----------------|----------|
| **Evaluating** | Running KubeStellar in a dev, test, or PoC environment; exploring it for a future production use case |
| **Workload Integration** | Integrated KubeStellar into at least one real workload (e.g., propagating actual Kubernetes resources across clusters); may still be pre-production |
| **Production** | Running KubeStellar in production with at least one business workload; clusters are managed by KubeStellar's BindingPolicy system |

**Academic, research, and non-profit organizations are fully welcome.** See Cornell University's entry as an example.

If you are unsure, open a PR anyway — the maintainers are happy to help determine the right tier.

---

## How to add yourself

### Option 1: Direct PR (fastest)

1. Fork [kubestellar/kubestellar](https://github.com/kubestellar/kubestellar)
2. Edit [ADOPTERS.md](../ADOPTERS.md) and add a row to the table using the template below
3. Open a pull request with the title: `📖 docs: add [Your Organization] to ADOPTERS.md`

### Option 2: Open an issue

If you'd prefer not to open a PR directly, open an issue with the label `kind/documentation` and include the information below. A maintainer will add your entry.

---

## Entry template

Copy and paste this row into the ADOPTERS.md table:

```markdown
| Your Organization | Brief description of how you use KubeStellar | Evaluating / Workload Integration / Production | [Link text](https://link-to-your-project-or-org) |
```

**Fields:**

- **Organization**: Your company, institution, or project name
- **Description**: 1–2 sentences on *how* you use KubeStellar (e.g., which problem it solves, what workloads you propagate, how many clusters)
- **Maturity Level**: Choose one: Evaluating / Workload Integration / Production
- **Further Information**: A link to your project repo, case study, blog post, or your organization's homepage

---

## What information do you need to share?

You only need to share what you are comfortable making public. At minimum:

- Your organization's name (or a pseudonym/project name if preferred)
- A short description of your use case
- Your maturity level

You do **not** need to share: cluster counts, workload details, architecture diagrams, team size, or any proprietary information.

---

## Privacy and consent

By opening a PR or issue to add your entry, you consent to your organization's name and use-case description being publicly displayed in the ADOPTERS.md file and on the KubeStellar website.

To remove or update your entry at any time, open a PR editing ADOPTERS.md or contact us on [CNCF Slack `#kubestellar-dev`](https://cloud-native.slack.com/archives/C097094RZ3M/).

---

## Example entries

From the current ADOPTERS.md:

| Organization | Description | Maturity Level | Further Information |
|---|---|---|---|
| Cornell University | Workload orchestration for the Software-Defined Farm (SDF) System | Workload Integration | [SDF Repo](https://github.com/Cornell-CIDA-Dev/Software-Defined-Farm) |

---

## Why does this matter?

KubeStellar is a [CNCF Sandbox project](https://www.cncf.io/projects/kubestellar/). Growing the ADOPTERS.md file:

- Helps other organizations evaluate KubeStellar with confidence
- Strengthens the project's case for CNCF Incubation
- Signals community health to potential contributors and sponsors
- Helps maintainers prioritize features that matter to real users

Every listing — even at the "Evaluating" tier — helps the project.

---

## Community

- [CNCF Slack `#kubestellar-dev`](https://cloud-native.slack.com/archives/C097094RZ3M/)
- [Community meetings](https://github.com/kubestellar/kubestellar/blob/main/CONTRIBUTING.md)
- [kubestellar-users mailing list](https://groups.google.com/g/kubestellar-users)
