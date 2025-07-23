# AI/ML Network Benchmarking Suite for KubeStellar

This tool benchmarks network performance for distributed AI/ML training patterns (AllReduce, Parameter Server, Bulk Transfer) across KubeStellar-managed clusters (WECs).

## Features
- Emulates AI/ML communication patterns: AllReduce, Parameter Server, Bulk Transfer
- Measures bandwidth, latency, jitter, collective op time, packet loss
- Configurable: node count, data size, topology, frequency, etc.
- Modular: easy to add new primitives
- Deployable via Helm across WECs

## Usage
1. Build the Go agent (`main.go`).
2. Deploy using the provided Helm chart (coming soon).
3. Configure test parameters in `values.yaml`.
4. Collect and analyze results.

See `main.go` for agent implementation details.
