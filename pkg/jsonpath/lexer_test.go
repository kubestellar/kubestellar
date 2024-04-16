/*
Copyright 2023 The KubeStellar Authors.

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

package jsonpath

import (
	"strconv"
	"testing"
)

func TestLexer(t *testing.T) {
CaseLoop:
	for _, testCase := range []struct {
		source  string
		results []Token
		goodEnd func(Token, error) bool
	}{
		{`'xyz`,
			nil,
			errIs(strconv.ErrSyntax)},
		{`'ab\\cd','ab\`,
			[]Token{{TokenString, `ab\cd`}, {TokenSpecial, ","}},
			errIs(strconv.ErrSyntax)},
		{`['foo']`,
			[]Token{{TokenSpecial, "["}, {TokenString, "foo"}, {TokenSpecial, "]"}},
			cleanEOF},
		{`i`,
			[]Token{{TokenIdentifier, "i"}}, cleanEOF},
		{`ijk`,
			[]Token{{TokenIdentifier, "ijk"}}, cleanEOF},
		{`i_k`,
			[]Token{{TokenIdentifier, "i_k"}}, cleanEOF},
		{`0`,
			[]Token{{TokenNumber, float64(0)}}, cleanEOF},
		{`.0`,
			[]Token{{TokenSpecial, "."}, {TokenNumber, float64(0)}}, cleanEOF},
		{`0.1`,
			[]Token{{TokenNumber, float64(0.1)}}, cleanEOF},
		{`[1,2:3]`,
			[]Token{{TokenSpecial, "["}, {TokenNumber, float64(1)}, {TokenSpecial, ","},
				{TokenNumber, float64(2)}, {TokenSpecial, ":"}, {TokenNumber, float64(3)}, {TokenSpecial, "]"}},
			cleanEOF},
		{`$.abc[*]..xyz['foo\'bar']`,
			[]Token{{TokenIdentifier, "$"}, {TokenSpecial, "."}, {TokenIdentifier, "abc"}, {TokenSpecial, "["},
				{TokenSpecial, "*"}, {TokenSpecial, "]"}, {TokenSpecial, ".."}, {TokenIdentifier, "xyz"},
				{TokenSpecial, "["}, {TokenString, "foo'bar"}, {TokenSpecial, "]"}},
			cleanEOF},
		{`@['a.b']`,
			[]Token{{TokenIdentifier, "@"}, {TokenSpecial, "["}, {TokenString, "a.b"}, {TokenSpecial, "]"}}, cleanEOF},
		{`$ ..[?@.x == 'it'||foo!=42]`,
			[]Token{{TokenIdentifier, "$"}, {TokenSpace, " "}, {TokenSpecial, ".."}, {TokenSpecial, "["},
				{TokenSpecial, "?"}, {TokenIdentifier, "@"}, {TokenSpecial, "."}, {TokenIdentifier, "x"},
				{TokenSpace, " "}, {TokenSpecial, "=="}, {TokenSpace, " "}, {TokenString, "it"},
				{TokenSpecial, "||"}, {TokenIdentifier, "foo"}, {TokenSpecial, "!="}, {TokenNumber, float64(42)},
				{TokenSpecial, "]"}}, cleanEOF},
	} {
		lxr := NewLexer(testCase.source)
		for idx, good := range testCase.results {
			token, err := lxr.Next()
			if err != nil {
				t.Errorf("For source %q, Next %d returned error %#+v", testCase.source, idx, err)
				continue CaseLoop
			}
			if token != good {
				t.Errorf("For source %q, Next %d returned token %#v but expected token %#v", testCase.source, idx, token, good)
				continue CaseLoop
			}
		}
		token, err := lxr.Next()
		if !testCase.goodEnd(token, err) {
			t.Errorf("For source %q, Next %d returned wrong token=%#v, err=%#+v", testCase.source, len(testCase.results), token, err)
		} else {
			t.Logf("Success for source %q", testCase.source)
		}
	}
}

func errIs(expected error) func(Token, error) bool {
	return func(tok Token, err error) bool {
		return err == expected
	}
}

func cleanEOF(tok Token, err error) bool {
	return tok.Type == TokenEOF && err == nil
}
