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
	"strings"

	"github.com/spf13/pflag"

	"k8s.io/apimachinery/pkg/util/sets"
)

// StringFilterOptions is a command-line oriented precursor to a StringFilter.
type StringFilterOptions struct {
	SpacesToo bool
	Literals  []string
}

func NewStringFilterOptions(literals ...string) StringFilterOptions {
	return StringFilterOptions{Literals: literals}
}

func (sfo StringFilterOptions) SeparateBySpacesToo() StringFilterOptions {
	sfo.SpacesToo = true
	return sfo
}

func (sfo *StringFilterOptions) AddToFlags(flagSet *pflag.FlagSet, flagName, what string) {
	sep := "comma"
	if sfo.SpacesToo {
		sep = "comma and/or space"
	}
	flagSet.StringSliceVar(&sfo.Literals, flagName, sfo.Literals, sep+" separated list of "+what)
}

// StringFilter is something that passes or rejects strings
type StringFilter struct {
	AllPass  bool
	Literals sets.Set[string]
}

func (sfo StringFilterOptions) ToFilter() (StringFilter, bool) {
	gotSome := len(sfo.Literals) > 0
	ansSet := sets.New[string]()
	for _, piece := range sfo.Literals {
		if len(piece) == 0 {
			continue
		}
		if !sfo.SpacesToo {
			ansSet.Insert(piece)
			continue
		}
		elts := strings.Split(piece, " ")
		for _, elt := range elts {
			if len(elt) != 0 {
				ansSet.Insert(elt)
			}
		}
	}
	if ansSet.Has("*") {
		return StringFilter{AllPass: true}, gotSome
	}
	return StringFilter{Literals: ansSet}, gotSome
}

func (sf *StringFilter) Passes(candidate string, passStar bool) bool {
	return sf.AllPass || passStar && candidate == "*" || sf.Literals.Has(candidate)
}

func (sf *StringFilter) FilterSlice(candidates []string, passStar bool) []string {
	ans := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if sf.Passes(candidate, passStar) {
			ans = append(ans, candidate)
		}
	}
	return ans
}
