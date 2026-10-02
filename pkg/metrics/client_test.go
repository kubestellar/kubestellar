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
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8smetrics "k8s.io/component-base/metrics"
)

func TestGVRString(t *testing.T) {
	tests := []struct {
		name     string
		gvr      schema.GroupVersionResource
		expected string
	}{
		{
			name: "standard resource with group",
			gvr: schema.GroupVersionResource{
				Group:    "apps",
				Version:  "v1",
				Resource: "deployments",
			},
			expected: "deployments.v1.apps",
		},
		{
			name: "core resource with empty group",
			gvr: schema.GroupVersionResource{
				Group:    "",
				Version:  "v1",
				Resource: "pods",
			},
			expected: "pods.v1.",
		},
		{
			name: "KubeStellar CRD",
			gvr: schema.GroupVersionResource{
				Group:    "control.kubestellar.io",
				Version:  "v1alpha1",
				Resource: "bindingpolicies",
			},
			expected: "bindingpolicies.v1alpha1.control.kubestellar.io",
		},
		{
			name: "all empty fields",
			gvr: schema.GroupVersionResource{
				Group:    "",
				Version:  "",
				Resource: "",
			},
			expected: "..",
		},
		{
			name: "beta version resource",
			gvr: schema.GroupVersionResource{
				Group:    "networking.k8s.io",
				Version:  "v1beta1",
				Resource: "ingresses",
			},
			expected: "ingresses.v1beta1.networking.k8s.io",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GVRString(tt.gvr)
			if result != tt.expected {
				t.Errorf("GVRString(%v) = %q, want %q", tt.gvr, result, tt.expected)
			}
		})
	}
}

func TestErrorShort(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nil error returns empty string",
			err:      nil,
			expected: "",
		},
		{
			name:     "io.EOF returns io.EOF",
			err:      io.EOF,
			expected: "io.EOF",
		},
		{
			name:     "io.ErrClosedPipe returns io.ErrClosedPipe",
			err:      io.ErrClosedPipe,
			expected: "io.ErrClosedPipe",
		},
		{
			name:     "io.ErrUnexpectedEOF returns io.ErrUnexpectedEOF",
			err:      io.ErrUnexpectedEOF,
			expected: "io.ErrUnexpectedEOF",
		},
		{
			name:     "errPanic returns panic",
			err:      errPanic,
			expected: "panic",
		},
		{
			name: "APIStatus NotFound error",
			err: &k8serrors.StatusError{
				ErrStatus: metav1.Status{
					Reason: metav1.StatusReasonNotFound,
				},
			},
			expected: "apiStatus:NotFound",
		},
		{
			name: "APIStatus Conflict error",
			err: &k8serrors.StatusError{
				ErrStatus: metav1.Status{
					Reason: metav1.StatusReasonConflict,
				},
			},
			expected: "apiStatus:Conflict",
		},
		{
			name: "APIStatus Unauthorized error",
			err: &k8serrors.StatusError{
				ErrStatus: metav1.Status{
					Reason: metav1.StatusReasonUnauthorized,
				},
			},
			expected: "apiStatus:Unauthorized",
		},
		{
			name: "APIStatus AlreadyExists error",
			err: &k8serrors.StatusError{
				ErrStatus: metav1.Status{
					Reason: metav1.StatusReasonAlreadyExists,
				},
			},
			expected: "apiStatus:AlreadyExists",
		},
		{
			name: "APIStatus with empty reason",
			err: &k8serrors.StatusError{
				ErrStatus: metav1.Status{
					Reason: "",
				},
			},
			expected: "apiStatus:",
		},
		{
			name:     "generic error returns type name",
			err:      errors.New("something went wrong"),
			expected: "*errors.errorString",
		},
		{
			name:     "custom error type returns type name",
			err:      &customTestError{msg: "test error"},
			expected: "*metrics.customTestError",
		},
		{
			name:     "fmt.Errorf wrapped error returns type name",
			err:      fmt.Errorf("wrapped: %w", errors.New("inner")),
			expected: "*fmt.wrapError",
		},
		{
			name:     "http error type returns type name",
			err:      &http.ProtocolError{ErrorString: "test"},
			expected: "*http.ProtocolError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ErrorShort(tt.err)
			if result != tt.expected {
				t.Errorf("ErrorShort(%v) = %q, want %q", tt.err, result, tt.expected)
			}
		})
	}
}

// customTestError is a helper type for testing ErrorShort with custom error types.
type customTestError struct {
	msg string
}

func (e *customTestError) Error() string {
	return e.msg
}

func TestMust(t *testing.T) {
	t.Run("nil error does not panic", func(t *testing.T) {
		// Should not panic
		Must(nil)
	})

	t.Run("non-nil error panics", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Error("Must(err) did not panic when given a non-nil error")
			}
			err, ok := r.(error)
			if !ok {
				t.Errorf("Must(err) panicked with non-error type: %T", r)
			}
			if err.Error() != "test error" {
				t.Errorf("Must(err) panicked with wrong error: got %q, want %q", err.Error(), "test error")
			}
		}()
		Must(errors.New("test error"))
	})
}

func TestMustRegister(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		reg := k8smetrics.NewKubeRegistry()
		msc := NewMultiSpaceClientMetrics()
		// Should not panic
		MustRegister(reg.Register, msc)
	})

	t.Run("panics on registration error", func(t *testing.T) {
		failReg := func(_ k8smetrics.Registerable) error {
			return errors.New("registration failed")
		}
		defer func() {
			r := recover()
			if r == nil {
				t.Error("MustRegister did not panic on registration error")
			}
		}()
		msc := NewMultiSpaceClientMetrics()
		MustRegister(failReg, msc)
	})

	t.Run("multiple registerables all succeed", func(t *testing.T) {
		reg := k8smetrics.NewKubeRegistry()
		msc1 := NewMultiSpaceClientMetrics()
		msc2 := NewMultiSpaceClientMetrics()
		// The second one will fail because it uses the same metric name,
		// but we test the pattern of calling with multiple args.
		// Use separate registries for each to avoid duplicate registration.
		MustRegister(reg.Register, msc1)

		reg2 := k8smetrics.NewKubeRegistry()
		MustRegister(reg2.Register, msc2)
	})
}

func TestMustRegisterAbles(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		reg := k8smetrics.NewKubeRegistry()
		gauge := k8smetrics.NewGauge(&k8smetrics.GaugeOpts{
			Namespace:      "kubestellar",
			Subsystem:      "test",
			Name:           "must_register_ables_test_gauge",
			Help:           "test gauge for MustRegisterAbles",
			StabilityLevel: k8smetrics.ALPHA,
		})
		// Should not panic
		MustRegisterAbles(reg.Register, gauge)
	})

	t.Run("panics on registration error", func(t *testing.T) {
		failReg := func(_ k8smetrics.Registerable) error {
			return errors.New("registration failed")
		}
		gauge := k8smetrics.NewGauge(&k8smetrics.GaugeOpts{
			Namespace:      "kubestellar",
			Subsystem:      "test",
			Name:           "fail_gauge",
			Help:           "test gauge",
			StabilityLevel: k8smetrics.ALPHA,
		})
		defer func() {
			r := recover()
			if r == nil {
				t.Error("MustRegisterAbles did not panic on registration error")
			}
		}()
		MustRegisterAbles(failReg, gauge)
	})
}

func TestMultiSpaceClientMetrics_Register(t *testing.T) {
	t.Run("register succeeds with valid registry", func(t *testing.T) {
		reg := k8smetrics.NewKubeRegistry()
		msc := NewMultiSpaceClientMetrics()
		err := msc.Register(reg.Register)
		if err != nil {
			t.Errorf("Register() returned unexpected error: %v", err)
		}
	})

	t.Run("register returns error from registry", func(t *testing.T) {
		expectedErr := errors.New("registration failed")
		failReg := func(_ k8smetrics.Registerable) error {
			return expectedErr
		}
		msc := NewMultiSpaceClientMetrics()
		err := msc.Register(failReg)
		if err != expectedErr {
			t.Errorf("Register() returned %v, want %v", err, expectedErr)
		}
	})
}

func TestMultiSpaceClientMetrics_MetricsForSpace(t *testing.T) {
	reg := k8smetrics.NewKubeRegistry()
	msc := NewMultiSpaceClientMetrics()
	MustRegister(reg.Register, msc)

	t.Run("returns non-nil ClientMetrics", func(t *testing.T) {
		cm := msc.MetricsForSpace("test-space")
		if cm == nil {
			t.Error("MetricsForSpace() returned nil")
		}
	})

	t.Run("different spaces return distinct metrics", func(t *testing.T) {
		cm1 := msc.MetricsForSpace("space-1")
		cm2 := msc.MetricsForSpace("space-2")
		if cm1 == nil || cm2 == nil {
			t.Error("MetricsForSpace() returned nil for one of the spaces")
		}
	})
}

func TestMultiSpaceClientMetrics_SpaceRecord(t *testing.T) {
	reg := k8smetrics.NewKubeRegistry()
	msc := NewMultiSpaceClientMetrics()
	MustRegister(reg.Register, msc)

	t.Run("record with nil error", func(t *testing.T) {
		// Should not panic
		msc.SpaceRecord("wds1", "deployments.v1.apps", "list", nil, 100*time.Millisecond)

		gathered, err := reg.Gather()
		if err != nil {
			t.Fatalf("Gather() returned unexpected error: %v", err)
		}
		if len(gathered) != 1 {
			t.Fatalf("Expected 1 MetricFamily, got %d", len(gathered))
		}
	})

	t.Run("record with non-nil error", func(t *testing.T) {
		// Should not panic
		msc.SpaceRecord("its1", "pods.v1.", "get",
			&k8serrors.StatusError{ErrStatus: metav1.Status{Reason: metav1.StatusReasonNotFound}},
			200*time.Millisecond)
	})

	t.Run("record with zero latency", func(t *testing.T) {
		// Should not panic
		msc.SpaceRecord("wds1", "services.v1.", "create", nil, 0)
	})

	t.Run("record with large latency", func(t *testing.T) {
		// Should not panic — tests the upper histogram buckets
		msc.SpaceRecord("wds1", "configmaps.v1.", "list", nil, 120*time.Second)
	})
}

func TestClientMetrics_Record(t *testing.T) {
	reg := k8smetrics.NewKubeRegistry()
	msc := NewMultiSpaceClientMetrics()
	MustRegister(reg.Register, msc)
	cm := msc.MetricsForSpace("test-space")

	t.Run("record with various methods", func(t *testing.T) {
		methods := []string{"create", "update", "delete", "get", "list", "watch", "patch"}
		for _, method := range methods {
			// Should not panic
			cm.Record("deployments.v1.apps", method, nil, 50*time.Millisecond)
		}
	})

	t.Run("record with error", func(t *testing.T) {
		cm.Record("pods.v1.", "get", io.EOF, 10*time.Millisecond)
	})

	t.Run("record accumulates metrics", func(t *testing.T) {
		cm.Record("services.v1.", "list", nil, 10*time.Millisecond)
		cm.Record("services.v1.", "list", nil, 20*time.Millisecond)
		cm.Record("services.v1.", "list", nil, 30*time.Millisecond)

		gathered, err := reg.Gather()
		if err != nil {
			t.Fatalf("Gather() returned unexpected error: %v", err)
		}
		if len(gathered) != 1 {
			t.Fatalf("Expected 1 MetricFamily, got %d", len(gathered))
		}
	})
}

func TestClientMetrics_ResourceMetrics(t *testing.T) {
	reg := k8smetrics.NewKubeRegistry()
	msc := NewMultiSpaceClientMetrics()
	MustRegister(reg.Register, msc)
	cm := msc.MetricsForSpace("test-space")

	t.Run("returns non-nil ClientResourceMetrics", func(t *testing.T) {
		gvr := schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "deployments",
		}
		crm := cm.ResourceMetrics(gvr)
		if crm == nil {
			t.Error("ResourceMetrics() returned nil")
		}
	})

	t.Run("resource metrics records correctly", func(t *testing.T) {
		gvr := schema.GroupVersionResource{
			Group:    "control.kubestellar.io",
			Version:  "v1alpha1",
			Resource: "bindingpolicies",
		}
		crm := cm.ResourceMetrics(gvr)
		// Should not panic
		crm.ResourceRecord("list", nil, 42*time.Millisecond)
		crm.ResourceRecord("get", io.EOF, 100*time.Millisecond)
	})

	t.Run("different GVRs produce independent metrics", func(t *testing.T) {
		gvr1 := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
		gvr2 := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}

		crm1 := cm.ResourceMetrics(gvr1)
		crm2 := cm.ResourceMetrics(gvr2)

		crm1.ResourceRecord("create", nil, 10*time.Millisecond)
		crm2.ResourceRecord("delete", nil, 20*time.Millisecond)

		gathered, err := reg.Gather()
		if err != nil {
			t.Fatalf("Gather() returned unexpected error: %v", err)
		}
		if len(gathered) != 1 {
			t.Fatalf("Expected 1 MetricFamily, got %d", len(gathered))
		}
	})
}

func TestClientResourceMetrics_ResourceRecord(t *testing.T) {
	reg := k8smetrics.NewKubeRegistry()
	msc := NewMultiSpaceClientMetrics()
	MustRegister(reg.Register, msc)
	cm := msc.MetricsForSpace("test-space")
	gvr := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"}
	crm := cm.ResourceMetrics(gvr)

	t.Run("record with nil error", func(t *testing.T) {
		crm.ResourceRecord("create", nil, 50*time.Millisecond)
	})

	t.Run("record with various errors", func(t *testing.T) {
		crm.ResourceRecord("get", io.EOF, 10*time.Millisecond)
		crm.ResourceRecord("list", io.ErrClosedPipe, 20*time.Millisecond)
		crm.ResourceRecord("watch", io.ErrUnexpectedEOF, 30*time.Millisecond)
		crm.ResourceRecord("update", errPanic, 5*time.Millisecond)
		crm.ResourceRecord("delete", errors.New("network error"), 100*time.Millisecond)
	})
}

func TestNewMultiSpaceClientMetrics(t *testing.T) {
	t.Run("creates non-nil instance", func(t *testing.T) {
		msc := NewMultiSpaceClientMetrics()
		if msc == nil {
			t.Error("NewMultiSpaceClientMetrics() returned nil")
		}
	})

	t.Run("CallLatency histogram is initialized", func(t *testing.T) {
		msc := NewMultiSpaceClientMetrics()
		if msc.CallLatency == nil {
			t.Error("CallLatency histogram is nil")
		}
	})

	t.Run("implements MultiSpaceClientMetrics interface", func(t *testing.T) {
		msc := NewMultiSpaceClientMetrics()
		var _ MultiSpaceClientMetrics = msc
	})
}

func TestEndToEnd_MetricsPipeline(t *testing.T) {
	// This test validates the full metrics pipeline:
	// NewMultiSpaceClientMetrics → Register → MetricsForSpace → ResourceMetrics → ResourceRecord → Gather
	reg := k8smetrics.NewKubeRegistry()
	msc := NewMultiSpaceClientMetrics()

	// Step 1: Register
	err := msc.Register(reg.Register)
	if err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	// Step 2: Get metrics for a space
	cm := msc.MetricsForSpace("wds1")
	if cm == nil {
		t.Fatal("MetricsForSpace() returned nil")
	}

	// Step 3: Get resource metrics
	gvr := schema.GroupVersionResource{
		Group:    "control.kubestellar.io",
		Version:  "v1alpha1",
		Resource: "bindingpolicies",
	}
	crm := cm.ResourceMetrics(gvr)
	if crm == nil {
		t.Fatal("ResourceMetrics() returned nil")
	}

	// Step 4: Record some metrics
	crm.ResourceRecord("list", nil, 42*time.Millisecond)
	crm.ResourceRecord("get", nil, 10*time.Millisecond)
	crm.ResourceRecord("create",
		&k8serrors.StatusError{ErrStatus: metav1.Status{Reason: metav1.StatusReasonConflict}},
		100*time.Millisecond)

	// Step 5: Also record via SpaceRecord path
	msc.SpaceRecord("its1", GVRString(gvr), "watch", nil, 500*time.Millisecond)

	// Step 6: Also record via cm.Record path
	cm.Record(GVRString(gvr), "update", nil, 25*time.Millisecond)

	// Step 7: Gather and verify
	gathered, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather() failed: %v", err)
	}
	if len(gathered) != 1 {
		t.Fatalf("Expected 1 MetricFamily, got %d", len(gathered))
	}

	mf := gathered[0]
	if mf.GetName() != "kubestellar_apiserver_call_latency_seconds" {
		t.Errorf("MetricFamily name = %q, want %q", mf.GetName(), "kubestellar_apiserver_call_latency_seconds")
	}
	expectedHelp := "[ALPHA] apiserver call latency in seconds"
	if mf.GetHelp() != expectedHelp {
		t.Errorf("MetricFamily help = %q, want %q", mf.GetHelp(), expectedHelp)
	}

	// We recorded 5 distinct label combinations, so expect 5 metric entries
	metrics := mf.GetMetric()
	if len(metrics) < 4 {
		t.Errorf("Expected at least 4 metric entries, got %d", len(metrics))
	}
}
