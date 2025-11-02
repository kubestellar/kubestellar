/*
Copyright YEAR The KubeStellar Authors.

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

package status

import (
	"context"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/pkg/util"
)

func (c *Controller) handleMultiWEC(ctx context.Context, wObjID util.ObjectIdentifier, nWECs int) error {
	// TODO: To be implemented
	logger := klog.FromContext(ctx)
	logger.V(4).Info("To be implemented", "object", wObjID, "nWECs", nWECs)

	return nil
}
