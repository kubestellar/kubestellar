# Observability in KubeStellar

KubeStellar provides several endpoints and integrations for observability and monitoring. This page lists the available observability features, how to access them, and how to use them in a typical deployment.

## Prometheus Metrics

KubeStellar exposes Prometheus-compatible metrics endpoints for its core components. These metrics can be scraped by a Prometheus server for monitoring and alerting.

- **Default Metrics Endpoint:**
  - Most KubeStellar controllers expose metrics at `/metrics` on their respective pods (usually on port 8080 or 9090).
  - Example: `http://<pod-ip>:8080/metrics`
- **How to enable:**
  - Prometheus scraping is enabled by default in the provided Helm chart values.
  - Example manifests for Prometheus configuration can be found in [`monitoring/prometheus/`](https://github.com/kubestellar/kubestellar/tree/main/monitoring/prometheus) and [`monitoring/configuration/`](https://github.com/kubestellar/kubestellar/tree/main/monitoring/configuration).
  - To install monitoring, see [`monitoring/install-ks-monitoring.sh`](https://github.com/kubestellar/kubestellar/blob/main/monitoring/install-ks-monitoring.sh).
- **What you get:**
  - Standard Go and process metrics (memory, CPU, GC, etc.)
  - KubeStellar-specific metrics (object reconciliation, errors, etc.)

## Debug Endpoints

Some KubeStellar components expose Go's built-in pprof debug endpoints for profiling and troubleshooting.

- **pprof Endpoints:**
  - Accessible at `/debug/pprof/` on the controller pods (if enabled).
  - Example: `http://<pod-ip>:8080/debug/pprof/`
  - See [Go pprof documentation](https://pkg.go.dev/net/http/pprof) for usage.

## Example: Accessing Metrics and Debug Endpoints

1. **Port-forward to a controller pod:**
   ```sh
   kubectl -n <namespace> port-forward deployment/kubestellar-controller-manager 8080:8080
   ```
2. **Access metrics:**
   - Open [http://localhost:8080/metrics](http://localhost:8080/metrics)
3. **Access pprof:**
   - Open [http://localhost:8080/debug/pprof/](http://localhost:8080/debug/pprof/)

## Grafana Dashboards

- Example Grafana dashboards and configuration can be found in [`monitoring/grafana/`](https://github.com/kubestellar/kubestellar/tree/main/monitoring/grafana).
- After deploying Prometheus and Grafana, you can import dashboards to visualize KubeStellar metrics.

## Additional Resources

- [KubeStellar Monitoring README](https://github.com/kubestellar/kubestellar/tree/main/monitoring/README.md)
- [Prometheus Operator Documentation](https://prometheus-operator.dev/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Go pprof Documentation](https://pkg.go.dev/net/http/pprof)

---

If you have suggestions for more observability features or documentation, please [open an issue](https://github.com/kubestellar/kubestellar/issues/new?labels=kind%2Fdocumentation&template=documentation_request.yaml).
