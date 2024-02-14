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

package abstract

import (
	"fmt"
	"strings"
)

type Pair[First, Second any] struct {
	First  First
	Second Second
}

func NewPair[First, Second any](first First, second Second) Pair[First, Second] {
	return Pair[First, Second]{first, second}
}

func (tup Pair[First, Second]) GetFirst() First   { return tup.First }
func (tup Pair[First, Second]) GetSecond() Second { return tup.Second }

func (tup Pair[First, Second]) String() string {
	var ans strings.Builder
	ans.WriteRune('(')
	ans.WriteString(fmt.Sprintf("%v", tup.First))
	ans.WriteString(", ")
	ans.WriteString(fmt.Sprintf("%v", tup.Second))
	ans.WriteRune(')')
	return ans.String()
}
