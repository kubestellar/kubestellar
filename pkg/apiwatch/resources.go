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

package apiwatch

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	upstreamdiscovery "k8s.io/client-go/discovery"
	cachediscovery "k8s.io/client-go/discovery/cached/memory"
	upstreamcache "k8s.io/client-go/tools/cache"
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
	"k8s.io/klog/v2"

	urmetav1a1 "github.com/kcp-dev/edge-mc/pkg/apis/meta/v1alpha1"
)

// Invalidatable is a cache that has to be explicitly invalidated
type Invalidatable interface {
	// Invalidate the cache
	Invalidate()
}

// ObjectNotifier is something that notifies the client like an informer does
type ObjectNotifier interface {
	AddEventHandler(handler upstreamcache.ResourceEventHandler)
}

// NewAPIResourceInformer creates an informer on the API resources
// revealed by the given client.  The objects delivered by the
// informer are of type `*urmetav1a1.APIResource`.
//
// The results from the given client are cached in memory and that
// cache has to be explicitly invalidated.  Invalidation can be done
// by calling the returned Invalidator.  Additionally, invalidation
// happens whenever any of the supplied invalidationNotifiers delivers
// a notification of an object addition.  Re-querying the given client
// is delayed by a few decaseconds (with Nagling) to support
// invalidations based on events that merely trigger some process of
// changing the set of API resources.
func NewAPIResourceInformer(ctx context.Context, clusterName string, client upstreamdiscovery.DiscoveryInterface, invalidationNotifiers ...ObjectNotifier) (upstreamcache.SharedInformer, Invalidatable) {
	logger := klog.FromContext(ctx).WithValues("cluster", clusterName)
	ctx = klog.NewContext(ctx, logger)
	rlw := &resourcesListWatcher{
		ctx:              ctx,
		logger:           logger,
		clusterName:      clusterName,
		client:           client,
		cache:            cachediscovery.NewMemCacheClient(client),
		resourceVersionI: 1,
	}
	rlw.cond = sync.NewCond(&rlw.mutex)
	go func() {
		doneCh := ctx.Done()
		for {
			select {
			case <-doneCh:
				return
			default:
			}
			var wait time.Duration
			func() {
				rlw.mutex.Lock()
				defer rlw.mutex.Unlock()
				if rlw.needRelist {
					now := time.Now()
					if now.Before(rlw.relistAfter) {
						wait = rlw.relistAfter.Sub(now)
					} else {
						logger.V(3).Info("Cycled APIResourceInformer")
						for _, cancel := range rlw.cancels {
							cancel()
						}
						rlw.needRelist = false
					}
					return
				}
				rlw.cond.Wait()
			}()
			if wait > 0 {
				time.Sleep(wait)
			}
		}
	}()
	for _, invalidator := range invalidationNotifiers {
		invalidator.AddEventHandler(upstreamcache.ResourceEventHandlerFuncs{
			AddFunc: func(obj any) {
				logger.V(3).Info("Notified of invalidator", "obj", obj)
				rlw.Invalidate()
			},
		})
	}
	return upstreamcache.NewSharedInformer(rlw, &urmetav1a1.APIResource{}, 0), rlw
}

type resourcesListWatcher struct {
	ctx         context.Context
	logger      klog.Logger
	clusterName string
	client      upstreamdiscovery.DiscoveryInterface
	cache       upstreamdiscovery.CachedDiscoveryInterface

	mutex            sync.Mutex
	cond             *sync.Cond
	resourceVersionI int64
	needRelist       bool
	relistAfter      time.Time
	cancels          []context.CancelFunc
}

func (rlw *resourcesListWatcher) Invalidate() {
	rlw.mutex.Lock()
	defer rlw.mutex.Unlock()
	rlw.resourceVersionI += 1
	rlw.relistAfter = time.Now().Add(time.Second * 20)
	rlw.needRelist = true
	rlw.cache.Invalidate()
	rlw.cond.Broadcast()
}

type resourceWatch struct {
	*resourcesListWatcher
	cancel  context.CancelFunc
	results chan watch.Event
}

func (rw *resourceWatch) ResultChan() <-chan watch.Event {
	return rw.results
}

func (rw *resourceWatch) Stop() {
	rw.cancel()
}

func (rlw *resourcesListWatcher) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	rlw.mutex.Lock()
	defer rlw.mutex.Unlock()
	resourceVersionS := strconv.FormatInt(rlw.resourceVersionI, 10)
	if resourceVersionS != opts.ResourceVersion {
		return nil, apierrors.NewResourceExpired(fmt.Sprintf("Requested version %s, have version %s in cluster %s", opts.ResourceVersion, resourceVersionS, rlw.clusterName))
	}
	timeout := time.Duration(*opts.TimeoutSeconds) * time.Second
	ctx, cancel := context.WithTimeout(rlw.ctx, timeout)
	rw := &resourceWatch{
		resourcesListWatcher: rlw,
		cancel:               cancel,
		results:              make(chan watch.Event),
	}
	rlw.cancels = append(rlw.cancels, cancel)
	go func() {
		<-ctx.Done()
		rlw.logger.V(3).Info("Ending an APIResource Watch")
		close(rw.results)
	}()
	return rw, nil
}

func (rlw *resourcesListWatcher) List(opts metav1.ListOptions) (runtime.Object, error) {
	resourceVersionI := func() int64 {
		rlw.mutex.Lock()
		defer rlw.mutex.Unlock()
		rlw.resourceVersionI = rlw.resourceVersionI + 1
		for _, cancel := range rlw.cancels {
			cancel()
		}
		return rlw.resourceVersionI
	}()
	resourceVersionS := strconv.FormatInt(resourceVersionI, 10)
	groupList, err := rlw.cache.ServerPreferredResources()
	if err != nil {
		return nil, err
	}
	ans := urmetav1a1.APIResourceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "APIResourceList",
			APIVersion: urmetav1a1.SchemeGroupVersion.String(),
		},
		ListMeta: metav1.ListMeta{ResourceVersion: resourceVersionS},
		Items:    []urmetav1a1.APIResource{},
	}
	for _, group := range groupList {
		gv, err := schema.ParseGroupVersion(group.GroupVersion)
		if err != nil {
			rlw.logger.Error(err, "Failed to parse a GroupVersion", "groupVersion", group.GroupVersion)
			continue
		}
		for _, rsc := range group.APIResources {
			ans.Items = append(ans.Items, urmetav1a1.APIResource{
				TypeMeta: metav1.TypeMeta{
					Kind:       "APIResource",
					APIVersion: urmetav1a1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:            group.GroupVersion + ":" + rsc.Name,
					ResourceVersion: resourceVersionS,
				},
				Spec: urmetav1a1.APIResourceSpec{
					Name:         rsc.Name,
					SingularName: rsc.SingularName,
					Namespaced:   rsc.Namespaced,
					Group:        gv.Group,
					Version:      gv.Version,
					Kind:         rsc.Kind,
					Verbs:        rsc.Verbs,
				},
			})
		}
	}
	return &ans, nil
}
