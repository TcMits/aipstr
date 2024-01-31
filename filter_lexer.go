// Code generated by Participle. DO NOT EDIT.
package aipstr

import (
	"fmt"
	"io"
	"regexp/syntax"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

var _ syntax.Op
var _ fmt.State

const _ = utf8.RuneError

var FilterBackRefCache sync.Map
var FilterLexer lexer.Definition = lexerFilterDefinitionImpl{}

type lexerFilterDefinitionImpl struct{}

func (lexerFilterDefinitionImpl) Symbols() map[string]lexer.TokenType {
	return map[string]lexer.TokenType{
		"And":        -3,
		"Comma":      -17,
		"Dot":        -18,
		"EOF":        -1,
		"False":      -11,
		"Float":      -13,
		"Ident":      -12,
		"Int":        -14,
		"LParen":     -6,
		"Neg":        -8,
		"Not":        -5,
		"Operator":   -9,
		"Or":         -4,
		"RParen":     -7,
		"String":     -16,
		"True":       -10,
		"Wildcard":   -15,
		"whitespace": -2,
	}
}

func (lexerFilterDefinitionImpl) LexString(filename string, s string) (lexer.Lexer, error) {
	return &lexerFilterImpl{
		s: s,
		pos: lexer.Position{
			Filename: filename,
			Line:     1,
			Column:   1,
		},
		states: []lexerFilterState{{name: "Root"}},
	}, nil
}

func (d lexerFilterDefinitionImpl) LexBytes(filename string, b []byte) (lexer.Lexer, error) {
	return d.LexString(filename, string(b))
}

func (d lexerFilterDefinitionImpl) Lex(filename string, r io.Reader) (lexer.Lexer, error) {
	s := &strings.Builder{}
	_, err := io.Copy(s, r)
	if err != nil {
		return nil, err
	}
	return d.LexString(filename, s.String())
}

type lexerFilterState struct {
	name   string
	groups []string
}

type lexerFilterImpl struct {
	s      string
	p      int
	pos    lexer.Position
	states []lexerFilterState
}

func (l *lexerFilterImpl) Next() (lexer.Token, error) {
	if l.p == len(l.s) {
		return lexer.EOFToken(l.pos), nil
	}
	var (
		state  = l.states[len(l.states)-1]
		groups []int
		sym    lexer.TokenType
	)
	switch state.name {
	case "Root":
		if match := matchFilterwhitespace(l.s, l.p, l.states[len(l.states)-1].groups); match[1] != 0 {
			sym = -2
			groups = match[:]
		} else if match := matchFilterAnd(l.s, l.p, l.states[len(l.states)-1].groups); match[1] != 0 {
			sym = -3
			groups = match[:]
		} else if match := matchFilterOr(l.s, l.p, l.states[len(l.states)-1].groups); match[1] != 0 {
			sym = -4
			groups = match[:]
		} else if match := matchFilterNot(l.s, l.p, l.states[len(l.states)-1].groups); match[1] != 0 {
			sym = -5
			groups = match[:]
		} else if match := matchFilterLParen(l.s, l.p, l.states[len(l.states)-1].groups); match[1] != 0 {
			sym = -6
			groups = match[:]
		} else if match := matchFilterRParen(l.s, l.p, l.states[len(l.states)-1].groups); match[1] != 0 {
			sym = -7
			groups = match[:]
		} else if match := matchFilterNeg(l.s, l.p, l.states[len(l.states)-1].groups); match[1] != 0 {
			sym = -8
			groups = match[:]
		} else if match := matchFilterOperator(l.s, l.p, l.states[len(l.states)-1].groups); match[1] != 0 {
			sym = -9
			groups = match[:]
		} else if match := matchFilterTrue(l.s, l.p, l.states[len(l.states)-1].groups); match[1] != 0 {
			sym = -10
			groups = match[:]
		} else if match := matchFilterFalse(l.s, l.p, l.states[len(l.states)-1].groups); match[1] != 0 {
			sym = -11
			groups = match[:]
		} else if match := matchFilterIdent(l.s, l.p, l.states[len(l.states)-1].groups); match[1] != 0 {
			sym = -12
			groups = match[:]
		} else if match := matchFilterFloat(l.s, l.p, l.states[len(l.states)-1].groups); match[1] != 0 {
			sym = -13
			groups = match[:]
		} else if match := matchFilterInt(l.s, l.p, l.states[len(l.states)-1].groups); match[1] != 0 {
			sym = -14
			groups = match[:]
		} else if match := matchFilterWildcard(l.s, l.p, l.states[len(l.states)-1].groups); match[1] != 0 {
			sym = -15
			groups = match[:]
		} else if match := matchFilterString(l.s, l.p, l.states[len(l.states)-1].groups); match[1] != 0 {
			sym = -16
			groups = match[:]
		} else if match := matchFilterComma(l.s, l.p, l.states[len(l.states)-1].groups); match[1] != 0 {
			sym = -17
			groups = match[:]
		} else if match := matchFilterDot(l.s, l.p, l.states[len(l.states)-1].groups); match[1] != 0 {
			sym = -18
			groups = match[:]
		}
	}
	if groups == nil {
		sample := []rune(l.s[l.p:])
		if len(sample) > 16 {
			sample = append(sample[:16], []rune("...")...)
		}
		return lexer.Token{}, participle.Errorf(l.pos, "invalid input text %q", string(sample))
	}
	pos := l.pos
	span := l.s[groups[0]:groups[1]]
	l.p = groups[1]
	l.pos.Advance(span)
	return lexer.Token{
		Type:  sym,
		Value: span,
		Pos:   pos,
	}, nil
}

func (l *lexerFilterImpl) sgroups(match []int) []string {
	sgroups := make([]string, len(match)/2)
	for i := 0; i < len(match)-1; i += 2 {
		sgroups[i/2] = l.s[l.p+match[i] : l.p+match[i+1]]
	}
	return sgroups
}

// [\t-\n\f-\r ]+
func matchFilterwhitespace(s string, p int, backrefs []string) (groups [2]int) {
	// [\t-\n\f-\r ] (CharClass)
	l0 := func(s string, p int) int {
		if len(s) <= p {
			return -1
		}
		rn := s[p]
		switch {
		case rn >= '\t' && rn <= '\n':
			return p + 1
		case rn >= '\f' && rn <= '\r':
			return p + 1
		case rn == ' ':
			return p + 1
		}
		return -1
	}
	// [\t-\n\f-\r ]+ (Plus)
	l1 := func(s string, p int) int {
		if p = l0(s, p); p == -1 {
			return -1
		}
		for len(s) > p {
			if np := l0(s, p); np == -1 {
				return p
			} else {
				p = np
			}
		}
		return p
	}
	np := l1(s, p)
	if np == -1 {
		return
	}
	groups[0] = p
	groups[1] = np
	return
}

// AND
func matchFilterAnd(s string, p int, backrefs []string) (groups [2]int) {
	if p+3 <= len(s) && s[p:p+3] == "AND" {
		groups[0] = p
		groups[1] = p + 3
	}
	return
}

// OR
func matchFilterOr(s string, p int, backrefs []string) (groups [2]int) {
	if p+2 <= len(s) && s[p:p+2] == "OR" {
		groups[0] = p
		groups[1] = p + 2
	}
	return
}

// NOT
func matchFilterNot(s string, p int, backrefs []string) (groups [2]int) {
	if p+3 <= len(s) && s[p:p+3] == "NOT" {
		groups[0] = p
		groups[1] = p + 3
	}
	return
}

// \(
func matchFilterLParen(s string, p int, backrefs []string) (groups [2]int) {
	if p < len(s) && s[p] == '(' {
		groups[0] = p
		groups[1] = p + 1
	}
	return
}

// \)
func matchFilterRParen(s string, p int, backrefs []string) (groups [2]int) {
	if p < len(s) && s[p] == ')' {
		groups[0] = p
		groups[1] = p + 1
	}
	return
}

// -
func matchFilterNeg(s string, p int, backrefs []string) (groups [2]int) {
	if p < len(s) && s[p] == '-' {
		groups[0] = p
		groups[1] = p + 1
	}
	return
}

// !=|>=|<=|[:<->]
func matchFilterOperator(s string, p int, backrefs []string) (groups [2]int) {
	// != (Literal)
	l0 := func(s string, p int) int {
		if p+2 <= len(s) && s[p:p+2] == "!=" {
			return p + 2
		}
		return -1
	}
	// >= (Literal)
	l1 := func(s string, p int) int {
		if p+2 <= len(s) && s[p:p+2] == ">=" {
			return p + 2
		}
		return -1
	}
	// <= (Literal)
	l2 := func(s string, p int) int {
		if p+2 <= len(s) && s[p:p+2] == "<=" {
			return p + 2
		}
		return -1
	}
	// [:<->] (CharClass)
	l3 := func(s string, p int) int {
		if len(s) <= p {
			return -1
		}
		rn := s[p]
		switch {
		case rn == ':':
			return p + 1
		case rn >= '<' && rn <= '>':
			return p + 1
		}
		return -1
	}
	// !=|>=|<=|[:<->] (Alternate)
	l4 := func(s string, p int) int {
		if np := l0(s, p); np != -1 {
			return np
		}
		if np := l1(s, p); np != -1 {
			return np
		}
		if np := l2(s, p); np != -1 {
			return np
		}
		if np := l3(s, p); np != -1 {
			return np
		}
		return -1
	}
	np := l4(s, p)
	if np == -1 {
		return
	}
	groups[0] = p
	groups[1] = np
	return
}

// true
func matchFilterTrue(s string, p int, backrefs []string) (groups [2]int) {
	if p+4 <= len(s) && s[p:p+4] == "true" {
		groups[0] = p
		groups[1] = p + 4
	}
	return
}

// false
func matchFilterFalse(s string, p int, backrefs []string) (groups [2]int) {
	if p+5 <= len(s) && s[p:p+5] == "false" {
		groups[0] = p
		groups[1] = p + 5
	}
	return
}

// [A-Z_a-z][0-9A-Z_a-z]*
func matchFilterIdent(s string, p int, backrefs []string) (groups [2]int) {
	// [A-Z_a-z] (CharClass)
	l0 := func(s string, p int) int {
		if len(s) <= p {
			return -1
		}
		rn := s[p]
		switch {
		case rn >= 'A' && rn <= 'Z':
			return p + 1
		case rn == '_':
			return p + 1
		case rn >= 'a' && rn <= 'z':
			return p + 1
		}
		return -1
	}
	// [0-9A-Z_a-z] (CharClass)
	l1 := func(s string, p int) int {
		if len(s) <= p {
			return -1
		}
		rn := s[p]
		switch {
		case rn >= '0' && rn <= '9':
			return p + 1
		case rn >= 'A' && rn <= 'Z':
			return p + 1
		case rn == '_':
			return p + 1
		case rn >= 'a' && rn <= 'z':
			return p + 1
		}
		return -1
	}
	// [0-9A-Z_a-z]* (Star)
	l2 := func(s string, p int) int {
		for len(s) > p {
			if np := l1(s, p); np == -1 {
				return p
			} else {
				p = np
			}
		}
		return p
	}
	// [A-Z_a-z][0-9A-Z_a-z]* (Concat)
	l3 := func(s string, p int) int {
		if p = l0(s, p); p == -1 {
			return -1
		}
		if p = l2(s, p); p == -1 {
			return -1
		}
		return p
	}
	np := l3(s, p)
	if np == -1 {
		return
	}
	groups[0] = p
	groups[1] = np
	return
}

// [\+\-]?[0-9]*(\.[0-9]+)
func matchFilterFloat(s string, p int, backrefs []string) (groups [4]int) {
	// [\+\-] (CharClass)
	l0 := func(s string, p int) int {
		if len(s) <= p {
			return -1
		}
		rn := s[p]
		if rn == '+' || rn == '-' {
			return p + 1
		}
		return -1
	}
	// [\+\-]? (Quest)
	l1 := func(s string, p int) int {
		if np := l0(s, p); np != -1 {
			return np
		}
		return p
	}
	// [0-9] (CharClass)
	l2 := func(s string, p int) int {
		if len(s) <= p {
			return -1
		}
		rn := s[p]
		switch {
		case rn >= '0' && rn <= '9':
			return p + 1
		}
		return -1
	}
	// [0-9]* (Star)
	l3 := func(s string, p int) int {
		for len(s) > p {
			if np := l2(s, p); np == -1 {
				return p
			} else {
				p = np
			}
		}
		return p
	}
	// \. (Literal)
	l4 := func(s string, p int) int {
		if p < len(s) && s[p] == '.' {
			return p + 1
		}
		return -1
	}
	// [0-9]+ (Plus)
	l5 := func(s string, p int) int {
		if p = l2(s, p); p == -1 {
			return -1
		}
		for len(s) > p {
			if np := l2(s, p); np == -1 {
				return p
			} else {
				p = np
			}
		}
		return p
	}
	// \.[0-9]+ (Concat)
	l6 := func(s string, p int) int {
		if p = l4(s, p); p == -1 {
			return -1
		}
		if p = l5(s, p); p == -1 {
			return -1
		}
		return p
	}
	// (\.[0-9]+) (Capture)
	l7 := func(s string, p int) int {
		np := l6(s, p)
		if np != -1 {
			groups[2] = p
			groups[3] = np
		}
		return np
	}
	// [\+\-]?[0-9]*(\.[0-9]+) (Concat)
	l8 := func(s string, p int) int {
		if p = l1(s, p); p == -1 {
			return -1
		}
		if p = l3(s, p); p == -1 {
			return -1
		}
		if p = l7(s, p); p == -1 {
			return -1
		}
		return p
	}
	np := l8(s, p)
	if np == -1 {
		return
	}
	groups[0] = p
	groups[1] = np
	return
}

// [\+\-]?[0-9]+
func matchFilterInt(s string, p int, backrefs []string) (groups [2]int) {
	// [\+\-] (CharClass)
	l0 := func(s string, p int) int {
		if len(s) <= p {
			return -1
		}
		rn := s[p]
		if rn == '+' || rn == '-' {
			return p + 1
		}
		return -1
	}
	// [\+\-]? (Quest)
	l1 := func(s string, p int) int {
		if np := l0(s, p); np != -1 {
			return np
		}
		return p
	}
	// [0-9] (CharClass)
	l2 := func(s string, p int) int {
		if len(s) <= p {
			return -1
		}
		rn := s[p]
		switch {
		case rn >= '0' && rn <= '9':
			return p + 1
		}
		return -1
	}
	// [0-9]+ (Plus)
	l3 := func(s string, p int) int {
		if p = l2(s, p); p == -1 {
			return -1
		}
		for len(s) > p {
			if np := l2(s, p); np == -1 {
				return p
			} else {
				p = np
			}
		}
		return p
	}
	// [\+\-]?[0-9]+ (Concat)
	l4 := func(s string, p int) int {
		if p = l1(s, p); p == -1 {
			return -1
		}
		if p = l3(s, p); p == -1 {
			return -1
		}
		return p
	}
	np := l4(s, p)
	if np == -1 {
		return
	}
	groups[0] = p
	groups[1] = np
	return
}

// \*
func matchFilterWildcard(s string, p int, backrefs []string) (groups [2]int) {
	if p < len(s) && s[p] == '*' {
		groups[0] = p
		groups[1] = p + 1
	}
	return
}

// '[^']*'|"[^"]*"
func matchFilterString(s string, p int, backrefs []string) (groups [2]int) {
	// ' (Literal)
	l0 := func(s string, p int) int {
		if p < len(s) && s[p] == '\'' {
			return p + 1
		}
		return -1
	}
	// [^'] (CharClass)
	l1 := func(s string, p int) int {
		if len(s) <= p {
			return -1
		}
		var (
			rn rune
			n  int
		)
		if s[p] < utf8.RuneSelf {
			rn, n = rune(s[p]), 1
		} else {
			rn, n = utf8.DecodeRuneInString(s[p:])
		}
		switch {
		case rn >= '\x00' && rn <= '&':
			return p + 1
		case rn >= '(' && rn <= '\U0010ffff':
			return p + n
		}
		return -1
	}
	// [^']* (Star)
	l2 := func(s string, p int) int {
		for len(s) > p {
			if np := l1(s, p); np == -1 {
				return p
			} else {
				p = np
			}
		}
		return p
	}
	// '[^']*' (Concat)
	l3 := func(s string, p int) int {
		if p = l0(s, p); p == -1 {
			return -1
		}
		if p = l2(s, p); p == -1 {
			return -1
		}
		if p = l0(s, p); p == -1 {
			return -1
		}
		return p
	}
	// " (Literal)
	l4 := func(s string, p int) int {
		if p < len(s) && s[p] == '"' {
			return p + 1
		}
		return -1
	}
	// [^"] (CharClass)
	l5 := func(s string, p int) int {
		if len(s) <= p {
			return -1
		}
		var (
			rn rune
			n  int
		)
		if s[p] < utf8.RuneSelf {
			rn, n = rune(s[p]), 1
		} else {
			rn, n = utf8.DecodeRuneInString(s[p:])
		}
		switch {
		case rn >= '\x00' && rn <= '!':
			return p + 1
		case rn >= '#' && rn <= '\U0010ffff':
			return p + n
		}
		return -1
	}
	// [^"]* (Star)
	l6 := func(s string, p int) int {
		for len(s) > p {
			if np := l5(s, p); np == -1 {
				return p
			} else {
				p = np
			}
		}
		return p
	}
	// "[^"]*" (Concat)
	l7 := func(s string, p int) int {
		if p = l4(s, p); p == -1 {
			return -1
		}
		if p = l6(s, p); p == -1 {
			return -1
		}
		if p = l4(s, p); p == -1 {
			return -1
		}
		return p
	}
	// '[^']*'|"[^"]*" (Alternate)
	l8 := func(s string, p int) int {
		if np := l3(s, p); np != -1 {
			return np
		}
		if np := l7(s, p); np != -1 {
			return np
		}
		return -1
	}
	np := l8(s, p)
	if np == -1 {
		return
	}
	groups[0] = p
	groups[1] = np
	return
}

// ,
func matchFilterComma(s string, p int, backrefs []string) (groups [2]int) {
	if p < len(s) && s[p] == ',' {
		groups[0] = p
		groups[1] = p + 1
	}
	return
}

// \.
func matchFilterDot(s string, p int, backrefs []string) (groups [2]int) {
	if p < len(s) && s[p] == '.' {
		groups[0] = p
		groups[1] = p + 1
	}
	return
}
