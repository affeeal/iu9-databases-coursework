package converter

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"unicode/utf8"

	"github.com/affeeal/iu9-databases-coursework/internal/rdf"
	"github.com/pkg/errors"
)

type file struct {
	Name                  string            `yaml:"name"`
	Headers               []string          `yaml:"headers"`
	Delimiter             string            `yaml:"delimiter"`
	Comment               string            `yaml:"comment"`
	Declarations          []declaration     `yaml:"declarations"`
	ArtificialDeclaration declaration       `yaml:"artificial_declaration"`
	EntityFacets          []entityFacetRule `yaml:"entity_facets"`
	Rdfs                  []rdfRule         `yaml:"rdfs"`
}

type schemaType struct {
	dt     dataType
	prefix string
}

type dataType uint

const (
	intType dataType = iota
	floatType
	stringType
	idType
)

var (
	dataTypes = map[string]dataType{
		"int":    intType,
		"float":  floatType,
		"string": stringType,
		"id":     idType,
	}

	termDecorations = []rdf.Decoration{
		rdf.QUOTES, // intType
		rdf.QUOTES, // floatType
		rdf.QUOTES, // stringType
		rdf.NONE,   // idType
	}

	// NOTE: idType cannot be a facet
	facetDecorations = []rdf.Decoration{
		rdf.NONE,   // intType
		rdf.NONE,   // floatType
		rdf.QUOTES, // stringType
	}
)

func (f *file) process(
	entitiesFacets map[string]entityFacets,
	output io.Writer,
	sourcesPath string,
) error {
	schema, err := f.validate()
	if err != nil {
		return err
	}

	source, err := os.Open(filepath.Join(sourcesPath, f.Name))
	if err != nil {
		return err
	}
	defer source.Close()

	reader := csv.NewReader(source)
	err = f.adjustReader(reader)
	if err != nil {
		return err
	}

	headers := append([]string(nil), f.Headers...)
	if len(headers) == 0 {
		headers, err = reader.Read()
		if err != nil {
			return err
		}
	} else {
		reader.FieldsPerRecord = len(headers)
	}

	if err := f.validateHeaders(headers); err != nil {
		return err
	}
	if !f.ArtificialDeclaration.empty() {
		headers = append(headers, f.ArtificialDeclaration.Name)
	}
	indices := make(map[string]uint)
	for i, header := range headers {
		indices[header] = uint(i)
	}

	for artificialId := 0; true; artificialId++ {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if !f.ArtificialDeclaration.empty() {
			record = append(record, fmt.Sprint(artificialId))
		}
		if err := f.saveFacets(entitiesFacets, record, schema, indices); err != nil {
			return fmt.Errorf("record %d: %w", artificialId+1, err)
		}
		err = f.writeRdfs(output, entitiesFacets, record, schema, indices)
		if err != nil {
			return fmt.Errorf("record %d: %w", artificialId+1, err)
		}
	}

	return nil
}

func (f *file) validateHeaders(headers []string) error {
	seen := make(map[string]struct{}, len(headers))
	for _, header := range headers {
		if _, exists := seen[header]; exists {
			return fmt.Errorf("%s: duplicate input header %q", f.Name, header)
		}
		seen[header] = struct{}{}
	}
	for _, declaration := range f.Declarations {
		if _, exists := seen[declaration.Name]; !exists {
			return fmt.Errorf(
				"%s: missing declared column %q",
				f.Name,
				declaration.Name,
			)
		}
	}
	if !f.ArtificialDeclaration.empty() {
		if _, exists := seen[f.ArtificialDeclaration.Name]; exists {
			return fmt.Errorf(
				"%s: artificial column %q duplicates an input header",
				f.Name,
				f.ArtificialDeclaration.Name,
			)
		}
	}
	return nil
}

func (f *file) validate() (map[string]schemaType, error) {
	if !filepath.IsLocal(f.Name) {
		return nil, fmt.Errorf("source name must be a nonempty relative path inside sources: %q", f.Name)
	}
	schema, err := f.validateDeclarations()
	if err != nil {
		return nil, err
	}

	for _, rule := range f.EntityFacets {
		if err = rule.validate(schema); err != nil {
			return nil, err
		}
	}

	for _, rule := range f.Rdfs {
		if err = rule.validate(schema); err != nil {
			return nil, err
		}
	}

	return schema, nil
}

func (f *file) validateDeclarations() (map[string]schemaType, error) {
	schema := make(map[string]schemaType)
	for _, decl := range f.Declarations {
		if err := decl.validate(schema); err != nil {
			return nil, err
		}
	}

	if !f.ArtificialDeclaration.empty() {
		if err := f.ArtificialDeclaration.validate(schema); err != nil {
			return nil, err
		}
	}

	return schema, nil
}

func (f *file) adjustReader(reader *csv.Reader) error {
	// reader.Comma == ',' by default
	if f.Delimiter != "" {
		delimiter, err := validateSymbol(f.Delimiter)
		if err != nil {
			return err
		}

		reader.Comma = delimiter
	}

	// reader.Comment == 0 by default
	if f.Comment != "" {
		comment, err := validateSymbol(f.Comment)
		if err != nil {
			return err
		}

		reader.Comment = comment
	}

	return nil
}

func validateSymbol(rawSymbol string) (rune, error) {
	if !utf8.ValidString(rawSymbol) || utf8.RuneCountInString(rawSymbol) != 1 {
		return 0, errors.New(
			"special symbol " + rawSymbol + " must be a single rune",
		)
	}

	symbol, _ := utf8.DecodeRuneInString(rawSymbol)
	if symbol == '\r' || symbol == '\n' || symbol == 0 || symbol == '"' || symbol == utf8.RuneError {
		return 0, fmt.Errorf("invalid CSV delimiter or comment character: %q", symbol)
	}

	return symbol, nil
}

func (file *file) saveFacets(
	entitiesFacets map[string]entityFacets,
	record []string,
	schema map[string]schemaType,
	indices map[string]uint,
) error {
	for _, rule := range file.EntityFacets {
		if record[indices[rule.Id]] == "" {
			return fmt.Errorf("entity facet ID %q must not be empty", rule.Id)
		}
		addFacet(
			entitiesFacets,
			makeEntityKey(
				schema[rule.Id].prefix,
				record[indices[rule.Id]],
			),
			rule.Key,
			rdf.NewTerm(
				record[indices[rule.Value]],
				facetDecorations[schema[rule.Value].dt],
			),
		)
	}
	return nil
}

func (f *file) writeRdfs(
	output io.Writer,
	entitiesFacets map[string]entityFacets,
	record []string,
	schema map[string]schemaType,
	indices map[string]uint,
) error {
	for _, rule := range f.Rdfs {
		objectIndex := indices[rule.Object]
		if record[objectIndex] == "" {
			continue
		}
		if record[indices[rule.Subject]] == "" {
			return fmt.Errorf("RDF subject %q must not be empty", rule.Subject)
		}

		subject := makeBlankNode(
			makeEntityKey(
				schema[rule.Subject].prefix,
				record[indices[rule.Subject]],
			),
		)

		objectType := schema[rule.Object]
		if rule.CastObjectTo != "" {
			objectType.dt = dataTypes[rule.CastObjectTo]
		}

		var object string
		if objectType.dt == idType {
			object = makeBlankNode(
				makeEntityKey(
					objectType.prefix,
					record[objectIndex],
				),
			)
		} else {
			object = record[objectIndex]
		}

		var facets []*rdf.Facet
		for _, rule := range rule.Facets {
			facets = append(
				facets,
				rdf.NewFacet(
					rule.Key,
					rdf.NewTerm(
						record[indices[rule.Value]],
						facetDecorations[schema[rule.Value].dt],
					),
				),
			)
		}

		if rule.EntityFacetsId != "" {
			entityKey := makeEntityKey(
				schema[rule.EntityFacetsId].prefix,
				record[indices[rule.EntityFacetsId]],
			)

			facets = append(facets, convertFacets(entitiesFacets[entityKey])...)
		}

		r := rdf.NewRdf(
			rdf.NewTerm(subject, rdf.NONE),
			rdf.NewTerm(rule.Predicate, rdf.ANGLE_BRACKETS),
			rdf.NewTerm(object, termDecorations[objectType.dt]),
			facets,
		)

		if _, err := io.WriteString(output, r.Stringln()); err != nil {
			return err
		}
	}

	return nil
}

func addFacet(
	entitiesFacets map[string]entityFacets,
	entityKey, facetKey string,
	term *rdf.Term,
) {
	ef, ok := entitiesFacets[entityKey]
	if !ok {
		ef = make(map[string]*rdf.Term)
		entitiesFacets[entityKey] = ef
	}

	ef[facetKey] = term
}

func convertFacets(ef entityFacets) []*rdf.Facet {
	s := make([]*rdf.Facet, 0, len(ef))
	keys := make([]string, 0, len(ef))
	for key := range ef {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		term := ef[key]
		s = append(s, rdf.NewFacet(key, term))
	}
	return s
}

func makeBlankNode(id string) string {
	return "_:" + id
}

func makeEntityKey(prefix, value string) string {
	return prefix + value
}
