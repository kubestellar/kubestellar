# Best Practices

**Note:**  This section is under construction and includes partial information. 
## Size considerations
As KubeStellar is built on top of Kubernetes all Kubernetes limitations and recommendations apply to KubeStellar as well. These recommendations can be found in [Kubernetes Considerations for large clusters](https://kubernetes.io/docs/setup/best-practices/cluster-large/).

In addition, KubeStellar Transport Plugin is built on top of OCM, so KubeStellar also comply to some of OCM's limitations. Users should take into account the following restrictions:

  * The ManifestWork shouldn't exceed 500KB
