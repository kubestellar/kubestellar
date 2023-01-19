# Contributing to edge-mc

edge-mc is [Apache 2.0 licensed](LICENSE) and we accept contributions via
GitHub pull requests.

Please read the following guide if you're interested in contributing to kcp.

## Certificate of Origin

By contributing to this project you agree to the Developer Certificate of
Origin (DCO). This document was created by the Linux Kernel community and is a
simple statement that you, as a contributor, have the legal right to make the
contribution. See the [DCO](DCO) file for details.

## Expectations around PRs

A branch submitted in a PR should consist only of a strict chain of
commits built on a commit that was in the `main` branch --- no merges
into the chain.  Rebase if necessary.  Do not make the chain
unnecessarily long; squash together commits that have no useful
distinction going forward.

Unlike the automation for Kubernetes, our CI automation does not
automatically add "approved" when the submited changes are all from
someone in the relevant OWNERS.  However, an author in OWNERS may
manually approve their own PR; this has to be done with an `/approve`
comment, because github does not allow the submitter to approve their
own PR.

Each PR should get assent from at least one other person besides the
author(s) before merging.  Typically this means an LGTM from someone
else who does a comprehensive review.  The author(s) should not LGTM
their own work.

A PR should get prompt review.  That includes an initial reaction
within one business day.

## Prerequisites

[Install Go](https://golang.org/doc/install) 1.19+.
