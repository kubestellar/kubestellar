/*
Copyright 2023 The KCP Authors.

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

package mailboxwatch

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	k8sapps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	machruntime "k8s.io/apimachinery/pkg/runtime"
	machschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	upstreamcache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	kcpk8sfake "github.com/kcp-dev/client-go/kubernetes/fake"
	tenancyv1a1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	kcpfake "github.com/kcp-dev/kcp/pkg/client/clientset/versioned/cluster/fake"
	kcpinformers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions"
	"github.com/kcp-dev/logicalcluster/v3"
)

func TestMain(m *testing.M) {
	klog.InitFlags(nil)
	os.Exit(m.Run())
}

type testObjectTracker struct {
	sync.Mutex
	objectsByGVK map[machschema.GroupVersionKind]map[objectName]mrObject
}

type objectName struct {
	cluster   logicalcluster.Name
	namespace string
	name      string
}

type mrObject interface {
	metav1.Object
	machruntime.Object
}

func NewTestObjectTracker() *testObjectTracker {
	return &testObjectTracker{
		objectsByGVK: map[machschema.GroupVersionKind]map[objectName]mrObject{},
	}
}

func (tot *testObjectTracker) OnAdd(objA any) {
	tot.handleObject(objA, true)
}

func (tot *testObjectTracker) OnUpdate(old, objA any) {
	tot.handleObject(objA, true)
}

func (tot *testObjectTracker) OnDelete(objA any) {
	if dfs, is := objA.(upstreamcache.DeletedFinalStateUnknown); is {
		objA = dfs.Obj
	}
	tot.handleObject(objA, false)
}

func (tot *testObjectTracker) handleObject(objA any, present bool) {
	objMR := objA.(mrObject)
	cluster := logicalcluster.From(objMR)
	gvk := objMR.GetObjectKind().GroupVersionKind()
	objName := objectName{cluster: cluster, namespace: objMR.GetNamespace(), name: objMR.GetName()}
	tot.Lock()
	defer tot.Unlock()
	if present {
		objectsByName := tot.objectsByGVK[gvk]
		if objectsByName == nil {
			objectsByName = map[objectName]mrObject{}
			tot.objectsByGVK[gvk] = objectsByName
		}
		objectsByName[objName] = objMR
	} else {
		objectsByName := tot.objectsByGVK[gvk]
		if objectsByName == nil {
			return
		}
		delete(objectsByName, objName)
		if len(objectsByName) == 0 {
			delete(tot.objectsByGVK, gvk)
		}
	}
}

func (tot *testObjectTracker) getObjects(gvk machschema.GroupVersionKind) map[objectName]mrObject {
	tot.Lock()
	defer tot.Unlock()
	return MapCopy(tot.objectsByGVK[gvk])
}

func (tot *testObjectTracker) objectsEqualCond(gvk machschema.GroupVersionKind, byName map[objectName]mrObject) func() (bool, error) {
	return func() (bool, error) {
		tot.Lock()
		defer tot.Unlock()
		return MapEqual(tot.objectsByGVK[gvk], byName), nil
	}
}

func TestMailboxInformer(t *testing.T) {
	resource := "replicasets"
	kind := "ReplicaSet"
	espwCluster := logicalcluster.Name("espw")
	rsGV := k8sapps.SchemeGroupVersion
	rsGVK := rsGV.WithKind(kind)
	rsGVR := rsGV.WithResource(resource)
	wsGV := tenancyv1a1.SchemeGroupVersion
	wsGVK := wsGV.WithKind("Workspace")
	for super := 1; super <= 3; super++ {
		replicaSets := map[objectName]mrObject{}
		workspaces := map[objectName]mrObject{}
		kcpClientset := kcpfake.NewSimpleClientset()
		kcpTracker := kcpClientset.Tracker()
		k8sClientset := kcpk8sfake.NewSimpleClientset()
		k8sTracker := k8sClientset.Tracker()
		ctx, stop := context.WithCancel(context.Background())
		actual := NewTestObjectTracker()
		wsPreInformer := kcpinformers.NewSharedInformerFactory(kcpClientset, 0).Tenancy().V1alpha1().Workspaces()
		go func() {
			wsPreInformer.Informer().Run(ctx.Done())
		}()
		rslw := NewListerWatcher[*k8sapps.ReplicaSetList](ctx, k8sClientset.AppsV1().ReplicaSets())
		rsInformer := NewSharedInformer(wsPreInformer.Cluster(espwCluster), rslw, &k8sapps.ReplicaSet{}, 0, upstreamcache.Indexers{})
		rsInformer.AddEventHandler(actual)
		go rsInformer.Run(ctx.Done())
		for iteration := 1; iteration <= 64; iteration++ {
			if len(replicaSets) > 0 && rand.Intn(2) == 0 {
				gonerIndex := rand.Intn(len(replicaSets))
				_, gonerObj := MapRemove(replicaSets, gonerIndex)
				gonerRS := gonerObj.(*k8sapps.ReplicaSet)
				cluster := logicalcluster.From(gonerRS)
				err := k8sTracker.Cluster(cluster.Path()).Delete(rsGVR, gonerRS.Namespace, gonerRS.Name)
				if err != nil {
					t.Fatalf("Failed to delete goner %+v: %v", gonerRS, err)
				} else {
					t.Logf("Removed from tracker: ReplicaSet %+v", gonerRS)
				}
			} else {
				clusterS := fmt.Sprintf("c%d", rand.Intn(int(math.Sqrt(float64(iteration)))))
				clusterN := logicalcluster.Name(clusterS)
				wsObjName := objectName{cluster: espwCluster, name: clusterS}
				workspace := workspaces[wsObjName]
				if workspace == nil {
					workspace = &tenancyv1a1.Workspace{
						TypeMeta: metav1.TypeMeta{
							Kind:       wsGVK.Kind,
							APIVersion: wsGV.String(),
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:        clusterS,
							Annotations: map[string]string{logicalcluster.AnnotationKey: espwCluster.String()},
						},
						Spec: tenancyv1a1.WorkspaceSpec{
							Cluster: clusterS,
						},
					}
					scopedTracker := kcpTracker.Cluster(espwCluster.Path())
					err := scopedTracker.Add(workspace)
					if err != nil {
						t.Fatalf("Failed to add workspace %+v: %v", workspace, err)
					} else {
						t.Logf("Added to tracker: Workspace named %q", workspace.GetName())
					}
					workspaces[wsObjName] = workspace
				}
				var objName objectName
				var obj *k8sapps.ReplicaSet
				for {
					objName = objectName{
						cluster:   clusterN,
						namespace: fmt.Sprintf("ns%d", rand.Intn(10000)),
						name:      fmt.Sprintf("rs%d", rand.Intn(10000)),
					}
					if _, has := replicaSets[objName]; !has {
						break
					}
				}
				obj = &k8sapps.ReplicaSet{
					TypeMeta: metav1.TypeMeta{
						Kind:       kind,
						APIVersion: rsGV.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: objName.namespace,
						Name:      objName.name,
						Annotations: map[string]string{
							logicalcluster.AnnotationKey: clusterS,
						},
					},
				}
				scopedTracker := k8sTracker.Cluster(clusterN.Path())
				err := scopedTracker.Add(obj)
				if err != nil {
					t.Fatalf("Failed to add %+v to testing tracker: %v", obj, err)
				} else {
					t.Logf("Added to tracker: ReplicaSet named %#v", objName)
				}
			}
			if wait.PollImmediate(time.Second, 5*time.Second, actual.objectsEqualCond(wsGVK, workspaces)) != nil {
				t.Fatalf("Workspaces did not settle in time: %+v != %+v", actual.getObjects(wsGVK), workspaces)
			}
			if wait.PollImmediate(time.Second, 5*time.Second, actual.objectsEqualCond(rsGVK, replicaSets)) != nil {
				t.Fatalf("Workspaces did not settle in time: %+v != %+v", actual.getObjects(rsGVK), replicaSets)
			}
		}
		stop()
	}
}

func MapRemove[Key comparable, Val any](from map[Key]Val, gonerIndex int) (Key, Val) {
	index := 0
	for key, val := range from {
		if index == gonerIndex {
			return key, val
		}
		index++
	}
	panic(from)
}

func MapCopy[Key comparable, Val any](old map[Key]Val) map[Key]Val {
	ans := map[Key]Val{}
	for key, val := range old {
		ans[key] = val
	}
	return ans
}

// MapEqual compares two maps for equality.
// `Val` has no type bound because bounding it by `comparable` does
// not work in go 1.19 (and that is the current version for this module).
// The values must be comparable, even though we can not say that in the type system.
func MapEqual[Key comparable, Val any](map1, map2 map[Key]Val) bool {
	if len(map1) != len(map2) {
		return false
	}
	for key, val1 := range map1 {
		val2, has := map2[key]
		if !has {
			return false
		}
		var v1 interface{} = val1
		var v2 interface{} = val2
		if v1 != v2 {
			return false
		}
	}
	return true
}
