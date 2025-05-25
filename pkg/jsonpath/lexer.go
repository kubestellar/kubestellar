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
	"strings"

	"github.com/dop251/goja"
	js_ast "github.com/dop251/goja/ast"
)

// Query represents a parsed query in a very restricted subset of JSONPath (RFC 9535).
// The only supported segment functionality is selecting one definite member of a JSON object.
type Query []string

// ParseQuery parses a very restricted form of JSONPath expression into a Query.
// This parser only accepts segments of the form `.name` or `[string]`.
// A string must be enclosed with double-quotes (`"`).
// No whitespace is allowed.
func ParseQuery(queryS string) (Query, error) {
	lexer, err := NewLexer(queryS, 0)
	if err != nil {
		return nil, err
	}
	query, err := lexer.ScanQuery()
	if err != nil {
		return query, err
	}
	if !lexer.eof {
		return query, fmt.Errorf("syntax error at %d: extra junk %q after last segment", lexer.chrPos, lexer.chr)
	}
	return query, nil
}

// Lexer supports scanning through a string.
// Lexer is intended ONLY for serialized usage, not concurrency.
type Lexer struct {
	source string
	reader io.RuneReader

	chr     rune // next rune to process
	chrPos  int  // index of start of chr
	nextPos int  // index after chr
	eof     bool // no more to process
}

func NewLexer(source string, startPos int) (*Lexer, error) {
	lxr := &Lexer{
		source:  source,
		reader:  strings.NewReader(source),
		chrPos:  startPos,
		nextPos: startPos,
		eof:     startPos >= len(source),
	}
	err := lxr.advance()
	return lxr, err
}

// GetPosition returns the postion of the character that the Lexer is currently looking at
// and whether the Lexer is looking at EOF.
func (lxr *Lexer) GetPosition() (int, bool) { return lxr.chrPos, lxr.eof }

// ScanQuery consumes a Query from the Lexer.
// As long a query as is available is consumed.
// The Lexer is left looking at the first character after the Query
// or EOF.
func (lxr *Lexer) ScanQuery() (Query, error) {
	query := Query{}
	if lxr.chr != '$' {
		return query, fmt.Errorf("syntax error at %d: missing root identifier (dollar sign)", lxr.chrPos)
	}
	if err := lxr.advance(); err != nil {
		return query, err
	}
	for !lxr.eof {
		if lxr.chr == '.' {
			if err := lxr.advance(); err != nil {
				return query, err
			}
			if !isNameFirst(lxr.chr) {
				return query, fmt.Errorf("syntax error at %d: expected member-name-shorthand, got %q", lxr.chrPos, lxr.chr)
			}
			if next, err := lxr.nextIdentifier(); err != nil {
				return query, err
			} else {
				query = append(query, next)
			}
		} else if lxr.chr == '[' {
			if err := lxr.advance(); err != nil {
				return query, err
			}
			if lxr.chr != '"' {
				return query, fmt.Errorf("syntax error at %d: expected open quote, got %q", lxr.chrPos, lxr.chr)
			}
			if next, err := lxr.nextString(); err != nil {
				return query, err
			} else {
				query = append(query, next)
			}
			if lxr.chr != ']' {
				return query, fmt.Errorf("syntax error at %d: missing close bracket, got %q", lxr.chrPos, lxr.chr)
			}
			if err := lxr.advance(); err != nil {
				return query, err
			}
		} else {
			break
		}
	}
	return query, nil
}

func (lxr *Lexer) advance() error {
	if lxr.eof {
		return io.EOF
	}
	lxr.chrPos = lxr.nextPos
	startPos := lxr.chrPos
	var err error
	var size int
	lxr.chr, size, err = lxr.reader.ReadRune()
	lxr.nextPos = lxr.chrPos + size
	if err == io.EOF {
		lxr.eof = true
	} else if err != nil {
		return fmt.Errorf("run reader error at %d: %w", startPos, err)
	}
	return nil
}

func (lxr *Lexer) nextString() (string, error) {
	startPos := lxr.chrPos
	close := lxr.chr
	for {
		if err := lxr.advance(); err != nil {
			return "", err
		}
		if lxr.eof {
			return lxr.source[startPos:], fmt.Errorf("syntax error at %d: missing close quote", lxr.chrPos)
		}
		if lxr.chr == close {
			if err := lxr.advance(); err != nil {
				return "", err
			}
			stringSrc := lxr.source[startPos:lxr.chrPos]
			stringProg, err := goja.Parse("lit", stringSrc)
			if err != nil {
				return stringSrc, fmt.Errorf("lexical error at %d from goja: %w", lxr.chrPos, err)
			}
			if len(stringProg.Body) != 1 {
				return stringSrc, fmt.Errorf("parsed goja body len is %d", len(stringProg.Body))
			}
			stringExpr, ok := stringProg.Body[0].(*js_ast.ExpressionStatement)
			if !ok {
				return stringSrc, fmt.Errorf("goja statement is not an expression statement, it is %#+v, at %T", stringProg.Body[0], stringProg.Body[0])
			}
			stringLit, ok := stringExpr.Expression.(*js_ast.StringLiteral)
			if !ok {
				return stringSrc, fmt.Errorf("goja statement is not a string literal, it is %#+v, at %T", stringProg.Body[0], stringProg.Body[0])
			}
			return stringLit.Value.String(), nil
		}
		if lxr.chr == '\\' {
			if err := lxr.advance(); err != nil {
				return "", err
			}
			if lxr.eof {
				return lxr.source[startPos:], fmt.Errorf("syntax error at %d: escape-EOF", lxr.chrPos)
			}
		}
	}
}

func (lxr *Lexer) nextIdentifier() (string, error) {
	startPos := lxr.chrPos
	for !lxr.eof {
		if err := lxr.advance(); err != nil {
			return "", err
		}
		if !isNameChar(lxr.chr) {
			break
		}
	}
	return lxr.source[startPos:lxr.chrPos], nil
}

func isNameFirst(r rune) bool {
	return isAlpha(r) || r == '_' || 0x80 <= r && r <= 0xD7FF || 0xE000 <= r && r <= 0x10FFFF
}

func isNameChar(r rune) bool {
	return isNameFirst(r) || isDigit(r)
}

func isAlpha(r rune) bool {
	return 'A' <= r && r <= 'Z' || 'a' <= r && r <= 'z'
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}
