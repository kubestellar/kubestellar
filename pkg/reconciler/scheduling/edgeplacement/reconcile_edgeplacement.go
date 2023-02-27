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

package edgeplacement

import (
	"context"
	"fmt"

	"github.com/kcp-dev/logicalcluster/v3"

	edgev1alpha1 "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
)

type edgePlacementReconciler struct{}

func (r *edgePlacementReconciler) reconcile(ctx context.Context, ep *edgev1alpha1.EdgePlacement) (reconcileStatus, *edgev1alpha1.EdgePlacement, error) {
	ws := logicalcluster.From(ep)
	fmt.Printf("reconciling EdgePlacement %s in Workspace %s\n", ep.Name, ws.String())
	return reconcileStatusContinue, ep, nil
}
