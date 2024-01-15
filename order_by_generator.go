package aipstr

import "fmt"

const (
	AscOp  = "_asc"
	DescOp = "_desc"
)

func (t *Declaration[T]) OrderByClause(filter *OrderBy) ([]T, error) {
	clauses, err := t.orderExpressionQuery(&filter.OrderExpression)
	if err != nil {
		return nil, err
	}
	return clauses, nil
}

func (t *Declaration[T]) orderExpressionQuery(expression *OrderExpression) (_ []T, err error) {
	results := make([]T, 0, 1)

	for is := range expression.OrderSequences {
		result, err := t.orderSequenceQuery(&expression.OrderSequences[is])
		if err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

func (t *Declaration[T]) orderSequenceQuery(seq *OrderSequence) (T, error) {
	if len(seq.Identities) != 1 {
		return zero[T](), fmt.Errorf("nested field is not supported")
	}

	col, ok := t.getColumnByField(seq.Identities[0], sortableSign)
	if !ok {
		return zero[T](), fmt.Errorf("unknown field '%s'", seq.Identities[0])
	}

	ascFunc, ok := t.getOperatorFunc(AscOp, fieldSign)
	if !ok {
		return zero[T](), fmt.Errorf("unknown operator '%s'", AscOp)
	}

	descFunc, ok := t.getOperatorFunc(DescOp, fieldSign)
	if !ok {
		return zero[T](), fmt.Errorf("unknown operator '%s'", DescOp)
	}

	if seq.Desc {
		return descFunc.field(col.field)
	}

	return ascFunc.field(col.field)
}
