/*
Copyright 2022 The KubeStellar Authors.

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

import edgev2alpha1 "github.com/kubestellar/kubestellar/pkg/apis/edge/v2alpha1"

type SyncerInterface interface {
	ReInitializeClients(resources []edgev2alpha1.EdgeSyncConfigResource, conversions []edgev2alpha1.EdgeSynConversion) error
	SyncOne(resource edgev2alpha1.EdgeSyncConfigResource, conversions []edgev2alpha1.EdgeSynConversion) error
	BackStatusOne(resource edgev2alpha1.EdgeSyncConfigResource, conversions []edgev2alpha1.EdgeSynConversion) error
	SyncMany(resource edgev2alpha1.EdgeSyncConfigResource, conversions []edgev2alpha1.EdgeSynConversion) error
}
