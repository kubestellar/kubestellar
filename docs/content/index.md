<!-- <img alt="" width="500px" align="left" src="../KubeStellar-with-Logo.png" />

<br/>
<br/>
<br/>
<br/>

# Multi-cluster Configuration Management for Edge, Multi-Cloud, and Hybrid Cloud
 -->
<p align="center">
<b>Did you know kubernetes manages multiple nodes but does not manage multiple clusters out of the box?  KubeStellar adds multi-cluster support to Kubernetes distributions.</b>
<br/>
<br/>
<b>Distinguishing features of KubeStellar</b>
</p>
- <b>[ multi-cluster down-syncing]({{ config.docs_url }}/{{ config.ks_branch }}/Getting-Started/quickstart/)</b> deploy, configure, and collect status <b>across pre-existing clusters</b>
- <b>[up-syncing]({{ config.docs_url }}/{{ config.ks_branch }}/Coding%20Milestones/PoC2023q1/kubestellar-syncer/)</b> from remote clusters (return <i>any</i> object, not just status)
- <b>lightweight logical cluster</b> support ([KubeFlex](https://github.com/kubestellar/kubeflex), [kcp](https://kcp.io), [kind](https://kind.io), etc.)
- <b>[resiliency]({{ config.docs_url }}/{{ config.ks_branch }}/Coding%20Milestones/PoC2023q1/mailbox-controller/)</b> to support disconnected operation and intermittent connectivity

<br/>
<p align="center">
<b>Additional features</b>
</p>
- non-wrapped / kubernetes-object-native denaturing (enables hierarchy) (no requirement to wrap objects)
- rule-based customization (grouping) - automate the customization of your deployments
- status summarization - summarize the status returned from all your deployments
- scalability - scale to a large number of objects, overcoming default Kubernetes limitations

[Learn more about KubeStellar](./readme.md)

[Try our QuickStart]({{ config.docs_url }}/{{ config.ks_branch }}/Getting-Started/quickstart/)

<br/>
<p align="center">
<img alt="" width="500px" align="center" src="./KubeStellar-with-Logo.png" />
</p>
<br/>
