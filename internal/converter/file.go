package converter

import (
	"encoding/csv"
	"io"
	"os"

	"github.com/affeeal/iu9-database-coursework/internal/converter/config"
	"github.com/affeeal/iu9-database-coursework/internal/converter/rdf"
	"github.com/pkg/errors"
)

type index uint8
type dataType uint8

type meta struct {
	raw *config.Record
	dt  dataType
}

type term struct {
	i     index
	dt    dataType
	extra map[string]string
}

type preparedFacet struct {
	raw    *config.Facet
	entity *term
	value  *term
}

type preparedRdf struct {
	raw         *config.Rdf
	subject     *term
	object      *term
	facetEntity *term
}

const prefix = "prefix"

const (
	intDt dataType = iota
	floatDt
	stringDt
	idDt
)

var (
	dataTypes = map[string]dataType{
		"int":    intDt,
		"float":  floatDt,
		"string": stringDt,
		"id":     idDt,
	}

	facetTypes = map[string]dataType{
		"int":    intDt,
		"float":  floatDt,
		"string": stringDt,
	}

	facetDecorations = []rdf.Decoration{
		rdf.None,   // intDt
		rdf.None,   // floatDt
		rdf.Quotes, // stringDt
	}

	termDecorations = []rdf.Decoration{
		rdf.Quotes, // intDt
		rdf.Quotes, // floatDt
		rdf.Quotes, // stringDt
		rdf.None,   // idDt
	}
)

func processFile(
	file *config.File,
	ef entitiesFacets,
	out *os.File,
	srcPath string,
	del, com rune,
) error {
	src, err := os.Open(makePath(srcPath, file.Name))
	if err != nil {
		return err
	}
	defer src.Close()

	reader := csv.NewReader(src)
	reader.Comma = del
	reader.Comment = com

	headers, err := reader.Read()
	if err != nil {
		return err
	}
	indices := make(map[string]index)
	for i, header := range headers {
		indices[header] = index(i)
	}

	schema, err := transformSchema(file.Schema)
	if err != nil {
		return err
	}
	fs, err := transformFacets(file.Facets, indices, schema)
	if err != nil {
		return err
	}
	rs, err := transformRdfs(file.Rdfs, indices, schema)
	if err != nil {
		return err
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		saveFacets(ef, fs, record)
		if err = writeRdfs(out, ef, rs, record); err != nil {
			return err
		}
	}

	return nil
}

func transformSchema(configSchema []config.Record) (map[string]meta, error) {
	schema := make(map[string]meta)

	for _, record := range configSchema {
		dt, ok := dataTypes[record.Type]
		if !ok {
			return nil, errors.New("undefined schema type " + record.Type)
		}

		schema[record.Name] = meta{raw: &record, dt: dt}
	}

	return schema, nil
}

func transformFacets(
	configFacets []config.Facet,
	indices map[string]index,
	schema map[string]meta,
) ([]preparedFacet, error) {
	facets := make([]preparedFacet, 0, len(configFacets))

	for _, cf := range configFacets {
		entity, err := newTerm(cf.Entity, schema, indices)
		if err != nil {
			return nil, err
		}

		value, err := newTerm(cf.Value, schema, indices)
		if err != nil {
			return nil, err
		}

		facets = append(
			facets,
			preparedFacet{raw: &cf, entity: entity, value: value},
		)
	}

	return facets, nil
}

func transformRdfs(
	configRdfs []config.Rdf,
	indices map[string]index,
	schema map[string]meta,
) ([]preparedRdf, error) {
	rs := make([]preparedRdf, 0, len(configRdfs))

	for _, r := range configRdfs {
		subject, err := newTerm(r.Subject, schema, indices)
		if err != nil {
			return nil, err
		}

		object, err := newTerm(r.Object, schema, indices)
		if err != nil {
			return nil, err
		}

		var facetEntity *term = nil
		if r.FacetEntity != "" {
			facetEntity, err = newTerm(r.FacetEntity, schema, indices)
			if err != nil {
				return nil, err
			}
		}

		rs = append(
			rs,
			preparedRdf{
				raw:         &r,
				subject:     subject,
				object:      object,
				facetEntity: facetEntity,
			},
		)
	}

	return rs, nil
}

func newTerm(
	name string,
	schema map[string]meta,
	indices map[string]index,
) (*term, error) {
	i, ok := indices[name]
	if !ok {
		return nil, errors.New("undefined index for name " + name)
	}

	m, ok := schema[name]
	if !ok {
		return nil, errors.New("undefined schema meta for " + name)
	}

	return &term{i: i, dt: m.dt, extra: m.raw.Extra}, nil
}

func saveFacets(ef entitiesFacets, fs []preparedFacet, record []string) {
	for _, f := range fs {
		add(
			ef,
			entityKey(f.entity.extra[prefix], record[f.entity.i]),
			f.raw.Key,
			rdf.NewTerm(record[f.value.i], facetDecorations[f.value.dt]),
		)
	}
}

func writeRdfs(
	output *os.File,
	ef entitiesFacets,
	rs []preparedRdf,
	record []string,
) error {
	for _, r := range rs {
		subject := blankNode(
			entityKey(r.subject.extra[prefix], record[r.subject.i]),
		)

		var object string
		if r.object.dt == idDt {
			object = blankNode(
				entityKey(r.object.extra[prefix], record[r.object.i]),
			)
		} else {
			object = record[r.object.i]
		}

		var fs []*rdf.Facet = nil
		if r.facetEntity != nil {
			fs = convert(
				ef[entityKey(r.facetEntity.extra[prefix], record[r.facetEntity.i])],
			)
		}

		r := rdf.NewRdf(
			rdf.NewTerm(subject, rdf.None),
			rdf.NewTerm(r.raw.Predicat, rdf.AngleBrackets),
			rdf.NewTerm(object, termDecorations[r.object.dt]),
			fs,
		)

		if _, err := output.WriteString(r.Stringln()); err != nil {
			return err
		}
	}

	return nil
}

func add(ef entitiesFacets, entity, key string, t *rdf.Term) {
	fs, ok := ef[entity]
	if !ok {
		fs = make(map[string]*rdf.Term)
		ef[entity] = fs
	}
	fs[key] = t
}

func convert(m map[string]*rdf.Term) []*rdf.Facet {
	s := make([]*rdf.Facet, 0, len(m))
	for key, term := range m {
		s = append(s, rdf.NewFacet(key, term))
	}
	return s
}

func blankNode(id string) string {
	return "_:" + id
}

func entityKey(prefix, value string) string {
	return prefix + value
}
