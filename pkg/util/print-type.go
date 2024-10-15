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

package util

import (
	"fmt"
)

// PrintType wraps any value with a `String()` method that uses
// `fmt.Sprintf` to print the value's type.
// It is less costly to construct one of these and pass it to an Info call
// on a disabled logr than to certainly invoke `fmt.Sprintf`.
type PrintType struct{ Val any }

// NewPrintType makes a new PrintType value
func NewPrintType(Val any) PrintType { return PrintType{Val} }

func (pt PrintType) String() string {
	return fmt.Sprintf("%T", pt.Val)
}
