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

// ExpandTemplates crawls over the input data structure and does
// template expansion on every `string` except those that are map keys.
// The input is made up of `map[string]any`, `[]any`, `string`, and other primitives.
// This side-effects every map and slice in the input.
// The template expansion treats an input `string` as a template
// as in `text/template` and expands it using the given `templateData`,
// which nothing mutates during this call.
// The given path is whatever the caller wants, and is extended in
// JSONPath style as the input data structure is traversed, ultimately being used
// as input to `text/template` to identify the template --- hence appearing in
// the resulting errors (if any).
// The returned `wantedChange` indicates whether there was any template syntax
// anywhere in the input.
func ExpandTemplates(path string, input any, templateData map[string]string) (output any, wantedChange bool, errors []string) {
	exp := expander{defs: templateData}
	output = exp.expandAny(path, input)
	return output, exp.wantedChange, exp.errors
}

// expander is something that can do template expansion on unmarshaled JSON data.
type expander struct {
	// errors is the `.Error()` of the errors encountered
	errors []string

	// wantedChange reports whether there was any template syntax
	// anywhere in the input
	wantedChange bool

	defs map[string]string
}

// expandAny side-effects the given JSON data to expand templates in leaf strings
func (exp *expander) expandAny(path string, data any) any {
	switch typed := data.(type) {
	case string:
		return exp.expandString(path, typed)
	case map[string]any:
		for key, val := range typed {
			newVal := exp.expandAny(path+"."+key, val)
			typed[key] = newVal
		}
		return typed
	case []any:
		for idx, val := range typed {
			newVal := exp.expandAny(fmt.Sprintf("%s[%d]", path, idx), val)
			typed[idx] = newVal
		}
		return typed
	default:
		return typed
	}
}

// expandString does template expansion on one string
func (exp *expander) expandString(path, input string) string {
	if !strings.Contains(input, "{{") {
		return input
	}
	exp.wantedChange = true
	tmpl := template.New(path).Option("missingkey=error")
	tmpl, err := tmpl.Parse(input)
	if err != nil {
		exp.errors = append(exp.errors, peel(err).Error())
		return ""
	}
	var builder bytes.Buffer
	err = tmpl.Execute(&builder, exp.defs)
	ans := builder.String()
	if err != nil {
		exp.errors = append(exp.errors, peel(err).Error())
	}
	return ans
}

func peel(err error) error {
	if templateErr, is := err.(*template.ExecError); is {
		return templateErr.Err
	}
	return err
}
