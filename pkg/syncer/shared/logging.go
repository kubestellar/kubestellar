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

package shared

import "github.com/go-logr/logr"

// WithReconciler adds the reconciler name to the logger.
func WithReconciler(logger logr.Logger, reconciler string) logr.Logger {
	return logger.WithValues("reconciler", reconciler)
}

// WithQueueKey adds the queue key to the logger.
func WithQueueKey(logger logr.Logger, key string) logr.Logger {
	return logger.WithValues("key", key)
}
