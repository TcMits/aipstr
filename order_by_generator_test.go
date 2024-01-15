package aipstr

import (
	"testing"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
)

func Test_Table_OrderByClause(t *testing.T) {
	parser := NewOrderByParser()

	decl := NewDeclaration(
		WithColumns(
			NewColumn("id", Filterable[Where]()),
			NewColumn("age", Filterable[Where](), Sortable[Where]()),
			NewColumn("active", Filterable[Where](), Sortable[Where]()),
			NewColumn("pets", Filterable[Where](), WithDeclaration(
				NewDeclaration[Where](
					WithColumns(
						NewColumn("id", Filterable[Where]()),
						NewColumn("name", Filterable[Where]())),
					WithOperatorFuncs(getOperatorFuncs()...),
				),
				func(w ...Where) (Where, error) {
					return func(s *sql.Selector) {
						petT := sql.
							Dialect(dialect.Postgres).
							Select("owner_id").
							From(sql.Table("pets"))

						for _, wi := range w {
							wi(petT)
						}

						s.Where(sql.In(s.C("id"), petT))
					}, nil
				},
			)),
		),
		WithOperatorFuncs(getOperatorFuncs()...),
	)

	type TestCase struct {
		input       string
		expectErr   bool
		expectOrder string
	}

	for _, tc := range []TestCase{
		{
			input:     "foo.bar",
			expectErr: true,
		},
		{
			input:     "foo",
			expectErr: true,
		},
		{
			input:       "active, age desc",
			expectErr:   false,
			expectOrder: `SELECT * FROM "users" ORDER BY "users"."active", "users"."age" DESC`,
		},
		{
			input:     "id",
			expectErr: true,
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			order, err := parser.ParseString("", tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			clauses, err := decl.OrderByClause(order)
			if tc.expectErr && err != nil {
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			s := sql.
				Dialect(dialect.Postgres).
				Select("*").
				From(sql.Table("users")).OrderBy()
			for _, o := range clauses {
				o(s)
			}

			query, _ := s.Query()
			if query != tc.expectOrder {
				t.Fatalf("expected %s, got %s", tc.expectOrder, query)
			}
		})
	}
}
