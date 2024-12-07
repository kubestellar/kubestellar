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
	"context"
	"fmt"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	k8smetrics "k8s.io/component-base/metrics"
	"k8s.io/klog/v2"

	ksmetrics "github.com/kubestellar/kubestellar/pkg/metrics"
	"github.com/kubestellar/kubestellar/pkg/util"
)

// This is an outline of a Kubernetes controller.
// In general, a controller takes its input from the state of objects of N various types I1, I2, ... IN
// and the controller's main job is to maintain some output in objects of M various types O1, O2, ... OM.
// The controller may also report into the status section of the I1, I2, ... IN objects.

// This is informed by https://github.com/kubernetes/sample-controller/blob/master/controller.go
// and https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/controllers.md
// and goes further by explicitly considering multiple types of input and multiple types of output
// and being somewhat opinionated.
// The primary opinions are as follows.
// - Because mutiple input objects affect one output object, the work on these objects is factored.
//   While the upstream controller outline does all the work when syncing a single input object,
//   this controller outline splits the processing into two phases: (1) syncing input objects,
//   which updates some internal data structures (which support the next phase), and (2) syncing
//   output objects.
// - Read API object state from the informer local caches (i.e., listers)
//   rather than querying the apiserver; the latter is more expensive and should
//   only be done when required for proper interlocking with other clients and/or concerns.
// - Use ksmetrics to collect metrics on the API objects and operations on them.
// - Some opinions on how various logger.V levels should be used.

// This exemplifies a single controller,
// meant to be one of several combined together in one "controller manager".

// controllerName distinguishes this controller from others in the system.
// Use it as the FieldManager value in the XXXOpts arguments to the calls on the apiservers.
// Also get it into the logger.
const controllerName = "example-controller"

// NewExampleController constructs a new controller.
// Call this before starting the informers.
func NewExampleController(
	ctx context.Context,
	clientMetrics ksmetrics.ClientMetrics,
	i1Client I1ClientInterface,
	i1PreInformer I1PreInformer,
	// and so on, for I2, ...
	iNClient INClientInterface,
	iNPreInformer INPreInformer,
	o1Client O1ClientInterface,
	o1PreInformer O1PreInformer,
	// and so on, for O2, ...
	oMClient OMClientInterface,
	oMPreInformer OMPreInformer,
) *controller {
	logger := klog.FromContext(ctx)

	measuredI1Client := ksmetrics.NewWrappedNamespacedClient[*I1Type, *I1List](clientMetrics, i1GVR,
		func(namespace string) ksmetrics.ClientModNamespace[*I1Type, *I1List] {
			return i1Client.ReplicaSets(namespace)
		})
	measuredINClient := ksmetrics.NewWrappedClusterScopedClient[*INType, *INList](clientMetrics, iNGVR, iNClient)

	measuredO1Client := ksmetrics.NewWrappedNamespacedClient[*O1Type, *O1List](clientMetrics, o1GVR,
		func(namespace string) ksmetrics.ClientModNamespace[*O1Type, *O1List] {
			return o1Client.Pods(namespace)
		})
	measuredOMClient := ksmetrics.NewWrappedClusterScopedClient[*OMType, *OMList](clientMetrics, oMGVR, oMClient)

	workqueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), controllerName)

	ctlr := &controller{
		i1Sampler: ksmetrics.NewListLenSampler(i1PreInformer.Informer().GetStore().List,
			&k8smetrics.KubeOpts{Namespace: "kubestellar", Subsystem: controllerName,
				Name: "i1s", Help: "number of I1 objects", StabilityLevel: k8smetrics.ALPHA}),
		i1Client:         measuredI1Client,
		i1InformerSynced: i1PreInformer.Informer().HasSynced,
		i1Lister:         i1PreInformer.Lister(),
		iNSampler: ksmetrics.NewListLenSampler(iNPreInformer.Informer().GetStore().List,
			&k8smetrics.KubeOpts{Namespace: "kubestellar", Subsystem: controllerName,
				Name: "ins", Help: "number of IN objects", StabilityLevel: k8smetrics.ALPHA}),
		iNClient:         measuredINClient,
		iNInformerSynced: iNPreInformer.Informer().HasSynced,
		iNLister:         iNPreInformer.Lister(),
		o1Sampler: ksmetrics.NewListLenSampler(o1PreInformer.Informer().GetStore().List,
			&k8smetrics.KubeOpts{Namespace: "kubestellar", Subsystem: controllerName,
				Name: "o1s", Help: "number of O1 objects", StabilityLevel: k8smetrics.ALPHA}),
		o1Client:         measuredO1Client,
		o1InformerSynced: o1PreInformer.Informer().HasSynced,
		o1Lister:         o1PreInformer.Lister(),
		oMSampler: ksmetrics.NewListLenSampler(oMPreInformer.Informer().GetStore().List,
			&k8smetrics.KubeOpts{Namespace: "kubestellar", Subsystem: controllerName,
				Name: "oNs", Help: "number of OM objects", StabilityLevel: k8smetrics.ALPHA}),
		oMClient:         measuredOMClient,
		oMInformerSynced: oMPreInformer.Informer().HasSynced,
		oMLister:         oMPreInformer.Lister(),
		workqueue:        workqueue,
	}
	i1PreInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			typed := obj.(*I1Type)
			objName := cache.MetaObjectToName(typed)
			logger.V(5).Info("Enqueuing reference to I1 due to informer notification of add", "ref", objName)
			ctlr.workqueue.Add(i1ref(objName))
			ctlr.i1Sampler.Prod()
		},
		UpdateFunc: func(oldObj, newObj any) {
			oldTyped := oldObj.(*I1Type)
			newTyped := newObj.(*I1Type)
			objName := cache.MetaObjectToName(newTyped)
			// The following is an example of ignoring an irrelevant change;
			// replace with what is actually appropriate.
			if oldTyped.Generation == newTyped.Generation {
				logger.V(5).Info("Not enqueuing reference to I1, got informed of irrelevant change", "ref", objName)
				return
			}
			logger.V(5).Info("Enqueuing reference to I1 due to informer notification of change", "ref", objName)
			ctlr.workqueue.Add(i1ref(objName))
			ctlr.i1Sampler.Prod()
		},
		DeleteFunc: func(obj any) {
			if typed, is := obj.(cache.DeletedFinalStateUnknown); is {
				obj = typed.Obj
			}
			typed := obj.(*I1Type)
			objName := cache.MetaObjectToName(typed)
			logger.V(5).Info("Enqueuing reference to I1 due to informer notification of deletion", "ref", objName)
			ctlr.workqueue.Add(i1ref(objName))
			ctlr.i1Sampler.Prod()
		},
	})
	iNPreInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			typed := obj.(*INType)
			objName := cache.MetaObjectToName(typed)
			logger.V(5).Info("Enqueuing reference to IN due to informer notification of add", "ref", objName)
			ctlr.workqueue.Add(iNref(objName))
			ctlr.iNSampler.Prod()
		},
		UpdateFunc: func(oldObj, newObj any) {
			oldTyped := oldObj.(*INType)
			newTyped := newObj.(*INType)
			objName := cache.MetaObjectToName(newTyped)
			// The following is an example of ignoring an irrelevant change;
			// replace with what is actually appropriate.
			if oldTyped.Generation == newTyped.Generation {
				logger.V(5).Info("Not enqueuing reference to IN, got informed of irrelevant change", "ref", objName)
				return
			}
			logger.V(5).Info("Enqueuing reference to IN due to informer notification of change", "ref", objName)
			ctlr.workqueue.Add(iNref(objName))
			ctlr.iNSampler.Prod()
		},
		DeleteFunc: func(obj any) {
			if typed, is := obj.(cache.DeletedFinalStateUnknown); is {
				obj = typed.Obj
			}
			typed := obj.(*INType)
			objName := cache.MetaObjectToName(typed)
			logger.V(5).Info("Enqueuing reference to IN due to informer notification of deletion", "ref", objName)
			ctlr.workqueue.Add(iNref(objName))
			ctlr.iNSampler.Prod()
		},
	})
	o1PreInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			typed := obj.(*O1Type)
			objName := cache.MetaObjectToName(typed)
			logger.V(5).Info("Enqueuing reference to O1 due to informer notification of add", "ref", objName)
			ctlr.workqueue.Add(o1ref(objName))
			ctlr.o1Sampler.Prod()
		},
		UpdateFunc: func(oldObj, newObj any) {
			oldTyped := oldObj.(*O1Type)
			newTyped := newObj.(*O1Type)
			objName := cache.MetaObjectToName(newTyped)
			// The following is an example of ignoring an irrelevant change;
			// replace with what is actually appropriate.
			if oldTyped.Generation == newTyped.Generation {
				logger.V(5).Info("Not enqueuing reference to O1, got informed of irrelevant change", "ref", objName)
				return
			}
			logger.V(5).Info("Enqueuing reference to O1 due to informer notification of change", "ref", objName)
			ctlr.workqueue.Add(o1ref(objName))
			ctlr.o1Sampler.Prod()
		},
		DeleteFunc: func(obj any) {
			if typed, is := obj.(cache.DeletedFinalStateUnknown); is {
				obj = typed.Obj
			}
			typed := obj.(*O1Type)
			objName := cache.MetaObjectToName(typed)
			logger.V(5).Info("Enqueuing reference to O1 due to informer notification of deletion", "ref", objName)
			ctlr.workqueue.Add(o1ref(objName))
			ctlr.o1Sampler.Prod()
		},
	})
	oMPreInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			typed := obj.(*OMType)
			objName := cache.MetaObjectToName(typed)
			logger.V(5).Info("Enqueuing reference to OM due to informer notification of add", "ref", objName)
			ctlr.workqueue.Add(oMref(objName))
			ctlr.oMSampler.Prod()
		},
		UpdateFunc: func(oldObj, newObj any) {
			oldTyped := oldObj.(*OMType)
			newTyped := newObj.(*OMType)
			objName := cache.MetaObjectToName(newTyped)
			// The following is an example of ignoring an irrelevant change;
			// replace with what is actually appropriate.
			if oldTyped.Generation == newTyped.Generation {
				logger.V(5).Info("Not enqueuing reference to OM, got informed of irrelevant change", "ref", objName)
				return
			}
			logger.V(5).Info("Enqueuing reference to OM due to informer notification of change", "ref", objName)
			ctlr.workqueue.Add(oMref(objName))
			ctlr.oMSampler.Prod()
		},
		DeleteFunc: func(obj any) {
			if typed, is := obj.(cache.DeletedFinalStateUnknown); is {
				obj = typed.Obj
			}
			typed := obj.(*OMType)
			objName := cache.MetaObjectToName(typed)
			logger.V(5).Info("Enqueuing reference to OM due to informer notification of deletion", "ref", objName)
			ctlr.workqueue.Add(oMref(objName))
			ctlr.oMSampler.Prod()
		},
	})

	return ctlr
}

// controller is the representation of the controller.
type controller struct {
	i1Sampler        ksmetrics.Sampler
	i1Client         ksmetrics.NamespacedClient[*I1Type, *I1List]
	i1InformerSynced func() bool
	i1Lister         I1Lister

	// And so on for I2, ...

	iNSampler        ksmetrics.Sampler
	iNClient         ksmetrics.ClientModNamespace[*INType, *INList]
	iNInformerSynced func() bool
	iNLister         INLister

	o1Sampler        ksmetrics.Sampler
	o1Client         ksmetrics.NamespacedClient[*O1Type, *O1List]
	o1InformerSynced func() bool
	o1Lister         O1Lister

	// And so on, for O2, ...

	oMSampler        ksmetrics.Sampler
	oMClient         ksmetrics.ClientModNamespace[*OMType, *OMList]
	oMInformerSynced func() bool
	oMLister         OMLister

	workqueue workqueue.RateLimitingInterface

	// A controller also has internal data structures to support the sync methods.
	// These might be formulated in any one of a variety of ways, possibly even a mixture.
	// One style would be to just hold the latest state seen for every I1, I2, ... IN object.
	// Another style would be to hold an internal representation of the state
	// that the controller wants to have written to each O1, O2, ... OM object.

	// Regarding concurrency: the pattern here ensures that there are no two concurrent
	// sync invocations working on the same object (reference);
	// other than that, there are no guarantees.
}

type i1ref cache.ObjectName
type iNref cache.ObjectName
type o1ref cache.ObjectName
type oMref cache.ObjectName

// Run animates the controller.
// Call this after starting the informers.
func (ctlr *controller) Run(ctx context.Context, concurrency int) error {
	defer utilruntime.HandleCrash()
	defer ctlr.workqueue.ShutDown()

	logger := klog.FromContext(ctx)

	logger.Info("Starting " + controllerName)

	// Wait for the caches to be synced before starting workers
	logger.Info("Waiting for informer caches to sync")

	if ok := cache.WaitForCacheSync(ctx.Done(), ctlr.i1InformerSynced /* and i2, ... */, ctlr.iNInformerSynced,
		ctlr.o1InformerSynced /* and o2, ... */, ctlr.oMInformerSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}
	// It is safe to omit the wait for sync of type TJ if the lister for TJ
	// is not used in ctlr.sync for anything other than pulling the referenced
	// tJ out of the TJ lister.

	logger.Info("Starting workers", "count", concurrency)
	// Launch workers to process Binding
	for i := 1; i <= concurrency; i++ {
		workerId := i // in go, there is one `i` variable that gets different values in different iterations of the loop
		go wait.UntilWithContext(ctx, func(ctx context.Context) { ctlr.runWorker(ctx, workerId) }, time.Second)
	}

	logger.Info("Started workers")
	<-ctx.Done()
	logger.Info("Shutting down workers")

	return nil
}

func (ctlr *controller) runWorker(ctx context.Context, workerId int) {
	logger := klog.FromContext(ctx).WithValues("workerID", workerId)
	ctx = klog.NewContext(ctx, logger)
	for ctlr.processNextWorkItem(ctx) {
	}
}

func (ctlr *controller) processNextWorkItem(ctx context.Context) bool {
	logger := klog.FromContext(ctx)
	objRef, shutdown := ctlr.workqueue.Get()
	if shutdown {
		return false
	}
	logger.V(5).Info("Popped object reference from workqueue", "objRef", objRef, "refType", util.NewPrintType(objRef))

	err := ctlr.sync(ctx, objRef)
	if err == nil {
		// If no error occurs then Forget this item so it does not
		// get queued again until another change happens.
		ctlr.workqueue.Forget(objRef)
		logger.V(4).Info("Processed workqueue item successfully.", "objRef", objRef, "refType", util.NewPrintType(objRef))
	} else {
		ctlr.workqueue.AddRateLimited(objRef)
		logger.V(4).Info("Encountered transient error while processing workqueue item; will retry later", "objRef", objRef, "refType", util.NewPrintType(objRef), "error", err)
	}
	return true
}

// sync returns a non-nil error when a retry is desired.
// Non-retryable errors are handled internally and `nil` is returned.
// Returns `nil` on success too.
func (ctlr *controller) sync(ctx context.Context, objRef any) error {
	switch typed := objRef.(type) {
	case i1ref:
		return ctlr.syncI1(ctx, cache.ObjectName(typed))
	// and so on, for I2, ...
	case iNref:
		return ctlr.syncIN(ctx, typed.Name)
	case o1ref:
		return ctlr.syncO1(ctx, cache.ObjectName(typed))
	// and so on, for O2, ...
	case oMref:
		return ctlr.syncOM(ctx, typed.Name)
	default:
		utilruntime.HandleError(fmt.Errorf("unexpected type of workqueue item: %T", objRef))
		return nil
	}
}

func (ctlr *controller) syncI1(ctx context.Context, ref cache.ObjectName) error {
	logger := klog.FromContext(ctx)
	i1, err := ctlr.i1Lister.ReplicaSets(ref.Namespace).Get(ref.Name)
	// How much do I hate the lister API? Let me count the ways...
	if err == nil {
	} else if i1 != nil {
		utilruntime.HandleError(fmt.Errorf("inconceivable! I1 Lister returned non-nil error and non-nil object, ref=%v", ref))
		return nil
	} else if !k8serrors.IsNotFound(err) {
		utilruntime.HandleError(fmt.Errorf("inconceivable! I1 Lister returned an error other than IsNotFound, ref=%v, err=%w", ref, err))
		return nil
	}
	if i1 != nil && i1.DeletionTimestamp != nil {
		logger.V(4).Info("Treating an I1 being deleted as if it is already gone", "ref", ref)
		i1 = nil
	}
	// Update internal data structures to account for the current state of i1.
	// For every object of type O1 or O2 or ... OM that needs to be updated,
	// enqueue a reference to that object and logger.V(5).Info that.
	// In the I and/or O sync methods be sure to break possible cycles such as
	// syncIJ(ctx, iJ) always enqueues a reference to oK
	// and syncOK(ctx, oK) always enqueues a reference to iJ.
	// If i1 is not nil then update its status section to do any or all of the following,
	// depending on the design of I1Status and this controller's role in that:
	// - report i1.Generation;
	// - report user errors in i1;
	// - summarize state of related O1, O2, ... OM objects.
	return nil // or an error, as appropriate
}

func (ctlr *controller) syncIN(ctx context.Context, name string) error {
	logger := klog.FromContext(ctx)
	iN, err := ctlr.iNLister.Get(name)
	// If you are willing to not consider the inconceivable outcomes, this is simple.
	if err == nil && iN.DeletionTimestamp != nil {
		logger.V(4).Info("Treating an IN being deleted as if it is already gone", "name", name)
		iN = nil
	}
	// Update internal data structures to account for the current state of iN.
	// For every object of type O1 or O2 or ... OM that needs to be updated,
	// enqueue a reference to that object and logger.V(5).Info that.
	// In the I and/or O sync methods be sure to break possible cycles such as
	// syncIJ(ctx, iJ) always enqueues a reference to oK
	// and syncOK(ctx, oK) always enqueues a reference to iJ.
	// If iN is not nil then update its status section to do any or all of the following,
	// depending on the design of INStatus and this controller's role in that:
	// - report iN.Generation;
	// - report user errors in iN;
	// - summarize state of related O1, O2, ... OM objects.
	return nil // or an error, as appropriate
}

func (ctlr *controller) syncO1(ctx context.Context, ref cache.ObjectName) error {
	logger := klog.FromContext(ctx)
	o1, err := ctlr.o1Lister.Pods(ref.Namespace).Get(ref.Name)
	if err == nil && o1.DeletionTimestamp != nil {
		logger.V(4).Info("Ignoring an O1 because it is being deleted", "ref", ref)
		// There will be another notification after o1 is gone.
		return nil
	}
	// Compare the state of o1 with what the controller's internal data structures say
	// o1's state should be; if different then create/update/delete o1 and logger.V(2).Info that.
	// If the design of I1Status or I2Status or ... INStatus makes it depend on o1's state
	// then enqueue a reference to the relevant I1 or I2 or ... IN object(s) and logger.V(5).Info that.

	// Comparing the state read for o1 with what the internal data structures say the state of o1
	// should be, this handles all sorts of scenarios --- including ones where stuff happend while
	// this controller was not running.
	return nil // or an error, as appropriate
}

func (ctlr *controller) syncOM(ctx context.Context, name string) error {
	logger := klog.FromContext(ctx)
	oM, err := ctlr.oMLister.Get(name)
	if err == nil && oM.DeletionTimestamp != nil {
		logger.V(4).Info("Ignoring an OM because it is being deleted", "name", name)
		// There will be another notification after oM is gone.
		return nil
	}
	// Compare the state of oM with what the controller's internal data structures say
	// oM's state should be; if different then create/update/delete oM and logger.V(2).Info that.
	// If the design of I1Status or I2Status or ... INStatus makes it depend on oM's state
	// then enqueue a reference to the relevant I1 or I2 or ... IN object(s) and logger.V(5).Info that.
	return nil // or an error, as appropriate
}
