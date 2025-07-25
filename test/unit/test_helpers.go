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

package unit

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/cache"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/binding"
	"github.com/kubestellar/kubestellar/pkg/util"
)

// Mock implementations for testing
type MockWorkloadEventHandler struct {
	mock.Mock
}

func (m *MockWorkloadEventHandler) HandleWorkloadObjectEvent(gvr schema.GroupVersionResource, oldObj, newObj util.MRObject, eventType binding.WorkloadEventType, wasDeletedFinalStateUnknown bool) {
	m.Called(gvr, oldObj, newObj, eventType, wasDeletedFinalStateUnknown)
}

type MockBindingPolicyResolver struct {
	mock.Mock
}

func (m *MockBindingPolicyResolver) GenerateBinding(bindingPolicyKey string) *v1alpha1.BindingSpec {
	args := m.Called(bindingPolicyKey)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*v1alpha1.BindingSpec)
}

func (m *MockBindingPolicyResolver) GetOwnerReference(bindingPolicyKey string) (metav1.OwnerReference, error) {
	args := m.Called(bindingPolicyKey)
	return args.Get(0).(metav1.OwnerReference), args.Error(1)
}

func (m *MockBindingPolicyResolver) CompareBinding(bindingPolicyKey string, bindingSpec *v1alpha1.BindingSpec) bool {
	args := m.Called(bindingPolicyKey, bindingSpec)
	return args.Bool(0)
}

func (m *MockBindingPolicyResolver) NoteBindingPolicy(bindingpolicy *v1alpha1.BindingPolicy) {
	m.Called(bindingpolicy)
}

func (m *MockBindingPolicyResolver) EnsureObjectData(bindingPolicyKey string, objIdentifier util.ObjectIdentifier, objUID, resourceVersion string, modulation binding.DownsyncModulation) (bool, error) {
	args := m.Called(bindingPolicyKey, objIdentifier, objUID, resourceVersion, modulation)
	return args.Bool(0), args.Error(1)
}

func (m *MockBindingPolicyResolver) RemoveObjectIdentifier(bindingPolicyKey string, objIdentifier util.ObjectIdentifier) bool {
	args := m.Called(bindingPolicyKey, objIdentifier)
	return args.Bool(0)
}

func (m *MockBindingPolicyResolver) GetObjectIdentifiers(bindingPolicyKey string) (sets.Set[util.ObjectIdentifier], error) {
	args := m.Called(bindingPolicyKey)
	return args.Get(0).(sets.Set[util.ObjectIdentifier]), args.Error(1)
}

func (m *MockBindingPolicyResolver) SetDestinations(bindingPolicyKey string, destinations sets.Set[string]) error {
	args := m.Called(bindingPolicyKey, destinations)
	return args.Error(0)
}

func (m *MockBindingPolicyResolver) ResolutionExists(bindingPolicyKey string) bool {
	args := m.Called(bindingPolicyKey)
	return args.Bool(0)
}

func (m *MockBindingPolicyResolver) GetSingletonReportedStateRequestForObject(objId util.ObjectIdentifier) (bool, int) {
	args := m.Called(objId)
	return args.Bool(0), args.Int(1)
}

func (m *MockBindingPolicyResolver) GetSingletonReportedStateRequestsForBinding(bindingPolicyKey string) []binding.SingletonReportedStateReturnStatus {
	args := m.Called(bindingPolicyKey)
	return args.Get(0).([]binding.SingletonReportedStateReturnStatus)
}

func (m *MockBindingPolicyResolver) DeleteResolution(bindingPolicyKey string) {
	m.Called(bindingPolicyKey)
}

func (m *MockBindingPolicyResolver) Broker() binding.ResolutionBroker {
	args := m.Called()
	return args.Get(0).(binding.ResolutionBroker)
}

type MockResolutionBroker struct {
	mock.Mock
}

func (m *MockResolutionBroker) RegisterCallbacks(callbacks binding.ResolutionCallbacks) error {
	args := m.Called(callbacks)
	return args.Error(0)
}

func (m *MockResolutionBroker) GetResolution(bindingPolicyKey string) binding.Resolution {
	args := m.Called(bindingPolicyKey)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(binding.Resolution)
}

func (m *MockResolutionBroker) NotifyBindingPolicyCallbacks(bindingPolicyKey string) {
	m.Called(bindingPolicyKey)
}

func (m *MockResolutionBroker) NotifySingletonRequestCallbacks(bindingPolicyKey string, objId util.ObjectIdentifier) {
	m.Called(bindingPolicyKey, objId)
}

type MockCombinedStatusResolver struct {
	mock.Mock
}

func (m *MockCombinedStatusResolver) ResolveCombinedStatus(statusCollector *v1alpha1.StatusCollector) (*v1alpha1.CombinedStatus, error) {
	args := m.Called(statusCollector)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v1alpha1.CombinedStatus), args.Error(1)
}

// Test data creation functions
func createTestBindingPolicy(name string) *v1alpha1.BindingPolicy {
	return &v1alpha1.BindingPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.BindingPolicySpec{
			ClusterSelectors: []metav1.LabelSelector{
				{
					MatchLabels: map[string]string{
						"env": "prod",
					},
				},
			},
			Downsync: []v1alpha1.DownsyncPolicyClause{
				{
					DownsyncObjectTest: v1alpha1.DownsyncObjectTest{
						APIGroup:   stringPtr("apps"),
						Resources:  []string{"deployments"},
						Namespaces: []string{"default"},
					},
					DownsyncModulation: v1alpha1.DownsyncModulation{
						CreateOnly: false,
					},
				},
			},
		},
		Status: v1alpha1.BindingPolicyStatus{
			ObservedGeneration: 1,
		},
	}
}

func createTestBinding(name string, bindingPolicyName string) *v1alpha1.Binding {
	return &v1alpha1.Binding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"binding-policy": bindingPolicyName,
			},
		},
		Spec: v1alpha1.BindingSpec{
			Workload: v1alpha1.DownsyncObjectClauses{
				NamespaceScope: []v1alpha1.NamespaceScopeDownsyncClause{
					{
						NamespaceScopeDownsyncObject: v1alpha1.NamespaceScopeDownsyncObject{
							GroupVersionResource: metav1.GroupVersionResource{
								Group:    "apps",
								Version:  "v1",
								Resource: "deployments",
							},
							Namespace:       "default",
							Name:            "test-deployment",
							ResourceVersion: "1",
						},
						DownsyncModulation: v1alpha1.DownsyncModulation{
							CreateOnly: false,
						},
					},
				},
			},
			Destinations: []v1alpha1.Destination{
				{
					ClusterId: "cluster-1",
				},
			},
		},
		Status: v1alpha1.BindingStatus{
			ObservedGeneration: 1,
		},
	}
}

func createTestWorkloadObject() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "test-deployment",
				"namespace": "default",
				"labels": map[string]interface{}{
					"app": "test",
				},
			},
			"spec": map[string]interface{}{
				"replicas": 3,
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app": "test",
					},
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"app": "test",
						},
					},
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"name":  "nginx",
								"image": "nginx:latest",
							},
						},
					},
				},
			},
		},
	}
}

func createTestWorkloadObjectWithSingletonStatus() *unstructured.Unstructured {
	obj := createTestWorkloadObject()
	obj.SetLabels(map[string]string{
		"app":                             "test",
		"kubestellar.io/singleton-status": "true",
	})
	return obj
}

func createTestStatusCollector(name string) *v1alpha1.StatusCollector {
	return &v1alpha1.StatusCollector{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.StatusCollectorSpec{
			Filter: (*v1alpha1.Expression)(stringPtr("true")),
			Select: []v1alpha1.NamedExpression{
				{
					Name: "status",
					Def:  v1alpha1.Expression("obj.status"),
				},
			},
			Limit: 20,
		},
		Status: v1alpha1.StatusCollectorStatus{
			ObservedGeneration: 1,
		},
	}
}

func createTestCombinedStatus(name, namespace string) *v1alpha1.CombinedStatus {
	return &v1alpha1.CombinedStatus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Results: []v1alpha1.NamedStatusCombination{
			{
				Name: "test-result",
				ColumnNames: []string{
					"status",
				},
				Rows: []v1alpha1.StatusCombinationRow{
					{
						Columns: []v1alpha1.Value{
							{
								Type:   "string",
								String: stringPtr("Ready"),
							},
						},
					},
				},
			},
		},
	}
}

func createTestWorkStatus(name, namespace string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "work.open-cluster-management.io/v1",
			"kind":       "WorkStatus",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
				"labels": map[string]interface{}{
					"transport.kubestellar.io/originWdsName": "test-wds",
				},
			},
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Applied",
						"status": "True",
					},
				},
			},
		},
	}
}

// Utility functions
func stringPtr(s string) *string {
	return &s
}

// Test reference types
type bindingPolicyRef string

func (bpr bindingPolicyRef) String() string {
	return string(bpr)
}

type bindingRef string

func (br bindingRef) String() string {
	return string(br)
}

type statusCollectorRef string

func (scr statusCollectorRef) String() string {
	return string(scr)
}

type combinedStatusRef string

func (csr combinedStatusRef) String() string {
	return string(csr)
}

type workloadObjectRef struct {
	util.ObjectIdentifier
}

func (wor workloadObjectRef) String() string {
	return wor.GVK.String() + "/" + wor.Resource + "/" + wor.ObjectName.Namespace + "/" + wor.ObjectName.Name
}

type workStatusRef struct {
	Name                   string
	WECName                string
	SourceObjectIdentifier util.ObjectIdentifier
}

func (wsr workStatusRef) String() string {
	return wsr.Name + ":" + wsr.WECName
}

func (wsr workStatusRef) ObjectName() cache.ObjectName {
	return cache.ObjectName{
		Namespace: wsr.WECName,
		Name:      wsr.Name,
	}
}

// Test logger
func createTestLogger(t *testing.T) logr.Logger {
	return logr.New(&testLogSink{t: t})
}

type testLogSink struct {
	t *testing.T
}

func (s *testLogSink) Init(info logr.RuntimeInfo) {}
func (s *testLogSink) Enabled(level int) bool     { return true }
func (s *testLogSink) Info(level int, msg string, keysAndValues ...interface{}) {
	s.t.Logf("INFO: %s %v", msg, keysAndValues)
}
func (s *testLogSink) Error(err error, msg string, keysAndValues ...interface{}) {
	s.t.Logf("ERROR: %s %v: %v", msg, keysAndValues, err)
}
func (s *testLogSink) WithValues(keysAndValues ...interface{}) logr.LogSink {
	return s
}
func (s *testLogSink) WithName(name string) logr.LogSink {
	return s
}
