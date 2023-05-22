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

package placement

import (
	"github.com/kcp-dev/logicalcluster/v3"
)

// ExternalName identifies a cluster-scoped object of some implicit kind
type ExternalName struct {
	// Cluster identifies the cluster.  It is the one-part ID, not a path.
	Cluster logicalcluster.Name

	Name string
}

// NewExternalName assumes the given cluster identifier is proper
func NewExternalName(cluster, name string) ExternalName {
	return ExternalName{Cluster: logicalcluster.Name(cluster), Name: name}
}

func (ExternalName) OfSPLocation(sp SinglePlacement) ExternalName {
	return ExternalName{Cluster: logicalcluster.Name(sp.Cluster), Name: sp.LocationName}
}

func (ExternalName) OfSPTarget(sp SinglePlacement) ExternalName {
	return ExternalName{Cluster: logicalcluster.Name(sp.Cluster), Name: sp.SyncTargetName}
}

func (en ExternalName) String() string {
	return en.Cluster.String() + ":" + en.Name
}
