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

type LexGood struct {
	tok Token
	val JPLiteralValue
}

func TestLexer(t *testing.T) {
CaseLoop:
	for _, testCase := range []struct {
		source  string
		results []LexGood
		goodEnd func(Token, JPLiteralValue, error) bool
	}{
		{`'xyz`,
			nil,
			errIs(strconv.ErrSyntax)},
		{`'ab\\cd','ab\`,
			[]LexGood{{TokenString, `ab\cd`}, {TokenSpecial, ","}},
			errIs(strconv.ErrSyntax)},
		{`['foo']`,
			[]LexGood{{TokenSpecial, "["}, {TokenString, "foo"}, {TokenSpecial, "]"}},
			cleanEOF},
		{`i`,
			[]LexGood{{TokenIdentifier, "i"}}, cleanEOF},
		{`ijk`,
			[]LexGood{{TokenIdentifier, "ijk"}}, cleanEOF},
		{`i_k`,
			[]LexGood{{TokenIdentifier, "i_k"}}, cleanEOF},
		{`0`,
			[]LexGood{{TokenNumber, int64(0)}}, cleanEOF},
		{`.0`,
			[]LexGood{{TokenSpecial, "."}, {TokenNumber, int64(0)}}, cleanEOF},
		{`0.1`,
			[]LexGood{{TokenNumber, int64(0)}, {TokenSpecial, "."}, {TokenNumber, int64(1)}}, cleanEOF},
		{`[1,2:3]`,
			[]LexGood{{TokenSpecial, "["}, {TokenNumber, int64(1)}, {TokenSpecial, ","},
				{TokenNumber, int64(2)}, {TokenSpecial, ":"}, {TokenNumber, int64(3)}, {TokenSpecial, "]"}},
			cleanEOF},
		{`$.abc[*]..xyz['foo\'bar']`,
			[]LexGood{{TokenIdentifier, "$"}, {TokenSpecial, "."}, {TokenIdentifier, "abc"}, {TokenSpecial, "["},
				{TokenSpecial, "*"}, {TokenSpecial, "]"}, {TokenSpecial, ".."}, {TokenIdentifier, "xyz"},
				{TokenSpecial, "["}, {TokenString, "foo'bar"}, {TokenSpecial, "]"}},
			cleanEOF},
	} {
		lxr := NewLexer(testCase.source)
		for idx, good := range testCase.results {
			tok, val, err := lxr.Next()
			if err != nil {
				t.Errorf("For source %q, Next %d returned error %#+v", testCase.source, idx, err)
				continue CaseLoop
			}
			if tok != good.tok {
				t.Errorf("For source %q, Next %d returned tok=%v, val=%#+v but expected tok=%v", testCase.source, idx, tok, val, good.tok)
				continue CaseLoop
			}
			if val != good.val {
				t.Errorf("For source %q, Next %d returned tok=%v, val=%#+v but expected val=%#+v", testCase.source, idx, tok, val, good.val)
				continue CaseLoop
			}
		}
		tok, val, err := lxr.Next()
		if !testCase.goodEnd(tok, val, err) {
			t.Errorf("For source %q, Next %d returned wrong tok=%v, val=%#+v, err=%#+v", testCase.source, len(testCase.results), tok, val, err)
		} else {
			t.Logf("Success for source %q", testCase.source)
		}
	}
}

func errIs(expected error) func(Token, JPLiteralValue, error) bool {
	return func(tok Token, val JPLiteralValue, err error) bool {
		return err == expected
	}
}

func cleanEOF(tok Token, val JPLiteralValue, err error) bool {
	return tok == TokenEOF && err == nil
}
