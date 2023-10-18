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

package msc

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	edgeapi "github.com/kubestellar/kubestellar/pkg/apis/edge/v2alpha1"
	edgeclientset "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned"
	edgeinformers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions"
	spacemsc "github.com/kubestellar/kubestellar/space-framework/pkg/msclientlib"
)

type EdgeClientGen = MultiSpaceInformerGen[*edgeclientset.Clientset, edgeclientset.Interface, edgeinformers.SharedInformerOption, edgeinformers.SharedScopedInformerFactory]

func NewEdgeClientGen(ksi spacemsc.KubestellarSpaceInterface) EdgeClientGen {
	return NewMSC(ksi, edgeclientset.NewForConfig, edgeinformers.NewSharedScopedInformerFactoryWithOptions)
}

func demoEdge(ksi spacemsc.KubestellarSpaceInterface) {
	gen := NewEdgeClientGen(ksi)
	client, _ := gen.NewForSpace("fred", "")
	var ep edgeapi.EdgePlacement
	client.EdgeV2alpha1().EdgePlacements().Create(context.Background(), &ep, metav1.CreateOptions{})
	infact := gen.NewInformerFactoryWithOptions(client, 0)
	infact.Edge().V2alpha1().EdgePlacements().Lister()
}
