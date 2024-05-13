package rdf

import "fmt"

type Decoration uint8

const (
	None Decoration = iota
	Quotes
	AngleBrackets
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
	case None:
		return term.val
	case Quotes:
		return `"` + term.val + `"`
	case AngleBrackets:
		return "<" + term.val + ">"
	}

	panic("bad term decoration type")
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

func (triple *Rdf) String() string {
	facets := ""
	if len(triple.facets) > 0 {
		facets = "("
		for i, facet := range triple.facets {
			if i > 0 {
				facets += ", "
			}
			facets += facet.key + "=" + facet.term.String()
		}
		facets += ") "
	}

	return fmt.Sprintf(
		"%s %s %s %s.",
		triple.subject.String(),
		triple.predicat.String(),
		triple.object.String(),
		facets,
	)
}

func (triple *Rdf) Stringln() string {
	return triple.String() + "\n"
}
