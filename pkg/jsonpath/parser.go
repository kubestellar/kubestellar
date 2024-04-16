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

package jsonpath

import (
	"fmt"
)

// This file has a parser for JSONPath queries.
// This is informed by RFC 9535 (JSONPath) and RFC 8259 (JSON).
// The grammar can be parsed by a recursive-descent parser.

type Parser struct {
	Lexer    *Lexer
	TokenPos int
	Token    Token
}

func NewParser(lexer *Lexer) (*Parser, error) {
	parser := &Parser{
		Lexer:    lexer,
		TokenPos: lexer.chrPos,
	}
	var err error
	parser.Token, err = lexer.Next()
	return parser, err
}

var rootIdentifier = Token{TokenIdentifier, "$"}

func (parser *Parser) ParseQuery() (Query, error) {
	if parser.Token != rootIdentifier {
		return nil, fmt.Errorf("syntax error at %d: expected root-identifier but got %#v", parser.TokenPos, parser.Token)
	}
	if err := parser.Advance(); err != nil {
		return nil, err
	}
	segs, err := parser.ParseSegments()
	return Query(segs), err
}

func (parser *Parser) ParseSegments() ([]Segment, error) {
	ans := []Segment{}
	for {
		if err := parser.SkipSpace(); err != nil {
			return ans, err
		}
		if !(parser.Token.Type == TokenSpecial && (parser.Token.Value == "[" ||
			parser.Token.Value == "." ||
			parser.Token.Value == "..")) {
			return ans, nil
		}
		if seg, err := parser.ParseSegment(); err != nil {
			return ans, err
		} else {
			ans = append(ans, seg)
		}
	}
}

func (parser *Parser) ParseSegment() (seg Segment, err error) {
	if err = parser.SkipSpace(); err != nil {
		return
	}
	if parser.Token == (Token{TokenSpecial, ".."}) {
		seg.Recurse = true
		if err = parser.Advance(); err != nil {
			return
		}
		if parser.Token == (Token{TokenSpecial, "["}) {
			seg.Selectors, err = parser.ParseBracketedSelection()
		} else if parser.Token == (Token{TokenSpecial, "*"}) {
			seg.Selectors = []Selector{{IsWildcard: true}}
			if err = parser.Advance(); err != nil {
				return
			}
		} else if parser.Token.Type == TokenIdentifier {
			name := parser.Token.Value.(string)
			seg.Selectors = []Selector{{Name: &name}}
			if err = parser.Advance(); err != nil {
				return
			}
		} else {
			err = fmt.Errorf("syntax error at %d: unexpected token after dotdot: %#v", parser.TokenPos, parser.Token)
		}
	} else if parser.Token == (Token{TokenSpecial, "["}) {
		seg.Selectors, err = parser.ParseBracketedSelection()
	} else if parser.Token == (Token{TokenSpecial, "."}) {
		if err = parser.Advance(); err != nil {
			return
		}
		if parser.Token == (Token{TokenSpecial, "*"}) {
			seg.Selectors = []Selector{{IsWildcard: true}}
			if err = parser.Advance(); err != nil {
				return
			}
		} else if parser.Token.Type == TokenIdentifier {
			name := parser.Token.Value.(string)
			seg.Selectors = []Selector{{Name: &name}}
			if err = parser.Advance(); err != nil {
				return
			}
		} else {
			err = fmt.Errorf("syntax error at %d: unexpected token after dot: %#v", parser.TokenPos, parser.Token)
		}
	} else {
		err = fmt.Errorf("syntax error at %d: unexpected token starting child-segment: %#v", parser.TokenPos, parser.Token)
	}
	return
}

func (parser *Parser) SkipSpace() (err error) {
	if parser.Token.Type == TokenSpace {
		err = parser.Advance()
	}
	return
}

func (parser *Parser) Advance() (err error) {
	parser.TokenPos = parser.Lexer.chrPos
	parser.Token, err = parser.Lexer.Next()
	return
}

func (parser *Parser) AdvanceSkipSpace() (err error) {
	if err = parser.Advance(); err != nil {
		return err
	}
	return parser.SkipSpace()
}

func (parser *Parser) ParseBracketedSelection() (sels []Selector, err error) {
	// parser.Token is opening bracket
	if err = parser.Advance(); err != nil {
		return
	}
	first := true
	for {
		if err = parser.SkipSpace(); err != nil {
			return
		}
		if parser.Token == (Token{TokenSpecial, "]"}) {
			return sels, parser.Advance()
		}
		if first {
			first = false
		} else if parser.Token == (Token{TokenSpecial, ","}) {
			if err = parser.AdvanceSkipSpace(); err != nil {
				return
			}
		} else {
			return sels, fmt.Errorf("syntax error at %d: got token %#v instead of comma between selectors", parser.TokenPos, parser.Token)
		}
		var sel Selector
		sel, err = parser.ParseSelector()
		sels = append(sels, sel)
		if err != nil {
			return
		}
	}
}

func (parser *Parser) ParseSelector() (sel Selector, err error) {
	if parser.Token.Type == TokenString {
		name := parser.Token.Value.(string)
		return Selector{Name: &name}, parser.Advance()
	}
	if parser.Token == (Token{TokenSpecial, "*"}) {
		return Selector{IsWildcard: true}, parser.Advance()
	}
	if parser.Token.Type == TokenNumber || parser.Token == (Token{TokenSpecial, ":"}) {
		// slice-selector or index-selector
		var start, end, step *int
		if parser.Token.Type == TokenNumber {
			startI := int(parser.Token.Value.(float64))
			start = &startI
			if err = parser.AdvanceSkipSpace(); err != nil {
				return
			}
		}
		if parser.Token != (Token{TokenSpecial, ":"}) {
			return Selector{Index: start}, nil
		}
		if err = parser.AdvanceSkipSpace(); err != nil {
			return
		}
		if parser.Token.Type == TokenNumber {
			endI := int(parser.Token.Value.(float64))
			end = &endI
			if err = parser.AdvanceSkipSpace(); err != nil {
				return
			}
		}
		if parser.Token == (Token{TokenSpecial, ":"}) {
			if err = parser.AdvanceSkipSpace(); err != nil {
				return
			}
			if parser.Token.Type == TokenNumber {
				stepI := int(parser.Token.Value.(float64))
				step = &stepI
				if err = parser.Advance(); err != nil {
					return
				}
			}
		}
		return Selector{Slice: &SliceSelector{start, end, step}}, nil
	}
	if parser.Token != (Token{TokenSpecial, "?"}) {
		return sel, fmt.Errorf("syntax error at %d: token %#v does not start a selector", parser.TokenPos, parser.Token)
	}
	if err = parser.AdvanceSkipSpace(); err != nil {
		return
	}
	var logicalOr LogicalOrExpr
	logicalOr, err = parser.ParseLogicalOrExpr()
	sel = Selector{Filter: &logicalOr}
	return
}

func (parser *Parser) ParseLogicalOrExpr() (logicalOr LogicalOrExpr, err error) {
	for {
		var term LogicalAndExpr
		if term, err = parser.ParseLogicalAndExpr(); err != nil {
			return
		}
		logicalOr.Terms = append(logicalOr.Terms, term)
		if err = parser.SkipSpace(); err != nil {
			return
		}
		if parser.Token != (Token{TokenSpecial, "||"}) {
			return
		}
		if err = parser.AdvanceSkipSpace(); err != nil {
			return
		}
	}
}

func (parser *Parser) ParseLogicalAndExpr() (logicalAnd LogicalAndExpr, err error) {
	for {
		var basic BasicExpr
		if basic, err = parser.ParseBasicExpr(); err != nil {
			return
		}
		logicalAnd.Factors = append(logicalAnd.Factors, basic)
		if err = parser.SkipSpace(); err != nil {
			return
		}
		if parser.Token != (Token{TokenSpecial, "&&"}) {
			return
		}
		if err = parser.AdvanceSkipSpace(); err != nil {
			return
		}
	}
}

func (parser *Parser) ParseBasicExpr() (basic BasicExpr, err error) {
	startPos := parser.TokenPos
	if parser.Token == (Token{TokenSpecial, "!"}) {
		basic.Negate = true
		if err = parser.AdvanceSkipSpace(); err != nil {
			return
		}
	}
	if parser.Token == (Token{TokenSpecial, "("}) {
		if err = parser.AdvanceSkipSpace(); err != nil {
			return
		}
		var inner LogicalOrExpr
		if inner, err = parser.ParseLogicalOrExpr(); err != nil {
			return
		}
		basic.Parenthesized = &inner
		if parser.Token == (Token{TokenSpecial, ")"}) {
			return basic, parser.Advance()
		}
		return basic, fmt.Errorf("syntax error at %d: missing closing paren, got %#v instead", parser.TokenPos, parser.Token)
	}
	// It is either comparison-expr or test-expr
	var left, right Comparable
	if left, err = parser.ParseFilterQueryOrComparable(); err != nil {
		return
	}
	if err = parser.SkipSpace(); err != nil {
		return
	}
	if !(parser.Token.Type == TokenSpecial && isComparisonOp(parser.Token.Value.(string))) {
		// It is a test-expr
		if left.Literal != nil {
			return basic, fmt.Errorf("syntax error at %d: literal %#v instead of filter-query", startPos, *left.Literal)
		}
		return BasicExpr{Test: &FilterQuery{Absolute: left.Absolute, Relative: left.Relative}}, nil
	}
	// It is a comparison-expr
	op := ComparisonOp(parser.Token.Value.(string))
	if basic.Negate {
		return basic, fmt.Errorf("syntax error at %d: negation not allowed on a comparison-expr", startPos)
	}
	if left.Absolute != nil && !isSingular(*left.Absolute) || left.Relative != nil && !isSingular(*left.Relative) {
		return basic, fmt.Errorf("syntax error at %d: comparison-expr but LHS segments are not all singular", startPos)
	}
	if err = parser.AdvanceSkipSpace(); err != nil {
		return
	}
	if right, err = parser.ParseFilterQueryOrComparable(); err != nil {
		return
	}
	if right.Absolute != nil && !isSingular(*right.Absolute) || right.Relative != nil && !isSingular(*right.Relative) {
		return basic, fmt.Errorf("syntax error at %d: comparison-expr but RHS segments are not all singular", startPos)
	}
	return BasicExpr{Compare: &ComparisonExpr{Left: left, Right: right, Op: op}}, nil
}

func isComparisonOp(op string) bool {
	return op == "==" || op == "!="
}

func isSingular(segs []Segment) bool {
	for _, seg := range segs {
		if seg.Recurse {
			return false
		}
		for _, sel := range seg.Selectors {
			if !(sel.Name != nil || sel.Index != nil) {
				return false
			}
		}
	}
	return true
}

func (parser *Parser) ParseFilterQueryOrComparable() (ans Comparable, err error) {
	if parser.Token.Type == TokenString || parser.Token.Type == TokenNumber {
		var val any = parser.Token.Value
		return Comparable{Literal: &val}, parser.Advance()
	}
	if parser.Token.Type == TokenKeyword {
		var val any
		switch parser.Token.Value {
		case "true":
			val = true
		case "false":
			val = false
		case "null":
		}
		return Comparable{Literal: &val}, parser.Advance()
	}
	if parser.Token == (Token{TokenIdentifier, "$"}) {
		query, err := parser.ParseQuery()
		return Comparable{Absolute: &query}, err
	}
	if parser.Token == (Token{TokenIdentifier, "@"}) {
		if err = parser.Advance(); err != nil {
			return
		}
		segs, err := parser.ParseSegments()
		return Comparable{Relative: &segs}, err
	}
	return ans, fmt.Errorf("syntax error at %d: token %#v does not start a filer-query or compatable", parser.TokenPos, parser.Token)
}
