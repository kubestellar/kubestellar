/*
Copyright 2021 The KCP Authors.

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

package syncers

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/kcp-dev/logicalcluster/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	kcptesting "github.com/kcp-dev/client-go/third_party/k8s.io/client-go/testing"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/util/yaml"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	ddsif "github.com/kcp-dev/kcp/pkg/informer"
	"github.com/kcp-dev/kcp/pkg/syncer/indexers"
)

//go:embed testdata/*.yaml
var testdata embed.FS

var scheme *runtime.Scheme

func init() {
	scheme = runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
}

func TestStatusSyncerProcess2(t *testing.T) {
	tests := map[string]struct {
		downstreamResources []*unstructured.Unstructured
		expectError         bool
		expectActionsOnFrom []clienttesting.Action
		expectActionsOnTo   []kcptesting.Action
	}{
		"StatusSyncer upsert to existing resource": {
			downstreamResources: []*unstructured.Unstructured{
				loadTemplateOrDie("configmap.yaml"),
				loadTemplateOrDie("deployment.yaml"),
				loadTemplateOrDie("crd.yaml"),
				loadTemplateOrDie("cr.yaml"),
			},
			expectActionsOnFrom: []clienttesting.Action{},
			expectActionsOnTo: []kcptesting.Action{
				updateDeploymentAction("test",
					toUnstructured(t, changeDeployment(
						deployment("theDeployment", "test", "root:org:ws", map[string]string{
							"state.workload.kcp.io/6ohB8yeXhwqTQVuBzJRgqcRJTpRjX7yTZu5g5g": "Sync",
						}, nil, nil),
						addDeploymentStatus(appsv1.DeploymentStatus{
							Replicas: 15,
						}))),
					"status"),
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			logger := klog.FromContext(ctx)

			allFromResources := []runtime.Object{}
			for _, r := range tc.downstreamResources {
				allFromResources = append(allFromResources, r)
			}
			downstreamClient := dynamicfake.NewSimpleDynamicClient(scheme, allFromResources...)
			ddsifForDownstream, err := ddsif.NewScopedDiscoveringDynamicSharedInformerFactory(downstreamClient, nil, nil,
				&mockedGVRSource{},
				cache.Indexers{
					indexers.ByNamespaceLocatorIndexName: indexers.IndexByNamespaceLocator,
				},
			)
			require.NoError(t, err)

			controller, err := NewUpsyncer(logger, downstreamClient, ddsifForDownstream)
			require.NoError(t, err)

			ddsifForDownstream.Start(ctx.Done())

			go ddsifForDownstream.StartWorker(ctx)

			// The only GVRs we care about are the 4 listed below
			t.Logf("waiting for upstream and downstream dynamic informer factories to be synced")
			gvrs := sets.NewString(
				schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}.String(),
				schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}.String(),
				schema.GroupVersionResource{Group: "my.domain", Version: "v1alpha1", Resource: "samples"}.String(),
				schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}.String(),
			)
			require.Eventually(t, func() bool {
				syncedDownstream, _ := ddsifForDownstream.Informers()
				foundDownstream := sets.NewString()
				for gvr := range syncedDownstream {
					foundDownstream.Insert(gvr.String())
				}
				return foundDownstream.IsSuperset(gvrs)
			}, wait.ForeverTestTimeout, 100*time.Millisecond)
			t.Logf("upstream and downstream dynamic informer factories are synced")

			// Now that we know the informer factories have the GVRs we care about synced, we need to clear the
			// actions so our expectations will be accurate.
			downstreamClient.ClearActions()

			for _, resource := range tc.downstreamResources {
				key := resource.GetNamespace() + "/" + resource.GetName()
				gvk := schema.FromAPIVersionAndKind(resource.GetAPIVersion(), resource.GetKind())
				err = controller.process(context.Background(),
					schema.GroupVersionResource{
						Group:    gvk.Group,
						Version:  gvk.Version,
						Resource: strings.ToLower(gvk.Kind) + "s", // TODO: find a way to convert GVK to GVR
					},
					key,
				)
			}
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Empty(t, cmp.Diff(tc.expectActionsOnFrom, downstreamClient.Actions()))
		})
	}
}

type mockedGVRSource struct {
}

func (s *mockedGVRSource) GVRs() map[schema.GroupVersionResource]ddsif.GVRPartialMetadata {
	return map[schema.GroupVersionResource]ddsif.GVRPartialMetadata{
		appsv1.SchemeGroupVersion.WithResource("deployments"): {
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Singular: "deployment",
				Kind:     "Deployment",
			},
		},
		{
			Version:  "v1",
			Resource: "configmaps",
		}: {
			Scope: apiextensionsv1.NamespaceScoped,
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Singular: "configmap",
				Kind:     "ConfigMap",
			},
		},
		{
			Group:    "my.domain",
			Version:  "v1alpha1",
			Resource: "samples",
		}: {
			Scope: apiextensionsv1.ClusterScoped,
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Singular: "sample",
				Kind:     "Sample",
			},
		},
		apiextensionsv1.SchemeGroupVersion.WithResource("customresourcedefinitions"): {
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Singular: "customresroucedefinition",
				Kind:     "CustomResourceDefinition",
			},
		},
	}
}

func (s *mockedGVRSource) Ready() bool {
	return true
}

func (s *mockedGVRSource) Subscribe() <-chan struct{} {
	return make(<-chan struct{})
}

func deployment(name, namespace, clusterName string, labels, annotations map[string]string, finalizers []string) *appsv1.Deployment {
	if clusterName != "" {
		if annotations == nil {
			annotations = make(map[string]string)
		}
		annotations[logicalcluster.AnnotationKey] = clusterName
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
			Finalizers:  finalizers,
		},
	}
}

type deploymentChange func(*appsv1.Deployment)

func changeDeployment(in *appsv1.Deployment, changes ...deploymentChange) *appsv1.Deployment {
	for _, change := range changes {
		change(in)
	}
	return in
}

func addDeploymentStatus(status appsv1.DeploymentStatus) deploymentChange {
	return func(d *appsv1.Deployment) {
		d.Status = status
	}
}

func toUnstructured(t require.TestingT, obj metav1.Object) *unstructured.Unstructured {
	var result unstructured.Unstructured
	err := scheme.Convert(obj, &result, nil)
	require.NoError(t, err)

	return &result
}

func deploymentAction(verb, namespace string, subresources ...string) kcptesting.ActionImpl {
	return kcptesting.ActionImpl{
		Namespace:   namespace,
		ClusterPath: logicalcluster.NewPath("root:org:ws"),
		Verb:        verb,
		Resource:    schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
		Subresource: strings.Join(subresources, "/"),
	}
}

func updateDeploymentAction(namespace string, object runtime.Object, subresources ...string) kcptesting.UpdateActionImpl {
	return kcptesting.UpdateActionImpl{
		ActionImpl: deploymentAction("update", namespace, subresources...),
		Object:     object,
	}
}

func loadTemplateOrDie(filename string) *unstructured.Unstructured {
	raw, err := testdata.ReadFile("testdata/" + filename)
	if err != nil {
		panic(fmt.Sprintf("failed to read file: %v", err))
	}
	decoder := yaml.NewYAMLToJSONDecoder(bytes.NewReader(raw))

	var u unstructured.Unstructured
	err = decoder.Decode(&u)
	if err != nil {
		panic(fmt.Sprintf("failed to decode file: %v", err))
	}
	return &u
}
