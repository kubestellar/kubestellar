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
	"fmt"
	"strings"
	"text/template"
)

// Expander is something that can do template expansion on unmarshaled JSON data.
type Expander struct {
	// Errors is the `.Error()` of the errors encountered
	Errors []string

	// ChangedSome reports whether parameter expansion made any changes to the data.
	ChangedSome bool

	defs map[string]string
}

// NewExpander makes a new one. Nobody mutates `defs` during a call on ExpandTemplates.
func NewExpander(defs map[string]string) *Expander {
	return &Expander{
		Errors: []string{},
		defs:   defs,
	}
}

// WantedChange tells whether a paremeter reference was seen
func (exp *Expander) WantedChange() bool {
	return exp.ChangedSome || len(exp.Errors) != 0
}

// ExpandTemplates side-effects the given JSON data to expand templates in leaf strings
func (exp *Expander) ExpandTemplates(path string, data any) any {
	switch typed := data.(type) {
	case string:
		return exp.ExpandString(path, typed)
	case map[string]any:
		for key, val := range typed {
			newVal := exp.ExpandTemplates(path+"."+key, val)
			typed[key] = newVal
		}
		return typed
	case []any:
		for idx, val := range typed {
			newVal := exp.ExpandTemplates(fmt.Sprintf("%s[%d]", path, idx), val)
			typed[idx] = newVal
		}
		return typed
	default:
		return typed
	}
}

// ExpandString does parameter expansion on one string
func (exp *Expander) ExpandString(path, input string) string {
	if !strings.Contains(input, "{{") {
		return input
	}
	tmpl := template.New(path).Option("missingkey=error")
	tmpl, err := tmpl.Parse(input)
	if err != nil {
		exp.Errors = append(exp.Errors, peel(err).Error())
		return ""
	}
	var builder bytes.Buffer
	err = tmpl.Execute(&builder, exp.defs)
	ans := builder.String()
	if err != nil {
		exp.Errors = append(exp.Errors, peel(err).Error())
	}
	exp.ChangedSome = true
	return ans
}

func peel(err error) error {
	if templateErr, is := err.(*template.ExecError); is {
		return templateErr.Err
	}
	return err
}
