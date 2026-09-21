package converter

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/affeeal/iu9-databases-coursework/internal/rdf"
	"gopkg.in/yaml.v3"
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
		{name: "NUL", input: "\x00", wantError: true},
		{name: "quote", input: "\"", wantError: true},
		{name: "replacement character", input: "\uFFFD", wantError: true},
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
	datasetPath := copyMinimalDataset(t)
	if err := ProcessDataset(datasetPath); err != nil {
		t.Fatal(err)
	}
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

func copyMinimalDataset(t *testing.T) string {
	t.Helper()
	destination := t.TempDir()
	if err := os.Mkdir(filepath.Join(destination, "sources"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"convert.yml", "expected.rdf", "sources/people.csv"} {
		contents, err := os.ReadFile(filepath.Join("testdata", "minimal", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(destination, name), contents, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return destination
}

func TestProcessDatasetRejectsInvalidConfigWithoutReplacingOutput(t *testing.T) {
	for _, config := range []string{
		"", "{}", "null", "files: []", "files: null", "unknown: true",
		"files: []\n---\nfiles: []\n",
	} {
		t.Run(config, func(t *testing.T) {
			directory := t.TempDir()
			if err := os.WriteFile(filepath.Join(directory, "convert.yml"), []byte(config), 0o644); err != nil {
				t.Fatal(err)
			}
			output := filepath.Join(directory, "output.rdf")
			if err := os.WriteFile(output, []byte("keep me"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := ProcessDataset(directory); err == nil {
				t.Fatal("invalid config unexpectedly succeeded")
			}
			got, err := os.ReadFile(output)
			if err != nil || string(got) != "keep me" {
				t.Fatalf("previous output was not preserved: %q, %v", got, err)
			}
		})
	}
}

func TestProcessDatasetRejectsAdditionalYAMLDocument(t *testing.T) {
	directory := copyMinimalDataset(t)
	path := filepath.Join(directory, "convert.yml")
	config, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	config = append(config, []byte("\n---\nfiles: []\n")...)
	if err := os.WriteFile(path, config, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ProcessDataset(directory); err == nil || !strings.Contains(err.Error(), "exactly one") {
		t.Fatalf("expected multiple-document error, got %v", err)
	}
}

func TestProcessHeaderlessRoadNetwork(t *testing.T) {
	directory := t.TempDir()
	if err := os.Mkdir(filepath.Join(directory, "sources"), 0o755); err != nil {
		t.Fatal(err)
	}
	config, err := os.ReadFile(filepath.Join("..", "..", "datasets", "roadNet-CA", "convert.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "convert.yml"), config, 0o644); err != nil {
		t.Fatal(err)
	}
	data := "# Directed graph\n# FromNodeId\tToNodeId\n0\t1\n1\t2\n"
	if err := os.WriteFile(filepath.Join(directory, "sources", "roadNet-CA.txt"), []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ProcessDataset(directory); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(directory, "output.rdf"))
	if err != nil {
		t.Fatal(err)
	}
	want := "_:0 <successors> _:1 .\n_:0 <id> \"0\" .\n_:1 <id> \"1\" .\n" +
		"_:1 <successors> _:2 .\n_:1 <id> \"1\" .\n_:2 <id> \"2\" .\n"
	if string(got) != want {
		t.Fatalf("unexpected headerless conversion:\n%s", got)
	}
}

func TestFileRejectsInvalidSourcePath(t *testing.T) {
	for _, name := range []string{"", "../outside.csv", "/outside.csv"} {
		input := file{Name: name}
		if _, err := input.validate(); err == nil {
			t.Errorf("invalid source path %q accepted", name)
		}
	}
}

func TestEmptyEntityIDsAreRejected(t *testing.T) {
	schema := map[string]schemaType{"id": {dt: idType, prefix: "p"}, "value": {dt: stringType}}
	indices := map[string]uint{"id": 0, "value": 1}
	record := []string{"", "Alice"}
	input := file{
		Rdfs:         []rdfRule{{Subject: "id", Predicate: "name", Object: "value"}},
		EntityFacets: []entityFacetRule{{Id: "id", facetRule: facetRule{Key: "name", Value: "value"}}},
	}
	var output bytes.Buffer
	if err := input.writeRdfs(&output, nil, record, schema, indices); err == nil {
		t.Fatal("empty subject was accepted")
	}
	if err := input.saveFacets(make(map[string]entityFacets), record, schema, indices); err == nil {
		t.Fatal("empty facet entity ID was accepted")
	}
}

func TestCommittedDatasetConfigurations(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("..", "..", "datasets", "*", "convert.yml"))
	if err != nil || len(paths) != 4 {
		t.Fatalf("expected four dataset configs, got %v (%v)", paths, err)
	}
	for _, path := range paths {
		t.Run(filepath.Base(filepath.Dir(path)), func(t *testing.T) {
			config, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer config.Close()
			decoder := yaml.NewDecoder(config)
			decoder.KnownFields(true)
			var ds dataset
			if err := decoder.Decode(&ds); err != nil {
				t.Fatal(err)
			}
			for _, input := range ds.Files {
				if _, err := input.validate(); err != nil {
					t.Errorf("%s: %v", input.Name, err)
				}
				if len(input.Headers) > 0 {
					if err := input.validateHeaders(input.Headers); err != nil {
						t.Error(err)
					}
				}
			}
		})
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
