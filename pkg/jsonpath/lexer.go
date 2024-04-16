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
	"strconv"
	"strings"

	"github.com/dop251/goja"
	js_ast "github.com/dop251/goja/ast"
)

// This file defines a lexer for JSONPath expressions.
// This is informed by RFC 9535 (JSONPath) and RFC 8259 (JSON).

type Token struct {
	Type  TokenType
	Value TokenValue
}

// TokenType is a type of token
type TokenType string

const (
	TokenEOF        TokenType = "EOF"
	TokenSpace      TokenType = "Space"
	TokenSpecial    TokenType = "Special"
	TokenString     TokenType = "String"
	TokenIdentifier TokenType = "ID" // `member-name-shorthand` of RFC 9535
	TokenNumber     TokenType = "Number"
	TokenKeyword    TokenType = "Keyword" // `true`, `false`, or `null`
)

type TokenValue interface{} // either string or float64

type Lexer struct {
	source string
	reader io.RuneReader

	chr     rune // next rune to process
	chrPos  int  // index of start of chr
	nextPos int  // index after chr
	eof     bool // no more to process
}

func NewLexer(source string) *Lexer {
	lxr := &Lexer{
		source: source,
		reader: strings.NewReader(source),
	}
	lxr.advance()
	return lxr
}

func (lxr *Lexer) Next() (Token, error) {
	if lxr.eof {
		return Token{TokenEOF, ""}, nil
	}
	chr := lxr.chr
	switch chr {
	case ' ', 9, 10, 13:
		return lxr.nextSpace()
	case '[', ']', '(', ')', '*', ',', ':', '?':
		if err := lxr.advance(); err != nil {
			return Token{TokenEOF, ""}, err
		}
		return Token{TokenSpecial, string(chr)}, nil
	case '.':
		if err := lxr.advance(); err != nil {
			return Token{TokenEOF, ""}, err
		}
		if lxr.chr == '.' {
			if err := lxr.advance(); err != nil {
				return Token{TokenEOF, ""}, err
			}
			return Token{TokenSpecial, ".."}, nil
		}
		return Token{TokenSpecial, "."}, nil
	case '$', '@':
		if err := lxr.advance(); err != nil {
			return Token{TokenEOF, ""}, err
		}
		return Token{TokenIdentifier, string(chr)}, nil
	case '|', '&', '=':
		if err := lxr.advance(); err != nil {
			return Token{TokenEOF, ""}, err
		}
		if lxr.chr != chr {
			return Token{TokenSpecial, ""}, fmt.Errorf("malformed symbol at %d: %q", lxr.chrPos, string(chr)+string(lxr.chr))
		}
		if err := lxr.advance(); err != nil {
			return Token{TokenEOF, ""}, err
		}
		return Token{TokenSpecial, string(chr) + string(chr)}, nil
	case '!':
		if err := lxr.advance(); err != nil {
			return Token{TokenEOF, ""}, err
		}
		if lxr.chr != '=' {
			return Token{TokenSpecial, string(chr)}, nil
		}
		if err := lxr.advance(); err != nil {
			return Token{TokenEOF, ""}, err
		}
		return Token{TokenSpecial, "!="}, nil
	}
	if chr == '-' || isDigit(chr) {
		return lxr.nextNumber()
	}
	if lxr.chr == '"' || lxr.chr == '\'' {
		return lxr.nextString()
	}
	if isNameFirst(chr) {
		return lxr.nextIdentifierOrKeyword()
	}
	return Token{TokenSpecial, ""}, fmt.Errorf("syntax error at pos %d char %q", lxr.chrPos, string(chr))
}

func (lxr *Lexer) advance() error {
	if lxr.eof {
		return io.EOF
	}
	lxr.chrPos = lxr.nextPos
	var err error
	var size int
	lxr.chr, size, err = lxr.reader.ReadRune()
	lxr.nextPos = lxr.chrPos + size
	if err == io.EOF {
		lxr.eof = true
	} else if err != nil {
		return err
	}
	return nil
}

func (lxr *Lexer) nextSpace() (Token, error) {
	startPos := lxr.chrPos
	for {
		if err := lxr.advance(); err != nil {
			return Token{TokenEOF, ""}, err
		}
		switch lxr.chr {
		case 9, 10, 13, ' ':
		default:
			return Token{TokenSpace, lxr.source[startPos:lxr.chrPos]}, nil
		}
	}
}

func (lxr *Lexer) nextString() (Token, error) {
	startPos := lxr.chrPos
	close := lxr.chr
	for {
		if err := lxr.advance(); err != nil {
			return Token{TokenEOF, ""}, err
		}
		if lxr.eof {
			return Token{TokenString, lxr.source[startPos:]}, strconv.ErrSyntax
		}
		if lxr.chr == close {
			if err := lxr.advance(); err != nil {
				return Token{TokenEOF, ""}, err
			}
			stringSrc := lxr.source[startPos:lxr.chrPos]
			stringProg, err := goja.Parse("lit", stringSrc)
			if err != nil {
				return Token{TokenString, stringSrc}, fmt.Errorf("lexical error at %d from goja: %w", lxr.chrPos, err)
			}
			if len(stringProg.Body) != 1 {
				return Token{TokenString, stringSrc}, fmt.Errorf("parsed goja body len is %d", len(stringProg.Body))
			}
			stringExpr, ok := stringProg.Body[0].(*js_ast.ExpressionStatement)
			if !ok {
				return Token{TokenString, stringSrc}, fmt.Errorf("goja statement is not an expression statement, it is %#+v, at %T", stringProg.Body[0], stringProg.Body[0])
			}
			stringLit, ok := stringExpr.Expression.(*js_ast.StringLiteral)
			if !ok {
				return Token{TokenString, stringSrc}, fmt.Errorf("goja statement is not a string literal, it is %#+v, at %T", stringProg.Body[0], stringProg.Body[0])
			}
			return Token{TokenString, stringLit.Value.String()}, nil
		}
		if lxr.chr == '\\' {
			if err := lxr.advance(); err != nil {
				return Token{TokenEOF, ""}, err
			}
			if lxr.eof {
				return Token{TokenString, lxr.source[startPos:]}, strconv.ErrSyntax
			}
		}
	}
}

func (lxr *Lexer) nextNumber() (Token, error) {
	startPos := lxr.chrPos
	dottable := true
	expable := true
	for !lxr.eof { // lxr.chr already included, look at next character
		if err := lxr.advance(); err != nil {
			return Token{TokenEOF, ""}, err
		}
		if isDigit(lxr.chr) {
			continue
		}
		if dottable && lxr.chr == '.' {
			dottable = false
			continue
		}
		if expable && (lxr.chr == 'e' || lxr.chr == 'E') {
			expable = false
			dottable = false
			if err := lxr.advance(); err != nil {
				return Token{TokenEOF, ""}, err
			}
			if lxr.chr == '-' || lxr.chr == '+' {
				if err := lxr.advance(); err != nil {
					return Token{TokenEOF, ""}, err
				}
			}
			if !isDigit(lxr.chr) {
				return Token{TokenNumber, lxr.source[startPos:lxr.nextPos]}, fmt.Errorf("floating-point literal lacks digit after 'e', at pos %d", lxr.chrPos)
			}
		}
		break
	}
	numSrc := lxr.source[startPos:lxr.chrPos]
	numFlt, err := strconv.ParseFloat(numSrc, 64)
	if err != nil {
		err = fmt.Errorf("lexical error at %d parsing number %q: %w", startPos, numSrc, err)
	}
	return Token{TokenNumber, numFlt}, err
}

func (lxr *Lexer) nextIdentifierOrKeyword() (Token, error) {
	tok, err := lxr.nextIdentifier()
	if tok.Type == TokenIdentifier && isKeyword(tok.Value.(string)) {
		tok.Type = TokenKeyword
	}
	return tok, err
}

func isKeyword(word string) bool {
	return word == "true" || word == "false" || word == "null"
}

func (lxr *Lexer) nextIdentifier() (Token, error) {
	startPos := lxr.chrPos
	for !lxr.eof {
		if err := lxr.advance(); err != nil {
			return Token{TokenEOF, ""}, err
		}
		if !isNameChar(lxr.chr) {
			break
		}
	}
	return Token{TokenIdentifier, lxr.source[startPos:lxr.chrPos]}, nil
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
