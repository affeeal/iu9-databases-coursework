package rdf

import "fmt"

type decoration uint8

const (
	None decoration = iota
	Quotes
	AngleBrackets
)

type Term struct {
	value string
	decor decoration
}

func NewTerm(value string, decor decoration) *Term {
	return &Term{
		value: value,
		decor: decor,
	}
}

func (term *Term) String() string {
	switch term.decor {
	case None:
		return term.value
	case Quotes:
		return `"` + term.value + `"`
	case AngleBrackets:
		return "<" + term.value + ">"
	}

	panic("Bad term decoration type")
}

type Facet struct {
	key  string
	term *Term
}

func NewFacet(key string, term *Term) *Facet {
	return &Facet{
		key:  key,
		term: term,
	}
}

type Triple struct {
	subject  *Term
	predicat *Term
	object   *Term
	facets   []*Facet
}

func NewTriple(
	subject *Term,
	predicat *Term,
	object *Term,
	facets []*Facet,
) *Triple {
	return &Triple{
		subject:  subject,
		predicat: predicat,
		object:   object,
		facets:   facets,
	}
}

func (triple *Triple) String() string {
	facets := ""
	if len(triple.facets) > 0 {
		facets = "("
		for i, f := range triple.facets {
			if i > 0 {
				facets += ", "
			}
			facets += f.key + "=" + f.term.String()
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

func BlankNode(id string) string {
	return "_:" + id
}
