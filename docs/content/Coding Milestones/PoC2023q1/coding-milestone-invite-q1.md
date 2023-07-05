---
title: "Invitation"
linkTitle: "Invitation"
---

Dear Contributors,

We are excited to invite you to join the first [KubeStellar opensource community]({{ config.repo_url }}) coding sprint. We will be focus on several key projects that are critical to the development of state-based edge solutions. Our collective work will be showcased to the opensource community on Thursday, April 27th.

This coding sprint will provide a great opportunity for you to showcase your skills, learn new techniques, and collaborate with other experienced engineers in the KubeStellar community. We believe that your contributions will be invaluable in helping us achieve our goals and making a lasting impact in the field of state-based edge technology.

The coding sprint will be dedicated to completing the following workload management elements:

- Implementing a Where Resolver and a Placement Translator, including customization options,
- Incorporating existing customization API into the [KubeStellar repo]({{ config.repo_url }}),
- Investigating implementation of a status summarizer, starting with basic implicit status, and later adding programmed summarization,
- Updating summarization API and integrating it into the [KubeStellar repo]({{ config.repo_url }}),
- Defining the API for identifying associated objects and its interaction with summarization, and implementing these,
- Streamlining the creation of workload management workspaces,
- Examining the use of Postgresql through Kine instead of etcd for scalability,
- Revising the milestone outline with regards to defining bootstrapping and support for cluster-scoped resources.

In addition to workload management, we will also be working on inventory management for the demo, as well as designing various demo scenarios, including a baseline demo with kubectl, demos with ArgoCD, FluxCD, and the European Space Agency (ESA). To support the engineers and demonstrations we will also need to automate the process of creating infrastructure, deploying demo pieces and instrumentation, bootstrapping, running scenarios, and collecting data.

If you are interested in joining us for this exciting coding sprint, please check out our ['good first issue' list]({{ config.repo_url }}/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22), or slack me [@Andy Anderson](https://kubernetes.slack.com/team/U0462LN24QJ) so I can connect you with others in your area of interest.  There is a place for every skillset to contribute. Not quite sure?  You can join our [bi-weekly community meetings](https://calendar.google.com/calendar/embed?src=b3d65c92bed7a9884ef7fe9e3f6c8fed16f6fb2f811f5750f547567a5dd58fed%40group.calendar.google.com&ctz=America%2FNew_York) to watch our progress.
