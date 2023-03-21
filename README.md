<img alt="KCP-Edge" width="500px" align="left" src="./contrib/logo/kcp-edge-5-white.png" />

<br/>

# KCP-Edge
## consistent, heterogeneous, and scalable edge configuration management
<br/><br/><br/>
[![github pages](https://github.com/kcp-dev/edge-mc/actions/workflows/gh-pages.yml/badge.svg)](https://github.com/kcp-dev/edge-mc/actions/workflows/gh-pages.yml)&nbsp;&nbsp;&nbsp;
[![PR Verifier](https://github.com/kcp-dev/edge-mc/actions/workflows/pr-verifier.yaml/badge.svg)](https://github.com/kcp-dev/edge-mc/actions/workflows/pr-verifier.yaml)&nbsp;&nbsp;&nbsp;
[![Open Source Helpers](https://www.codetriage.com/kcp-dev/edge-mc/badges/users.svg)](https://www.codetriage.com/kcp-dev/edge-mc)&nbsp;&nbsp;&nbsp;
[![first-timers-only](https://img.shields.io/badge/first--timers--only-friendly-blue.svg?style=flat-square)](https://www.firsttimersonly.com/)&nbsp;&nbsp;&nbsp;
<a href="https://app.slack.com/client/T09NY5SBT/C021U8WSAFK"> 
    <img alt="Join Slack" src="https://img.shields.io/badge/KCP--Edge-Join%20Slack-blue?logo=slack">
  </a>
## Overview
KCP-Edge is a subproject of kcp focusing on concerns arising from edge multicluster use cases:

- Hierarchy, infrastructure & platform, roles & responsibilities, integration architecture, security issues
- Runtime in[ter]dependence: An edge location may need to operate independently of the center and other edge locations​
- Non-namespaced objects: need general support
- Cardinality of destinations: A source object may propagate to many thousands of destinations. ​ 

## Goals

- collaboratively design a component set similar to those found in the current kcp TMC implementation (dedicated Workspace type, scheduler, syncer-like mechanism, edge placement object definition, status collection strategy, etc.)
- Specify a multi-phased proof-of-concept inclusive of component architecture, interfaces, and example workloads
- Validate phases of proof-of-concept with kcp, Kube SIG-Multicluster, and CNCF community members interested in Edge

## Areas of exploration

- Desired placement expression​: Need a way for one center object to express large number of desired copies​
- Scheduling/syncing interface​: Need something that scales to large number of destinations​
- Rollout control​: Client needs programmatic control of rollout, possibly including domain-specific logic​
- Customization: Need a way for one pattern in the center to express how to customize for all the desired destinations​
- Status from many destinations​: Center clients may need a way to access status from individual edge copies
- Status summarization​: Client needs a way to specify how statuses from edge copies are processed/reduced along the way from edge to center​.

## Quickstart

Checkout our [Quickstart Guide](https://docs.kcp-edge.io/docs/getting-started/quickstart/)

## Contributing

We ❤️ our contributors! If you're interested in helping us out, please head over to our [Contributing](CONTRIBUTING.md) guide.

## Getting in touch

There are several ways to communicate with us:

- The [`#kcp-dev` channel](https://app.slack.com/client/T09NY5SBT/C021U8WSAFK) in the [Kubernetes Slack workspace](https://slack.k8s.io)
- Our mailing lists:
    - [kcp-dev](https://groups.google.com/g/kcp-dev) for development discussions
    - [kcp-users](https://groups.google.com/g/kcp-users) for discussions among users and potential users
- Subscribe to the [community calendar](https://calendar.google.com/calendar/embed?src=ujjomvk4fa9fgdaem32afgl7g0%40group.calendar.google.com) for community meetings and events
    - The kcp-dev mailing list is subscribed to this calendar
- See recordings of past KCP-Edge community meetings on [YouTube](https://www.youtube.com/playlist?list=PL1ALKGr_qZKc9jyv1EfOFNfoAJo9Q6Ebd)
- See [upcoming](https://github.com/kcp-dev/edge-mc/issues?q=is%3Aissue+is%3Aopen+label%3Acommunity-meeting) and [past](https://github.com/kcp-dev/edge-mc/issues?q=is%3Aissue+is%3Aclosed+label%3Acommunity-meeting) community meeting agendas and notes
- Browse the [shared Google Drive](https://drive.google.com/drive/folders/1FN7AZ_Q1CQor6eK0gpuKwdGFNwYI517M?usp=sharing) to share design docs, notes, etc.
    - Members of the kcp-dev mailing list can view this drive
- Read our [documentation](https://kcp-edge.io)
- Follow us on:
   - LinkedIn - [#kcpedge](https://www.linkedin.com/feed/hashtag/?keywords=kcpedge)
   - Medium - [kcp-edge.medium.com](https://medium.com/@kcp-edge/list/predefined:e785a0675051:READING_LIST)
   
   
## ❤️ Contributors

Thanks go to these wonderful people:

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tr>
    <
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
  </tr>

</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->
