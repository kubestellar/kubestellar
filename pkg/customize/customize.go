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

package customize

import (
	"io"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"

	a "github.com/kubestellar/kubestellar/pkg/abstract"
)

// Expander is something that can do parameter expansion on unmarshaled JSON data.
type Expander struct {
	// Undefined is the set of parameters that were referenced but had no value
	Undefined sets.Set[string]

	// ChangedSome reports whether parameter expansion made any changes to the data.
	// When the value of a parameter is not found, that expansion does not happen.
	ChangedSome bool

	loadDefs func() a.Getter[string, string]

	defs a.Getter[string, string]
}

func NewExpander(loadDefs func() a.Getter[string, string]) *Expander {
	return &Expander{
		Undefined: sets.New[string](),
		loadDefs:  loadDefs,
	}
}

func (exp *Expander) WantedChange() bool {
	return exp.ChangedSome || len(exp.Undefined) != 0
}

// ExpandParameters side-effects the given JSON data to expand parameters in leaf strings
func (exp *Expander) ExpandParameters(data any) any {
	switch typed := data.(type) {
	case string:
		return exp.ExpandString(typed)
	case map[string]any:
		for key, val := range typed {
			newVal := exp.ExpandParameters(val)
			typed[key] = newVal
		}
		return typed
	case []any:
		for idx, val := range typed {
			newVal := exp.ExpandParameters(val)
			typed[idx] = newVal
		}
		return typed
	default:
		return typed
	}
}

// ExpandString does parameter expansion on one string
func (exp *Expander) ExpandString(input string) string {
	if !strings.ContainsRune(input, '%') {
		return input
	}
	var builder strings.Builder
	inputReader := strings.NewReader(input)
	for {
		next, _, err := inputReader.ReadRune()
		if err == io.EOF {
			break
		}
		if next != '%' {
			builder.WriteRune(next)
			continue
		}
		next, _, err = inputReader.ReadRune()
		if err == io.EOF {
			builder.WriteRune('%')
			break
		}
		if next != '(' {
			builder.WriteRune('%')
			builder.WriteRune(next)
			continue
		}
		name := readNameToReplace(inputReader)
		if exp.defs == nil {
			exp.defs = exp.loadDefs()
		}
		replacement, found := exp.defs.Get(name)
		if found {
			builder.WriteString(replacement)
			exp.ChangedSome = true
		} else {
			exp.Undefined.Insert(name)
			builder.WriteString("%(")
			builder.WriteString(name)
			builder.WriteString(")")
		}
	}
	return builder.String()
}

func readNameToReplace(inputReader io.RuneReader) string {
	var builder strings.Builder
	for {
		next, _, err := inputReader.ReadRune()
		if err != nil {
			return builder.String()
		}
		if next == ')' {
			return builder.String()
		}
		builder.WriteRune(next)
	}
}
