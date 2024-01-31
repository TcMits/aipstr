package aipstr

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type IntoValue interface {
	IntoInt() (int64, bool)
	IntoFloat() (float64, bool)
	IntoString() (string, bool)
	IntoBool() (bool, bool)
	IntoIdent() (string, bool)
	IsWildcard() bool
}

type unimplementedIntoValue struct{}

func (unimplementedIntoValue) IntoInt() (int64, bool) {
	return 0, false
}

func (unimplementedIntoValue) IntoFloat() (float64, bool) {
	return 0, false
}

func (unimplementedIntoValue) IntoString() (string, bool) {
	return "", false
}

func (unimplementedIntoValue) IntoBool() (bool, bool) {
	return false, false
}

func (unimplementedIntoValue) IntoIdent() (string, bool) {
	return "", false
}

func (unimplementedIntoValue) IsWildcard() bool {
	return false
}

type ZeroIntoValue struct{ unimplementedIntoValue }

func (ZeroIntoValue) IntoInt() (int64, bool) {
	return 0, true
}

func (ZeroIntoValue) IntoFloat() (float64, bool) {
	return 0, true
}

func (ZeroIntoValue) IntoString() (string, bool) {
	return "", true
}

func (ZeroIntoValue) IntoBool() (bool, bool) {
	return false, true
}

func NewFilterLexer() *lexer.StatefulDefinition {
	return lexer.MustSimple([]lexer.SimpleRule{
		{Name: "whitespace", Pattern: `\s+`},
		{Name: "And", Pattern: `AND`},
		{Name: "Or", Pattern: `OR`},
		{Name: "Not", Pattern: `NOT`},
		{Name: "LParen", Pattern: `\(`},
		{Name: "RParen", Pattern: `\)`},
		{Name: "Neg", Pattern: `-`},
		{Name: "Operator", Pattern: `!=|>=|<=|>|<|=|:`},
		{Name: "True", Pattern: `true`},
		{Name: "False", Pattern: `false`},
		{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},
		{Name: "Float", Pattern: `[-+]?[0-9]*(\.[0-9]+)`},
		{Name: "Int", Pattern: `[-+]?[0-9]+`},
		{Name: "Wildcard", Pattern: `\*`},
		{Name: "String", Pattern: `'[^']*'|"[^"]*"`},
		{Name: "Comma", Pattern: `,`},
		{Name: "Dot", Pattern: `\.`},
	})
}

func NewFilterParser() *participle.Parser[Filter] {
	return participle.MustBuild[Filter](
		participle.Lexer(FilterLexer),
		participle.Unquote("String"),
		participle.Elide("whitespace"),
		participle.Union[IntoValue](WildcardValue{}, FloatValue{}, IntValue{}, StringValue{}, BooleanValue{}, IdentValue{}),
	)
}

type Boolean bool

func (b *Boolean) Capture(values []string) error {
	*b = values[0] == "true"
	return nil
}

type BooleanValue struct {
	unimplementedIntoValue

	Value Boolean `parser:"@(True | False)"`
}

func (b BooleanValue) IntoBool() (bool, bool) {
	return bool(b.Value), true
}

type FloatValue struct {
	unimplementedIntoValue

	Value float64 `parser:"@Float"`
}

func (f FloatValue) IntoFloat() (float64, bool) {
	return f.Value, true
}

type IntValue struct {
	unimplementedIntoValue

	Value int64 `parser:"@Int"`
}

func (i IntValue) IntoInt() (int64, bool) {
	return i.Value, true
}

type StringValue struct {
	unimplementedIntoValue

	Value string `parser:"@String"`
}

func (s StringValue) IntoString() (string, bool) {
	return s.Value, true
}

type IdentValue struct {
	unimplementedIntoValue

	Value string `parser:"@Ident"`
}

func (i IdentValue) IntoIdent() (string, bool) {
	return i.Value, true
}

type WildcardValue struct {
	unimplementedIntoValue

	Value bool `parser:"@Wildcard"`
}

func (w WildcardValue) IsWildcard() bool {
	return w.Value
}

type Filter struct {
	Expression Expression `parser:"@@"`
}

func (f *Filter) String() string {
	return f.Expression.String()
}

type Expression struct {
	Sequences []Sequence `parser:"@@ (And @@)*"`
}

func (e *Expression) String() string {
	var s strings.Builder
	for i := range e.Sequences {
		if i > 0 {
			s.WriteString(" AND ")
		}
		s.WriteString(e.Sequences[i].String())
	}

	return s.String()
}

type Sequence struct {
	Terms []Term `parser:"@@ (Or @@)*"`
}

func (s *Sequence) String() string {
	var b strings.Builder
	for i := range s.Terms {
		if i > 0 {
			b.WriteString(" OR ")
		}
		b.WriteString(s.Terms[i].String())
	}

	return b.String()
}

type Term struct {
	Negated bool   `parser:"@(Not | Neg)?"`
	Simple  Simple `parser:"@@"`
}

func (t *Term) String() string {
	var b strings.Builder
	if t.Negated {
		b.WriteString("-")
	}
	b.WriteString(t.Simple.String())
	return b.String()
}

type Simple struct {
	IsComposite bool        `parser:"(@LParen"`
	Composite   Expression  `parser:"@@ RParen)"`
	Restriction Restriction `parser:"| @@"`
}

func (s *Simple) String() string {
	if !s.IsComposite {
		return s.Restriction.String()
	}

	var b strings.Builder
	b.WriteString("(")
	b.WriteString(s.Composite.String())
	b.WriteString(")")
	return b.String()
}

type Restriction struct {
	Comparable Comparable `parser:"@@"`
	Operator   string     `parser:"(@Operator"`
	Arg        Arg        `parser:"@@)?"`
}

func (r *Restriction) String() string {
	var b strings.Builder
	b.WriteString(r.Comparable.String())
	if r.Operator != "" {
		b.WriteString(r.Operator)
		b.WriteString(r.Arg.String())
	}
	return b.String()
}

type Comparable struct {
	Value    Value   `parser:"@@"`
	Fields   []Field `parser:"(Dot @@)*"`
	Callable bool    `parser:"(@LParen"`
	Args     []Arg   `parser:"(@@ (Comma @@)*)? RParen)?"`
}

func (c *Comparable) String() string {
	var b strings.Builder
	b.WriteString(c.Value.String())
	for i := range c.Fields {
		b.WriteString(".")
		b.WriteString(c.Fields[i].String())
	}

	if c.Callable {
		b.WriteString("(")
		for i := range c.Args {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(c.Args[i].String())
		}
		b.WriteString(")")
	}
	return b.String()
}

type Value struct {
	Inner IntoValue `parser:"@@"`
}

func cmp[T int64 | string | float64](op string, a, b func() (T, bool)) bool {
	if v1, ok := a(); ok {
		if v2, ok := b(); ok {
			switch op {
			case EqOp:
				return v1 == v2
			case NeOp:
				return v1 != v2
			case GtOp:
				return v1 > v2
			case GeOp:
				return v1 >= v2
			case LtOp:
				return v1 < v2
			case LeOp:
				return v1 <= v2
			case HasOp:
				return strings.Contains(fmt.Sprint(v1), fmt.Sprint(v2))
			}
		}
	}

	return false
}

func (v *Value) op(op string, v2 *Value) bool {
	if ok := cmp(op, v.Inner.IntoInt, v2.Inner.IntoInt); ok {
		return true
	}

	if ok := cmp(op, v.Inner.IntoFloat, v2.Inner.IntoFloat); ok {
		return true
	}

	if ok := cmp(op, v.Inner.IntoString, v2.Inner.IntoString); ok {
		return true
	}

	if v1, ok := v.Inner.IntoBool(); ok {
		if v2, ok := v2.Inner.IntoBool(); ok {
			switch op {
			case EqOp:
				return v1 == v2
			case NeOp:
				return v1 != v2
			}
		}
	}

	return false
}

func (v *Value) String() string {
	if v.Inner.IsWildcard() {
		return "*"
	}

	if isInt, ok := v.Inner.IntoInt(); ok {
		return strconv.FormatInt(isInt, 10)
	}

	if isFloat, ok := v.Inner.IntoFloat(); ok {
		return strconv.FormatFloat(isFloat, 'f', -1, 64)
	}

	if isBool, ok := v.Inner.IntoBool(); ok {
		return strconv.FormatBool(isBool)
	}

	if isString, ok := v.Inner.IntoString(); ok {
		var b strings.Builder
		b.WriteString("\"")
		b.WriteString(isString)
		b.WriteString("\"")
		return b.String()
	}

	ident, _ := v.Inner.IntoIdent()
	return ident
}

type Field struct {
	Keyword string `parser:"@(Not | And | Or)"`
	Value   Value  `parser:"| @@"`
}

func (f *Field) String() string {
	if f.Keyword != "" {
		return f.Keyword
	}

	return f.Value.String()
}

type Arg struct {
	IsComposite bool       `parser:"(@LParen"`
	Composite   Expression `parser:"@@ RParen)"`
	Comparable  Comparable `parser:"| @@"`
}

func (a *Arg) String() string {
	if !a.IsComposite {
		return a.Comparable.String()
	}

	var b strings.Builder
	b.WriteString("(")
	b.WriteString(a.Composite.String())
	b.WriteString(")")
	return b.String()
}
