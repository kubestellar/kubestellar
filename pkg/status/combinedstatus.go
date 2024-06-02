/*
Copyright 2024 The KubeStellar Authors.

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

import "github.com/kubestellar/kubestellar/pkg/util"

// TODO: implement syncing CombinedStatus objects

func getCombinedStatusIdentifier(bindingName string, objectIdentifier util.ObjectIdentifier) util.ObjectIdentifier {
	// The name of the CombinedStatus object is the concatenation of:
	// - the UID of the workload object
	// - the string ":"
	// - the UID of the BindingPolicy object.
	return util.IdentifierForCombinedStatus("", objectIdentifier.ObjectName.Namespace) // TODO
}
