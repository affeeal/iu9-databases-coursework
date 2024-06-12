package converter

import (
	"encoding/csv"
	"io"
	"os"

	"github.com/affeeal/iu9-database-coursework/internal/rdf"
	"github.com/pkg/errors"
)

type File struct {
	Name         string        `yaml:"name"`
	Delimiter    string        `yaml:"delimiter"`
	Comment      string        `yaml:"comment"`
	Declarations []Declaration `yaml:"declarations"`
	FacetsToSave []FacetRule   `yaml:"facets_to_save"`
	RdfsToWrite  []RdfRule     `yaml:"rdfs_to_write"`
}

type Declaration struct {
	Name  string            `yaml:"name"`
	Type  string            `yaml:"type"`
	Extra map[string]string `yaml:"extra"`
}

type FacetRule struct {
	Entity string `yaml:"entity"`
	Key    string `yaml:"key"`
	Value  string `yaml:"value"`
}

type RdfRule struct {
	Subject     string `yaml:"subject"`
	Predicat    string `yaml:"predicat"`
	Object      string `yaml:"object"`
	FacetEntity string `yaml:"facet_entity"`
}

type schemaType struct {
	dt    dataType
	extra map[string]string
}

type dataType uint

const (
	INT dataType = iota
	FLOAT
	STRING
	ID
)

var (
	dataTypes = map[string]dataType{
		"int":    INT,
		"float":  FLOAT,
		"string": STRING,
		"id":     ID,
	}

	termDecorations = []rdf.Decoration{
		rdf.QUOTES, // INT
		rdf.QUOTES, // FLOAT
		rdf.QUOTES, // STRING
		rdf.NONE,   // ID
	}

	// NOTE: no ID
	facetDecorations = []rdf.Decoration{
		rdf.NONE,   // INT
		rdf.NONE,   // FLOAT
		rdf.QUOTES, // STRING
	}
)

const PREFIX_OPTION = "prefix"

func (file *File) process(
	entitiesFacets map[string]entityFacets,
	output *os.File,
	sourcesPath string,
) error {
	schema, err := file.validate()
	if err != nil {
		return err
	}

	source, err := os.Open(makePath(sourcesPath, file.Name))
	if err != nil {
		return err
	}
	defer source.Close()

	reader := csv.NewReader(source)

	if file.Delimiter != "" {
		delimiter, err := validateSymbol(file.Delimiter)
		if err != nil {
			return err
		}

		reader.Comma = delimiter
	}

	if file.Comment != "" {
		comment, err := validateSymbol(file.Comment)
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

		file.saveFacets(entitiesFacets, record, schema, indices)
		if err = file.writeRdfs(output, entitiesFacets, record, schema, indices); err != nil {
			return err
		}
	}

	return nil
}

func (file *File) validate() (map[string]schemaType, error) {
	schema, err := file.validateDeclarations()
	if err != nil {
		return nil, err
	}

	if err = file.validateFacetRules(schema); err != nil {
		return nil, err
	}

	if err = file.validateRdfRules(schema); err != nil {
		return nil, err
	}

	return schema, nil
}

func (file *File) validateDeclarations() (map[string]schemaType, error) {
	schema := make(map[string]schemaType)

	for _, decl := range file.Declarations {
		if _, ok := schema[decl.Name]; ok {
			return nil, errors.New("Schema name " + decl.Name + " redefinition")
		}

		dt, ok := dataTypes[decl.Type]
		if !ok {
			return nil, errors.New("Unknown data type " + decl.Type)
		}

		schema[decl.Name] = schemaType{
			dt:    dt,
			extra: decl.Extra,
		}
	}

	return schema, nil
}

func (file *File) validateFacetRules(schema map[string]schemaType) error {
	for _, rule := range file.FacetsToSave {
		if err := validateId(schema, rule.Entity, "facet entity"); err != nil {
			return err
		}

		if _, err := validateName(schema, rule.Value, "facet value"); err != nil {
			return err
		}
	}

	return nil
}

func (file *File) validateRdfRules(schema map[string]schemaType) error {
	for _, rule := range file.RdfsToWrite {
		if err := validateId(schema, rule.Subject, "RDF subject"); err != nil {
			return err
		}

		if _, err := validateName(schema, rule.Object, "RDF object"); err != nil {
			return err
		}

		if rule.FacetEntity == "" {
			continue
		}

		if err := validateId(schema, rule.FacetEntity, "RDF facet entity"); err != nil {
			return err
		}
	}

	return nil
}

func validateName(
	schema map[string]schemaType,
	name, role string,
) (schemaType, error) {
	st, ok := schema[name]
	if !ok {
		return st, errors.New("Unknown " + role + " " + name)
	}

	return st, nil
}

func validateId(schema map[string]schemaType, name, role string) error {
	st, err := validateName(schema, name, role)
	if err != nil {
		return err
	}

	if st.dt != ID {
		return errors.New(
			"Data type of " + role + " " + name + " must be an id",
		)
	}

	return nil
}

func validateSymbol(rawSymbol string) (rune, error) {
	if len(rawSymbol) != 1 {
		return 0, errors.New(
			"Special symbol " + rawSymbol + " must be a single rune",
		)
	}

	symbol := rune(rawSymbol[0])

	if symbol == '\r' || symbol == '\n' {
		return 0, errors.New(`Special symbol must not be \r, \n`)
	}

	return symbol, nil
}

func (file *File) saveFacets(
	entitiesFacets map[string]entityFacets,
	record []string,
	schema map[string]schemaType,
	indices map[string]uint,
) {
	for _, rule := range file.FacetsToSave {
		addFacet(
			entitiesFacets,
			makeEntityKey(
				schema[rule.Entity].extra[PREFIX_OPTION],
				record[indices[rule.Entity]],
			),
			rule.Key,
			rdf.NewTerm(
				record[indices[rule.Value]],
				facetDecorations[schema[rule.Value].dt],
			),
		)
	}
}

func (file *File) writeRdfs(
	output *os.File,
	entitiesFacets map[string]entityFacets,
	record []string,
	schema map[string]schemaType,
	indices map[string]uint,
) error {
	for _, rule := range file.RdfsToWrite {
		objectIndex := indices[rule.Object]
		if record[objectIndex] == "" {
			continue
		}

		subject := makeBlankNode(
			makeEntityKey(
				schema[rule.Subject].extra[PREFIX_OPTION],
				record[indices[rule.Subject]],
			),
		)

		var object string
		objectType := schema[rule.Object]
		if objectType.dt == ID {
			object = makeBlankNode(
				makeEntityKey(
					objectType.extra[PREFIX_OPTION],
					record[objectIndex],
				),
			)
		} else {
			object = record[objectIndex]
		}

		var facets []*rdf.Facet = nil
		if rule.FacetEntity != "" {
			entityKey := makeEntityKey(
				schema[rule.FacetEntity].extra[PREFIX_OPTION],
				record[indices[rule.FacetEntity]],
			)

			facets = convertFacets(entitiesFacets[entityKey])
		}

		r := rdf.NewRdf(
			rdf.NewTerm(subject, rdf.NONE),
			rdf.NewTerm(rule.Predicat, rdf.ANGLE_BRACKETS),
			rdf.NewTerm(object, termDecorations[objectType.dt]),
			facets,
		)

		if _, err := output.WriteString(r.Stringln()); err != nil {
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
