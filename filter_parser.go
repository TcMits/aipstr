package aipstr

import (
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

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
		participle.Lexer(NewFilterLexer()),
		participle.Unquote("String"),
		participle.Elide("whitespace"),
	)
}

type Boolean bool

func (b *Boolean) Capture(values []string) error {
	*b = values[0] == "true"
	return nil
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
	Restriction *Restriction `parser:"@@"`
	Composite   *Expression  `parser:"| (LParen @@ RParen)"`
}

func (s *Simple) String() string {
	if s.Restriction != nil {
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
	Arg        *Arg       `parser:"@@)?"`
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
	Wildcard bool     `parser:"@Wildcard"`
	Float    *float64 `parser:"| @Float"`
	Int      *int64   `parser:"| @Int"`
	Str      *string  `parser:"| @String"`
	Boolean  *Boolean `parser:"| @(True | False)"`
	Ident    string   `parser:"| @Ident"`
}

func (v *Value) op(op string, v2 *Value) bool {
	if v.Int != nil && v2.Int != nil {
		switch op {
		case EqOp:
			return *v.Int == *v2.Int
		case NeOp:
			return *v.Int != *v2.Int
		case GtOp:
			return *v.Int > *v2.Int
		case GeOp:
			return *v.Int >= *v2.Int
		case LtOp:
			return *v.Int < *v2.Int
		case LeOp:
			return *v.Int <= *v2.Int
		}
	}

	if v.Float != nil && v2.Float != nil {
		switch op {
		case EqOp:
			return *v.Float == *v2.Float
		case NeOp:
			return *v.Float != *v2.Float
		case GtOp:
			return *v.Float > *v2.Float
		case GeOp:
			return *v.Float >= *v2.Float
		case LtOp:
			return *v.Float < *v2.Float
		case LeOp:
			return *v.Float <= *v2.Float
		}
	}

	if v.Boolean != nil && v2.Boolean != nil {
		switch op {
		case EqOp:
			return *v.Boolean == *v2.Boolean
		case NeOp:
			return *v.Boolean != *v2.Boolean
		}
	}

	if v.Str != nil && v2.Str != nil {
		switch op {
		case EqOp:
			return *v.Str == *v2.Str
		case NeOp:
			return *v.Str != *v2.Str
		case GtOp:
			return *v.Str > *v2.Str
		case GeOp:
			return *v.Str >= *v2.Str
		case LtOp:
			return *v.Str < *v2.Str
		case LeOp:
			return *v.Str <= *v2.Str
		case HasOp:
			return strings.Contains(*v.Str, *v2.Str)
		}
	}

	return false
}

func (v *Value) String() string {
	if v.Wildcard {
		return "*"
	}

	if v.Int != nil {
		return strconv.FormatInt(*v.Int, 10)
	}

	if v.Float != nil {
		return strconv.FormatFloat(*v.Float, 'f', -1, 64)
	}

	if v.Boolean != nil {
		return strconv.FormatBool(bool(*v.Boolean))
	}

	if v.Str != nil {
		var b strings.Builder
		b.WriteString("\"")
		b.WriteString(*v.Str)
		b.WriteString("\"")
		return b.String()
	}

	return v.Ident
}

type Field struct {
	Value   *Value `parser:"@@"`
	Keyword string `parser:"| @(Not | And | Or)"`
}

func (f *Field) String() string {
	if f.Value != nil {
		return f.Value.String()
	}
	return f.Keyword
}

type Arg struct {
	Comparable *Comparable `parser:"(@@"`
	Composite  *Expression `parser:"| LParen @@ RParen)"`
}

func (a *Arg) String() string {
	if a.Comparable != nil {
		return a.Comparable.String()
	}

	var b strings.Builder
	b.WriteString("(")
	b.WriteString(a.Composite.String())
	b.WriteString(")")
	return b.String()
}
