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
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	restclient "k8s.io/client-go/rest"
	upstreamcache "k8s.io/client-go/tools/cache"

	kcpcache "github.com/kcp-dev/apimachinery/v2/pkg/cache"
	kcpinformers "github.com/kcp-dev/apimachinery/v2/third_party/informers"
	kcpscopedclientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned"
	kcpapiinformers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions"
	tenancyv1a1informers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions/tenancy/v1alpha1"
)

// NewSharedInformer constructs a shared index informer on objects of a given kind in mailbox workspaces.
// It follows the usual pattern for a constructor of informers except for an additional parameter
// at the front, which it uses to keep appraised of the mailbox workspaces.
// The ListerWatcher can be constructed using `NewListerWatcher`, for example
// as is done in `impl_test.go`.
func NewSharedInformer(
	mailboxWorkspacePreInformer tenancyv1a1informers.WorkspaceInformer,
	listerWatcher upstreamcache.ListerWatcher,
	exampleObject runtime.Object,
	defaultEventHandlerResyncPeriod time.Duration,
	indexers upstreamcache.Indexers,
) kcpcache.ScopeableSharedIndexInformer {
	// Implementation outline:
	// Use the informer on mailbox workspaces to stay appraised of the logicalcluster.Name for
	// each mailbox workspace qua logical cluster.
	// Wrap the given ListerWatcher with filtering that passes only objects in one of those logical clusters.
	// At first stage of development, do not take special care when mailbox workspaces arrive or depart,
	// because objects in them will likely arrive later and depart sooner.
	// At a later stage of development, do not rely on happy timing.
	wrappedListerWatcher := newFilteredListerWatcher(mailboxWorkspacePreInformer, listerWatcher)
	return kcpinformers.NewSharedIndexInformer(wrappedListerWatcher, exampleObject, defaultEventHandlerResyncPeriod, indexers)
}

// NewSharedInformerForEdgeConfig is like NewSharedInformer but takes a REST Config for the
// edge service provider workspace and constructs the pre-informer for the mailbox workspaces.
// That pre-informer is then used to call NewSharedInformer and is also returned,
// so that the caller can wait on HasSynced of the mailbox workspace informer.
func NewSharedInformerForEdgeConfig(
	edgeServiceProviderWorkspaceClientConfig *restclient.Config,
	listerWatcher upstreamcache.ListerWatcher,
	exampleObject runtime.Object,
	defaultEventHandlerResyncPeriod time.Duration,
	indexers upstreamcache.Indexers,
) (tenancyv1a1informers.WorkspaceInformer, kcpcache.ScopeableSharedIndexInformer, error) {
	workspaceScopedClientset, err := kcpscopedclientset.NewForConfig(edgeServiceProviderWorkspaceClientConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create clientset for mailbox workspaces: %w", err)
	}
	workspaceScopedInformerFactory := kcpapiinformers.NewSharedScopedInformerFactoryWithOptions(workspaceScopedClientset, 0)
	workspaceScopedPreInformer := workspaceScopedInformerFactory.Tenancy().V1alpha1().Workspaces()
	return workspaceScopedPreInformer, NewSharedInformer(workspaceScopedPreInformer, listerWatcher, exampleObject, defaultEventHandlerResyncPeriod, indexers), nil
}
