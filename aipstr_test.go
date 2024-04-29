package aipstr_test

import (
	"errors"
	"reflect"
	"testing"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/TcMits/aipstr"
	"go.einride.tech/aip/filtering"
)

type p = *sql.Predicate

func valueExceptIdent(v aipstr.Value) (any, bool) {
	if _, ok := v.IntoIdent(); ok {
		return nil, false
	}

	if i, ok := v.IntoInt(); ok {
		return i, true
	}

	if f, ok := v.IntoFloat(); ok {
		return f, true
	}

	if b, ok := v.IntoBool(); ok {
		return b, true
	}

	if s, ok := v.IntoDuration(); ok {
		return s, true
	}

	if t, ok := v.IntoTime(); ok {
		return t, true
	}

	if s, ok := v.IntoString(); ok {
		return s, true
	}

	return nil, true
}

var _ aipstr.Builder[p] = builder{}

type builder struct {
	aipstr.UnimplementedBuilder[p]

	s *sql.Selector
}

func (bu builder) Standalone(value aipstr.Value) (p, error) {
	ident, _ := value.IntoIdent()
	if len(ident) == 1 && (ident[0] == "true" || ident[0] == "false") {
		value = aipstr.BoolValue{Value: ident[0] == "true"}
	} else if len(ident) == 1 && ident[0] == "*" {
		return nil, errors.New("unexpected")
	} else if len(ident) > 1 {
		return nil, errors.New("unexpected")
	} else if len(ident) > 0 {
		return sql.P(func(b *sql.Builder) { b.Ident(bu.s.C(ident[0])) }), nil
	}

	v, ok := valueExceptIdent(value)
	if !ok {
		return nil, errors.New("unimplemented")
	}

	return sql.P(func(b *sql.Builder) { b.Arg(v) }), nil
}

func (builder) Function(op string, _ p, args ...p) (p, error) {
	switch op {
	case filtering.FunctionAnd:
		return sql.And(args...), nil
	case filtering.FunctionOr:
		return sql.Or(args...), nil
	case filtering.FunctionNot:
		if len(args) != 1 {
			return sql.Not(sql.And(args...)), nil
		}

		return sql.Not(args[0]), nil
	}

	if len(args) != 2 {
		return nil, errors.New("unexpected number of args")
	}

	switch op {
	case filtering.FunctionEquals:
		return sql.P(func(b *sql.Builder) {
			b.Join(args[0]).WriteOp(sql.OpEQ).Join(args[1])
		}), nil
	case filtering.FunctionNotEquals:
		return sql.P(func(b *sql.Builder) {
			b.Join(args[0]).WriteOp(sql.OpNEQ).Join(args[1])
		}), nil
	case filtering.FunctionGreaterThan:
		return sql.P(func(b *sql.Builder) {
			b.Join(args[0]).WriteOp(sql.OpGT).Join(args[1])
		}), nil
	case filtering.FunctionGreaterEquals:
		return sql.P(func(b *sql.Builder) {
			b.Join(args[0]).WriteOp(sql.OpGTE).Join(args[1])
		}), nil
	case filtering.FunctionLessThan:
		return sql.P(func(b *sql.Builder) {
			b.Join(args[0]).WriteOp(sql.OpLT).Join(args[1])
		}), nil
	case filtering.FunctionLessEquals:
		return sql.P(func(b *sql.Builder) {
			b.Join(args[0]).WriteOp(sql.OpLTE).Join(args[1])
		}), nil
	case filtering.FunctionHas:
		return sql.P(func(b *sql.Builder) {
			b.Join(args[1]).WriteOp(sql.OpIn).Join(args[0])
		}), nil
	default:
		return nil, errors.New("unimplemented function" + op)
	}
}

func (bu builder) HasNestedBuilder(ident []string, exc func(aipstr.Builder[p]) (p, error)) (p, error) {
	if len(ident) == 1 && ident[0] == "pets" {
		petT := sql.
			Dialect(dialect.Postgres).
			Select("owner_id").
			From(sql.Table("pets"))

		newP, err := exc(builder{s: petT})
		if err != nil {
			return nil, err
		}

		petT.Where(newP)
		return sql.In(bu.s.C("id"), petT), nil
	}

	return nil, errors.New("unimplemented")
}

func Test_ParseFilter(t *testing.T) {
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
			input:       "test123",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE "users"."test123"`,
		},
		{
			input:       "123",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE $1`,
			args:        []any{int64(123)},
		},
		{
			input:       "0",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE $1`,
			args:        []any{int64(0)},
		},
		{
			input:       "1.0",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE $1`,
			args:        []any{1.0},
		},
		{
			input:       "0.0",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE $1`,
			args:        []any{0.0},
		},
		{
			input:       "'test'",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE $1`,
			args:        []any{"test"},
		},
		{
			input:       "''",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE $1`,
			args:        []any{""},
		},
		{
			input:       "true",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE $1`,
			args:        []any{true},
		},
		{
			input:       "false",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE $1`,
			args:        []any{false},
		},
		{
			input:       "pets:(name='cat' OR name='dog')",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE "users"."id" IN (SELECT "owner_id" FROM "pets" WHERE "pets"."name" = $1 OR "pets"."name" = $2)`,
			args:        []any{"cat", "dog"},
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
			input:       "id = test123",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE "users"."id" = "users"."test123"`,
			args:        nil,
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
			input:       "id = 1",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE "users"."id" = $1`,
			args:        []any{int64(1)},
		},
		{
			input:       "1 = id",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE $1 = "users"."id"`,
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
			input:       "id:1",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE $1 IN "users"."id"`,
			args:        []any{int64(1)},
		},
		{
			input:       "1 = 1",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE $1 = $2`,
			args:        []any{int64(1), int64(1)},
		},
		{
			input:       "1 = 0",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE $1 = $2`,
			args:        []any{int64(1), int64(0)},
		},
		{
			input:       "1 != 1",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE $1 <> $2`,
			args:        []any{int64(1), int64(1)},
		},
		{
			input:       "pets = 1",
			expectErr:   false,
			expectWhere: `SELECT * FROM "users" WHERE "users"."pets" = $1`,
			args:        []any{int64(1)},
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			var parser filtering.Parser
			parser.Init(tc.input)

			expr, err := parser.Parse()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			s := sql.
				Dialect(dialect.Postgres).
				Select("*").
				From(sql.Table("users"))

			var b aipstr.Builder[p] = builder{s: s}
			where, err := aipstr.ParseFilter(b, expr.Expr)
			if tc.expectErr && err != nil {
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			s.Where(where)
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
