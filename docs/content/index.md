<!-- <img alt="" width="500px" align="left" src="../KubeStellar-with-Logo.png" />

<br/>
<br/>
<br/>
<br/>

# Multicluster Configuration Management for Edge, Multi-Cloud, and Hybrid Cloud
 -->
<p align="center">
<b>Distinguishing features of KubeStellar</b>
</p>
- <b>[ multicluster down-syncing](https://kubestellar.io/quickstart)</b> deploy, configure, and collect status <b>across pre-existing clusters</b>
- <b>[up-syncing](https://docs.kubestellar.io/{{ config.ks_branch }}/Coding%20Milestones/PoC2023q1/kubestellar-syncer/)</b> from remote clusters (return <i>any</i> object, not just status)
- <b>lightweight logical cluster</b> support ([KubeFlex](https://github.com/kubestellar/kubeflex), [kcp](https://kcp.io), [kind](https://kind.io), etc.)
- <b>[resiliency](https://docs.kubestellar.io/{{ config.ks_branch }}/Coding%20Milestones/PoC2023q1/mailbox-controller/)</b> to support disconnected operation and intermittent connectivity

<br/>
<p align="center">
<b>Additional features</b>
</p>
- non-wrapped / kubernetes-object-native denaturing (enables hierarchy) (no requirement to wrap objects)
- rule-based customization (grouping) - automate the customization of your deployments
- status summarization - summarize the status returned from all your deployments
- scalability - scale to a large number of objects, overcoming default Kubernetes limitations

[Learn more about KubeStellar](./readme.md)
[Try our QuickStart](https://kubestellar.io/quickstart)

<br/>
<p align="center">
<img alt="" width="500px" align="center" src="./KubeStellar-with-Logo.png" />
</p>
<br/>