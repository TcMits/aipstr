package aipstr

import "testing"

func TestFilter(t *testing.T) {
	parser := NewFilterParser()

	type TestCase struct {
		input    string
		expected string
	}

	for _, tc := range []TestCase{
		{"foo", "foo"},
		{"foo AND bar", "foo AND bar"},
		{"foo OR bar", "foo OR bar"},
		{"foo AND bar OR baz", "foo AND bar OR baz"},
		{"foo AND (bar OR baz)", "foo AND (bar OR baz)"},
		{"foo AND bar OR baz AND qux", "foo AND bar OR baz AND qux"},
		{"foo AND (bar OR baz) AND qux", "foo AND (bar OR baz) AND qux"},
		{"foo AND (bar OR baz) OR qux", "foo AND (bar OR baz) OR qux"},
		{"foo AND (bar OR baz) OR qux AND quux", "foo AND (bar OR baz) OR qux AND quux"},
		{"foo AND (bar OR baz) OR qux AND (quux OR corge)", "foo AND (bar OR baz) OR qux AND (quux OR corge)"},
		{"foo AND (bar OR baz) OR qux AND (quux OR corge) OR grault", "foo AND (bar OR baz) OR qux AND (quux OR corge) OR grault"},
		{"foo=1", "foo=1"},
		{"foo=1 AND bar=2", "foo=1 AND bar=2"},
		{"foo=1 OR bar=2", "foo=1 OR bar=2"},
		{"foo=1 AND bar=2 OR baz=3", "foo=1 AND bar=2 OR baz=3"},
		{"foo=1 AND (bar=2 OR baz=3)", "foo=1 AND (bar=2 OR baz=3)"},
		{"foo=1 AND bar=2 OR baz=3 AND qux=4", "foo=1 AND bar=2 OR baz=3 AND qux=4"},
		{"foo=1 AND (bar=2 OR baz=3) AND qux=4", "foo=1 AND (bar=2 OR baz=3) AND qux=4"},
		{"foo>1", "foo>1"},
		{"foo>=1", "foo>=1"},
		{"foo<1", "foo<1"},
		{"foo<=1", "foo<=1"},
		{"foo!=1", "foo!=1"},
		{"foo=1 AND bar>2", "foo=1 AND bar>2"},
		{"foo=1 AND bar>=2", "foo=1 AND bar>=2"},
		{"foo=1 AND bar<2", "foo=1 AND bar<2"},
		{"foo=1 AND bar<=2", "foo=1 AND bar<=2"},
		{"foo=1 AND bar!=2", "foo=1 AND bar!=2"},
		{"foo=1 AND bar=2 OR baz>3", "foo=1 AND bar=2 OR baz>3"},
		{"foo=1 AND (bar=2 OR baz>3)", "foo=1 AND (bar=2 OR baz>3)"},
		{"foo:bar", "foo:bar"},
		{"foo:bar AND baz:qux", "foo:bar AND baz:qux"},
		{"foo:bar OR baz:qux", "foo:bar OR baz:qux"},
		{"foo:\"bar baz\"", "foo:\"bar baz\""},
		{"foo:\"bar baz\" AND baz:qux", "foo:\"bar baz\" AND baz:qux"},
		{"-foo", "-foo"},
		{"NOT foo", "-foo"},
		{"NOT foo AND NOT bar", "-foo AND -bar"},
		{"NOT foo OR NOT bar", "-foo OR -bar"},
		{"NOT foo AND bar", "-foo AND bar"},
		{"NOT foo OR bar", "-foo OR bar"},
		{"NOT foo AND NOT bar OR NOT baz", "-foo AND -bar OR -baz"},

		// traversal
		{"foo.bar=1", "foo.bar=1"},
		{"foo.bar.baz=1", "foo.bar.baz=1"},
		{"foo.bar.baz.qux=1", "foo.bar.baz.qux=1"},
		{"foo.bar.baz.qux.corge=1", "foo.bar.baz.qux.corge=1"},
		{"foo.bar.baz.qux.corge.grault=1", "foo.bar.baz.qux.corge.grault=1"},
		{"foo.bar.baz.qux.corge.grault.garply=1", "foo.bar.baz.qux.corge.grault.garply=1"},

		// function
		{"foo.bar(1)=1", "foo.bar(1)=1"},
		{"foo.bar(1, 2)=1", "foo.bar(1, 2)=1"},
		{"foo.bar(1, 2, 3)=1", "foo.bar(1, 2, 3)=1"},
		{"foo() AND bar()", "foo() AND bar()"},
		{"foo(1) AND bar(2)", "foo(1) AND bar(2)"},
		{"foo() AND bar(1, 2)", "foo() AND bar(1, 2)"},

		// type
		{"1.1", "1.1"},
		{"1", "1"},
		{"'test'", "\"test\""},
		{"*", "*"},
		{"true", "true"},
		{"false", "false"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			filter, err := parser.ParseString("", tc.input)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			if got := filter.String(); got != tc.expected {
				t.Errorf("got %q; expected %q", got, tc.expected)
			}
		})
	}
}
