package rdf

import "fmt"

type Decoration uint8

const (
	NONE Decoration = iota
	QUOTES
	ANGLE_BRACKETS
)

type Term struct {
	val string
	dec Decoration
}

func NewTerm(val string, dec Decoration) *Term {
	return &Term{val: val, dec: dec}
}

func (term *Term) String() string {
	switch term.dec {
	case NONE:
		return term.val
	case QUOTES:
		return `"` + term.val + `"`
	case ANGLE_BRACKETS:
		return "<" + term.val + ">"
	}

	return ""
}

type Facet struct {
	key  string
	term *Term
}

func NewFacet(key string, term *Term) *Facet {
	return &Facet{key: key, term: term}
}

type Rdf struct {
	subject  *Term
	predicat *Term
	object   *Term
	facets   []*Facet
}

func NewRdf(
	subject *Term,
	predicat *Term,
	object *Term,
	facets []*Facet,
) *Rdf {
	return &Rdf{
		subject:  subject,
		predicat: predicat,
		object:   object,
		facets:   facets,
	}
}

func (rdf *Rdf) String() string {
	facets := ""
	if len(rdf.facets) > 0 {
		facets = "("
		for i, facet := range rdf.facets {
			if i > 0 {
				facets += ", "
			}
			facets += facet.key + "=" + facet.term.String()
		}
		facets += ") "
	}

	return fmt.Sprintf(
		"%s %s %s %s.",
		rdf.subject.String(),
		rdf.predicat.String(),
		rdf.object.String(),
		facets,
	)
}

func (rdf *Rdf) Stringln() string {
	return rdf.String() + "\n"
}
