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

package outline

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	appsclient "k8s.io/client-go/kubernetes/typed/apps/v1"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	corelisters "k8s.io/client-go/listers/core/v1"

	ksapi "github.com/kubestellar/kubestellar/api/control/v1alpha1"
	ksclient "github.com/kubestellar/kubestellar/pkg/generated/clientset/versioned/typed/control/v1alpha1"
	ksinformers "github.com/kubestellar/kubestellar/pkg/generated/informers/externalversions/control/v1alpha1"
	kslisters "github.com/kubestellar/kubestellar/pkg/generated/listers/control/v1alpha1"
)

// I1Type is an example type of object that is input to the controller.
// For this example we use ReplicaSet.
type I1Type = appsv1.ReplicaSet
type I1List = appsv1.ReplicaSetList

var i1GVR = appsv1.SchemeGroupVersion.WithResource("replicasets")

type I1ClientInterface appsclient.ReplicaSetsGetter

// I1PreInformer is something that can produce both an I1 informer and an I1Lister.
type I1PreInformer = appsinformers.ReplicaSetInformer

type I1Lister = appslisters.ReplicaSetLister

// INType is another example type of object that is input to the controller.
// For this example we use BindingPolicy.
// We deliberately use one example that is namespaced and one that is not, to
// illustrate the pattern for each.
type INType = ksapi.BindingPolicy
type INList = ksapi.BindingPolicyList

var iNGVR = ksapi.SchemeGroupVersion.WithResource("bindingpolicies")

type INClientInterface ksclient.BindingPolicyInterface

// INPreInformer is something that can produce both an IN informer and an INLister.
type INPreInformer = ksinformers.BindingPolicyInformer

type INLister = kslisters.BindingPolicyLister

// O1Type is an example type of object that gets controller output.
// For this example we use Pod.
type O1Type = corev1.Pod
type O1List = corev1.PodList

var o1GVR = corev1.SchemeGroupVersion.WithResource("pods")

type O1ClientInterface = coreclient.PodsGetter

// O1PreInformer is something that can produce both an O1 informer and an O1Lister.
type O1PreInformer = coreinformers.PodInformer

type O1Lister = corelisters.PodLister

// OMType is an example type of object that gets controller output.
// For this example we use Binding.
// We deliberately use one example that is namespaced and one that is not, to
// illustrate the pattern for each.
type OMType = ksapi.Binding
type OMList = ksapi.BindingList

var oMGVR = ksapi.SchemeGroupVersion.WithResource("bindings")

type OMClientInterface = ksclient.BindingInterface

// OMPreInformer is something that can produce both an OM informer and an OMLister.
type OMPreInformer = ksinformers.BindingInformer

type OMLister = kslisters.BindingLister
