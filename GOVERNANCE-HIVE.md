# Hive Project Governance

Hive is a subproject of [KubeStellar](https://github.com/kubestellar), a CNCF
Sandbox project. As a subproject, Hive follows KubeStellar's overall governance
structure while maintaining its own Maintainer Committee for day-to-day
project decisions.

- [Values](#values)
- [Subproject Governance](#subproject-governance)
- [Maintainer Committee](#maintainer-committee)
  - [Maintainer Committee Duties](#maintainer-committee-duties)
  - [Current Maintainers](#current-maintainers)
- [Contributor Ladder](#contributor-ladder)
- [Code of Conduct](#code-of-conduct)
- [Decision-Making](#decision-making)
- [Adding New Components](#adding-new-components)
- [Amendments](#amendments)

## Values

Hive and its leadership embrace the following values:

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

Hive is one of KubeStellar's subprojects, alongside:

* [KubeStellar Core](https://github.com/kubestellar/kubestellar) — multi-cluster
  configuration management for edge, multi-cloud, and hybrid environments
* [KubeStellar Console](https://github.com/kubestellar/console) — multi-cluster
  dashboard and observability

As a subproject, Hive defers to the KubeStellar Steering Committee on
cross-project matters, CNCF compliance, and trademark usage. All other
governance is handled by Hive's own Maintainer Committee.

## Maintainer Committee

All active Maintainers of Hive, as defined in the Contributor Ladder, are
members of the Maintainer Committee, which governs the project.

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

Should a member of the Maintainer Committee cease being active in the project,
violate the Code of Conduct, or need to be removed for some other reason, they
may be removed by a 2/3 majority vote of the other Committee members, or a
majority vote of the KubeStellar Steering Committee.

### Current Maintainers

| Name           | GitHub                                            | Role       |
|----------------|---------------------------------------------------|------------|
| Andy Anderson  | [@clubanderson](https://github.com/clubanderson)  | Maintainer |

## Contributor Ladder

Hive follows the [KubeStellar Contributor Ladder](https://github.com/kubestellar/community/blob/main/CONTRIBUTOR_LADDER.md)
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

Hive follows the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/main/code-of-conduct.md).
Violations may be reported to the KubeStellar Code of Conduct Committee or
directly to the CNCF at conduct@cncf.io.

## Decision-Making

Decisions within Hive are made using a consensus-seeking process:

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

## Adding New Components

New components (agent types, policy engines, integrations) may be proposed by
any contributor via a GitHub issue. The Maintainer Committee will evaluate
proposals based on:

* Alignment with Hive's mission of AI agent orchestration for open source;
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

Amendments that affect Hive's relationship to the KubeStellar umbrella project
require approval from the KubeStellar Steering Committee.
