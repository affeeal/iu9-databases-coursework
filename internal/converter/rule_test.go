package converter

import "testing"

func TestRuleRejectsEmptyNames(t *testing.T) {
	schema := map[string]schemaType{"id": {dt: idType}, "value": {dt: stringType}}
	triple := rdfRule{Subject: "id", Object: "value"}
	if err := triple.validate(schema); err == nil {
		t.Fatal("empty predicate accepted")
	}
	facet := facetRule{Value: "value"}
	if err := facet.validate(schema, "facet"); err == nil {
		t.Fatal("empty facet key accepted")
	}
}
