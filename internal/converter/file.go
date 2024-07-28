package converter

import (
	"encoding/csv"
	"io"
	"os"

	"github.com/affeeal/iu9-databases-coursework/internal/rdf"
	"github.com/pkg/errors"
)

type file struct {
	Name         string            `yaml:"name"`
	Delimiter    string            `yaml:"delimiter"`
	Comment      string            `yaml:"comment"`
	Declarations []declaration     `yaml:"declarations"`
	EntityFacets []entityFacetRule `yaml:"entity_facets"`
	Rdfs         []rdfRule         `yaml:"rdfs"`
}

type declaration struct {
	Name  string            `yaml:"name"`
	Type  string            `yaml:"type"`
	Extra map[string]string `yaml:"extra"`
}

type schemaType struct {
	dt    dataType
	extra map[string]string
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

const prefixOption = "prefix"

func (f *file) process(
	entitiesFacets map[string]entityFacets,
	output *os.File,
	sourcesPath string,
) error {
	schema, err := f.validate()
	if err != nil {
		return err
	}

	source, err := os.Open(makePath(sourcesPath, f.Name))
	if err != nil {
		return err
	}
	defer source.Close()

	reader := csv.NewReader(source)

	// reader.Delimiter == ',' by default
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

	headers, err := reader.Read()
	if err != nil {
		return err
	}

	indices := make(map[string]uint)
	for i, header := range headers {
		indices[header] = uint(i)
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		f.saveFacets(entitiesFacets, record, schema, indices)
		err = f.writeRdfs(output, entitiesFacets, record, schema, indices)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *file) validate() (map[string]schemaType, error) {
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
		if _, ok := schema[decl.Name]; ok {
			return nil, errors.New("schema name " + decl.Name + " redefinition")
		}

		dt, ok := dataTypes[decl.Type]
		if !ok {
			return nil, errors.New("unknown data type " + decl.Type)
		}

		schema[decl.Name] = schemaType{
			dt:    dt,
			extra: decl.Extra,
		}
	}

	return schema, nil
}

func validateSymbol(rawSymbol string) (rune, error) {
	if len(rawSymbol) != 1 {
		return 0, errors.New(
			"special symbol " + rawSymbol + " must be a single rune",
		)
	}

	symbol := rune(rawSymbol[0])

	if symbol == '\r' || symbol == '\n' {
		return 0, errors.New(`special symbol must not be \r, \n`)
	}

	return symbol, nil
}

func (file *file) saveFacets(
	entitiesFacets map[string]entityFacets,
	record []string,
	schema map[string]schemaType,
	indices map[string]uint,
) {
	for _, rule := range file.EntityFacets {
		addFacet(
			entitiesFacets,
			makeEntityKey(
				schema[rule.Id].extra[prefixOption],
				record[indices[rule.Id]],
			),
			rule.Key,
			rdf.NewTerm(
				record[indices[rule.Value]],
				facetDecorations[schema[rule.Value].dt],
			),
		)
	}
}

func (file *file) writeRdfs(
	output *os.File,
	entitiesFacets map[string]entityFacets,
	record []string,
	schema map[string]schemaType,
	indices map[string]uint,
) error {
	for _, rule := range file.Rdfs {
		objectIndex := indices[rule.Object]
		if record[objectIndex] == "" {
			continue
		}

		subject := makeBlankNode(
			makeEntityKey(
				schema[rule.Subject].extra[prefixOption],
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
					objectType.extra[prefixOption],
					record[objectIndex],
				),
			)
		} else {
			object = record[objectIndex]
		}

		var facets []*rdf.Facet = nil
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
				schema[rule.EntityFacetsId].extra[prefixOption],
				record[indices[rule.EntityFacetsId]],
			)

			facets = append(facets, convertFacets(entitiesFacets[entityKey])...)
		}

		r := rdf.NewRdf(
			rdf.NewTerm(subject, rdf.NONE),
			rdf.NewTerm(rule.Predicat, rdf.ANGLE_BRACKETS),
			rdf.NewTerm(object, termDecorations[objectType.dt]),
			facets,
		)

		_, err := output.WriteString(r.Stringln())
		if err != nil {
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
	for key, term := range ef {
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
