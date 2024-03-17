<!--readme-for-root-start-->

<img alt="" width="500px" align="left" src="../KubeStellar-with-Logo.png" />

<br/>
<br/>
<br/>
<br/>

## Multi-cluster Configuration Management for Edge, Multi-Cloud, and Hybrid Cloud

[![](https://github.com/kubestellar/kubestellar/actions/workflows/docs-gen-and-push.yml/badge.svg?branch={{ config.ks_branch }})](https://github.com/kubestellar/kubestellar/actions/workflows/docs-gen-and-push.yml)&nbsp;&nbsp;&nbsp;
[![](https://img.shields.io/badge/first--timers--only-friendly-blue.svg?style=flat-square)](https://www.firsttimersonly.com/)&nbsp;&nbsp;&nbsp;
[![](https://github.com/kubestellar/kubestellar/actions/workflows/broken-links-crawler.yml/badge.svg)](https://github.com/kubestellar/kubestellar/actions/workflows/broken-links-crawler.yml)
<a href="https://kubernetes.slack.com/archives/C058SUSL5AA"> 
    <img alt="Join Slack" src="https://img.shields.io/badge/KubeStellar-Join%20Slack-blue?logo=slack">
  </a>

Imagine KubeStellar as a post office for your Kubernetes resources. When you drop packages at the post office, they don't open them; they deliver them to the right recipients. Similarly, KubeStellar works like this for your Kubernetes resources. Instead of running resources right away, KubeStellar safely stores and sends resources to selected clusters across the globe—whether they're in public clouds, private clouds, or on the edge of your network. It's a super useful tool for spreading your Kubernetes resources wherever you need them without disrupting your existing tools and workflows.

How does KubeStellar resist the temptation to run your Kubernetes resources right away? KubeStellar accepts your applied resources in a special staging area (virtual cluster) where pods can't be created. Then, at your direction, KubeStellar transfers your applied resources to remote clusters where they can create pods and other required resource dependencies. KubeStellar does this using many different lightweight virtual cluster providers (Kind, KubeFlex, KCP, etc.) to create this special staging area. 

KubeStellar is an innovative way to stage inactive Kubernetes resources and then apply them to any cluster to run. KubeStellar introduces a native way to expand, optimize, and protect your Kubernetes resources from individual cluster misconfiguration, utilization, and failure. 

__Don't change anything, just add KubeStellar!__


## KubeStellar treats multiple Kubernetes clusters as one so you can:

- __Centrally__ apply Kubernetes resources for selective deployment across multiple clusters 
- Use __standard Kubernetes native deployment tools__ (kubectl, Helm, Kustomize, ArgoCD, Flux); no resource bundling required
- __Discover__ dynamically created objects created on remote clusters
- Make __disconnected__ cluster operation possible
- __Scale__ with 1:many and many:1 scenarios
- __Remain compatible__ with cloud-native solutions

## Use The Source, Luke

**NOTE**: This website is still under construction, and not all of the documentation appears on this website. The rest, which is also still under construction, is designed to be viewed directly from the GitHub repository; see the README of [the KubeStellar GitHub repository](http://github.com/kubestellar/kubestellar). To read the documentation for release `X.Y.Z`, view the Git content tagged `vX.Y.Z`.

## Contributing

We ❤️ our contributors! If you're interested in helping us out, please head over to our [Contributing]({{ config.docs_url }}/{{ config.ks_branch }}/Contribution%20guidelines/CONTRIBUTING/) guide.

## Getting in touch

There are several ways to communicate with us:

Instantly get access to our documents and meeting invites [http://kubestellar.io/joinus](http://kubestellar.io/joinus)

- The [`#kubestellar-dev` channel](https://kubernetes.slack.com/archives/C058SUSL5AA) in the [Kubernetes Slack workspace](https://slack.k8s.io)
- Our mailing lists:
    - [kubestellar-dev](https://groups.google.com/g/kubestellar-dev) for development discussions
    - [kubestellar-users](https://groups.google.com/g/kubestellar-users) for discussions among users and potential users
- Subscribe to the [community calendar](https://calendar.google.com/calendar/event?action=TEMPLATE&tmeid=MWM4a2loZDZrOWwzZWQzZ29xanZwa3NuMWdfMjAyMzA1MThUMTQwMDAwWiBiM2Q2NWM5MmJlZDdhOTg4NGVmN2ZlOWUzZjZjOGZlZDE2ZjZmYjJmODExZjU3NTBmNTQ3NTY3YTVkZDU4ZmVkQGc&tmsrc=b3d65c92bed7a9884ef7fe9e3f6c8fed16f6fb2f811f5750f547567a5dd58fed%40group.calendar.google.com&scp=ALL) for community meetings and events
    - The [kubestellar-dev](https://groups.google.com/g/kubestellar-dev) mailing list is subscribed to this calendar
- See recordings of past KubeStellar community meetings on [YouTube](https://www.youtube.com/@kubestellar)
- See [upcoming](https://github.com/kubestellar/kubestellar/issues?q=is%3Aissue+is%3Aopen+label%3Acommunity-meeting) and [past](https://github.com/kubestellar/kubestellar/issues?q=is%3Aissue+is%3Aclosed+label%3Acommunity-meeting) community meeting agendas and notes
- Browse the [shared Google Drive](https://drive.google.com/drive/folders/1p68MwkX0sYdTvtup0DcnAEsnXElobFLS?usp=sharing) to share design docs, notes, etc.
    - Members of the [kubestellar-dev](https://groups.google.com/g/kubestellar-dev) mailing list can view this drive
- Read our [documentation]({{ config.docs_url }})
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
    <td align="center"><a href="https://github.com/waltforme"><img src="https://avatars.githubusercontent.com/u/8633434?v=4" width="100px;" alt=""/><br /><sub><b>Jun Duan</b></sub></a><br /><a href="https://github.com/kubestellar/kubestellar/issues?q=assignee%3Awaltforme+" title="Contributed PRs">👀</a></td>
    <td align="center"><a href="https://github.com/dumb0002"><img src="https://avatars.githubusercontent.com/u/25727844?v=4" width="100px;" alt=""/><br /><sub><b>Braulio Dumba</b></sub></a><br /><a href="https://github.com/kubestellar/kubestellar/issues?q=assignee%3Adumb0002+" title="Contributed PRs">👀</a></td>
    <td align="center"><a href="https://github.com/MikeSpreitzer"><img src="https://avatars.githubusercontent.com/u/14296719?v=4" width="100px;" alt=""/><br /><sub><b>Mike Spreitzer</b></sub></a><br /><a href="https://github.com/kubestellar/kubestellar/pulls?q=is%3Apr+reviewed-by%3AMikeSpreitzer" title="Reviewed Pull Requests">👀</a></td>
    <td align="center"><a href="https://github.com/pdettori"><img src="https://avatars.githubusercontent.com/u/6678093?v=4" width="100px;" alt=""/><br /><sub><b>Paolo Dettori</b></sub></a><br /><a href=https://github.com/kubestellar/kubestellar/issues?q=assignee%3Apdettori+" title="Contributed PRs">👀</a></td>
    <td align="center"><a href="https://github.com/clubanderson"><img src="https://avatars.githubusercontent.com/u/407614?v=4" width="100px;" alt=""/><br /><sub><b>Andy Anderson</b></sub></a><br /><a href="https://github.com/kubestellar/kubestellar/pulls?q=is%3Apr+reviewed-by%3Aclubanderson" title="Reviewed Pull Requests">👀</a></td>
    <td align="center"><a href="https://github.com/francostellari"><img src="https://avatars.githubusercontent.com/u/50019234?v=4" width="100px;" alt=""/><br /><sub><b>Franco Stellari</b></sub></a><br /><a href="https://github.com/kubestellar/kubestellar/issues?q=assignee%3Afrancostellari+" title="Contributed PRs">👀</a></td>
    <td align="center"><a href="https://github.com/ezrasilvera"><img src="https://avatars.githubusercontent.com/u/13567561?v=4" width="100px;" alt=""/><br /><sub><b>Ezra Silvera</b></sub></a><br /><a href="https://github.com/kubestellar/kubestellar/pulls?q=is%3Apr+reviewed-by%3Aezrasilvera" title="Reviewed Pull Requests">👀</a></td>
  </tr>
  <tr>
    <td align="center"><a href="https://github.com/fileppb"><img src="https://avatars.githubusercontent.com/u/124100147?v=4" width="100px;" alt=""/><br /><sub><b>Bob Filepp</b></sub></a><br /><a href="https://github.com/kubestellar/kubestellar/issues?q=assignee%3Afileppb+" title="Contributed PRs">👀</a></td>
    <td align="center"><a href="https://github.com/thinkahead"><img src="https://avatars.githubusercontent.com/u/7507482?v=4" width="100px;" alt=""/><br /><sub><b>Alexei Karve</b></sub></a><br /><a href="https://github.com/kubestellar/kubestellar/issues?q=assignee%3Athinkahead+" title="Contributed PRs">👀</a></td>
    <td align="center"><a href="https://github.com/mra-ruiz"><img src="https://avatars.githubusercontent.com/u/16118462?v=4" width="100px;" alt=""/><br /><sub><b>Maria Camila Ruiz Cardenas</b></sub></a><br /><a href="https://github.com/kubestellar/kubestellar/issues?q=assignee%3Amra-ruiz+" title="Contributed PRs">👀</a></td>
    <td align="center"><a href="https://github.com/aslom"><img src="https://avatars.githubusercontent.com/u/1648338?v=4" width="100px;" alt=""/><br /><sub><b>Aleksander Slominski</b></sub></a><br /><a href="https://github.com/kubestellar/kubestellar/issues?q=assignee%3Aaslom+" title="Contributed PRs">👀</a></td>
    <td align="center"><a href="https://github.com/amanroa"><img src="https://avatars.githubusercontent.com/u/26678552?v=4" width="100px;" alt=""/><br /><sub><b>Aashni Manroa</b></sub></a><br /><a href="https://github.com/kubestellar/kubestellar/issues?q=assignee%3Aamanroa+" title="Contributed PRs">👀</a></td>
    <td align="center"><a href="https://github.com/KPRoche"><img src="https://avatars.githubusercontent.com/u/25445603?v=4" width="100px;" alt=""/><br /><sub><b>Kevin Roche</b></sub></a><br /><a href="https://github.com/kubestellar/kubestellar/issues?q=assignee%3AKPRoche+" title="Contributed PRs">👀</a></td>
    <td align="center"><a href="https://github.com/namasl"><img src="https://avatars.githubusercontent.com/u/144150872?v=4" width="100px;" alt=""/><br /><sub><b>Nick Masluk</b></sub></a><br /><a href="https://github.com/kubestellar/kubestellar/issues?q=assignee%3Anamasl+" title="Contributed PRs">👀</a></td>
  </tr>
  <tr>
    <td align="center"><a href="https://github.com/fab7"><img src="https://avatars.githubusercontent.com/u/15231306?v=4" width="100px;" alt=""/><br /><sub><b>Francois Abel</b></sub></a><br /><a href="https://github.com/kubestellar/kubestellar/issues?q=assignee%3Afab7+" title="Contributed PRs">👀</a></td>
  </tr>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->
<!--readme-for-root-end-->
