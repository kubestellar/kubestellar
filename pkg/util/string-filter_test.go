/*
Copyright 2025 The KubeStellar Authors.

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
	"reflect"
	"testing"

	"github.com/spf13/pflag"

	"k8s.io/apimachinery/pkg/util/sets"
)

func TestNewStringFilterOptions(t *testing.T) {
	sfo := NewStringFilterOptions("alpha", "beta")
	expected := []string{"alpha", "beta"}
	if !reflect.DeepEqual(sfo.Literals, expected) {
		t.Errorf("expected %v, got %v", expected, sfo.Literals)
	}
	if sfo.SpacesToo {
		t.Errorf("expected SpacesToo to default to false")
	}

	sfoSpaces := sfo.SeparateBySpacesToo()
	if !sfoSpaces.SpacesToo {
		t.Errorf("expected SeparateBySpacesToo to set SpacesToo to true")
	}
}

func TestAddToFlags(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	sfo := NewStringFilterOptions("default-val")
	sfo.AddToFlags(fs, "my-filter", "items")

	err := fs.Parse([]string{"--my-filter=val1,val2"})
	if err != nil {
		t.Fatalf("unexpected error parsing flags: %v", err)
	}

	expected := []string{"val1", "val2"}
	if !reflect.DeepEqual(sfo.Literals, expected) {
		t.Errorf("expected literals %v, got %v", expected, sfo.Literals)
	}

	fs2 := pflag.NewFlagSet("test2", pflag.ContinueOnError)
	sfo2 := NewStringFilterOptions().SeparateBySpacesToo()
	sfo2.AddToFlags(fs2, "spaced-filter", "tokens")

	flag := fs2.Lookup("spaced-filter")
	if flag == nil {
		t.Fatalf("expected flag spaced-filter to be registered")
	}
	if flag.Usage != "comma and/or space separated list of tokens" {
		t.Errorf("unexpected usage string: %s", flag.Usage)
	}
}

func TestToFilter(t *testing.T) {
	tests := []struct {
		name            string
		options         StringFilterOptions
		expectedAllPass bool
		expectedSet     sets.Set[string]
		expectedGotSome bool
	}{
		{
			name:            "empty options",
			options:         NewStringFilterOptions(),
			expectedAllPass: false,
			expectedSet:     sets.New[string](),
			expectedGotSome: false,
		},
		{
			name:            "empty string literals only",
			options:         NewStringFilterOptions("", ""),
			expectedAllPass: false,
			expectedSet:     sets.New[string](),
			expectedGotSome: false,
		},
		{
			name:            "whitespace only with SpacesToo",
			options:         NewStringFilterOptions("   ", "\t").SeparateBySpacesToo(),
			expectedAllPass: false,
			expectedSet:     sets.New[string](),
			expectedGotSome: false,
		},
		{
			name:            "single literal without SpacesToo",
			options:         NewStringFilterOptions("hello world"),
			expectedAllPass: false,
			expectedSet:     sets.New("hello world"),
			expectedGotSome: true,
		},
		{
			name:            "multiple literals with SpacesToo",
			options:         NewStringFilterOptions("alpha beta", "gamma \t delta").SeparateBySpacesToo(),
			expectedAllPass: false,
			expectedSet:     sets.New("alpha", "beta", "gamma", "delta"),
			expectedGotSome: true,
		},
		{
			name:            "wildcard literal",
			options:         NewStringFilterOptions("*"),
			expectedAllPass: true,
			expectedSet:     nil,
			expectedGotSome: true,
		},
		{
			name:            "wildcard among other literals with SpacesToo",
			options:         NewStringFilterOptions("foo * bar").SeparateBySpacesToo(),
			expectedAllPass: true,
			expectedSet:     nil,
			expectedGotSome: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			filter, gotSome := tc.options.ToFilter()
			if gotSome != tc.expectedGotSome {
				t.Errorf("gotSome = %v, expected %v", gotSome, tc.expectedGotSome)
			}
			if filter.AllPass != tc.expectedAllPass {
				t.Errorf("AllPass = %v, expected %v", filter.AllPass, tc.expectedAllPass)
			}
			if !tc.expectedAllPass {
				if !filter.Literals.Equal(tc.expectedSet) {
					t.Errorf("Literals = %v, expected %v", filter.Literals, tc.expectedSet)
				}
			}
		})
	}
}

func TestStringFilterPasses(t *testing.T) {
	var nilFilter *StringFilter
	if nilFilter.Passes("anything", false) {
		t.Errorf("expected nil StringFilter.Passes to return false")
	}
	if nilFilter.Passes("*", true) {
		t.Errorf("expected nil StringFilter.Passes to return false even with passStar=true")
	}

	allPassFilter := &StringFilter{AllPass: true}
	if !allPassFilter.Passes("foo", false) {
		t.Errorf("expected AllPass filter to pass arbitrary candidate")
	}
	if !allPassFilter.Passes("*", false) {
		t.Errorf("expected AllPass filter to pass '*' candidate")
	}

	setFilter := &StringFilter{
		Literals: sets.New("user1", "user2"),
	}

	if !setFilter.Passes("user1", false) {
		t.Errorf("expected user1 to pass")
	}
	if !setFilter.Passes("user2", false) {
		t.Errorf("expected user2 to pass")
	}
	if setFilter.Passes("user3", false) {
		t.Errorf("did not expect user3 to pass")
	}
	if setFilter.Passes("*", false) {
		t.Errorf("did not expect '*' to pass when passStar is false")
	}
	if !setFilter.Passes("*", true) {
		t.Errorf("expected '*' to pass when passStar is true")
	}
}

func TestStringFilterFilterSlice(t *testing.T) {
	var nilFilter *StringFilter
	if res := nilFilter.FilterSlice([]string{"a", "b"}, false); res != nil {
		t.Errorf("expected nil filter to return nil, got %v", res)
	}

	setFilter := &StringFilter{
		Literals: sets.New("a", "c"),
	}

	if res := setFilter.FilterSlice(nil, false); res != nil {
		t.Errorf("expected nil slice to return nil, got %v", res)
	}
	if res := setFilter.FilterSlice([]string{}, false); res != nil {
		t.Errorf("expected empty slice to return nil, got %v", res)
	}

	candidates := []string{"a", "b", "c", "*", "d"}
	resWithoutStar := setFilter.FilterSlice(candidates, false)
	expectedWithoutStar := []string{"a", "c"}
	if !reflect.DeepEqual(resWithoutStar, expectedWithoutStar) {
		t.Errorf("expected %v, got %v", expectedWithoutStar, resWithoutStar)
	}

	resWithStar := setFilter.FilterSlice(candidates, true)
	expectedWithStar := []string{"a", "c", "*"}
	if !reflect.DeepEqual(resWithStar, expectedWithStar) {
		t.Errorf("expected %v, got %v", expectedWithStar, resWithStar)
	}
}
