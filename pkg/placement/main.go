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

package placement

import (
	"context"

	k8scache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

func NewPlacementTranslator(
	ctx context.Context,
	mailboxWorkspaceInformer k8scache.SharedInformer,
	// TODO: add the other needed arguments
) {
	logger := klog.FromContext(ctx)
	_, mbPathToName := NewNameAndPath(logger, mailboxWorkspaceInformer, true)
	// TODO: make all the other needed infrastructure

	// TODO: replace all these dummies
	whatResolver := RelayWhatResolver()
	whereResolver := RelayWhereResolver()
	setBinder := NewDummySetBinder()
	workloadProjector := NewDummyWorkloadProjector(mbPathToName)
	placementProjector := NewDummyWorkloadProjector(mbPathToName)
	AssemplePlacementTranslator(whatResolver, whereResolver, setBinder, workloadProjector, placementProjector)
}
