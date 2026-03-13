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
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
)

// TestExpandTemplatesOversizedInput verifies that strings exceeding maxTemplateInputSize
// are rejected with an error and do not cause a DoS.
func TestExpandTemplatesOversizedInput(t *testing.T) {
	// A string of exactly maxTemplateInputSize bytes (with template syntax) should be allowed.
	atLimit := "{{.key}}" + strings.Repeat("A", maxTemplateInputSize-len("{{.key}}"))
	_, _, errs := ExpandTemplates("test", atLimit, map[string]string{"key": "v"})
	for _, e := range errs {
		if strings.Contains(e, "exceeds maximum allowed size") {
			t.Errorf("expected no size-limit error at exact limit, got: %v", errs)
			break
		}
	}

	// A string of exactly maxTemplateInputSize+1 bytes should be rejected.
	oversize := "{{.key}}" + strings.Repeat("A", maxTemplateInputSize-len("{{.key}}")+1)
	input := map[string]any{"field": oversize}
	defs := map[string]string{"key": "value"}

	_, _, errs = ExpandTemplates("test", input, defs)
	if len(errs) == 0 {
		t.Error("expected an error for oversized template input, got none")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e, "exceeds maximum allowed size") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'exceeds maximum allowed size' error, got: %v", errs)
	}
}

// TestExpandTemplatesRestrictedFunctions verifies that the disabled template functions
// (call, html, js, urlquery) return errors instead of executing.
func TestExpandTemplatesRestrictedFunctions(t *testing.T) {
	tests := []struct {
		name     string
		template string
	}{
		{"html function disabled", `{{html "<b>test</b>"}}`},
		{"js function disabled", `{{js "alert(1)"}}`},
		{"urlquery function disabled", `{{urlquery "a=1&b=2"}}`},
		// call is overridden to prevent indirect function invocation; calling a
		// built-in like print via call should be blocked.
		{"call function disabled", `{{call print "hello"}}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, errs := ExpandTemplates("test", tc.template, map[string]string{})
			if len(errs) == 0 {
				t.Errorf("expected error for restricted function %q, got none", tc.name)
			}
		})
	}
}

// TestExpandTemplatesMissingKeyError verifies that referencing a non-existent template key
// returns an error gracefully without panicking.
func TestExpandTemplatesMissingKeyError(t *testing.T) {
	input := "{{.nonexistentKey}}"
	defs := map[string]string{}

	_, _, errs := ExpandTemplates("test", input, defs)
	if len(errs) == 0 {
		t.Error("expected an error for missing key, got none")
	}
}

// TestExpandTemplatesValidSubstitution verifies that normal template substitution
// still works correctly after applying security restrictions.
func TestExpandTemplatesValidSubstitution(t *testing.T) {
	input := map[string]any{
		"greeting": "Hello, {{.name}}!",
		"static":   "no template here",
	}
	defs := map[string]string{"name": "World"}

	output, wantedChange, errs := ExpandTemplates("test", input, defs)
	if len(errs) != 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
	if !wantedChange {
		t.Error("expected wantedChange=true, got false")
	}
	outMap := output.(map[string]any)
	if outMap["greeting"] != "Hello, World!" {
		t.Errorf("expected 'Hello, World!', got %q", outMap["greeting"])
	}
	if outMap["static"] != "no template here" {
		t.Errorf("expected 'no template here', got %q", outMap["static"])
	}
}

func FuzzTestCustomize(f *testing.F) {
	f.Add(int64(42)) // Add seed values - must match the parameter types
	f.Add(int64(19))

	f.Fuzz(func(t *testing.T, seed int64) {
		rg := rand.New(rand.NewSource(seed))
		rg.Uint64()
		rg.Uint64()
		rg.Uint64()

		for try := 1; try <= 100; try++ {
			gen := &generator{rg: rg, defs: map[string]string{}, undefined: sets.New[string]()}
			input, expected := gen.generateData()
			inputCopy := runtime.DeepCopyJSONValue(input)
			actual, wantedChange, errs := ExpandTemplates(fmt.Sprintf("try%d", try), inputCopy, gen.defs)
			t.Logf("Tested input=%#v, defs=%#v", input, gen.defs)
			fail := false
			if len(gen.errors) == len(errs) {
				t.Logf("Got expected number of errors; errors=%#v", errs)
			} else {
				t.Errorf("Expected errors=%v, got errors %v", gen.errors, errs)
				fail = true
			}
			if apiequality.Semantic.DeepEqual(expected, actual) {
				t.Logf("Got expected output %#v", actual)
			} else {
				t.Errorf("Expected %#v, got %#v", expected, actual)
				fail = true
			}
			if (gen.changeSome || len(gen.errors) > 0) != wantedChange {
				t.Errorf("Expected WantedChange=%v, got %v", (gen.changeSome || len(gen.errors) > 0), wantedChange)
				fail = true
			}
			if fail {
				t.FailNow()
			}
			t.Log("success")
		}
	})
}

type generator struct {
	rg         *rand.Rand
	defs       map[string]string
	undefined  sets.Set[string]
	errors     []error
	changeSome bool
}

func (gen *generator) generateData() (any, any) {
	typeI := gen.rg.Intn(100)
	switch {
	case typeI < 25:
		size := gen.rg.Intn(3) + gen.rg.Intn(2)
		input := make(map[string]any, size)
		expected := make(map[string]any, size)
		for i := 0; i < size; i++ {
			key := fmt.Sprintf("k%d", (i+1)*10+gen.rg.Intn(10))
			inputVal, expectedVal := gen.generateData()
			input[key] = inputVal
			expected[key] = expectedVal
		}
		return input, expected
	case typeI < 50:
		size := gen.rg.Intn(3) + gen.rg.Intn(2)
		input := make([]any, size)
		expected := make([]any, size)
		for i := 0; i < size; i++ {
			input[i], expected[i] = gen.generateData()
		}
		return input, expected
	case typeI < 75:
		return gen.generateString(true)
	case typeI < 82:
		x := gen.rg.Int63()
		return x, x
	case typeI < 89:
		x := gen.rg.Float64()
		return x, x
	case typeI < 96:
		x := gen.rg.Intn(2) == 1
		return x, x
	default:
		return nil, nil
	}
}

func (gen *generator) generateString(withParm bool) (string, string) {
	size := gen.rg.Intn(4) + gen.rg.Intn(4)
	var input strings.Builder
	var expected strings.Builder
	expectMore := true
	var err error
	expectSyntaxError := false
	gendParm := false
	for i := 0; i < size; i++ {
		if withParm && gen.rg.Intn(60) == 0 { // generate a syntax error
			input.WriteString("{{ + }}")
			expectMore = false
			expectSyntaxError = true
		} else if withParm && gen.rg.Intn(60) == 0 { // generate a syntax error
			input.WriteString("{{if true}}")
			expectMore = false
			expectSyntaxError = true
		} else if withParm && gen.rg.Intn(5) == 0 { // generate a request for parameter expansion
			parmName := fmt.Sprintf("p%d", gen.rg.Intn(10))
			call := "{{." + parmName + "}}"
			input.WriteString(call)
			gendParm = true
			var parmVal *string
			if val, have := gen.defs[parmName]; have { // value already decided
				parmVal = &val
			} else if gen.undefined.Has(parmName) { // already decided to be undefined
				if err == nil {
					err = fmt.Errorf("Undefined: %q", parmName)
				}
				expectMore = false
			} else if gen.rg.Intn(3) > 0 { // define a new parameter value
				val, _ := gen.generateString(false)
				parmVal = &val
				gen.defs[parmName] = val
			} else { // make this one undefined
				gen.undefined.Insert(parmName)
				if err == nil {
					err = fmt.Errorf("Undefined: %q", parmName)
				}
				expectMore = false
			}
			if expectMore {
				expected.WriteString(*parmVal)
			}
		} else {
			chr := 'A' + gen.rg.Intn(26)
			input.WriteByte(byte(chr))
			if expectMore {
				expected.WriteByte(byte(chr))
			}
		}
	}
	if expectSyntaxError {
		gen.errors = append(gen.errors, errors.New("syntax error"))
		return input.String(), ""

	}
	gen.changeSome = gen.changeSome || gendParm
	if err != nil {
		gen.errors = append(gen.errors, err)
	}
	return input.String(), expected.String()
}
