<!-- Note that this repo has two readme files, with content that is as nearly identical as is practical: `/README.md` and `/docs/content/readme.md` -->

<img alt="" width="500px" align="left" src="KubeStellar-with-Logo.png" />

<br/>
<br/>
<br/>
<br/>

## Multi-cluster Configuration Management for Edge, Multi-Cloud, and Hybrid Cloud

[![](https://img.shields.io/badge/first--timers--only-friendly-blue.svg?style=flat-square)](https://www.firsttimersonly.com/)&nbsp;&nbsp;&nbsp;
[![](https://github.com/kubestellar/kubestellar/actions/workflows/broken-links-crawler.yml/badge.svg)](https://github.com/kubestellar/kubestellar/actions/workflows/broken-links-crawler.yml)
[![](https://www.bestpractices.dev/projects/8266/badge)](https://www.bestpractices.dev/projects/8266)
[![](https://api.scorecard.dev/projects/github.com/kubestellar/kubestellar/badge)](https://scorecard.dev/viewer/?uri=github.com/kubestellar/kubestellar)
[![](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/kubestellar)](https://artifacthub.io/packages/search?repo=kubestellar)
<a href="https://cloud-native.slack.com/archives/C097094RZ3M"> 
    <img alt="Join Slack" src="https://img.shields.io/badge/KubeStellar-Join%20Slack-blue?logo=slack">
  </a>
<a href="https://deepwiki.com/kubestellar/kubestellar">
  <img src="https://deepwiki.com/badge.svg" alt="Ask DeepWiki">
</a>

**KubeStellar** is a Cloud Native Computing Foundation (CNCF) Sandbox project that simplifies the deployment and configuration of applications across multiple Kubernetes clusters. It provides a seamless experience akin to using a single cluster, and it integrates with the tools you're already familiar with, eliminating the need to modify existing resources.

KubeStellar is particularly beneficial if you're currently deploying in a single cluster and are looking to expand to multiple clusters, or if you're already using multiple clusters and are seeking a more streamlined developer experience.


![KubeStellar High Level View](docs/content/images/kubestellar-high-level.png)


The use of multiple clusters offers several advantages, including:

- Separation of environments (e.g., development, testing, staging)
- Isolation of groups, teams, or departments
- Compliance with enterprise security or data governance requirements
- Enhanced resiliency, including across different clouds
- Improved resource availability
- Access to heterogeneous resources
- Capability to run applications on the edge, including in disconnected environments

In a single-cluster setup, developers typically access the cluster and deploy Kubernetes objects directly. Without KubeStellar, multiple clusters are usually deployed and configured individually, which can be time-consuming and complex.

KubeStellar simplifies this process by allowing developers to define a binding policy between clusters and Kubernetes objects. It then uses your regular single-cluster tooling to deploy and configure each cluster based on these binding policies, making multi-cluster operations as straightforward as managing a single cluster. This approach enhances productivity and efficiency, making KubeStellar a valuable tool in a multi-cluster Kubernetes environment.

## Website

For usage, architecture, and other documentation, see [the website](https://kubestellar.io).

## Contributing

We ❤️ our contributors! If you're interested in helping us out, please head over to our [Contributing](https://github.com/kubestellar/kubestellar/blob/main/CONTRIBUTING.md) guide and be sure to look at `main` or the release of interest to you.

This community has a [Code of Conduct](./CODE_OF_CONDUCT.md). Please make sure to follow it.

## Our Roadmap
Have a look at what we are working on next, see our [Roadmap](docs/content/direct/roadmap.md) 

## Getting in touch

There are several ways to communicate with us:

Instantly get access to our documents and meeting invites at http://kubestellar.io/joinus

- The [`#kubestellar-dev` channel](https://cloud-native.slack.com/archives/C097094RZ3M) in the [CNCF Slack workspace](https://communityinviter.com/apps/cloud-native/cncf)
- Our mailing lists:
    - [kubestellar-dev](https://groups.google.com/g/kubestellar-dev) for development discussions
    - [kubestellar-users](https://groups.google.com/g/kubestellar-users) for discussions among users and potential users
- Subscribe to the [community meeting calendar](https://calendar.google.com/calendar/event?action=TEMPLATE&tmeid=MWM4a2loZDZrOWwzZWQzZ29xanZwa3NuMWdfMjAyMzA1MThUMTQwMDAwWiBiM2Q2NWM5MmJlZDdhOTg4NGVmN2ZlOWUzZjZjOGZlZDE2ZjZmYjJmODExZjU3NTBmNTQ3NTY3YTVkZDU4ZmVkQGc&tmsrc=b3d65c92bed7a9884ef7fe9e3f6c8fed16f6fb2f811f5750f547567a5dd58fed%40group.calendar.google.com&scp=ALL) for community meetings and events
    - The [kubestellar-dev](https://groups.google.com/g/kubestellar-dev) mailing list is subscribed to this calendar
- See recordings of past KubeStellar community meetings on [YouTube](https://www.youtube.com/@kubestellar)
- See [upcoming](https://github.com/kubestellar/kubestellar/issues?q=is%3Aissue+is%3Aopen+label%3Acommunity-meeting) and [past](https://github.com/kubestellar/kubestellar/issues?q=is%3Aissue+is%3Aclosed+label%3Acommunity-meeting) community meeting agendas and notes
- Browse the [shared Google Drive](https://drive.google.com/drive/folders/1p68MwkX0sYdTvtup0DcnAEsnXElobFLS?usp=sharing) to share design docs, notes, etc.
    - Members of the [kubestellar-dev](https://groups.google.com/g/kubestellar-dev) mailing list can view this drive
- Follow us on:
   - LinkedIn - [#kubestellar](https://www.linkedin.com/feed/hashtag/?keywords=kubestellar)
   - Medium - [kubestellar.medium.com](https://medium.com/@kubestellar/list/predefined:e785a0675051:READING_LIST)


<div>
<h2><font size="6"><img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Smilies/Red%20Heart.png" alt="Red Heart" width="40" height="40" /> Contributors </font></h2>
</div>
<br>

<center>
<a href="https://github.com/kubestellar/kubestellar/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=kubestellar/kubestellar" />
</a>
</center>
<br>
<br>

[![CLOMonitor report summary](https://clomonitor.io/api/projects/cncf/kubestellar/report-summary?theme=light)](https://clomonitor.io/projects/cncf/kubestellar)
<br>
<br>

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fkubestellar%2Fkubestellar.svg?type=large&issueType=license)](https://app.fossa.com/projects/git%2Bgithub.com%2Fkubestellar%2Fkubestellar?ref=badge_large&issueType=license)
<br>
<br>

<td>
    <a href="https://landscape.cncf.io">
        <img src="/docs/overrides/images/cncf-color.png" width="300px;" alt="Cloud Native Computing Foundation Logo"/>
    </a>
</td>
<br>We are a Cloud Native Computing Foundation sandbox project.
<br>Kubernetes and the Kubernetes logo are registered trademarks of The Linux Foundation® (TLF).
<br>The Linux Foundation has registered trademarks and uses trademarks. For a list of trademarks of The Linux Foundation, please see our <a href="https://www.linuxfoundation.org/legal/trademark-usage">Trademark Usage page</a>.
<br>© 2022-2025. The KubeStellar Authors.

