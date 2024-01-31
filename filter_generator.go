package aipstr

import "fmt"

const (
	EqOp    = "="
	NeOp    = "!="
	GtOp    = ">"
	GeOp    = ">="
	LtOp    = "<"
	LeOp    = "<="
	HasOp   = ":"
	TrueOp  = "_true"
	FalseOp = "_false"
	AndOp   = "_and"
	OrOp    = "_or"
	NotOp   = "_not"
)

func (t *Declaration[T]) WhereClause(filter *Filter) (T, error) {
	clause, err := t.expressionQuery(&filter.Expression)
	if err != nil {
		return zero[T](), err
	}
	return clause, nil
}

func (t *Declaration[T]) expressionQuery(expression *Expression) (_ T, err error) {
	var exprW T
	andFunc, ok := t.getOperatorFunc(AndOp, combineSign)
	if !ok {
		return zero[T](), fmt.Errorf("unknown operator '%s'", AndOp)
	}

	orFunc, ok := t.getOperatorFunc(OrOp, combineSign)
	if !ok {
		return zero[T](), fmt.Errorf("unknown operator '%s'", OrOp)
	}

	is := -1
	for is = range expression.Sequences {
		var seqW T
		it := -1
		seq := &expression.Sequences[is]
		for it = range seq.Terms {
			f, err := t.termQuery(&seq.Terms[it])
			if err != nil {
				return zero[T](), err
			}

			if it == 0 {
				seqW = f
			} else if seqW, err = orFunc.combine(seqW, f); err != nil {
				return zero[T](), err
			}
		}

		if it == -1 {
			continue
		}

		if is == 0 {
			exprW = seqW
		} else if exprW, err = andFunc.combine(exprW, seqW); err != nil {
			return zero[T](), err
		}
	}

	if is == -1 {
		trueFunc, ok := t.getOperatorFunc(TrueOp, noFieldSign)
		if !ok {
			return zero[T](), fmt.Errorf("unknown operator '%s'", TrueOp)
		}

		return trueFunc.noField(), nil
	}

	return exprW, nil
}

func (t *Declaration[T]) termQuery(term *Term) (T, error) {
	simpleQuery, err := t.simpleQuery(&term.Simple)
	if err != nil {
		return zero[T](), err
	}
	if term.Negated {
		notFunc, ok := t.getOperatorFunc(NotOp, combineSign)
		if !ok {
			return zero[T](), fmt.Errorf("unknown operator '%s'", NotOp)
		}

		return notFunc.combine(simpleQuery)
	}
	return simpleQuery, nil
}

func (t *Declaration[T]) simpleQuery(simple *Simple) (T, error) {
	if !simple.IsComposite {
		return t.restrictionQuery(&simple.Restriction)
	}

	return t.expressionQuery(&simple.Composite)
}

func (t *Declaration[T]) restrictionQuery(restriction *Restriction) (_ T, err error) {
	comparableValue, err := t.valueQueryFromComparable(&restriction.Comparable)
	if err != nil {
		return zero[T](), err
	}

	comparableCol, hasComparableCol, err := t.columnFromValueQuery(comparableValue)
	if err != nil {
		return zero[T](), err
	}

	trueFunc, ok := t.getOperatorFunc(TrueOp, noFieldSign)
	if !ok {
		return zero[T](), fmt.Errorf("unknown operator '%s'", TrueOp)
	}

	falseFunc, ok := t.getOperatorFunc(FalseOp, noFieldSign)
	if !ok {
		return zero[T](), fmt.Errorf("unknown operator '%s'", FalseOp)
	}

	switch {
	case restriction.Operator == "":
		if hasComparableCol {
			eqBoolFunc, ok := t.getOperatorFunc(EqOp, fieldWithValueBoolSign)
			if !ok {
				return zero[T](), fmt.Errorf("unknown operator '%s'", EqOp)
			}

			return eqBoolFunc.fieldWithValueBool(comparableCol.field, true)
		}

		if comparableValue.op(EqOp, &Value{Inner: ZeroIntoValue{}}) {
			return falseFunc.noField(), nil
		}

		return trueFunc.noField(), nil
	case (restriction.Arg.IsComposite &&
		hasComparableCol &&
		comparableCol.combineDeclFilter != nil &&
		comparableCol.declaration != nil &&
		restriction.Operator == HasOp):
		// allow nested filter:
		// example: "pets:(name = 'cat' OR name = 'dog')"
		w, err := comparableCol.declaration.expressionQuery(&restriction.Arg.Composite)
		if err != nil {
			return zero[T](), err
		}

		return comparableCol.combineDeclFilter(w)
	case restriction.Arg.IsComposite:
		return zero[T](), fmt.Errorf("nested filter is not supported")
	case hasComparableCol && (comparableCol.combineDeclFilter != nil || comparableCol.declaration != nil):
		return zero[T](), fmt.Errorf("nested filter is not supported")
	}

	argValue, err := t.valueQueryFromComparable(&restriction.Arg.Comparable)
	if err != nil {
		return zero[T](), err
	}

	argCol, hasArgCol, err := t.columnFromValueQuery(argValue)
	if err != nil {
		return zero[T](), err
	}

	switch {
	case hasComparableCol && hasArgCol:
		fwfFunc, ok := t.getOperatorFunc(restriction.Operator, fieldWithFieldSign)
		if !ok {
			return zero[T](), fmt.Errorf("unknown operator '%s'", restriction.Operator)
		}

		return fwfFunc.fieldWithField(comparableCol.field, argCol.field)
	case hasComparableCol || hasArgCol:
		var field string
		var v *Value
		if hasComparableCol {
			field = comparableCol.field
			v = argValue
		} else {
			field = argCol.field
			v = comparableValue
		}

		if isInt, ok := v.Inner.IntoInt(); ok {
			fwiFunc, ok := t.getOperatorFunc(restriction.Operator, fieldWithValueIntSign)
			if !ok {
				return zero[T](), fmt.Errorf("unknown operator '%s'", restriction.Operator)
			}

			return fwiFunc.fieldWithValueInt(field, isInt)
		} else if isFloat, ok := v.Inner.IntoFloat(); ok {
			fwfFunc, ok := t.getOperatorFunc(restriction.Operator, fieldWithValueFloatSign)
			if !ok {
				return zero[T](), fmt.Errorf("unknown operator '%s'", restriction.Operator)
			}

			return fwfFunc.fieldWithValueFloat(field, isFloat)
		} else if isBool, ok := v.Inner.IntoBool(); ok {
			fwbFunc, ok := t.getOperatorFunc(restriction.Operator, fieldWithValueBoolSign)
			if !ok {
				return zero[T](), fmt.Errorf("unknown operator '%s'", restriction.Operator)
			}

			return fwbFunc.fieldWithValueBool(field, isBool)
		} else if isString, ok := v.Inner.IntoString(); ok {
			fwsFunc, ok := t.getOperatorFunc(restriction.Operator, fieldWithValueStringSign)
			if !ok {
				return zero[T](), fmt.Errorf("unknown operator '%s'", restriction.Operator)
			}

			return fwsFunc.fieldWithValueString(field, isString)
		}

		return zero[T](), fmt.Errorf("unknown type of value")
	default:
		if comparableValue.op(restriction.Operator, argValue) {
			return trueFunc.noField(), nil
		}

		return falseFunc.noField(), nil
	}
}

func (t *Declaration[T]) valueQueryFromComparable(comparable *Comparable) (*Value, error) {
	if comparable.Callable {
		return nil, fmt.Errorf("callable is not supported")
	}

	if len(comparable.Fields) > 0 {
		return nil, fmt.Errorf("nested fields are not supported")
	}

	if comparable.Value.Inner.IsWildcard() {
		return nil, fmt.Errorf("wildcard is not supported")
	}

	return &comparable.Value, nil
}

func (t *Declaration[T]) columnFromValueQuery(v *Value) (*Column[T], bool, error) {
	ident, ok := v.Inner.IntoIdent()
	if !ok {
		return nil, false, nil
	}

	column, ok := t.getColumnByField(ident, filterableSign)
	if ok {
		return column, true, nil
	}

	return nil, false, fmt.Errorf("unknown value '%s'", ident)
}
