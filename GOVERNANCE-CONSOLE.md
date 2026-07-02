# KubeStellar Console Project Governance

KubeStellar Console is a subproject of [KubeStellar](https://github.com/kubestellar),
a CNCF Sandbox project. As a subproject, Console follows KubeStellar's overall
governance structure while maintaining its own Maintainer Committee for
day-to-day project decisions.

- [Values](#values)
- [Subproject Governance](#subproject-governance)
- [Maintainer Committee](#maintainer-committee)
  - [Maintainer Committee Duties](#maintainer-committee-duties)
  - [Current Maintainers](#current-maintainers)
  - [Becoming a Maintainer](#becoming-a-maintainer)
  - [Removing a Maintainer](#removing-a-maintainer)
- [Contributor Ladder](#contributor-ladder)
- [Code of Conduct](#code-of-conduct)
- [Security Response Team](#security-response-team)
- [Decision-Making](#decision-making)
- [Adding New Components](#adding-new-components)
- [Amendments](#amendments)

## Values

KubeStellar Console and its leadership embrace the following values:

* Openness: Communication and decision-making happens in the open and is
  discoverable for future reference. As much as possible, all discussions and
  work take place in public forums and open repositories.

* Fairness: All stakeholders have the opportunity to provide feedback and submit
  contributions, which will be considered on their merits.

* Community over Product or Company: Sustaining and growing our community takes
  priority over shipping code or sponsors' organizational goals. Each
  contributor participates in the project as an individual.

* Inclusivity: We innovate through different perspectives and skill sets, which
  can only be accomplished in a welcoming and respectful environment.

* Participation: Responsibilities within the project are earned through
  participation, and there is a clear path up the contributor ladder into
  leadership positions.

## Subproject Governance

KubeStellar Console is one of KubeStellar's subprojects, alongside:

* [KubeStellar Core](https://github.com/kubestellar/kubestellar) — multi-cluster
  configuration management for edge, multi-cloud, and hybrid environments
* [Hive](https://github.com/kubestellar/hive) — AI agent orchestration for
  open source project maintenance

As a subproject, Console defers to the KubeStellar Steering Committee on
cross-project matters, CNCF compliance, and trademark usage. All other
governance is handled by Console's own Maintainer Committee.

## Maintainer Committee

All active Maintainers of KubeStellar Console, as defined in the Contributor
Ladder, are members of the Maintainer Committee, which governs the project.
Maintainers have write access to the [project GitHub repository](https://github.com/kubestellar/console)
and can merge their own patches or patches from others. The current maintainers
can be found as top-level approvers in [OWNERS](OWNERS).

This privilege is granted with some expectation of responsibility: maintainers
are people who care about the KubeStellar Console project and want to help it
grow and improve. A maintainer is not just someone who can make changes, but
someone who has demonstrated their ability to collaborate with the team, get the
most knowledgeable people to review code and docs, contribute high-quality code,
and follow through to fix issues (in code or tests).

### Maintainer Committee Duties

The Maintainer Committee is responsible for the following governance activities:

* Ensuring that the project creates and publishes regular releases;
* Holding regular, project-wide discussions on issues and planning;
* Monthly review of project contributors for advancement on the Contributor
  Ladder;
* Making final decisions on project changes that involve controversial
  trade-offs;
* Responding to security compromise reports;
* Supporting the Code of Conduct within the project and referring violations to
  the KubeStellar Code of Conduct Committee;
* Selecting one representative to the KubeStellar Steering Committee annually.

### Current Maintainers

| Name           | GitHub                                            | Role       |
|----------------|---------------------------------------------------|------------|
| Andy Anderson  | [@clubanderson](https://github.com/clubanderson)  | Maintainer |

### Becoming a Maintainer

To become a Maintainer you need to demonstrate the following:

* commitment to the project:
  * participate in discussions, contributions, code and documentation reviews
    for 3 months or more,
  * perform reviews for 5 non-trivial pull requests,
  * contribute 5 non-trivial pull requests and have them merged,
* ability to write quality code and/or documentation,
* ability to collaborate with the team,
* understanding of how the team works (policies, processes for testing and code
  review, etc),
* understanding of the project's code base and coding and documentation style.

A new Maintainer must be proposed by an existing maintainer by sending a message
to the [developer mailing list](https://groups.google.com/g/kubestellar-dev). A
simple majority vote of existing Maintainers approves the application.

### Removing a Maintainer

Maintainers may resign at any time if they feel that they will not be able to
continue fulfilling their project duties.

Maintainers may also be removed after being inactive, failure to fulfill their
Maintainer responsibilities, violating the Code of Conduct, or other reasons.
Inactivity is defined as a period of very low or no activity in the project for
a year or more, with no definite schedule to return to full Maintainer activity.

A Maintainer may be removed at any time by a 2/3 vote of the remaining
maintainers, or a majority vote of the KubeStellar Steering Committee.

## Contributor Ladder

KubeStellar Console follows the [KubeStellar Contributor Ladder](https://github.com/kubestellar/community/blob/main/CONTRIBUTOR_LADDER.md)
with the following project-specific roles:

* **Contributor**: Anyone who has submitted a pull request, filed an issue, or
  participated in project discussions.
* **Organization Member**: Contributors who have made sustained contributions
  and are recognized by the Maintainer Committee. Organization Members may be
  granted triage permissions.
* **Reviewer**: Organization Members who have demonstrated expertise in one or
  more areas of the codebase and are listed in the OWNERS file as reviewers.
* **Maintainer**: Reviewers who have demonstrated broad project knowledge,
  sustained contribution, and sound technical judgment. Maintainers have write
  access and are listed as approvers in the OWNERS file.

Advancement on the Contributor Ladder is reviewed monthly by the Maintainer
Committee. Contributors may request advancement by opening an issue or
contacting a Maintainer directly.

## Code of Conduct

KubeStellar Console follows the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/main/code-of-conduct.md).
Violations by community members will be discussed and resolved on the
[private Maintainer mailing list](https://groups.google.com/u/1/g/kubestellar-dev-private).
Violations may also be reported directly to the CNCF at conduct@cncf.io.

## Security Response Team

The Maintainers will appoint a Security Response Team to handle security
reports. This committee may simply consist of the Maintainer Committee
themselves. The Security Response Team is responsible for handling all reports
of security holes and breaches according to the [security policy](SECURITY.md).

## Decision-Making

Decisions within KubeStellar Console are made using a consensus-seeking process:

1. **Proposals** are submitted as GitHub issues or pull requests.
2. **Discussion** happens in the open on the issue or PR.
3. **Consensus** is sought among Maintainers. If consensus cannot be reached,
   decisions are made by majority vote of the Maintainer Committee.
4. **Escalation**: Decisions that affect other KubeStellar subprojects or that
   the Maintainer Committee cannot resolve may be escalated to the KubeStellar
   Steering Committee.

Lazy consensus applies to routine decisions: if no objection is raised within
a reasonable period (typically 72 hours for non-trivial changes), the proposal
is accepted.

Votes can be taken on [the developer mailing list](https://groups.google.com/g/kubestellar-dev)
or [the private Maintainer mailing list](https://groups.google.com/u/1/g/kubestellar-dev-private)
for security or conduct matters. Most votes require a simple majority of all
Maintainers to succeed.

## Adding New Components

New components (dashboard cards, integrations, API endpoints) may be proposed by
any contributor via a GitHub issue. The Maintainer Committee will evaluate
proposals based on:

* Alignment with Console's mission of multi-cluster observability and management;
* Whether the component is appropriately licensed (Apache-2.0);
* Whether the component is under active development;
* Code and design quality.

Experimental components may be accepted with an "Experimental" designation and
will be reviewed periodically for promotion to stable status.

## Amendments

This governance document can be amended with a 2/3 majority vote of the
Maintainer Committee. Unless time is of the essence, amendments should be
circulated in the contributor community for comment for at least one week
before voting.

Amendments that affect Console's relationship to the KubeStellar umbrella
project require approval from the KubeStellar Steering Committee.
