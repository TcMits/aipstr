package aipstr

import (
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

func NewOrderByLexer() *lexer.StatefulDefinition {
	return lexer.MustSimple([]lexer.SimpleRule{
		{Name: "whitespace", Pattern: `\s+`},
		{Name: "Desc", Pattern: `desc`},
		{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},
		{Name: "Dot", Pattern: `\.`},
		{Name: "Comma", Pattern: `,`},
	})
}

func NewOrderByParser() *participle.Parser[OrderBy] {
	return participle.MustBuild[OrderBy](
		participle.Lexer(OrderByLexer),
		participle.Elide("whitespace"),
	)
}

type OrderBy struct {
	OrderExpression OrderExpression `parser:"@@"`
}

func (ob *OrderBy) String() string {
	return ob.OrderExpression.String()
}

type OrderExpression struct {
	OrderSequences []OrderSequence `parser:"@@ (Comma @@)*"`
}

func (oe *OrderExpression) String() string {
	var sb strings.Builder
	for i, oc := range oe.OrderSequences {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(oc.String())
	}
	return sb.String()
}

type OrderSequence struct {
	Identities []string `parser:"@Ident (Dot Ident)*"`
	Desc       bool     `parser:"@Desc?"`
}

func (oc *OrderSequence) String() string {
	var sb strings.Builder
	for i, ident := range oc.Identities {
		if i > 0 {
			sb.WriteString(".")
		}
		sb.WriteString(ident)
	}
	if oc.Desc {
		sb.WriteString(" desc")
	}
	return sb.String()
}
