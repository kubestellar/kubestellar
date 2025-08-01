# Observability in KubeStellar

KubeStellar provides endpoints and integrations for observability and monitoring. This page describes the available observability features, how to access them, and how to use them in a typical deployment.

## Metrics Endpoints

KubeStellar controllers expose Prometheus-compatible metrics endpoints. These endpoints respond to HTTP requests for metrics and can be queried by any monitoring system; KubeStellar does not mandate how metrics are collected or scraped. For an example of collecting both metrics and debug endpoint data using Prometheus, see the [monitoring](https://github.com/kubestellar/kubestellar/tree/main/monitoring/) directory. This is just one possible approach; KubeStellar does not require or mandate any specific monitoring tool or method.

### Metrics Endpoint Table

| Controller                     | Protocol | Port  | Path     | AuthN/AuthZ                                               | Notes                                                         |
|-------------------------------|----------|-------|----------|-----------------------------------------------------------|---------------------------------------------------------------|
| kubestellar-controller-manager | HTTPS    | 8443  | /metrics | Kubernetes client authentication required (Service or Pod) | Service: `kubestellar-controller-manager-metrics-service`.    |
| kubestellar-controller-manager | HTTP     | 8080  | /metrics | None (debug; not recommended for production)               | Debug endpoint; typically disabled in production.              |
| ks-transport-controller        | HTTP     | 8090  | /metrics | None (in-cluster)                                          | Port configurable via Helm values.                            |
| status-addon-controller        | HTTP     | 9280  | /metrics | None (in-cluster)                                          |                                                               |
| status-addon-agent             | HTTP     | 8080  | /metrics | None (in-cluster)                                          | Port configurable via Helm values.                            |

**Note:** The listed ports are defaults. Only the `ks-transport-controller` and `status-addon-agent` metrics ports are configurable via Helm values (see [issue #2158](https://github.com/kubestellar/kubestellar/issues/2158)).

## Debug/Profiling Endpoints

Some KubeStellar components expose Go's built-in pprof debug endpoints for profiling and troubleshooting.

### pprof Endpoint Table

| Controller                      | Protocol | Port  | Path            | AuthN/AuthZ | Notes |
|----------------------------------|----------|-------|-----------------|-------------|-------|
| kubestellar-controller-manager   | HTTP     | 8082  | /debug/pprof/   | None        |  |
| ks-transport-controller         | HTTP     | 8092  | /debug/pprof/   | None        | Port configurable via Helm values. |
| status-addon-controller         | HTTP     | 9282  | /debug/pprof/   | None        |  |
| status-addon-agent              | HTTP     | 8082  | /debug/pprof/   | None        | Port configurable via Helm values. |

**Note:** The listed ports are defaults. Only the `ks-transport-controller` and `status-addon-agent` pprof ports are configurable via Helm values.

## Example: Accessing KubeStellar Controller Metrics and Debug Endpoints

**Note:** The following example assumes you have a running KubeStellar controller-manager pod and access to the appropriate Kubernetes context and namespace. The Deployment name is always `kubestellar-controller-manager`, but you may need to adjust the context and namespace for your environment.

```sh
kubectl --context kind-kubeflex port-forward -n wds1-system deployment/kubestellar-controller-manager 8443:8443 8082:8082
```

Access metrics: [https://localhost:8443/metrics](https://localhost:8443/metrics) (Kubernetes client authentication required)

Access pprof: [http://localhost:8082/debug/pprof/](http://localhost:8082/debug/pprof/)

## Grafana Dashboards

- Example Grafana dashboards and configuration can be found in [`monitoring/grafana/`](https://github.com/kubestellar/kubestellar/tree/main/monitoring/grafana).
- After deploying Prometheus and Grafana (or your preferred stack), you can import dashboards to visualize KubeStellar metrics.

## Additional Resources

- [KubeStellar Monitoring](https://github.com/kubestellar/kubestellar/tree/main/monitoring/) (one possible way to collect metrics and profiles)
- [Prometheus Operator Documentation](https://prometheus-operator.dev/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Go pprof Documentation](https://pkg.go.dev/net/http/pprof)

---

If you have suggestions for more observability features or documentation, please [open an issue](https://github.com/kubestellar/kubestellar/issues/new?labels=kind%2Fdocumentation&template=documentation_request.yaml).
