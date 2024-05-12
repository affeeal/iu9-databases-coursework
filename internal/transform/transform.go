package transform

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/affeeal/iu9-database-coursework/internal/transform/rdf"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Dataset struct {
	Name      string `yaml:"name"`
	Delimiter string `yaml:"delimiter"`
	Comment   string `yaml:"comment"`
	Files     []File `yaml:"files"`
}

type File struct {
	Name   string `yaml:"name"`
	Schema []struct {
		Name string `yaml:"name"`
		Type string `yaml:"type"`
	}
	Facets []Facet `yaml:"facets"`
	RDFs   []RDF   `yaml:"RDFs"`
}

type Facet struct {
	Entity string `yaml:"entity"`
	Key    string `yaml:"key"`
	Value  string `yaml:"value"`
	Type   string `yaml:"type"`
}

type RDF struct {
	Subject     string `yaml:"subject"`
	Predicat    string `yaml:"predicat"`
	Object      string `yaml:"object"`
	FacetEntity string `yaml:"facetEntity"`
}

type entitiesFacets map[string]map[string]*rdf.Term

type index uint8
type dataType uint8
type facetType uint8

type facetAux struct {
	entity index
	value  index
	ft     facetType
}

type rdfAux struct {
	subject termAux
	object  termAux
	entityI index
}

type termAux struct {
	i  index
	dt dataType
}

const (
	configName = "transform.yml"
	outputName = "transformed.rdf"
	sourceName = "source"
)

const (
	intFt facetType = iota
	floatFt
	stringFt
)

const (
	intDt dataType = iota
	floatDt
	idDt
	stringDt
)

var (
	dataTypes = map[string]dataType{
		"int":    intDt,
		"float":  floatDt,
		"id":     idDt,
		"string": stringDt,
	}

	facetTypes = map[string]facetType{
		"int":    intFt,
		"float":  floatFt,
		"string": stringFt,
	}

	facetTypeToDecoration = []rdf.Decoration{
		rdf.None,   // intFt
		rdf.None,   // floatFt
		rdf.Quotes, // stringFt
	}

	dataTypeToDecoration = []rdf.Decoration{
		rdf.Quotes, // intDt
		rdf.Quotes, // floatDt
		rdf.None,   // idDt
		rdf.Quotes, // stringDt
	}
)

func ProcessDatasets(datasetsPath string) error {
	entires, err := os.ReadDir(datasetsPath)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(len(entires))

	for _, entry := range entires {
		if !entry.IsDir() {
			wg.Done()
			continue
		}

		go func(datasetPath string) {
			defer wg.Done()

			if err := ProcessDataset(datasetPath); err != nil {
				log.Println(err)
				return
			}

			fmt.Printf("successfully processed %s\n", datasetPath)
		}(datasetsPath + "/" + entry.Name())
	}

	wg.Wait()

	return nil
}

func ProcessDataset(datasetPath string) error {
	configFile, err := os.Open(datasetPath + "/" + configName)
	if err != nil {
		return err
	}
	defer configFile.Close()

	decoder := yaml.NewDecoder(configFile)
	decoder.KnownFields(true)

	var ds Dataset
	if err = decoder.Decode(&ds); err != nil {
		return err
	}

	if err = ds.Process(datasetPath); err != nil {
		return err
	}

	return nil
}

func (ds *Dataset) Process(datasetPath string) error {
	sourcePath := datasetPath + "/" + sourceName

	output, err := os.Create(datasetPath + "/" + outputName)
	if err != nil {
		return ds.wrap(err)
	}
	defer output.Close()

	delimiter, err := ds.validateDelimiter()
	if err != nil {
		return ds.wrap(err)
	}

	comment, err := ds.validateComment()
	if err != nil {
		return ds.wrap(err)
	}

	esFs := make(entitiesFacets)
	for _, file := range ds.Files {
		if err = file.transform(esFs, output, sourcePath, delimiter, comment); err != nil {
			return ds.wrap(err)
		}
	}

	return nil
}

func (ds *Dataset) wrap(err error) error {
	return errors.Wrap(err, "dataset "+ds.Name)
}

func (ds *Dataset) validateDelimiter() (rune, error) {
	return validateSymbol("delimiter", ds.Delimiter, ',')
}

func (ds *Dataset) validateComment() (rune, error) {
	return validateSymbol("comment", ds.Comment, 0)
}

func validateSymbol(name, rawSym string, defaultSym rune) (rune, error) {
	if rawSym == "" {
		return defaultSym, nil
	} else if len(rawSym) != 1 {
		return 0, errors.New(name + " must be a valid rune")
	}

	sym := rune(rawSym[0])
	if sym == '\r' || sym == '\n' {
		return 0, errors.New(name + "must not be \\r, \\n")
	}

	return sym, nil
}

func (file *File) transform(
	esFs entitiesFacets,
	output *os.File,
	sourcePath string,
	delimiter, comment rune,
) error {
	schema, err := file.transformSchema()
	if err != nil {
		return err
	}

	source, err := os.Open(sourcePath + "/" + file.Name)
	if err != nil {
		return file.wrap(err)
	}
	defer source.Close()

	reader := csv.NewReader(source)
	reader.Comma = delimiter
	reader.Comment = comment

	headers, err := reader.Read()
	if err != nil {
		return file.wrap(err)
	}

	indices := make(map[string]index)
	for i, header := range headers {
		indices[header] = index(i)
	}

	fsAux, err := file.getFacetsAux(indices)
	if err != nil {
		return file.wrap(err)
	}

	rsAux, err := file.getRdfsAux(indices, schema)
	if err != nil {
		return file.wrap(err)
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return file.wrap(err)
		}

		file.saveFacets(esFs, record, fsAux)
		file.writeRdfs(output, esFs, record, rsAux)
	}

	return nil
}

func (file *File) transformSchema() (map[string]dataType, error) {
	schema := make(map[string]dataType)

	for _, item := range file.Schema {
		dt, ok := dataTypes[item.Type]
		if !ok {
			return nil, errors.New("undefined schema data type " + item.Type)
		}

		schema[item.Name] = dt
	}

	return schema, nil
}

func (file *File) saveFacets(
	esFs entitiesFacets,
	record []string,
	fsAux []facetAux,
) {
	for i, f := range file.Facets {
		fAux := fsAux[i]

		add(
			esFs,
			entityKey(f.Entity, record[fAux.entity]),
			f.Key,
			rdf.NewTerm(record[fAux.value], facetTypeToDecoration[fAux.ft]),
		)
	}
}

func (file *File) writeRdfs(
	output *os.File,
	esFs entitiesFacets,
	record []string,
	rsAux []rdfAux,
) error {
	for i, r := range file.RDFs {
		rAux := rsAux[i]

		subject := transform(
			rAux.subject.dt,
			r.Subject,
			record[rAux.subject.i],
		)

		object := transform(
			rAux.object.dt,
			r.Object,
			record[rAux.object.i],
		)

		var fs []*rdf.Facet = nil
		if r.FacetEntity != "" {
			fs = convert(esFs[entityKey(r.FacetEntity, record[rAux.entityI])])
		}

		r := rdf.NewRDF(
			rdf.NewTerm(subject, dataTypeToDecoration[rAux.subject.dt]),
			rdf.NewTerm(r.Predicat, rdf.AngleBrackets),
			rdf.NewTerm(object, dataTypeToDecoration[rAux.object.dt]),
			fs,
		)

		if _, err := output.WriteString(r.Stringln()); err != nil {
			return err
		}
	}

	return nil
}

func (file *File) wrap(err error) error {
	return errors.Wrap(err, "file "+file.Name)
}

func (file *File) getFacetsAux(indices map[string]index) ([]facetAux, error) {
	fsAux := make([]facetAux, 0, len(file.Facets))

	for _, f := range file.Facets {
		entity, ok := indices[f.Entity]
		if !ok {
			return nil, errors.New("undefined facet entity " + f.Entity)
		}

		value, ok := indices[f.Value]
		if !ok {
			return nil, errors.New("undefined facet value " + f.Value)
		}

		ft, ok := facetTypes[f.Type]
		if !ok {
			return nil, errors.New("undefined facet type " + f.Type)
		}

		fsAux = append(fsAux, facetAux{entity: entity, value: value, ft: ft})
	}

	return fsAux, nil
}

func (file *File) getRdfsAux(
	indices map[string]index,
	schema map[string]dataType,
) ([]rdfAux, error) {
	rsAux := make([]rdfAux, 0, len(file.RDFs))

	for _, r := range file.RDFs {
		subjectI, ok := indices[r.Subject]
		if !ok {
			return nil, errors.New(
				"undefined index of RDF subject " + r.Subject,
			)
		}

		subjectDt, ok := schema[r.Subject]
		if !ok {
			return nil, errors.New(
				"undefined data type of RDF subject " + r.Subject,
			)
		}

		objectI, ok := indices[r.Object]
		if !ok {
			return nil, errors.New("undefined index of RDF object " + r.Object)
		}

		objectDT, ok := schema[r.Object]
		if !ok {
			return nil, errors.New(
				"undefined data type of RDF object " + r.Object,
			)
		}

		var entityI index = 0
		if r.FacetEntity != "" {
			entityI, ok = indices[r.FacetEntity]
			if !ok {
				return nil, errors.New(
					"undefined index of RDF facet entity " + r.FacetEntity,
				)
			}
		}

		rsAux = append(
			rsAux,
			rdfAux{
				subject: termAux{i: subjectI, dt: subjectDt},
				object:  termAux{i: objectI, dt: objectDT},
				entityI: entityI,
			},
		)
	}

	return rsAux, nil
}

func transform(dt dataType, name, value string) string {
	if dt == idDt {
		return "_:" + entityKey(name, value)
	}
	return value
}

func add(esFs entitiesFacets, entity, key string, term *rdf.Term) {
	eFs, ok := esFs[entity]
	if !ok {
		eFs = make(map[string]*rdf.Term)
		esFs[entity] = eFs
	}
	eFs[key] = term
}

func convert(m map[string]*rdf.Term) []*rdf.Facet {
	s := make([]*rdf.Facet, 0, len(m))
	for key, term := range m {
		s = append(s, rdf.NewFacet(key, term))
	}
	return s
}

func entityKey(name, value string) string {
	return name + value
}
