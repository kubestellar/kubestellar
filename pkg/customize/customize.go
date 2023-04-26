/*
Copyright 2023 The KCP Authors.

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
	"encoding/json"
	"io"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"

	schedulingv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/scheduling/v1alpha1"

	edgeapi "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
	"github.com/kcp-dev/edge-mc/pkg/jsonpath"
)

func Customize(logger klog.Logger, input *unstructured.Unstructured, customizer *edgeapi.Customizer, loc *schedulingv1alpha1.Location) *unstructured.Unstructured {
	expandInput := input.GetAnnotations()[edgeapi.ParameterExpansionAnnotationKey] == "true"
	expandCustomizer := customizer != nil && customizer.Annotations[edgeapi.ParameterExpansionAnnotationKey] == "true"
	if customizer == nil && !expandInput {
		return input
	}
	output := input.DeepCopy()
	outputU := output.UnstructuredContent()
	defs := Definitions{loc.GetLabels(), loc.GetAnnotations()}
	if expandInput {
		outputA := expandParameters(outputU, defs)
		outputU = outputA.(map[string]any)
	}
	output.SetUnstructuredContent(outputU)
	if customizer != nil {
		jrs := []jsonpath.Replacement{}
		for _, repl := range customizer.Replacements {
			where := repl.Path
			if expandCustomizer {
				where = expandString(where, defs)
			}
			jp, err := jsonpath.ParsePath(where)
			if err != nil {
				logger.Error(err, "Failed to parse replacement path")
				continue
			}
			valueStr := repl.Value
			if expandCustomizer {
				valueStr = expandString(valueStr, defs)
			}
			var valueAny any
			err = json.Unmarshal([]byte(valueStr), &valueAny)
			if err != nil {
				logger.Error(err, "Failed to unmarshal replacement value", "replacementPath", repl.Path, "replacementValue", repl.Value, "valueStr", valueStr)
				continue
			}
			jrs = append(jrs, jsonpath.Replacement{Path: jp, Value: valueAny})
		}
		outputU, err := jsonpath.Update(outputU, jrs...)
		if err != nil {
			logger.Error(err, "Failed to do update")
		} else {
			output.SetUnstructuredContent(outputU)
		}
	}
	return output
}

type Definitions []map[string]string

func (defs Definitions) Get(key string) (string, bool) {
	for _, def := range defs {
		if val, have := def[key]; have {
			return val, true
		}
	}
	return "", false
}

func expandString(input string, defs Definitions) string {
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
		replacement, found := defs.Get(name)
		if found {
			builder.WriteString(replacement)
		} else {
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

func expandParameters(data any, defs Definitions) any {
	switch typed := data.(type) {
	case string:
		return expandString(typed, defs)
	case map[string]any:
		for key, val := range typed {
			newVal := expandParameters(val, defs)
			typed[key] = newVal
		}
		return typed
	case []any:
		for idx, val := range typed {
			newVal := expandParameters(val, defs)
			typed[idx] = newVal
		}
		return typed
	default:
		return typed
	}
}
