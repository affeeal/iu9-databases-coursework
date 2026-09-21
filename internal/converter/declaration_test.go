package converter

import "testing"

func TestDeclarationValidation(t *testing.T) {
	tests := []struct {
		name         string
		declarations []declaration
	}{
		{
			name:         "empty name",
			declarations: []declaration{{Type: "id"}},
		},
		{
			name: "duplicate name",
			declarations: []declaration{
				{Name: "id", Type: "id"},
				{Name: "id", Type: "string"},
			},
		},
		{
			name:         "unknown type",
			declarations: []declaration{{Name: "value", Type: "decimal"}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := file{Declarations: test.declarations}
			if _, err := input.validateDeclarations(); err == nil {
				t.Fatal("invalid declaration unexpectedly succeeded")
			}
		})
	}
}
