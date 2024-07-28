package converter

import "github.com/pkg/errors"

type rdfRule struct {
	Subject        string      `yaml:"subject"`
	Predicat       string      `yaml:"predicat"`
	Object         string      `yaml:"object"`
	CastObjectTo   string      `yaml:"cast_object_to"`
	Facets         []facetRule `yaml:"facets"`
	EntityFacetsId string      `yaml:"entity_facets_id"`
}

type entityFacetRule struct {
	facetRule `       yaml:",inline"`
	Id        string `yaml:"id"`
}

type facetRule struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

func (rule *rdfRule) validate(schema map[string]schemaType) error {
	err := validateSchemaType(schema, "RDF subject", rule.Subject, true)
	if err != nil {
		return err
	}

	_, err = validateSchemaName(schema, "RDF object", rule.Object)
	if err != nil {
		return err
	}

	// TODO: validate cast_object_to type

	for _, rule := range rule.Facets {
		if err = rule.validate(schema, "RDF facet"); err != nil {
			return err
		}
	}

	if rule.EntityFacetsId != "" {
		return validateSchemaType(
			schema,
			"RDF entity facets id",
			rule.EntityFacetsId,
			true,
		)
	}

	return nil
}

func (rule *entityFacetRule) validate(schema map[string]schemaType) error {
	err := validateSchemaType(schema, "entity facet id", rule.Id, true)
	if err != nil {
		return err
	}

	return rule.facetRule.validate(schema, "entity facet")
}

func (rule *facetRule) validate(
	schema map[string]schemaType,
	context string,
) error {
	return validateSchemaType(schema, context+" value", rule.Value, false)
}

func validateSchemaType(
	schema map[string]schemaType,
	context, name string,
	supposedToBeId bool,
) error {
	st, err := validateSchemaName(schema, context, name)
	if err != nil {
		return err
	}

	if isId := (st.dt == idType); isId != supposedToBeId {
		if isId {
			return errors.New(context + " " + name + " data type must be an id")
		}

		return errors.New(context + " " + name + " data type must not be an id")
	}

	return nil
}

func validateSchemaName(
	schema map[string]schemaType,
	context, name string,
) (schemaType, error) {
	st, ok := schema[name]
	if !ok {
		return st, errors.New("unknown " + context + " " + name)
	}

	return st, nil
}
