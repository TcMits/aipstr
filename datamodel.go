package aipstr

const (
	filterableSign = 1 << iota
	sortableSign
)

type Column[T any] struct {
	sign              int
	declaration       *Declaration[T]
	field             string
	combineDeclFilter func(...T) (T, error)
}

type ColOption[T any] func(*Column[T])

func Filterable[T any]() ColOption[T] {
	return func(c *Column[T]) {
		c.sign |= filterableSign
	}
}

func Sortable[T any]() ColOption[T] {
	return func(c *Column[T]) {
		c.sign |= sortableSign
	}
}

func WithDeclaration[T any](decl *Declaration[T], filter func(...T) (T, error)) ColOption[T] {
	return func(c *Column[T]) {
		c.declaration = decl
		c.combineDeclFilter = filter
	}
}

func NewColumn[T any](field string, opts ...ColOption[T]) *Column[T] {
	if field == "" {
		panic("nil argument")
	}

	c := Column[T]{field: field}
	for _, opt := range opts {
		opt(&c)
	}
	return &c
}

const (
	fieldWithValueIntSign = 1 << iota
	fieldWithValueFloatSign
	fieldWithValueBoolSign
	fieldWithValueStringSign
	fieldWithFieldSign
	fieldSign
	noFieldSign
	combineSign
)

type DeclarationOperatorFunc[T any] struct {
	name                 string
	fieldWithValueInt    func(string, int64) (T, error)
	fieldWithValueFloat  func(string, float64) (T, error)
	fieldWithValueBool   func(string, bool) (T, error)
	fieldWithValueString func(string, string) (T, error)
	fieldWithField       func(string, string) (T, error)
	field                func(string) (T, error)
	noField              func() T
	combine              func(...T) (T, error)
}

type OperatorFuncOption[T any] func(*DeclarationOperatorFunc[T])

func WithFieldWithValueInt[T any](f func(string, int64) (T, error)) OperatorFuncOption[T] {
	return func(of *DeclarationOperatorFunc[T]) {
		of.fieldWithValueInt = f
	}
}

func WithFieldWithValueFloat[T any](f func(string, float64) (T, error)) OperatorFuncOption[T] {
	return func(of *DeclarationOperatorFunc[T]) {
		of.fieldWithValueFloat = f
	}
}

func WithFieldWithValueBool[T any](f func(string, bool) (T, error)) OperatorFuncOption[T] {
	return func(of *DeclarationOperatorFunc[T]) {
		of.fieldWithValueBool = f
	}
}

func WithFieldWithValueString[T any](f func(string, string) (T, error)) OperatorFuncOption[T] {
	return func(of *DeclarationOperatorFunc[T]) {
		of.fieldWithValueString = f
	}
}

func WithFieldWithField[T any](f func(string, string) (T, error)) OperatorFuncOption[T] {
	return func(of *DeclarationOperatorFunc[T]) {
		of.fieldWithField = f
	}
}

func WithNoField[T any](f func() T) OperatorFuncOption[T] {
	return func(dof *DeclarationOperatorFunc[T]) {
		dof.noField = f
	}
}

func WithCombine[T any](f func(...T) (T, error)) OperatorFuncOption[T] {
	return func(of *DeclarationOperatorFunc[T]) {
		of.combine = f
	}
}

func WithFieldWithValueAny[T any](f func(string, any) (T, error)) OperatorFuncOption[T] {
	return func(of *DeclarationOperatorFunc[T]) {
		of.fieldWithValueInt = func(field string, v int64) (T, error) {
			return f(field, v)
		}
		of.fieldWithValueFloat = func(field string, v float64) (T, error) {
			return f(field, v)
		}
		of.fieldWithValueBool = func(field string, v bool) (T, error) {
			return f(field, v)
		}
		of.fieldWithValueString = func(field string, v string) (T, error) {
			return f(field, v)
		}
	}
}

func WithFieldWithValueAnyNoErr[T any](f func(string, any) T) OperatorFuncOption[T] {
	return func(of *DeclarationOperatorFunc[T]) {
		of.fieldWithValueInt = func(field string, v int64) (T, error) {
			return f(field, v), nil
		}
		of.fieldWithValueFloat = func(field string, v float64) (T, error) {
			return f(field, v), nil
		}
		of.fieldWithValueBool = func(field string, v bool) (T, error) {
			return f(field, v), nil
		}
		of.fieldWithValueString = func(field string, v string) (T, error) {
			return f(field, v), nil
		}
	}
}

func WithFieldWithFieldNoErr[T any](f func(string, string) T) OperatorFuncOption[T] {
	return func(of *DeclarationOperatorFunc[T]) {
		of.fieldWithField = func(field string, v string) (T, error) {
			return f(field, v), nil
		}
	}
}

func WithCombineNoErr[T any](f func(...T) T) OperatorFuncOption[T] {
	return func(of *DeclarationOperatorFunc[T]) {
		of.combine = func(args ...T) (T, error) {
			return f(args...), nil
		}
	}
}

func WithField[T any](f func(string) (T, error)) OperatorFuncOption[T] {
	return func(dof *DeclarationOperatorFunc[T]) {
		dof.field = f
	}
}

func WithFieldNoErr[T any](f func(string) T) OperatorFuncOption[T] {
	return func(dof *DeclarationOperatorFunc[T]) {
		dof.field = func(field string) (T, error) {
			return f(field), nil
		}
	}
}

func NewOperatorFunc[T any](name string, opts ...OperatorFuncOption[T]) *DeclarationOperatorFunc[T] {
	of := DeclarationOperatorFunc[T]{name: name}
	for _, opt := range opts {
		opt(&of)
	}
	return &of
}

func (f *DeclarationOperatorFunc[T]) withSignature(sign ...int) bool {
	for _, s := range sign {
		if s&fieldWithValueIntSign != 0 && f.fieldWithValueInt == nil {
			return false
		}

		if s&fieldWithValueFloatSign != 0 && f.fieldWithValueFloat == nil {
			return false
		}

		if s&fieldWithValueBoolSign != 0 && f.fieldWithValueBool == nil {
			return false
		}

		if s&fieldWithValueStringSign != 0 && f.fieldWithValueString == nil {
			return false
		}

		if s&fieldWithFieldSign != 0 && f.fieldWithField == nil {
			return false
		}

		if s&fieldSign != 0 && f.field == nil {
			return false
		}

		if s&noFieldSign != 0 && f.noField == nil {
			return false
		}

		if s&combineSign != 0 && f.combine == nil {
			return false
		}
	}

	return true
}

type Declaration[T any] struct {
	columns map[string]*Column[T]
	ops     map[string]*DeclarationOperatorFunc[T]
}

type DeclarationOption[T any] func(*Declaration[T])

func WithColumns[T any](c ...*Column[T]) DeclarationOption[T] {
	return func(d *Declaration[T]) {
		for _, col := range c {
			d.columns[col.field] = col
		}
	}
}

func WithOperatorFuncs[T any](of ...*DeclarationOperatorFunc[T]) DeclarationOption[T] {
	return func(d *Declaration[T]) {
		for _, f := range of {
			d.ops[f.name] = f
		}
	}
}

func NewDeclaration[T any](opts ...DeclarationOption[T]) *Declaration[T] {
	d := Declaration[T]{columns: map[string]*Column[T]{}, ops: map[string]*DeclarationOperatorFunc[T]{}}
	for _, opt := range opts {
		opt(&d)
	}
	return &d
}

func (d *Declaration[T]) getColumnByField(field string, sign ...int) (*Column[T], bool) {
	column, ok := d.columns[field]
	if !ok || column == nil {
		return nil, false
	}

	for _, s := range sign {
		if column.sign&s == 0 {
			return nil, false
		}
	}

	return column, true
}

func (d *Declaration[T]) getOperatorFunc(name string, sign ...int) (*DeclarationOperatorFunc[T], bool) {
	of, ok := d.ops[name]
	if !ok || of == nil {
		return nil, false
	}

	if !of.withSignature(sign...) {
		return nil, false
	}

	return of, true
}
