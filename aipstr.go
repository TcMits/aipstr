package aipstr

import (
	"fmt"
	"time"

	"go.einride.tech/aip/filtering"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

func zero[T any]() T {
	var t T
	return t
}

type Builder[T any] interface {
	Standalone(value Value) (T, error)
	Function(op string, target T, args ...T) (T, error)
}

type NestedBuilder[T any] interface {
	// special func, allow special use-case
	// example "variants:(id = 1 OR name = 'foo')" means
	// filter records which has variants with id = 1 or name = 'foo'
	HasNestedBuilder(field []string, exc func(Builder[T]) (T, error)) (T, error)
}

var _ Builder[any] = (*UnimplementedBuilder[any])(nil)

type UnimplementedBuilder[T any] struct{}

func (UnimplementedBuilder[T]) Standalone(value Value) (T, error) {
	return zero[T](), fmt.Errorf("unimplemented Standalone")
}

func (UnimplementedBuilder[T]) Function(op string, target T, args ...T) (T, error) {
	return zero[T](), fmt.Errorf("unimplemented StdComparisonCall")
}

type filterParser[T any] struct {
	builder Builder[T]
	filter  *expr.Expr
}

func ParseFilter[T any](b Builder[T], filter *expr.Expr) (T, error) {
	if filter == nil {
		return zero[T](), nil
	}

	return (&filterParser[T]{b, filter}).parseExpr(filter)
}

func (p *filterParser[T]) parseExpr(e *expr.Expr) (T, error) {
	switch e.GetExprKind().(type) {
	case *expr.Expr_CallExpr:
		return p.parseCallExpr(e)
	case *expr.Expr_IdentExpr:
		return p.parseIdentExpr(e)
	case *expr.Expr_SelectExpr:
		return p.parseSelectExpr(e)
	case *expr.Expr_ConstExpr:
		return p.parseConstExpr(e)
	}

	return zero[T](), fmt.Errorf("unsupported expr: %v", e)
}

func parseConstExpr(e *expr.Expr) (Value, error) {
	constExpr := e.GetConstExpr()
	switch constExpr.GetConstantKind().(type) {
	case *expr.Constant_BoolValue:
		return BoolValue{Value: constExpr.GetBoolValue()}, nil
	case *expr.Constant_DoubleValue:
		return FloatValue{Value: constExpr.GetDoubleValue()}, nil
	case *expr.Constant_Int64Value:
		return IntValue{Value: constExpr.GetInt64Value()}, nil
	case *expr.Constant_StringValue:
		return StringValue{Value: constExpr.GetStringValue()}, nil
	case *expr.Constant_Uint64Value:
		return IntValue{Value: int64(constExpr.GetUint64Value())}, nil
	case *expr.Constant_DurationValue:
		return StringValue{Value: constExpr.GetDurationValue().AsDuration().String()}, nil
	case *expr.Constant_TimestampValue:
		return StringValue{Value: constExpr.GetTimestampValue().AsTime().Format(time.RFC3339Nano)}, nil
	case *expr.Constant_NullValue:
		return unimplementedValue{}, nil
	}

	return unimplementedValue{}, fmt.Errorf("unsupported constant: %v", e)
}

func (p *filterParser[T]) parseConstExpr(e *expr.Expr) (T, error) {
	value, err := parseConstExpr(e)
	if err != nil {
		return zero[T](), err
	}

	return p.builder.Standalone(value)
}

func parseSelectExpr(e *expr.Expr, out *[]string) error {
	selectExpr := e.GetSelectExpr()
	operand := selectExpr.GetOperand()
	switch operand.GetExprKind().(type) {
	case *expr.Expr_IdentExpr:
		*out = append(*out, operand.GetIdentExpr().GetName(), selectExpr.GetField())
	case *expr.Expr_SelectExpr:
		err := parseSelectExpr(operand, out)
		if err != nil {
			return err
		}
		*out = append(*out, selectExpr.GetField())
	default:
		return fmt.Errorf("unsupported select expr: %v", e)
	}

	return nil
}

func (p *filterParser[T]) parseSelectExpr(e *expr.Expr) (T, error) {
	path := make([]string, 0, 2)
	if err := parseSelectExpr(e, &path); err != nil {
		return zero[T](), err
	}

	return p.builder.Standalone(IdentValue{Value: path})
}

func (p *filterParser[T]) parseIdentExpr(e *expr.Expr) (T, error) {
	return p.builder.Standalone(IdentValue{Value: []string{e.GetIdentExpr().GetName()}})
}

func (p *filterParser[T]) parseCallExpr(e *expr.Expr) (_ T, err error) {
	switch e.GetCallExpr().GetFunction() {
	case filtering.FunctionHas:
		return p.parseComparisonHasCallExpr(e)
	default:
		return p.parseFunctionCallExpr(e)
	}
}

func (p *filterParser[T]) parseComparisonHasCallExpr(e *expr.Expr) (_ T, err error) {
	nestedHas, ok := p.builder.(NestedBuilder[T])
	if !ok {
		return p.parseFunctionCallExpr(e)
	}

	callExpr := e.GetCallExpr()
	args := callExpr.GetArgs()

	if len(args) != 2 {
		return zero[T](), fmt.Errorf("invalid number of args: %v", e)
	}

	var lhsV Value = unimplementedValue{}
	switch args[0].GetExprKind().(type) {
	case *expr.Expr_IdentExpr:
		lhsV = IdentValue{Value: []string{args[0].GetIdentExpr().GetName()}}
	case *expr.Expr_SelectExpr:
		out := make([]string, 0, 2)
		err := parseSelectExpr(args[0], &out)
		if err != nil {
			return zero[T](), err
		}

		lhsV = IdentValue{Value: out}
	default:
		return p.parseFunctionCallExpr(e)
	}

	switch args[1].GetExprKind().(type) {
	case *expr.Expr_CallExpr:
		ident, _ := lhsV.IntoIdent()
		return nestedHas.HasNestedBuilder(ident, func(b Builder[T]) (T, error) {
			return ParseFilter(b, args[1])
		})
	default:
		return p.parseFunctionCallExpr(e)
	}
}

func (p *filterParser[T]) parseFunctionCallExpr(e *expr.Expr) (_ T, err error) {
	callExpr := e.GetCallExpr()
	args := callExpr.GetArgs()
	results := make([]T, 0, len(args))
	targetExp := callExpr.GetTarget()
	target := zero[T]()
	if targetExp != nil {
		target, err = p.parseExpr(targetExp)
		if err != nil {
			return zero[T](), err
		}
	}

	for _, arg := range args {
		result, err := p.parseExpr(arg)
		if err != nil {
			return zero[T](), err
		}

		results = append(results, result)
	}

	return p.builder.Function(callExpr.GetFunction(), target, results...)
}
