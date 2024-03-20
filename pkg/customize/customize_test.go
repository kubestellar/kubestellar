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
	"fmt"
	"math/rand"
	"strings"
	"testing"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"

	a "github.com/kubestellar/kubestellar/pkg/abstract"
)

func TestCustomize(t *testing.T) {
	rg := rand.New(rand.NewSource(42))
	rg.Uint64()
	rg.Uint64()
	rg.Uint64()
	for try := 1; try <= 100; try++ {
		gen := &generator{rg: rg, defs: map[string]string{}, undefined: sets.New[string]()}
		input, expected := gen.generateData()
		exp := NewExpander(func() a.Getter[string, string] { return a.MapFromLang(gen.defs) })
		inputCopy := runtime.DeepCopyJSONValue(input)
		actual := exp.ExpandParameters(inputCopy)
		if !apiequality.Semantic.DeepEqual(expected, actual) {
			t.Fatalf("Expected %#v, got %#v; input=%#v, defs=%v", expected, actual, input, gen.defs)
		}
		if gen.wantSome != exp.WantedChange() {
			t.Fatalf("Expected WantedChange=%v, got %v; input=%#v", gen.wantSome, exp.WantedChange(), input)
		}
		if gen.changeSome != exp.ChangedSome {
			t.Fatalf("Expected ChangedSome=%v, got %v; input=%#v, defs=%v", gen.changeSome, exp.ChangedSome, input, gen.defs)
		}
		if !gen.undefined.Equal(exp.Undefined) {
			t.Fatalf("Expected Undefined=%v, got %v; input=%#v, defs=%v", gen.undefined, exp.Undefined, input, gen.defs)

		}
		t.Logf("Success expanding %#v to %#v, defs=%v", input, expected, gen.defs)
	}
}

type generator struct {
	rg         *rand.Rand
	defs       map[string]string
	undefined  sets.Set[string]
	wantSome   bool
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
	for i := 0; i < size; i++ {
		if withParm && gen.rg.Intn(5) == 0 { // generate a request for parameter expansion
			parmName := fmt.Sprintf("p%d", gen.rg.Intn(10))
			call := "%(" + parmName + ")"
			input.WriteString(call)
			gen.wantSome = true
			var parmVal *string
			if val, have := gen.defs[parmName]; have { // value already decided
				parmVal = &val
			} else if gen.undefined.Has(parmName) { // already decided to be undefined
			} else if gen.rg.Intn(3) > 0 { // define a new parameter value
				val, _ := gen.generateString(false)
				parmVal = &val
				gen.defs[parmName] = val
			} else { // make thsi one undefined
				gen.undefined.Insert(parmName)
			}
			if parmVal != nil {
				expected.WriteString(*parmVal)
				gen.changeSome = true
			} else {
				expected.WriteString(call)
			}
		} else {
			chr := 'A' + gen.rg.Intn(26)
			input.WriteByte(byte(chr))
			expected.WriteByte(byte(chr))
		}
	}
	return input.String(), expected.String()
}
