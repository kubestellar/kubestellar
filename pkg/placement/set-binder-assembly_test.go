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
	"context"
	"os"
	"testing"

	"k8s.io/klog/v2"
)

func TestMain(m *testing.M) {
	klog.InitFlags(nil)
	os.Exit(m.Run())
}

func TestSetBinder(t *testing.T) {
	ctx := context.Background()
	logger := klog.FromContext(ctx)
	amp := NewTestAPIMapProvider(logger)
	binder := NewSetBinder(logger, NewWorkloadPartsDifferencer, NewUpsyncDifferencer, NewResolvedWhereDifferencer,
		SimpleBindingOrganizer(logger),
		amp, DefaultResourceModes, nil)
	exerciseSetBinder(t, logger, amp.AsResourceReceiver(), binder)
}
