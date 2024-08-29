package converter

import "github.com/pkg/errors"

type declaration struct {
	Name   string `yaml:"name"`
	Type   string `yaml:"type"`
	Prefix string `yaml:"prefix"`
}

func (d *declaration) validate(schema map[string]schemaType) error {
	if _, ok := schema[d.Name]; ok {
		return errors.New("Schema name " + d.Name + " redefinition")
	}

	dt, ok := dataTypes[d.Type]
	if !ok {
		return errors.New("Unknown data type " + d.Type)
	}

	schema[d.Name] = schemaType{
		dt:     dt,
		prefix: d.Prefix,
	}

	return nil
}

func (d *declaration) empty() bool {
	return *d == declaration{}
}
