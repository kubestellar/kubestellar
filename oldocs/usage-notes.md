# User Guide fragments

Status of this document: this is material that needs to go elsewhere.

This document is for users of a release. Examples of using the latest stable release are in [the examples document](examples.md). This document adds information not conveyed in the examples.

## Using a pre-existing cluster as the hosting cluster

See [Using an existing hosting cluster](./hosting-cluster.md)

## When everything is not on the same machine

Thus far we can only say how to handle this when the hosting cluster is OpenShift. The problem is getting URLs that work from everywhere. OpenShift is a hosted product, your clusters have domain names that are resolvable from everywhere. In other words, if you use an OpenShift cluster as your hosting cluster then this problem is already solved.
