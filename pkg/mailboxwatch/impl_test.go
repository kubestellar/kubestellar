/*
Copyright 2023 The KubeStellar Authors.

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
	"strconv"
	"sync"
	"testing"
	"time"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	machruntime "k8s.io/apimachinery/pkg/runtime"
	machschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	upstreamcache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	tenancyv1a1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	kcpfake "github.com/kcp-dev/kcp/pkg/client/clientset/versioned/cluster/fake"
	kcpinformers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions"
	"github.com/kcp-dev/logicalcluster/v3"

	edgeapi "github.com/kubestellar/kubestellar/pkg/apis/edge/v1alpha1"
	edgefakeclient "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned/cluster/fake"
	edgeclusterclient "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned/cluster/typed/edge/v1alpha1"
	edgescopedclient "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned/typed/edge/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/placement"
)

var _ edgeapi.SyncerConfig
var _ edgescopedclient.SyncerConfigInterface
var _ edgeclusterclient.SyncerConfigClusterInterface

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

func (tot *testObjectTracker) objectsEqualCond(gvk machschema.GroupVersionKind, byName map[objectName]mrObject) func(context.Context) (bool, error) {
	return func(context.Context) (bool, error) {
		tot.Lock()
		defer tot.Unlock()
		have := tot.objectsByGVK[gvk]
		return apiequality.Semantic.DeepEqual(have, byName), nil
	}
}

func TestMailboxInformer(t *testing.T) {
	kind := "SyncerConfig"
	espwParentClusterS := "root"
	espwCluster := logicalcluster.Name("espw")
	scGV := edgeapi.SchemeGroupVersion
	scGVK := scGV.WithKind(kind)
	sclGVK := scGV.WithKind(kind + "List")
	wsGV := tenancyv1a1.SchemeGroupVersion
	wsGVK := wsGV.WithKind("Workspace")
	var nextRV int64 = 13
	nextRVS := func() string {
		nextRV++
		return strconv.FormatInt(nextRV, 10)
	}
	for super := 1; super <= 3; super++ {
		ctx, stop := context.WithCancel(context.Background())
		actual := NewTestObjectTracker()
		expectedWorkspaces := map[objectName]mrObject{}
		espwParentClusterN := logicalcluster.Name(espwParentClusterS)
		espwObjName := objectName{cluster: espwParentClusterN, name: "espw"}
		espwW := &tenancyv1a1.Workspace{
			TypeMeta: metav1.TypeMeta{
				Kind:       wsGVK.Kind,
				APIVersion: wsGV.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:            espwObjName.name,
				Annotations:     map[string]string{logicalcluster.AnnotationKey: espwParentClusterS},
				ResourceVersion: nextRVS(),
			},
			Spec: tenancyv1a1.WorkspaceSpec{
				Cluster: "espwclu",
			}}
		expectedWorkspaces[espwObjName] = espwW
		kcpClientset := kcpfake.NewSimpleClientset()
		kcpClusterInformerFactory := kcpinformers.NewSharedInformerFactory(kcpClientset, 0)
		wsPreInformer := kcpClusterInformerFactory.Tenancy().V1alpha1().Workspaces()
		wsPreInformer.Informer().AddEventHandler(actual)
		kcpClusterInformerFactory.Start(ctx.Done())
		wsesOK := actual.objectsEqualCond(wsGVK, expectedWorkspaces)
		// The fake clientset (kcpfake.NewSimpleClientset) does not implement WATCH properly;
		// regardless of arguments, it notifies the client of changes that come in after the
		// start of the WATCH and nothing before.
		// To cope, try an add-sleep-check sequence with progressively longer sleeps until
		// one succeeds or exhaustion.
		setupWSOK := backoffPoll(t, ctx, "create ESPW", "delete failed ESPW",
			func(ctx context.Context) error {
				_, err := kcpClientset.TenancyV1alpha1().Cluster(espwParentClusterN.Path()).Workspaces().Create(ctx, espwW, metav1.CreateOptions{FieldManager: "test"})
				return err
			},
			func(ctx context.Context) error {
				return kcpClientset.TenancyV1alpha1().Cluster(espwParentClusterN.Path()).Workspaces().Delete(ctx, espwW.Name, metav1.DeleteOptions{})
			},
			wsesOK, time.Second, 4)
		if !setupWSOK {
			t.Fatal("ESPW startup failed")
			return
		}
		t.Logf("ESPW startup succeeded")

		syncerConfigs := map[objectName]mrObject{}
		scsOK := actual.objectsEqualCond(scGVK, syncerConfigs)
		dummySCON := objectName{cluster: "foo", name: "bar"}
		dummySC := &edgeapi.SyncerConfig{
			TypeMeta: metav1.TypeMeta{
				Kind:       kind,
				APIVersion: scGV.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: dummySCON.name,
				Annotations: map[string]string{
					"the-mbws-name":              "dummymbws",
					logicalcluster.AnnotationKey: "dummycluster",
				},
				ResourceVersion: nextRVS(),
			},
		}
		t.Logf("Dummy SyncerConfig=%v", dummySC)
		// syncerConfigs[dummySCON] = dummySC
		edgeClientset := edgefakeclient.NewSimpleClientset( /*dummySC*/ )
		scInformer := NewSharedInformer[edgescopedclient.SyncerConfigInterface, *edgeapi.SyncerConfigList](ctx, sclGVK, wsPreInformer.Cluster(espwCluster),
			edgeClientset.EdgeV1alpha1().SyncerConfigs(), &edgeapi.SyncerConfig{}, 0, upstreamcache.Indexers{})
		scInformer.AddEventHandler(actual)
		go func() {
			doneCh := ctx.Done()
			scInformer.Run(doneCh)
		}()

		// TODO: create a fake clientset for SyncerConfig that has enough fidelity in LIST and WATCH
		// to make a reasonable test possible.
		time.Sleep(5 * time.Second)

		for iteration := 1; iteration <= 64; iteration++ {
			if len(syncerConfigs) > 0 && rand.Intn(2) == 0 {
				gonerIndex := rand.Intn(len(syncerConfigs))
				_, gonerObj := MapRemove(syncerConfigs, gonerIndex)
				gonerSC := gonerObj.(*edgeapi.SyncerConfig)
				cluster := logicalcluster.From(gonerSC)
				//mbwsName := gonerSC.Annotations["the-mbws-name"]
				//wsObjName := objectName{cluster: espwCluster, name: mbwsName}
				//delete(expectedWorkspaces, wsObjName)
				err := edgeClientset.Cluster(cluster.Path()).EdgeV1alpha1().SyncerConfigs().Delete(ctx, gonerSC.Name, metav1.DeleteOptions{})
				if err != nil {
					t.Fatalf("Failed to delete goner %+v: %v", gonerSC, err)
				} else {
					t.Logf("At super=%v, iteration=%v: removed from tracker: SyncerConfig %+v", super, iteration, gonerSC)
				}
			} else {
				invWSNum := rand.Intn(int(math.Sqrt(float64(iteration))))
				invWSClusterS := fmt.Sprintf("ic%d", invWSNum)
				syncTargetNum := rand.Intn(int(math.Sqrt(float64(iteration))))
				syncTargetUID := fmt.Sprintf("beef-%d", syncTargetNum)
				mbwsName := invWSClusterS + placement.WSNameSep + syncTargetUID
				mbwsNum := invWSNum*100 + syncTargetNum
				mbwsClusterS := fmt.Sprintf("mc%d", mbwsNum)
				mbwsClusterN := logicalcluster.Name(mbwsClusterS)
				wsObjName := objectName{cluster: espwCluster, name: mbwsName}
				workspace := expectedWorkspaces[wsObjName]
				if workspace == nil {
					workspaceW := &tenancyv1a1.Workspace{
						TypeMeta: metav1.TypeMeta{
							Kind:       wsGVK.Kind,
							APIVersion: wsGV.String(),
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:            mbwsName,
							Annotations:     map[string]string{logicalcluster.AnnotationKey: espwCluster.String()},
							ResourceVersion: nextRVS(),
						},
						Spec: tenancyv1a1.WorkspaceSpec{
							Cluster: mbwsClusterS,
						}}
					workspace = workspaceW
					var err error
					_, err = kcpClientset.TenancyV1alpha1().Cluster(espwCluster.Path()).Workspaces().Create(ctx, workspaceW, metav1.CreateOptions{FieldManager: "test"})
					if err != nil {
						t.Fatalf("Failed to add workspace %+v: %v", workspace, err)
					} else {
						t.Logf("Added to tracker: Workspace named %q", workspace.GetName())
					}
					expectedWorkspaces[wsObjName] = workspace
				}
				var objName objectName
				var obj *edgeapi.SyncerConfig
				for {
					objName = objectName{
						cluster: mbwsClusterN,
						name:    fmt.Sprintf("rs%d", rand.Intn(10000)),
					}
					if _, has := syncerConfigs[objName]; !has {
						break
					}
				}
				obj = &edgeapi.SyncerConfig{
					TypeMeta: metav1.TypeMeta{
						Kind:       kind,
						APIVersion: scGV.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: objName.namespace,
						Name:      objName.name,
						Annotations: map[string]string{
							"the-mbws-name":              mbwsName,
							logicalcluster.AnnotationKey: mbwsClusterS,
						},
						ResourceVersion: nextRVS(),
					},
				}
				syncerConfigs[objName] = obj
				_, err := edgeClientset.Cluster(mbwsClusterN.Path()).EdgeV1alpha1().SyncerConfigs().Create(ctx, obj, metav1.CreateOptions{FieldManager: "test"})
				if err != nil {
					t.Fatalf("Failed to add %+v to testing tracker: %v", obj, err)
				} else {
					t.Logf("At super=%v, iteration=%v: added to tracker: SyncerConfig named %#v", super, iteration, objName)
				}
			}
			if wait.PollWithContext(ctx, time.Second, 5*time.Second, wsesOK) != nil {
				t.Fatalf("Workspaces did not settle in time: super=%v, iteration=%v, %+v != %+v", super, iteration, actual.getObjects(wsGVK), expectedWorkspaces)
			}
			if wait.PollWithContext(ctx, time.Second, 5*time.Second, scsOK) != nil {
				t.Fatalf("SyncerConfigs did not settle in time: super=%v, iteration=%v, %+v != %+v", super, iteration, actual.getObjects(scGVK), syncerConfigs)
			}
		}
		stop()
	}
}

func backoffPoll(t *testing.T, ctx context.Context, setupTask, teardownTask string, setup, teardown func(context.Context) error, cond wait.ConditionWithContextFunc, pause1 time.Duration, tries int) bool {
	pause := pause1
	for try := 1; try <= tries; try++ {
		time.Sleep(pause)
		err := setup(ctx)
		if err != nil {
			t.Fatalf("Failed to %s on try %v: %v", setupTask, try, err)
			return false
		}
		time.Sleep(pause * 3)
		ok, err := cond(ctx)
		if ok && err == nil {
			return true
		}
		err = teardown(ctx)
		if err != nil {
			t.Fatalf("Failed to %s on try %v: %v", teardownTask, try, err)
			return false
		}
		pause = pause * 2
	}
	return false
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
