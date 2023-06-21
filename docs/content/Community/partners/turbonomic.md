# Check out KubeStellar working with Turbonomic:
Medium - [Make Multi-Cluster Scheduling a No-Brainer](https://medium.com/@waltforme/make-multi-cluster-scheduling-a-no-brainer-e1979ba5b9b2)<br/>

### Turbonomic and KubeStellar Demo Day
<p align=center>
<div id="spinner1">
  <img width="140" height="140" src="../../../images/spinner.gif" class="centerImage">
</div>
<iframe class="centerImage" id="embed1" width="0" height="0" src="https://www.youtube.com/embed/B3jZTnu1LDo?controls=0" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen style="visibility:hidden;" onload= "document.getElementById('spinner1').style.display='none';document.getElementById('embed1').style.visibility='visible';document.getElementById('embed1').width='720';document.getElementById('embed1').height='400';"></iframe>
</p>

### How do I get this working with my KubeStellar instance?
As we can see from the blog and the demo, Turbonomic talks to KubeStellar via GitOps. The scheduling decisions are passed from Turbonomic to KubeStellar in two steps:
1. Turbo -> GitHub repository.
2. GitHub repository -> KubeStellar.

For the 1st step (Turbonomic -> GitHub repository), a controller named "[change reconciler](https://github.com/irfanurrehman/change-reconciler)" creates PRs against the GitHub repository, where the PRs contains changes to scheduling decisions.

There's also [a piece of code](https://github.com/edge-experiments/turbonomic-integrations) which intercepts Turbonomic actions and creates CRs for the above change reconciler.

For the 2nd step (GitHub repository-> KubeStellar), we can use Argo CD. The detailed procedure to integrate Argo CD with KubeStellar is documented [here](./argocd.md).

As we can see from the blog and the demo, Turbonomic collects data from edge clusters. This is made possible by installing [kubeturbo](https://github.com/turbonomic/kubeturbo) into each of the edge clusters.


### Turbonomic and KubeStellar in the news
<p align=center>
<div id="spinner2">
    <img width="140" height="140" src="../../../images/spinner.gif" class="centerImage">
</div>
<iframe class="centerImage" id="embed2" src="https://www.linkedin.com/embed/feed/update/urn:li:share:7066466334334668800" scrolling=no height="0" width="0" frameborder="0" allowfullscreen="" title="Embedded post" style="visibility:hidden;" onload= "document.getElementById('spinner2').style.display='none';document.getElementById('embed2').style.visibility='visible';document.getElementById('embed2').width='740';document.getElementById('embed2').height='400';"></iframe>
</p>

<style type="text/css">
.centerImage
{
 display: block;
 margin: auto;
}
</style>