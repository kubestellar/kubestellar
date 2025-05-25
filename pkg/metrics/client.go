/*
Copyright 2024 The KubeStellar Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	"fmt"
	"io"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8smetrics "k8s.io/component-base/metrics"
)

// MultiSpaceClientMetrics maintains apiserver call metrics for a collection of spaces.
// A space is the service provided by a collection of equivalent apiservers.
type MultiSpaceClientMetrics interface {
	Register(RegisterFn) error

	// MetricsForSpace returns the metrics interface for a given space
	MetricsForSpace(string) ClientMetrics

	// SpaceRecord observes one round-trip latency.
	// resource is a value returned from GVRString
	SpaceRecord(space, resource, method string, err error, latency time.Duration)
}

// ClientMetrics is the client-of-apiserver metrics for various kinds of object
type ClientMetrics interface {
	// ResourceMetrics returns the metrics interface specialized to the given resource
	ResourceMetrics(schema.GroupVersionResource) ClientResourceMetrics

	// Record observes one round-trip latency.
	// resource is a value returned from GVRString
	Record(resource string, method string, err error, latency time.Duration)
}

type ClientResourceMetrics interface {
	// ResourceRecord observes one round-trip latency.
	// resource is a value returned from GVRString
	ResourceRecord(method string, err error, latency time.Duration)
}

// GVRString renders a GroupVersionResource as a domain name in a simple, regular way
func GVRString(gvr schema.GroupVersionResource) string {
	return gvr.Resource + "." + gvr.Version + "." + gvr.Group
}

type RegisterFn = func(k8smetrics.Registerable) error

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func MustRegister(reg RegisterFn, registerable ...interface{ Register(RegisterFn) error }) {
	for _, ra := range registerable {
		Must(ra.Register(reg))
	}
}

func MustRegisterAbles(reg RegisterFn, registerable ...k8smetrics.Registerable) {
	for _, ra := range registerable {
		Must(reg(ra))
	}
}

type multiSpaceClientMetrics struct {
	// CallLatency measures round-trip seconds.
	// Labels are:
	// - space
	// - resource
	// - method
	// - error
	CallLatency *k8smetrics.HistogramVec
}

type clientMetrics struct {
	// CallLatency measures round-trip seconds.
	// Labels are:
	// - resource
	// - method
	// - error
	CallLatency prometheus.ObserverVec
}

type clientResourceMetrics struct {
	Resource string
	Base     clientMetrics
}

var _ MultiSpaceClientMetrics = &multiSpaceClientMetrics{}

func NewMultiSpaceClientMetrics() *multiSpaceClientMetrics {
	return &multiSpaceClientMetrics{
		CallLatency: k8smetrics.NewHistogramVec(&k8smetrics.HistogramOpts{
			Namespace:      "kubestellar",
			Subsystem:      "apiserver_call",
			Name:           "latency_seconds",
			Help:           "apiserver call latency in seconds",
			Buckets:        []float64{0.01, 0.1, 0.2, 0.5, 1, 2, 5, 10, 20, 50, 100},
			StabilityLevel: k8smetrics.ALPHA,
		},
			[]string{"space", "resource", "method", "error"}),
	}
}

func (msc *multiSpaceClientMetrics) Register(reg RegisterFn) error {
	return reg(msc.CallLatency)
}

func (msc *multiSpaceClientMetrics) MetricsForSpace(space string) ClientMetrics {
	sub := msc.CallLatency.MustCurryWith(map[string]string{"space": space})
	return &clientMetrics{sub}
}

func (msc *multiSpaceClientMetrics) SpaceRecord(space, resource, method string, err error, latency time.Duration) {
	errStr := ErrorShort(err)
	msc.CallLatency.WithLabelValues(space, resource, method, errStr).Observe(latency.Seconds())
}

func (cm *clientMetrics) ResourceMetrics(gvr schema.GroupVersionResource) ClientResourceMetrics {
	return &clientResourceMetrics{
		Resource: GVRString(gvr),
		Base:     *cm,
	}
}
func (cm *clientMetrics) Record(resource, method string, err error, latency time.Duration) {
	errStr := ErrorShort(err)
	cm.CallLatency.WithLabelValues(resource, method, errStr).Observe(latency.Seconds())
}

func (crm *clientResourceMetrics) ResourceRecord(method string, err error, latency time.Duration) {
	crm.Base.Record(crm.Resource, method, err, latency)
}

func ErrorShort(err error) string {
	if err == nil {
		return ""
	}
	if apiStatus, is := err.(k8serrors.APIStatus); is {
		status := apiStatus.Status()
		return "apiStatus:" + string(status.Reason)
	}
	switch err {
	case errPanic:
		return "panic"
	case io.EOF:
		return "io.EOF"
	case io.ErrClosedPipe:
		return "io.ErrClosedPipe"
	case io.ErrUnexpectedEOF:
		return "io.ErrUnexpectedEOF"
	}
	return fmt.Sprintf("%T", err)
}
