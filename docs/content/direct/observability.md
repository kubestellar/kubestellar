# Observability in KubeStellar

KubeStellar provides endpoints and integrations for observability and monitoring. This page describes the available observability features, how to access them, and how to use them in a typical deployment.

## Metrics Endpoints

KubeStellar controllers expose Prometheus-compatible metrics endpoints. These endpoints respond to HTTP requests for metrics and can be queried by any monitoring system; KubeStellar does not mandate how metrics are collected.

### Metrics Endpoint Details

- **Protocol:** HTTP
- **Path:** `/metrics`
- **Typical Port:** 8080 (sometimes 9090)
- **Authentication/Authorization:**
  - By default, metrics endpoints are not authenticated within the cluster. If you expose them outside the cluster, secure them appropriately.
- **Example:**
  - `http://<pod-ip>:8080/metrics`

### Enabling Metrics

KubeStellar controllers respond to requests for metrics at the `/metrics` endpoint by default. For one possible way to collect these metrics using Prometheus and Helm, see the [monitoring/README.md](https://github.com/kubestellar/kubestellar/tree/main/monitoring/README.md). You may use any monitoring stack you prefer.

## Debug/Profiling Endpoints

Some KubeStellar components expose Go's built-in pprof debug endpoints for profiling and troubleshooting.

### pprof Endpoint Details

- **Protocol:** HTTP
- **Path:** `/debug/pprof/`
- **Typical Port:** 8082
- **Authentication/Authorization:**
  - By default, pprof endpoints are not authenticated within the cluster. Secure them if exposed externally.
- **Example:**
  - `http://<pod-ip>:8082/debug/pprof/`
- **Reference:** [Go pprof documentation](https://pkg.go.dev/net/http/pprof)

## Example: Accessing Metrics and Debug Endpoints

> **Note:** The following examples assume you have a running KubeStellar controller pod and access to the appropriate Kubernetes context and namespace. Adjust the context, namespace, and deployment name as needed for your environment.

1. **Port-forward to a controller pod:**
   ```sh
   # Example for a generic cluster/namespace:
   kubectl -n <namespace> port-forward deployment/kubestellar-controller-manager 8080:8080 8082:8082

   # Example for a kind-based test environment:
   kubectl --context kind-kubeflex port-forward -n wds1-system deployment/kubestellar-controller-manager 8080:8080 8082:8082
   ```
2. **Access metrics:**
   - Open [http://localhost:8080/metrics](http://localhost:8080/metrics)
3. **Access pprof:**
   - Open [http://localhost:8082/debug/pprof/](http://localhost:8082/debug/pprof/)

## Grafana Dashboards

- Example Grafana dashboards and configuration can be found in [`monitoring/grafana/`](https://github.com/kubestellar/kubestellar/tree/main/monitoring/grafana).
- After deploying Prometheus and Grafana (or your preferred stack), you can import dashboards to visualize KubeStellar metrics.

## Additional Resources

- [KubeStellar Monitoring README](https://github.com/kubestellar/kubestellar/tree/main/monitoring/README.md) (one possible way to collect metrics)
- [Prometheus Operator Documentation](https://prometheus-operator.dev/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Go pprof Documentation](https://pkg.go.dev/net/http/pprof)

---

If you have suggestions for more observability features or documentation, please [open an issue](https://github.com/kubestellar/kubestellar/issues/new?labels=kind%2Fdocumentation&template=documentation_request.yaml).
