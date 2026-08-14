package rdf

import "testing"

func TestTermString(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		decoration Decoration
		expected   string
	}{
		{name: "none", value: "_:node", decoration: NONE, expected: "_:node"},
		{name: "angle brackets", value: "name", decoration: ANGLE_BRACKETS, expected: "<name>"},
		{
			name:       "escaped string",
			value:      "quote \" slash \\ line\n\ttab\x01",
			decoration: QUOTES,
			expected:   `"quote \" slash \\ line\n\ttab\u0001"`,
		},
		{name: "unicode", value: "Привет", decoration: QUOTES, expected: `"Привет"`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := NewTerm(test.value, test.decoration).String()
			if actual != test.expected {
				t.Fatalf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestRdfString(t *testing.T) {
	tests := []struct {
		name     string
		facets   []*Facet
		expected string
	}{
		{
			name:     "without facets",
			expected: `_:alice <name> "Alice" .`,
		},
		{
			name: "with facets",
			facets: []*Facet{
				NewFacet("rank", NewTerm("7", NONE)),
				NewFacet("note", NewTerm("a\"b", QUOTES)),
			},
			expected: `_:alice <name> "Alice" (rank=7, note="a\"b") .`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := NewRdf(
				NewTerm("_:alice", NONE),
				NewTerm("name", ANGLE_BRACKETS),
				NewTerm("Alice", QUOTES),
				test.facets,
			).String()
			if value != test.expected {
				t.Fatalf("expected %q, got %q", test.expected, value)
			}
		})
	}
}
