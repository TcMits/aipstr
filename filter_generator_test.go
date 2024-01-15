package aipstr

import (
	"reflect"
	"testing"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
)

type Where = func(*sql.Selector)

func getOperatorFuncs() []*DeclarationOperatorFunc[Where] {
	return []*DeclarationOperatorFunc[Where]{
		NewOperatorFunc(EqOp, WithFieldWithValueAnyNoErr(sql.FieldEQ), WithFieldWithFieldNoErr(sql.FieldsEQ)),
		NewOperatorFunc(NeOp, WithFieldWithValueAnyNoErr(sql.FieldNEQ), WithFieldWithFieldNoErr(sql.FieldsNEQ)),
		NewOperatorFunc(GtOp, WithFieldWithValueAnyNoErr(sql.FieldGT), WithFieldWithFieldNoErr(sql.FieldsGT)),
		NewOperatorFunc(GeOp, WithFieldWithValueAnyNoErr(sql.FieldGTE), WithFieldWithFieldNoErr(sql.FieldsGTE)),
		NewOperatorFunc(LtOp, WithFieldWithValueAnyNoErr(sql.FieldLT), WithFieldWithFieldNoErr(sql.FieldsLT)),
		NewOperatorFunc(LeOp, WithFieldWithValueAnyNoErr(sql.FieldLTE), WithFieldWithFieldNoErr(sql.FieldsLTE)),
		NewOperatorFunc(AndOp, WithCombineNoErr(sql.AndPredicates[Where])),
		NewOperatorFunc(OrOp, WithCombineNoErr(sql.OrPredicates[Where])),
		NewOperatorFunc(NotOp, WithCombineNoErr(sql.NotPredicates[Where])),
		NewOperatorFunc(AscOp, WithFieldNoErr(func(s string) Where {
			return sql.OrderByField(s).ToFunc()
		})),
		NewOperatorFunc(DescOp, WithFieldNoErr(func(s string) Where {
			return sql.OrderByField(s, sql.OrderDesc()).ToFunc()
		})),
		NewOperatorFunc(
			HasOp, WithFieldWithValueString(func(s1, s2 string) (Where, error) {
				return sql.FieldContains(s1, s2), nil
			}),
		),
		NewOperatorFunc(
			TrueOp, WithNoField(func() Where {
				return func(s *sql.Selector) {
					s.Where(sql.P().Append(func(b *sql.Builder) {
						b.WriteString("TRUE")
					}))
				}
			}),
		),
		NewOperatorFunc(
			FalseOp, WithNoField(func() Where {
				return func(s *sql.Selector) {
					s.Where(sql.False())
				}
			}),
		),
	}
}

func Test_Table_WhereClause(t *testing.T) {
	parser := NewFilterParser()

	decl := NewDeclaration(
		WithColumns(
			NewColumn("id", Filterable[Where]()),
			NewColumn("age", Filterable[Where]()),
			NewColumn("active", Filterable[Where]()),
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
		expectWhere string
		args        []any
	}

	for _, tc := range []TestCase{
		{
			input:       "active OR -active",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE "users"."active" OR (NOT ("users"."active"))`,
			args:        nil,
		},
		{
			input:       "active AND -active",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE "users"."active" AND (NOT ("users"."active"))`,
			args:        nil,
		},
		{
			input:       "(active)",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE "users"."active"`,
			args:        nil,
		},
		{
			input:     "*",
			expectErr: true,
		},
		{
			input:     "test.test",
			expectErr: true,
		},
		{
			input:     "test()",
			expectErr: true,
		},
		{
			input:     "test123",
			expectErr: true,
		},
		{
			input:       "123",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE TRUE`,
			args:        nil,
		},
		{
			input:       "0",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE FALSE`,
			args:        nil,
		},
		{
			input:       "1.0",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE TRUE`,
			args:        nil,
		},
		{
			input:       "0.0",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE FALSE`,
			args:        nil,
		},
		{
			input:       "'test'",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE TRUE`,
			args:        nil,
		},
		{
			input:       "''",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE FALSE`,
			args:        nil,
		},
		{
			input:       "true",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE TRUE`,
			args:        nil,
		},
		{
			input:       "false",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE FALSE`,
			args:        nil,
		},
		{
			input:       "pets:(name='cat' OR name='dog')",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE "users"."id" IN (SELECT "owner_id" FROM "pets" WHERE "pets"."name" = $1 OR "pets"."name" = $2)`,
			args:        []any{"cat", "dog"},
		},
		{
			input:     "id = (id = 'cat')",
			expectErr: true,
		},
		{
			input:     "id = test.test",
			expectErr: true,
		},
		{
			input:     "id = test()",
			expectErr: true,
		},
		{
			input:     "id = test123",
			expectErr: true,
		},
		{
			input:       "id = id",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE "users"."id" = "users"."id"`,
			args:        nil,
		},
		{
			input:       "id > id",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE "users"."id" > "users"."id"`,
			args:        nil,
		},
		{
			input:       "id >= id",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE "users"."id" >= "users"."id"`,
			args:        nil,
		},
		{
			input:       "id < id",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE "users"."id" < "users"."id"`,
			args:        nil,
		},
		{
			input:       "id <= id",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE "users"."id" <= "users"."id"`,
			args:        nil,
		},
		{
			input:       "id != id",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE "users"."id" <> "users"."id"`,
			args:        nil,
		},
		{
			input:     "id:id",
			expectErr: true,
		},
		{
			input:       "id = 1",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE "users"."id" = $1`,
			args:        []any{int64(1)},
		},
		{
			input:       "1 = id",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE "users"."id" = $1`,
			args:        []any{int64(1)},
		},
		{
			input:     "id = *",
			expectErr: true,
		},
		{
			input:     "* = id",
			expectErr: true,
		},
		{
			input:       "id != 1",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE "users"."id" <> $1`,
			args:        []any{int64(1)},
		},
		{
			input:       "id > 1",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE "users"."id" > $1`,
			args:        []any{int64(1)},
		},
		{
			input:       "id >= 1",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE "users"."id" >= $1`,
			args:        []any{int64(1)},
		},
		{
			input:       "id < 1",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE "users"."id" < $1`,
			args:        []any{int64(1)},
		},
		{
			input:       "id <= 1",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE "users"."id" <= $1`,
			args:        []any{int64(1)},
		},
		{
			input:     "id:1",
			expectErr: true,
		},
		{
			input:       "id:'test'",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE "users"."id" LIKE $1`,
			args:        []any{"%test%"},
		},
		{
			input:       "1 = 1",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE TRUE`,
			args:        nil,
		},
		{
			input:       "1 = 0",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE FALSE`,
			args:        nil,
		},
		{
			input:       "1 != 1",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE FALSE`,
			args:        nil,
		},
		{
			input:     "pets = 1",
			expectErr: true,
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			filter, err := parser.ParseString("", tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			where, err := decl.WhereClause(filter)
			if tc.expectErr && err != nil {
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			s := sql.
				Dialect(dialect.Postgres).
				Select("*").
				From(sql.Table("users"))
			where(s)

			query, args := s.Query()
			if query != tc.expectWhere {
				t.Fatalf("expected %s, got %s", tc.expectWhere, query)
			}

			if !reflect.DeepEqual(args, tc.args) {
				t.Fatalf("expected %v, got %v", tc.args, args)
			}
		})
	}
}
