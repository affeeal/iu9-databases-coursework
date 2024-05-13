package config

type Dataset struct {
	Name      string `yaml:"name"`
	Delimiter string `yaml:"delimiter"`
	Comment   string `yaml:"comment"`
	Files     []File `yaml:"files"`
}

type File struct {
	Name   string   `yaml:"name"`
	Schema []Record `yaml:"schema"`
	Facets []Facet  `yaml:"facets"`
	Rdfs   []Rdf    `yaml:"rdfs"`
}

type Record struct {
	Name  string            `yaml:"name"`
	Type  string            `yaml:"type"`
	Extra map[string]string `yaml:"extra"`
}

type Facet struct {
	Entity string `yaml:"entity"`
	Key    string `yaml:"key"`
	Value  string `yaml:"value"`
}

type Rdf struct {
	Subject     string `yaml:"subject"`
	Predicat    string `yaml:"predicat"`
	Object      string `yaml:"object"`
	FacetEntity string `yaml:"facet_entity"`
}
