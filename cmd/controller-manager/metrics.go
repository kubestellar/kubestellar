package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	reconcileErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "reconcile_errors_total",
			Help: "Total number of reconcile errors per controller.",
		},
	)
	reconcileQueueLength = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "reconcile_queue_length",
			Help: "Length of reconcile queue per controller.",
		},
	)
	reconcileDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "reconcile_time_seconds",
			Help:    "Length of time per reconcile per controller.",
			Buckets: prometheus.DefBuckets,
		},
	)
)

func initMetrics() {
	// Register custom metrics
	prometheus.MustRegister(reconcileErrors)
	prometheus.MustRegister(reconcileQueueLength)
	prometheus.MustRegister(reconcileDuration)

	// Register built-in Go metrics
	prometheus.MustRegister(prometheus.NewGoCollector())         // go_goroutines, go_info
	prometheus.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{})) // optional but includes memory, CPU, etc.

	// Expose /metrics endpoint
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			panic("Failed to start metrics HTTP server: " + err.Error())
		}
	}()
}

