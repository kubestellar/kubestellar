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
	"unicode"

	"github.com/dop251/goja"
	js_ast "github.com/dop251/goja/ast"
)

type Lexer struct {
	source string
	reader io.RuneReader

	chr     rune // next rune to process
	chrPos  int  // index of start of chr
	nextPos int  // index after chr
	eof     bool // no more to process
}

// Token is a type of token
type Token string

const (
	TokenEOF        Token = "EOF"
	TokenSpecial    Token = "Special"
	TokenString     Token = "String"
	TokenIdentifier Token = "ID"
	TokenNumber     Token = "Number"
)

type JPLiteralValue interface{} // either string or int64

func NewLexer(source string) *Lexer {
	lxr := &Lexer{
		source: source,
		reader: strings.NewReader(source),
	}
	lxr.advance()
	return lxr
}

func (lxr *Lexer) Next() (Token, JPLiteralValue, error) {
	if lxr.eof {
		return TokenEOF, "", nil
	}
	chr := lxr.chr
	switch chr {
	case '[', ']', '*', ',', ':':
		if err := lxr.advance(); err != nil {
			return TokenEOF, "", err
		}
		return TokenSpecial, string(chr), nil
	}
	if chr == '.' {
		if err := lxr.advance(); err != nil {
			return TokenEOF, "", err
		}
		if lxr.chr == '.' {
			if err := lxr.advance(); err != nil {
				return TokenEOF, "", err
			}
			return TokenSpecial, "..", nil
		}
		return TokenSpecial, ".", nil
	}
	if '0' <= lxr.chr && lxr.chr <= '9' {
		return lxr.nextNumber(0)
	}
	if lxr.chr == '"' || lxr.chr == '\'' {
		return lxr.nextString()
	}
	if isIdentifierStart(chr) {
		return lxr.nextIdentifierName()
	}
	return TokenSpecial, "", fmt.Errorf("syntax error at %q", string(chr))
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

func (lxr *Lexer) nextString() (Token, string, error) {
	startPos := lxr.chrPos
	close := lxr.chr
	for {
		if err := lxr.advance(); err != nil {
			return TokenEOF, "", err
		}
		if lxr.eof {
			return TokenString, lxr.source[startPos:], strconv.ErrSyntax
		}
		if lxr.chr == close {
			if err := lxr.advance(); err != nil {
				return TokenEOF, "", err
			}
			stringSrc := lxr.source[startPos:lxr.chrPos]
			stringProg, err := goja.Parse("lit", stringSrc)
			if err != nil {
				return TokenString, stringSrc, err
			}
			if len(stringProg.Body) != 1 {
				return TokenString, stringSrc, fmt.Errorf("parsed goja body len is %d", len(stringProg.Body))
			}
			stringExpr, ok := stringProg.Body[0].(*js_ast.ExpressionStatement)
			if !ok {
				return TokenString, stringSrc, fmt.Errorf("goja statement is not an expression statement, it is %#+v, at %T", stringProg.Body[0], stringProg.Body[0])
			}
			stringLit, ok := stringExpr.Expression.(*js_ast.StringLiteral)
			if !ok {
				return TokenString, stringSrc, fmt.Errorf("goja statement is not a string literal, it is %#+v, at %T", stringProg.Body[0], stringProg.Body[0])
			}
			return TokenString, stringLit.Value.String(), nil
		}
		if lxr.chr == '\\' {
			if err := lxr.advance(); err != nil {
				return TokenEOF, "", err
			}
			if lxr.eof {
				return TokenString, lxr.source[startPos:], strconv.ErrSyntax
			}
		}
	}
}

func (lxr *Lexer) nextNumber(startOffset int) (Token, JPLiteralValue, error) {
	startPos := lxr.chrPos + startOffset
	isHex := false
	isOctal := false
	baseable := lxr.chr == '0' && startOffset == 0
	for !lxr.eof {
		if err := lxr.advance(); err != nil {
			return TokenEOF, "", err
		}
		if baseable {
			baseable = false
			if lxr.chr == 'o' {
				isOctal = true
				continue
			} else if lxr.chr == 'x' {
				isHex = true
				continue
			}
		}
		if '0' <= lxr.chr && lxr.chr < '8' {
			continue
		}
		if (!isOctal) && (lxr.chr == '8' || lxr.chr == '9') {
			continue
		}
		if isHex && ('a' <= lxr.chr && lxr.chr <= 'f' || 'A' <= lxr.chr && lxr.chr <= 'F') {
			continue
		}
		break
	}
	numSrc := lxr.source[startPos:lxr.chrPos]
	numInt, err := strconv.ParseInt(numSrc, 10, 64)
	return TokenNumber, numInt, err
}

func (lxr *Lexer) nextIdentifierName() (Token, JPLiteralValue, error) {
	startPos := lxr.chrPos
	for !lxr.eof {
		if err := lxr.advance(); err != nil {
			return TokenEOF, "", err
		}
		if !isIdentifierPart(lxr.chr) {
			break
		}
	}
	return TokenIdentifier, lxr.source[startPos:lxr.chrPos], nil
}

func isIdentifierStart(r rune) bool {
	return r == '$' || r == '_' || (unicode.In(r, unicode.Letter, unicode.Title, unicode.Lm, unicode.Other_ID_Start) &&
		!unicode.In(r, unicode.Pattern_Syntax, unicode.Pattern_White_Space))
}

func isIdentifierPart(r rune) bool {
	return r == '$' || r == '_' || r == 0x200C || r == 0x200D || unicode.IsDigit(r) || (unicode.In(r, unicode.Letter, unicode.Title, unicode.Lm, unicode.Other_ID_Continue) &&
		!unicode.In(r, unicode.Pattern_Syntax, unicode.Pattern_White_Space))
}
