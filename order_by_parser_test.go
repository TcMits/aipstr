package aipstr

import "testing"

func TestOrderBy(t *testing.T) {
	parser := NewOrderByParser()

	type TestCase struct {
		input    string
		expected string
	}

	for _, tc := range []TestCase{
		{"foo", "foo"},
		{"foo,bar", "foo, bar"},
		{"foo,bar,baz", "foo, bar, baz"},
		{"foo desc,bar,baz", "foo desc, bar, baz"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			orderBy, err := parser.ParseString("", tc.input)
			if err != nil {
				t.Fatalf("failed to parse %q: %v", tc.input, err)
			}
			got := orderBy.String()
			if got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}
