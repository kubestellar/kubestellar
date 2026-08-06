/*
Copyright 2025 The KubeStellar Authors.

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

package binding

import (
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/util"
)

func TestLabelsMatchAny(t *testing.T) {
	logger := klog.Background()

	cases := []struct {
		name      string
		labelSet  map[string]string
		selectors []metav1.LabelSelector
		want      bool
	}{
		{
			name:      "empty selectors returns false",
			labelSet:  map[string]string{"app": "nginx"},
			selectors: nil,
			want:      false,
		},
		{
			name:     "single matchLabels selector matches",
			labelSet: map[string]string{"app": "nginx", "env": "prod"},
			selectors: []metav1.LabelSelector{
				{MatchLabels: map[string]string{"app": "nginx"}},
			},
			want: true,
		},
		{
			name:     "matchLabels mismatch returns false",
			labelSet: map[string]string{"app": "nginx"},
			selectors: []metav1.LabelSelector{
				{MatchLabels: map[string]string{"app": "postgres"}},
			},
			want: false,
		},
		{
			name:     "first selector fails but second matches (OR semantics)",
			labelSet: map[string]string{"app": "nginx"},
			selectors: []metav1.LabelSelector{
				{MatchLabels: map[string]string{"app": "postgres"}},
				{MatchLabels: map[string]string{"app": "nginx"}},
			},
			want: true,
		},
		{
			name:     "empty matchLabels selector matches everything",
			labelSet: map[string]string{"app": "nginx"},
			selectors: []metav1.LabelSelector{
				{},
			},
			want: true,
		},
		{
			name:     "matchExpressions In operator matches",
			labelSet: map[string]string{"env": "staging"},
			selectors: []metav1.LabelSelector{
				{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{Key: "env", Operator: metav1.LabelSelectorOpIn, Values: []string{"staging", "prod"}},
					},
				},
			},
			want: true,
		},
		{
			name:     "matchExpressions NotIn rejects matching value",
			labelSet: map[string]string{"env": "staging"},
			selectors: []metav1.LabelSelector{
				{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{Key: "env", Operator: metav1.LabelSelectorOpNotIn, Values: []string{"staging"}},
					},
				},
			},
			want: false,
		},
		{
			name:     "nil label set does not match non-empty requirement",
			labelSet: nil,
			selectors: []metav1.LabelSelector{
				{MatchLabels: map[string]string{"app": "nginx"}},
			},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := labelsMatchAny(logger, tc.labelSet, tc.selectors)
			if got != tc.want {
				t.Errorf("labelsMatchAny() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestALabelSelectorIsEmpty(t *testing.T) {
	cases := []struct {
		name      string
		selectors []metav1.LabelSelector
		want      bool
	}{
		{
			name:      "no selectors",
			selectors: nil,
			want:      false,
		},
		{
			name: "one non-empty matchLabels",
			selectors: []metav1.LabelSelector{
				{MatchLabels: map[string]string{"k": "v"}},
			},
			want: false,
		},
		{
			name: "one empty selector",
			selectors: []metav1.LabelSelector{
				{},
			},
			want: true,
		},
		{
			name: "first non-empty then empty — empty wins",
			selectors: []metav1.LabelSelector{
				{MatchLabels: map[string]string{"k": "v"}},
				{},
			},
			want: true,
		},
		{
			name: "both non-empty",
			selectors: []metav1.LabelSelector{
				{MatchLabels: map[string]string{"a": "1"}},
				{MatchLabels: map[string]string{"b": "2"}},
			},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ALabelSelectorIsEmpty(tc.selectors...)
			if got != tc.want {
				t.Errorf("ALabelSelectorIsEmpty() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDestinationsMatch(t *testing.T) {
	cases := []struct {
		name          string
		resolvedDests sets.Set[string]
		bindingDests  []v1alpha1.Destination
		want          bool
	}{
		{
			name:          "both empty",
			resolvedDests: sets.New[string](),
			bindingDests:  nil,
			want:          true,
		},
		{
			name:          "length mismatch",
			resolvedDests: sets.New("cluster1", "cluster2"),
			bindingDests:  []v1alpha1.Destination{{ClusterId: "cluster1"}},
			want:          false,
		},
		{
			name:          "same elements",
			resolvedDests: sets.New("cluster1", "cluster2"),
			bindingDests: []v1alpha1.Destination{
				{ClusterId: "cluster1"},
				{ClusterId: "cluster2"},
			},
			want: true,
		},
		{
			name:          "same length but different elements",
			resolvedDests: sets.New("cluster1", "cluster3"),
			bindingDests: []v1alpha1.Destination{
				{ClusterId: "cluster1"},
				{ClusterId: "cluster2"},
			},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := destinationsMatch(tc.resolvedDests, tc.bindingDests)
			if got != tc.want {
				t.Errorf("destinationsMatch() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDestinationsStringSetToSortedDestinations(t *testing.T) {
	cases := []struct {
		name  string
		input sets.Set[string]
	}{
		{name: "nil set", input: nil},
		{name: "empty set", input: sets.New[string]()},
		{name: "single element", input: sets.New("cluster1")},
		{name: "multiple elements sorted lexicographically", input: sets.New("zz", "aa", "mm")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := destinationsStringSetToSortedDestinations(tc.input)

			if len(got) != tc.input.Len() {
				t.Fatalf("expected %d destinations, got %d", tc.input.Len(), len(got))
			}

			gotSet := sets.New[string]()
			for _, d := range got {
				gotSet.Insert(d.ClusterId)
			}
			if !gotSet.Equal(tc.input) {
				t.Errorf("element mismatch: got %v, want %v", gotSet, tc.input)
			}

			for i := 1; i < len(got); i++ {
				if got[i-1].ClusterId > got[i].ClusterId {
					t.Errorf("not sorted: %q > %q at indices %d,%d",
						got[i-1].ClusterId, got[i].ClusterId, i-1, i)
				}
			}
		})
	}
}

func TestSortBindingWorkloadObjects(t *testing.T) {
	t.Run("cluster-scope sorted by GVR then name", func(t *testing.T) {
		input := v1alpha1.DownsyncObjectClauses{
			ClusterScope: []v1alpha1.ClusterScopeDownsyncClause{
				{ClusterScopeDownsyncObject: v1alpha1.ClusterScopeDownsyncObject{
					GroupVersionResource: metav1.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
					Name:                 "zzz",
				}},
				{ClusterScopeDownsyncObject: v1alpha1.ClusterScopeDownsyncObject{
					GroupVersionResource: metav1.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
					Name:                 "aaa",
				}},
			},
		}
		sortBindingWorkloadObjects(&input)
		if input.ClusterScope[0].Name != "aaa" || input.ClusterScope[1].Name != "zzz" {
			t.Errorf("expected [aaa, zzz], got [%s, %s]",
				input.ClusterScope[0].Name, input.ClusterScope[1].Name)
		}
	})

	t.Run("namespace-scope sorted by namespace", func(t *testing.T) {
		input := v1alpha1.DownsyncObjectClauses{
			NamespaceScope: []v1alpha1.NamespaceScopeDownsyncClause{
				{NamespaceScopeDownsyncObject: v1alpha1.NamespaceScopeDownsyncObject{
					GroupVersionResource: metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"},
					Namespace:            "z-ns", Name: "pod1",
				}},
				{NamespaceScopeDownsyncObject: v1alpha1.NamespaceScopeDownsyncObject{
					GroupVersionResource: metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"},
					Namespace:            "a-ns", Name: "pod1",
				}},
			},
		}
		sortBindingWorkloadObjects(&input)
		if input.NamespaceScope[0].Namespace != "a-ns" || input.NamespaceScope[1].Namespace != "z-ns" {
			t.Errorf("expected [a-ns, z-ns], got [%s, %s]",
				input.NamespaceScope[0].Namespace, input.NamespaceScope[1].Namespace)
		}
	})

	t.Run("already sorted is stable", func(t *testing.T) {
		input := v1alpha1.DownsyncObjectClauses{
			ClusterScope: []v1alpha1.ClusterScopeDownsyncClause{
				{ClusterScopeDownsyncObject: v1alpha1.ClusterScopeDownsyncObject{
					GroupVersionResource: metav1.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
					Name:                 "aaa",
				}},
				{ClusterScopeDownsyncObject: v1alpha1.ClusterScopeDownsyncObject{
					GroupVersionResource: metav1.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
					Name:                 "zzz",
				}},
			},
		}
		sortBindingWorkloadObjects(&input)
		if input.ClusterScope[0].Name != "aaa" || input.ClusterScope[1].Name != "zzz" {
			t.Errorf("expected [aaa, zzz], got [%s, %s]",
				input.ClusterScope[0].Name, input.ClusterScope[1].Name)
		}
	})
}

func TestDownsyncModulationEqual(t *testing.T) {
	cases := []struct {
		name  string
		left  DownsyncModulation
		right DownsyncModulation
		want  bool
	}{
		{
			name:  "identical zero values",
			left:  DownsyncModulation{StatusCollectors: sets.New[string]()},
			right: DownsyncModulation{StatusCollectors: sets.New[string]()},
			want:  true,
		},
		{
			name:  "same status collectors",
			left:  DownsyncModulation{StatusCollectors: sets.New("sc1", "sc2")},
			right: DownsyncModulation{StatusCollectors: sets.New("sc1", "sc2")},
			want:  true,
		},
		{
			name:  "different status collectors",
			left:  DownsyncModulation{StatusCollectors: sets.New("sc1", "sc2")},
			right: DownsyncModulation{StatusCollectors: sets.New("sc1")},
			want:  false,
		},
		{
			name:  "createOnly differs",
			left:  DownsyncModulation{CreateOnly: true, StatusCollectors: sets.New[string]()},
			right: DownsyncModulation{CreateOnly: false, StatusCollectors: sets.New[string]()},
			want:  false,
		},
		{
			name:  "wantSingleton differs",
			left:  DownsyncModulation{WantSingletonReportedState: true, StatusCollectors: sets.New[string]()},
			right: DownsyncModulation{WantSingletonReportedState: false, StatusCollectors: sets.New[string]()},
			want:  false,
		},
		{
			name:  "wantMultiWEC differs",
			left:  DownsyncModulation{WantMultiWECReportedState: true, StatusCollectors: sets.New[string]()},
			right: DownsyncModulation{WantMultiWECReportedState: false, StatusCollectors: sets.New[string]()},
			want:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.left.Equal(tc.right); got != tc.want {
				t.Errorf("Equal() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDownsyncModulationAddExternal(t *testing.T) {
	t.Run("ORs boolean flags and unions StatusCollectors", func(t *testing.T) {
		dm := ZeroDownsyncModulation()
		dm.AddExternal(v1alpha1.DownsyncModulation{
			CreateOnly:                 true,
			StatusCollectors:           []string{"sc1", "sc2"},
			WantSingletonReportedState: true,
		})

		if !dm.CreateOnly {
			t.Error("expected CreateOnly = true")
		}
		if !dm.WantSingletonReportedState {
			t.Error("expected WantSingletonReportedState = true")
		}
		if !dm.StatusCollectors.HasAll("sc1", "sc2") {
			t.Errorf("expected sc1, sc2 in StatusCollectors, got %v", dm.StatusCollectors)
		}
	})

	t.Run("accumulates across multiple calls", func(t *testing.T) {
		dm := ZeroDownsyncModulation()
		dm.AddExternal(v1alpha1.DownsyncModulation{StatusCollectors: []string{"sc1"}})
		dm.AddExternal(v1alpha1.DownsyncModulation{StatusCollectors: []string{"sc2"}})
		dm.AddExternal(v1alpha1.DownsyncModulation{WantMultiWECReportedState: true})

		if !dm.StatusCollectors.HasAll("sc1", "sc2") {
			t.Errorf("expected sc1, sc2, got %v", dm.StatusCollectors)
		}
		if !dm.WantMultiWECReportedState {
			t.Error("expected WantMultiWECReportedState = true")
		}
	})
}

func TestDownsyncModulationRoundTrip(t *testing.T) {
	orig := v1alpha1.DownsyncModulation{
		CreateOnly:                 true,
		StatusCollectors:           []string{"sc-a", "sc-b"},
		WantSingletonReportedState: true,
		WantMultiWECReportedState:  false,
	}
	internal := DownsyncModulationFromExternal(orig)
	rt := internal.ToExternal()

	if rt.CreateOnly != orig.CreateOnly {
		t.Errorf("CreateOnly: got %v, want %v", rt.CreateOnly, orig.CreateOnly)
	}
	if rt.WantSingletonReportedState != orig.WantSingletonReportedState {
		t.Errorf("WantSingletonReportedState: got %v, want %v", rt.WantSingletonReportedState, orig.WantSingletonReportedState)
	}
	if rt.WantMultiWECReportedState != orig.WantMultiWECReportedState {
		t.Errorf("WantMultiWECReportedState: got %v, want %v", rt.WantMultiWECReportedState, orig.WantMultiWECReportedState)
	}
	if got, want := sets.New(rt.StatusCollectors...), sets.New(orig.StatusCollectors...); !got.Equal(want) {
		t.Errorf("StatusCollectors: got %v, want %v", got, want)
	}
}

func TestErrorIsBindingPolicyResolutionNotFound(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "has the expected prefix",
			err:  fmt.Errorf("%s - bindingpolicy-key: my-policy", bindingPolicyResolutionNotFoundErrorPrefix),
			want: true,
		},
		{
			name: "unrelated error",
			err:  fmt.Errorf("some other error"),
			want: false,
		},
		{
			name: "partial prefix text does not match",
			err:  fmt.Errorf("bindingpolicy resolution"),
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := errorIsBindingPolicyResolutionNotFound(tc.err)
			if got != tc.want {
				t.Errorf("errorIsBindingPolicyResolutionNotFound() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestBindingPolicyResolverLifecycle(t *testing.T) {
	resolver := NewBindingPolicyResolver()
	policy := &v1alpha1.BindingPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "test-policy", UID: "uid-123"},
	}

	if resolver.ResolutionExists(policy.Name) {
		t.Fatal("resolution should not exist before NoteBindingPolicy")
	}

	resolver.NoteBindingPolicy(policy)
	if !resolver.ResolutionExists(policy.Name) {
		t.Fatal("resolution should exist after NoteBindingPolicy")
	}

	resolver.NoteBindingPolicy(policy)
	if !resolver.ResolutionExists(policy.Name) {
		t.Fatal("resolution should still exist after second NoteBindingPolicy")
	}

	dests := sets.New("cluster1", "cluster2")
	if err := resolver.SetDestinations(policy.Name, dests); err != nil {
		t.Fatalf("SetDestinations: %v", err)
	}

	spec := resolver.GenerateBinding(policy.Name)
	if spec == nil {
		t.Fatal("GenerateBinding returned nil")
	}
	gotDests := sets.New[string]()
	for _, d := range spec.Destinations {
		gotDests.Insert(d.ClusterId)
	}
	if !gotDests.Equal(dests) {
		t.Errorf("destinations: got %v, want %v", gotDests, dests)
	}

	resolver.DeleteResolution(policy.Name)
	if resolver.ResolutionExists(policy.Name) {
		t.Fatal("resolution should be gone after DeleteResolution")
	}
	if resolver.GenerateBinding(policy.Name) != nil {
		t.Error("GenerateBinding should return nil after DeleteResolution")
	}
}

func TestBindingPolicyResolverSetDestinationsOnMissingKey(t *testing.T) {
	resolver := NewBindingPolicyResolver()
	err := resolver.SetDestinations("nonexistent", sets.New("c1"))
	if err == nil {
		t.Fatal("expected error for missing resolution, got nil")
	}
	if !errorIsBindingPolicyResolutionNotFound(err) {
		t.Errorf("expected resolution-not-found error, got: %v", err)
	}
}

func TestBindingPolicyResolverEnsureAndRemoveObject(t *testing.T) {
	resolver := NewBindingPolicyResolver()
	policy := &v1alpha1.BindingPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "pol", UID: "uid-pol"},
	}
	resolver.NoteBindingPolicy(policy)

	objID := makeObjID("default", "nginx")
	mod := ZeroDownsyncModulation()

	changed, err := resolver.EnsureObjectData(policy.Name, objID, "uid-obj", "rv1", mod)
	if err != nil {
		t.Fatalf("EnsureObjectData: %v", err)
	}
	if !changed {
		t.Error("expected resolution to be changed on first EnsureObjectData")
	}

	changed, err = resolver.EnsureObjectData(policy.Name, objID, "uid-obj", "rv1", mod)
	if err != nil {
		t.Fatalf("EnsureObjectData (no-op): %v", err)
	}
	if changed {
		t.Error("expected no change on identical EnsureObjectData")
	}

	if removed := resolver.RemoveObjectIdentifier(policy.Name, objID); !removed {
		t.Error("expected RemoveObjectIdentifier to report a change")
	}
	if removed := resolver.RemoveObjectIdentifier(policy.Name, objID); removed {
		t.Error("expected RemoveObjectIdentifier to report no change when object is absent")
	}
}

func makeObjID(ns, name string) util.ObjectIdentifier {
	return util.ObjectIdentifier{
		Resource:   "pods",
		ObjectName: cache.ObjectName{Namespace: ns, Name: name},
	}
}
