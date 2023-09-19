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

package v1alpha1

// WantSingletonReportKey is used to request return of reported state to
// the Workload Description Space while the number of executing copies is 1.
// This request is indicated by a workload object having an annotation with
// this key and the value "true".
const WantSingletonReportKey string = "kubestellar.io/want-singleton-report"

// ExecutingCountKey is used to report on the number of executing copies of
// a workload object while it has the "kubestellar.io/want-singleton-report=true" annotation.
// This is the key of an annotation whose value is a string representing the number of
// executing copies of the annotated workload object.  While this annotation is
// present with the value "1", the reported state is being returned into this workload object.
const ExecutingCountKey string = "kubestellar.io/executing-count"
