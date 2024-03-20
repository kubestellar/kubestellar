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
	"bytes"
	"html/template"
)

// Expander is something that can do parameter expansion on unmarshaled JSON data.
type Expander struct {
	// Errors is the set of errors encountered
	Errors []error

	// ChangedSome reports whether parameter expansion made any changes to the data.
	// When the value of a parameter is not found, that expansion does not happen.
	ChangedSome bool

	loadDefs func() map[string]string

	defs map[string]string
}

func NewExpander(loadDefs func() map[string]string) *Expander {
	return &Expander{
		Errors:   []error{},
		loadDefs: loadDefs,
	}
}

// WantedChange tells whether a paremeter reference was seen
func (exp *Expander) WantedChange() bool {
	return exp.ChangedSome || len(exp.Errors) != 0
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
	tmpl := template.New("").Option("missingkey=error")
	tmpl, err := tmpl.Parse(input)
	if err != nil {
		exp.Errors = append(exp.Errors, err)
		return ""
	}
	if exp.defs == nil {
		exp.defs = exp.loadDefs()
	}
	var builder bytes.Buffer
	err = tmpl.Execute(&builder, exp.defs)
	ans := builder.String()
	if err != nil {
		exp.Errors = append(exp.Errors, err)
	}
	exp.ChangedSome = exp.ChangedSome || (ans != input)
	return ans
}
