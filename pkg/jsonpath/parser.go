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
	"fmt"
	"io"
)

// This file parses a subset of JSONPath with the following syntax.

// Path = `$`
// Path = Path Selector
// Path = `..` `*`

// Selector = `.` Identifier
// Selector = `.` Integer
// Selector = `.` `*`
// Selector = `[` List `]`
// Selector = `[` Range `]`
// Selector = `[` `*` `]`
// Selector = `..`

// List = StringList
// List = IntList
// StringList = String
// StringList = StringList `,` String
// IntList = Integer
// IntList = Intlist, Integer

// Range = [ Integer ] `:` [ Integer ] [ `:` [ Integer ] ]

// Parsed is the result of parsing a JSONPath expression.
type Parsed []Selector

type SelectorType string

const (
	SelectorName       SelectorType = "Name"
	SelectorRange      SelectorType = "Range"
	SelectorList       SelectorType = "List"
	SelectorRecurse    SelectorType = "Recurse" // every descendant object
	SelectorEveryChild SelectorType = "EveryChild"
)

type Selector struct {
	Type  SelectorType
	Name  string
	Range *Range
	List  []Selector
}

type Range struct {
	start    int
	afterEnd *int
	stride   int
}

func (left Parsed) Equals(right Parsed) bool {
	if len(left) != len(right) {
		return false
	}
	for idx, leftSel := range left {
		if !leftSel.Equals(right[idx]) {
			return false
		}
	}
	return true
}

func (left Selector) Equals(right Selector) bool {
	if left.Type != right.Type || left.Name != right.Name {
		return false
	}
	if !left.Range.Equals(right.Range) {
		return false
	}
	if !Parsed(left.List).Equals(Parsed(right.List)) {
		return false
	}
	return true
}

func (left *Range) Equals(right *Range) bool {
	if left == nil {
		return right == nil
	}
	if right == nil {
		return false
	}
	if left.start != right.start || left.stride != right.stride {
		return false
	}
	return ptrEqual(left.afterEnd, right.afterEnd)
}

func ParseString(source string) (Parsed, error) {
	lxr := NewLexer(source)
	return Parse(lxr)
}

func Parse(lxr *Lexer) (Parsed, error) {
	tok, val, err := lxr.Next()
	if err != nil {
		return nil, err
	}
	if tok != TokenIdentifier || val != "$" {
		return nil, fmt.Errorf("syntax error: did not start with $")
	}
	parsed := Parsed{}
	nextTakeStar := false
	for !lxr.eof {
		takeStar := nextTakeStar
		nextTakeStar = false
		startPos := lxr.chrPos
		tok, val, err := lxr.Next()
		if err == io.EOF {
			return parsed, nil
		}
		if tok == TokenSpecial && val == "." {
			namePos := lxr.chrPos
			tok, val, err := lxr.Next()
			if err != nil {
				return parsed, err
			}
			if tok == TokenSpecial && val == "*" {
				parsed = append(parsed, Selector{Type: SelectorEveryChild})
				continue
			}
			if tok == TokenIdentifier {
				parsed = append(parsed, Selector{
					Type: SelectorName,
					Name: val.(string),
				})
				continue
			}
			if tok == TokenNumber {
				num1, ok := val.(int64)
				if !ok {
					return parsed, fmt.Errorf("syntax error at position %d: expected int but got float %v", namePos, val)
				}
				parsed = append(parsed, Selector{Type: SelectorRange, Range: rangeOf1(int(num1))})
				continue
			}
			return parsed, fmt.Errorf("syntax error at position %d: expected star or identifer but got tok=%v val=%v", namePos, tok, val)
		} else if tok == TokenSpecial && val == "[" {
			subPos := lxr.chrPos
			tok, val, err := lxr.Next()
			if err != nil {
				return parsed, err
			}
			if tok == TokenSpecial && val == "*" {
				parsed = append(parsed, Selector{Type: SelectorEveryChild})
				closePos := lxr.chrPos
				tok, val, err := lxr.Next()
				if err != nil {
					return parsed, err
				}
				if tok != TokenSpecial || val != "]" {
					return parsed, fmt.Errorf("syntax error at position %d: expected close bracket but got tok=%v val=%v", closePos, tok, val)
				}
				continue
			}
			if tok == TokenSpecial && val == ":" {
				sel, err := parseRange(lxr, 0)
				if err != nil {
					return parsed, err
				}
				parsed = append(parsed, sel)
				continue
			}
			if tok == TokenNumber {
				sel, err := parseNumericSubscripts(lxr, subPos, val)
				if err != nil {
					return parsed, err
				}
				parsed = append(parsed, sel)
				continue
			} else if tok == TokenString {
				sel, err := parseStringSubscripts(lxr, val.(string))
				if err != nil {
					return parsed, err
				}
				parsed = append(parsed, sel)
				continue
			} else {
				return parsed, fmt.Errorf("syntax error at position %d: expected number or string but got tok=%v val=%v", subPos, tok, val)
			}
		} else if tok == TokenSpecial && val == ".." {
			parsed = append(parsed, Selector{Type: SelectorRecurse})
			nextTakeStar = true
		} else if takeStar && tok == TokenSpecial && val == "*" {
			parsed = append(parsed, Selector{Type: SelectorEveryChild})
		} else {
			return parsed, fmt.Errorf("syntax error at position %d: tok=%v val=%v", startPos, tok, val)
		}
	}
	return parsed, nil
}

// parseStringSubscripts finishes parsing `[string,string,<and so on>]`
// after having consumed the first string
func parseStringSubscripts(lxr *Lexer, str1 string) (Selector, error) {
	subs := []Selector{{Type: SelectorName, Name: str1}}
	for {
		nextPos := lxr.chrPos
		tok, val, err := lxr.Next()
		if err != nil {
			return Selector{}, err
		}
		if tok == TokenSpecial && val == "]" {
			if len(subs) == 1 {
				return subs[0], nil
			}
			return Selector{Type: SelectorList, List: subs}, nil
		}
		if tok != TokenSpecial || val != "," {
			return Selector{}, fmt.Errorf("syntax error at position %d: expected comma or close bracket but got tok=%v val=%v", nextPos, tok, val)
		}
		nextPos = lxr.chrPos
		tok, val, err = lxr.Next()
		if err != nil {
			return Selector{}, err
		}
		if tok != TokenString {
			return Selector{}, fmt.Errorf("syntax error at position %d: expected string but got tok=%v val=%v", nextPos, tok, val)
		}
		subs = append(subs, Selector{Type: SelectorName, Name: val.(string)})
	}
}

// parseNumericSubscripts finishes parsing a square bracket selector
// after having consumed the left bracket and a number.
func parseNumericSubscripts(lxr *Lexer, sub1Pos int, numVal JPLiteralValue) (Selector, error) {
	num1, ok := numVal.(int64)
	num1i := int(num1)
	if !ok {
		return Selector{}, fmt.Errorf("syntax error at position %d: expected int but got float", sub1Pos)
	}
	nextPos := lxr.chrPos
	tok, val, err := lxr.Next()
	if err != nil {
		return Selector{}, err
	}
	if tok == TokenSpecial && val == ":" {
		return parseRange(lxr, num1i)
	}
	subs := []Selector{{Type: SelectorRange, Range: rangeOf1(num1i)}}
	for { // Consume `,number` until `]`
		if tok == TokenSpecial && val == "]" {
			if len(subs) == 1 {
				return subs[0], nil
			}
			return Selector{Type: SelectorList, List: subs}, nil
		}
		if tok != TokenSpecial || val != "," {
			return Selector{}, fmt.Errorf("syntax error at position %d: expected comma or close bracket but got tok=%v val=%v", nextPos, tok, val)
		}
		nextPos = lxr.chrPos
		tok, val, err = lxr.Next()
		if err != nil {
			return Selector{}, err
		}
		if tok != TokenNumber {
			return Selector{}, fmt.Errorf("syntax error at position %d: expected number but got tok=%v val=%v", nextPos, tok, val)
		}
		num, ok := val.(int64)
		if !ok {
			return Selector{}, fmt.Errorf("syntax error at position %d: expected int but got %v", nextPos, val)
		}
		numi := int(num)
		subs = append(subs, Selector{Type: SelectorRange, Range: rangeOf1(numi)})
		nextPos = lxr.chrPos
		tok, val, err = lxr.Next()
		if err != nil {
			return Selector{}, err
		}
	}
}

// parseRange finishes parsing a `[start:end:stride]` selector.
// The first colon has been consumed.
func parseRange(lxr *Lexer, start int) (Selector, error) {
	sel := Selector{Type: SelectorRange, Range: &Range{start, nil, 1}}
	endPos := lxr.chrPos
	tok, val, err := lxr.Next()
	if err != nil {
		return sel, err
	}
	col2pos := endPos
	if tok == TokenNumber {
		afterEnd, ok := val.(int64)
		if !ok {
			return sel, fmt.Errorf("syntax error at %d: expected int but got float %v", endPos, val)
		}
		sel.Range.afterEnd = ptr(int(afterEnd))
		col2pos = lxr.chrPos
		tok, val, err = lxr.Next()
		if err != nil {
			return sel, err
		}
	}
	if tok == TokenSpecial && val == "]" {
		return sel, nil
	}
	if tok != TokenSpecial || val != ":" {
		return sel, fmt.Errorf("syntax error at %d: expected colon or close bracket but got tok=%v val=%v", col2pos, tok, val)
	}
	stridePos := lxr.chrPos
	tok, val, err = lxr.Next()
	if err != nil {
		return sel, err
	}
	if tok == TokenSpecial && val == "]" {
		return sel, nil
	}
	if tok != TokenNumber {
		return sel, fmt.Errorf("syntax error at %d: expected number but got tok=%v val=%v", stridePos, tok, val)
	}
	stride64, ok := val.(int64)
	if !ok {
		return sel, fmt.Errorf("syntax error at %d: expected int but got float %v", stridePos, val)
	}
	sel.Range.stride = int(stride64)
	closePos := lxr.chrPos
	tok, val, err = lxr.Next()
	if err != nil {
		return sel, err
	}
	if tok != TokenSpecial || val != "]" {
		return sel, fmt.Errorf("syntax error at %d: expected close bracket but got tok=%v val=%v", closePos, tok, val)
	}
	return sel, nil
}

func rangeOf1(index int) *Range {
	return &Range{index, ptr(index + 1), 1}
}

func ptr[T any](val T) *T {
	return &val
}

func ptrEqual[T comparable](left, right *T) bool {
	if left == nil {
		return right == nil
	}
	if right == nil {
		return false
	}
	return (*left) == (*right)
}
