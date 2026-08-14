package converter

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/affeeal/iu9-databases-coursework/internal/rdf"
)

func TestValidateSymbol(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  rune
		wantError bool
	}{
		{name: "ASCII", input: ";", expected: ';'},
		{name: "Unicode", input: "→", expected: '→'},
		{name: "empty", wantError: true},
		{name: "two runes", input: "ab", wantError: true},
		{name: "newline", input: "\n", wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := validateSymbol(test.input)
			if (err != nil) != test.wantError {
				t.Fatalf("unexpected error: %v", err)
			}
			if actual != test.expected {
				t.Fatalf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestValidateHeaders(t *testing.T) {
	tests := []struct {
		name    string
		file    file
		headers []string
	}{
		{
			name:    "duplicate",
			file:    file{Name: "data.csv"},
			headers: []string{"id", "id"},
		},
		{
			name: "missing declaration",
			file: file{
				Name:         "data.csv",
				Declarations: []declaration{{Name: "id", Type: "id"}},
			},
			headers: []string{"name"},
		},
		{
			name: "artificial collision",
			file: file{
				Name:                  "data.csv",
				ArtificialDeclaration: declaration{Name: "id", Type: "id"},
			},
			headers: []string{"id"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.file.validateHeaders(test.headers); err == nil {
				t.Fatal("invalid headers unexpectedly succeeded")
			} else if !strings.Contains(err.Error(), test.file.Name) {
				t.Fatalf("error does not identify input file: %v", err)
			}
		})
	}
}

func TestConvertFacetsIsDeterministic(t *testing.T) {
	input := entityFacets{
		"zeta":  rdf.NewTerm("2", rdf.NONE),
		"alpha": rdf.NewTerm("1", rdf.NONE),
	}
	actual := convertFacets(input)
	expected := []*rdf.Facet{
		rdf.NewFacet("alpha", rdf.NewTerm("1", rdf.NONE)),
		rdf.NewFacet("zeta", rdf.NewTerm("2", rdf.NONE)),
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("facets are not sorted: %#v", actual)
	}
}

func TestValidateSchemaTypeDiagnostics(t *testing.T) {
	schema := map[string]schemaType{
		"id":    {dt: idType},
		"value": {dt: stringType},
	}
	if err := validateSchemaType(schema, "subject", "value", true); err == nil ||
		!strings.Contains(err.Error(), "must be an id") {
		t.Fatalf("unexpected required-ID diagnostic: %v", err)
	}
	if err := validateSchemaType(schema, "facet", "id", false); err == nil ||
		!strings.Contains(err.Error(), "must not be an id") {
		t.Fatalf("unexpected forbidden-ID diagnostic: %v", err)
	}
}

func TestProcessDatasetFixture(t *testing.T) {
	datasetPath := filepath.Join("testdata", "minimal")
	if err := ProcessDataset(datasetPath); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(filepath.Join(datasetPath, "output.rdf")) })

	actual, err := os.ReadFile(filepath.Join(datasetPath, "output.rdf"))
	if err != nil {
		t.Fatal(err)
	}
	expected, err := os.ReadFile(filepath.Join(datasetPath, "expected.rdf"))
	if err != nil {
		t.Fatal(err)
	}
	if string(actual) != string(expected) {
		t.Fatalf("unexpected RDF output:\n%s", actual)
	}
}

func TestProcessDatasetDoesNotReplaceOutputOnFailure(t *testing.T) {
	datasetPath := t.TempDir()
	if err := os.Mkdir(filepath.Join(datasetPath, "sources"), 0o755); err != nil {
		t.Fatal(err)
	}
	config := `files:
  - name: data.csv
    declarations:
      - name: missing
        type: id
    rdfs: []
`
	if err := os.WriteFile(filepath.Join(datasetPath, "convert.yml"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(datasetPath, "sources", "data.csv"), []byte("present\n1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	original := []byte("previous complete output\n")
	outputPath := filepath.Join(datasetPath, "output.rdf")
	if err := os.WriteFile(outputPath, original, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := ProcessDataset(datasetPath); err == nil {
		t.Fatal("invalid conversion unexpectedly succeeded")
	}
	actual, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(actual) != string(original) {
		t.Fatalf("output was replaced after failure: %q", actual)
	}
	temporaryFiles, err := filepath.Glob(filepath.Join(datasetPath, ".output-*.rdf"))
	if err != nil {
		t.Fatal(err)
	}
	if len(temporaryFiles) != 0 {
		t.Fatalf("temporary output remains: %v", temporaryFiles)
	}
}

func TestProcessDatasetAddsConfiguredArtificialID(t *testing.T) {
	datasetPath := t.TempDir()
	if err := os.Mkdir(filepath.Join(datasetPath, "sources"), 0o755); err != nil {
		t.Fatal(err)
	}
	config := `files:
  - name: data.csv
    declarations:
      - name: name
        type: string
    artificial_declaration:
      name: row
      type: id
      prefix: r
    rdfs:
      - subject: row
        predicate: name
        object: name
`
	if err := os.WriteFile(filepath.Join(datasetPath, "convert.yml"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(datasetPath, "sources", "data.csv"), []byte("name\nAlice\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ProcessDataset(datasetPath); err != nil {
		t.Fatal(err)
	}
	actual, err := os.ReadFile(filepath.Join(datasetPath, "output.rdf"))
	if err != nil {
		t.Fatal(err)
	}
	if expected := "_:r0 <name> \"Alice\" .\n"; string(actual) != expected {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
}
