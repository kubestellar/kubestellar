<!--readme-for-root-start-->

<img alt="" width="500px" align="left" src="/KubeStellar-with-Logo.png" />

<br/>
<br/>
<br/>
<br/>

## Multi-cluster Configuration Management for Edge, Multi-Cloud, and Hybrid Cloud
<br/>

[![Generate and push docs](https://github.com/kcp-dev/edge-mc/actions/workflows/docs-gen-and-push.yml/badge.svg?branch=release-0.2)](https://github.com/kcp-dev/edge-mc/actions/workflows/docs-gen-and-push.yml)&nbsp;&nbsp;&nbsp;
[![PR Verifier](https://github.com/kcp-dev/edge-mc/actions/workflows/pr-verifier.yaml/badge.svg)](https://github.com/kcp-dev/edge-mc/actions/workflows/pr-verifier.yaml)&nbsp;&nbsp;&nbsp;
[![Open Source Helpers](https://www.codetriage.com/kcp-dev/edge-mc/badges/users.svg)](https://www.codetriage.com/kcp-dev/edge-mc)&nbsp;&nbsp;&nbsp;
[![first-timers-only](https://img.shields.io/badge/first--timers--only-friendly-blue.svg?style=flat-square)](https://www.firsttimersonly.com/)&nbsp;&nbsp;&nbsp;
[![Broken Links Crawler](https://github.com/kcp-dev/edge-mc/actions/workflows/broken-links-crawler.yml/badge.svg)](https://github.com/kcp-dev/edge-mc/actions/workflows/broken-links-crawler.yml)
[![QuickStart test](https://github.com/kcp-dev/edge-mc/actions/workflows/docs-ecutable-qs.yml/badge.svg)](https://github.com/kcp-dev/edge-mc/actions/workflows/docs-ecutable-qs.yml)&nbsp;&nbsp;&nbsp;
[![docs-ecutable - example1](https://github.com/kcp-dev/edge-mc/actions/workflows/docs-ecutable-example1.yml/badge.svg?branch=release-0.2)](https://github.com/kcp-dev/edge-mc/actions/workflows/docs-ecutable-example1.yml)&nbsp;&nbsp;&nbsp;
[![docs-ecutable - placement-translator](https://github.com/kcp-dev/edge-mc/actions/workflows/docs-ecutable-placement.yml/badge.svg?branch=release-0.2)](https://github.com/kcp-dev/edge-mc/actions/workflows/docs-ecutable-placement.yml)&nbsp;&nbsp;&nbsp;
[![docs-ecutable - mailbox-controller](https://github.com/kcp-dev/edge-mc/actions/workflows/docs-ecutable-mailbox.yml/badge.svg?branch=release-0.2)](https://github.com/kcp-dev/edge-mc/actions/workflows/docs-ecutable-mailbox.yml)&nbsp;&nbsp;&nbsp;
[![docs-ecutable - kubestellar-scheduler](https://github.com/kcp-dev/edge-mc/actions/workflows/docs-ecutable-scheduler.yml/badge.svg?branch=release-0.2)](https://github.com/kcp-dev/edge-mc/actions/workflows/docs-ecutable-scheduler.yml)&nbsp;&nbsp;&nbsp;
[![docs-ecutable - kubestellar-syncer](https://github.com/kcp-dev/edge-mc/actions/workflows/docs-ecutable-syncer.yml/badge.svg?branch=release-0.2)](https://github.com/kcp-dev/edge-mc/actions/workflows/docs-ecutable-syncer.yml)&nbsp;&nbsp;&nbsp;
<a href="https://kubernetes.slack.com/archives/C058SUSL5AA"> 
    <img alt="Join Slack" src="https://img.shields.io/badge/KubeStellar-Join%20Slack-blue?logo=slack">
  </a>
## Overview
KubeStellar is an opensource project focused on concerns arising from multi-cluster configuration management for edge, multi-cloud, and hybrid cloud use cases:

- Hierarchy, infrastructure & platform, roles & responsibilities, integration architecture, security issues
- Runtime in[ter]dependence: An edge location may need to operate independently of the center and other edge locations​
- Non-namespaced objects: need general support
- Cardinality of destinations: A source object may propagate to many thousands of destinations. 

## Goals

- Collaboratively design a component set similar to those found in the current kcp TMC implementation (dedicated Workspace type, scheduler, syncer-like mechanism, edge placement object definition, status collection strategy, etc.)
- Specify a multi-phased proof-of-concept inclusive of component architecture, interfaces, and example workloads
- Validate phases of proof-of-concept with kcp, Kube SIG-Multicluster, and CNCF community members interested in Edge

## Areas of exploration

- Desired placement expression​: Need a way for one center object to express large number of desired copies​
- Scheduling/syncing interface​: Need something that scales to large number of destinations​
- Rollout control​: Client needs programmatic control of rollout, possibly including domain-specific logic​
- Customization: Need a way for one pattern in the center to express how to customize for all the desired destinations​
- Status from many destinations​: Center clients may need a way to access status from individual edge copies
- Status summarization​: Client needs a way to specify how statuses from edge copies are processed/reduced along the way from edge to center​.

## QuickStart

Checkout our [QuickStart Guide](https://docs.kubestellar.io/stable/Getting-Started/quickstart/)

## Contributing

We ❤️ our contributors! If you're interested in helping us out, please head over to our [Contributing](https://docs.kubestellar.io/stable/Contribution%20guidelines/CONTRIBUTING/) guide.

## Getting in touch

There are several ways to communicate with us:

- The [`#kcp-dev` channel](https://app.slack.com/client/T09NY5SBT/C021U8WSAFK) in the [Kubernetes Slack workspace](https://slack.k8s.io)
- Our mailing lists:
    - [kubestellar-dev](https://groups.google.com/g/kubestellar-dev) for development discussions
    - [kubestellar-users](https://groups.google.com/g/kubestellar-users) for discussions among users and potential users
- Subscribe to the [community calendar](https://calendar.google.com/calendar/event?action=TEMPLATE&tmeid=MWM4a2loZDZrOWwzZWQzZ29xanZwa3NuMWdfMjAyMzA1MThUMTQwMDAwWiBiM2Q2NWM5MmJlZDdhOTg4NGVmN2ZlOWUzZjZjOGZlZDE2ZjZmYjJmODExZjU3NTBmNTQ3NTY3YTVkZDU4ZmVkQGc&tmsrc=b3d65c92bed7a9884ef7fe9e3f6c8fed16f6fb2f811f5750f547567a5dd58fed%40group.calendar.google.com&scp=ALL) for community meetings and events
    - The [kubestellar-dev](https://groups.google.com/g/kubestellar-dev) mailing list is subscribed to this calendar
- See recordings of past KubeStellar community meetings on [YouTube](https://www.youtube.com/@kubestellar)
- See [upcoming](https://github.com/kcp-dev/edge-mc/issues?q=is%3Aissue+is%3Aopen+label%3Acommunity-meeting) and [past](https://github.com/kcp-dev/edge-mc/issues?q=is%3Aissue+is%3Aclosed+label%3Acommunity-meeting) community meeting agendas and notes
- Browse the [shared Google Drive](https://drive.google.com/drive/folders/1p68MwkX0sYdTvtup0DcnAEsnXElobFLS?usp=sharing) to share design docs, notes, etc.
    - Members of the [kubestellar-dev](https://groups.google.com/g/kubestellar-dev) mailing list can view this drive
- Read our [documentation](https://kubestellar.io)
- Follow us on:
   - LinkedIn - [#kubestellar](https://www.linkedin.com/feed/hashtag/?keywords=kubestellar)
   - Medium - [kubestellar.medium.com](https://medium.com/@kubestellar/list/predefined:e785a0675051:READING_LIST)
   
   
## ❤️ Contributors

Thanks go to these wonderful people:

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tr>
    <td align="center"><a href="https://github.com/waltforme"><img src="https://avatars.githubusercontent.com/u/8633434?v=4" width="100px;" alt=""/><br /><sub><b>Jun Duan</b></sub></a><br /><a href="https://github.com/kcp-dev/edge-mc/issues?q=assignee%3Awaltforme+" title="Contributed PRs">👀</a></td>
    <td align="center"><a href="https://github.com/dumb0002"><img src="https://avatars.githubusercontent.com/u/25727844?v=4" width="100px;" alt=""/><br /><sub><b>Braulio Dumba</b></sub></a><br /><a href="https://github.com/kcp-dev/edge-mc/issues?q=assignee%3Adumb0002+" title="Contributed PRs">👀</a></td>
    <td align="center"><a href="https://github.com/MikeSpreitzer"><img src="https://avatars.githubusercontent.com/u/14296719?v=4" width="100px;" alt=""/><br /><sub><b>Mike Spreitzer</b></sub></a><br /><a href="https://github.com/kcp-dev/edge-mc/pulls?q=is%3Apr+reviewed-by%3AMikeSpreitzer" title="Reviewed Pull Requests">👀</a></td>
    <td align="center"><a href="https://github.com/pdettori"><img src="https://avatars.githubusercontent.com/u/6678093?v=4" width="100px;" alt=""/><br /><sub><b>Paolo Dettori</b></sub></a><br /><a href=https://github.com/kcp-dev/edge-mc/issues?q=assignee%3Apdettori+" title="Contributed PRs">👀</a></td>
    <td align="center"><a href="https://github.com/clubanderson"><img src="https://avatars.githubusercontent.com/u/407614?v=4" width="100px;" alt=""/><br /><sub><b>Andy Anderson</b></sub></a><br /><a href="https://github.com/kcp-dev/edge-mc/pulls?q=is%3Apr+reviewed-by%3Aclubanderson" title="Reviewed Pull Requests">👀</a></td>
    <td align="center"><a href="https://github.com/francostellari"><img src="https://avatars.githubusercontent.com/u/50019234?v=4" width="100px;" alt=""/><br /><sub><b>Franco Stellari</b></sub></a><br /><a href="https://github.com/kcp-dev/edge-mc/issues?q=assignee%3Afrancostellari+" title="Contributed PRs">👀</a></td>
    <td align="center"><a href="https://github.com/ezrasilvera"><img src="https://avatars.githubusercontent.com/u/13567561?v=4" width="100px;" alt=""/><br /><sub><b>Ezra Silvera</b></sub></a><br /><a href="https://github.com/kcp-dev/edge-mc/pulls?q=is%3Apr+reviewed-by%3Aezrasilvera" title="Reviewed Pull Requests">👀</a></td>
  </tr>
  <tr>
    <td align="center"><a href="https://github.com/fileppb"><img src="https://avatars.githubusercontent.com/u/124100147?v=4" width="100px;" alt=""/><br /><sub><b>Bob Filepp</b></sub></a><br /><a href="https://github.com/kcp-dev/edge-mc/issues?q=assignee%3Afileppb+" title="Contributed PRs">👀</a></td>
    <td align="center"><a href="https://github.com/thinkahead"><img src="https://avatars.githubusercontent.com/u/7507482?v=4" width="100px;" alt=""/><br /><sub><b>Alexei Karve</b></sub></a><br /><a href="https://github.com/kcp-dev/edge-mc/issues?q=assignee%3Athinkahead+" title="Contributed PRs">👀</a></td>
    <td align="center"><a href="https://github.com/mra-ruiz"><img src="https://avatars.githubusercontent.com/u/16118462?v=4" width="100px;" alt=""/><br /><sub><b>Maria Camila Ruiz Cardenas</b></sub></a><br /><a href="https://github.com/kcp-dev/edge-mc/issues?q=assignee%3Amra-ruiz+" title="Contributed PRs">👀</a></td>
    <td align="center"><a href="https://github.com/aslom"><img src="https://avatars.githubusercontent.com/u/1648338?v=4" width="100px;" alt=""/><br /><sub><b>Aleksander Slominski</b></sub></a><br /><a href="https://github.com/kcp-dev/edge-mc/issues?q=assignee%3Aaslom+" title="Contributed PRs">👀</a></td>
        
  </tr>

</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->
<!--readme-for-root-end-->
