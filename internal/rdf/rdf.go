package rdf

import "fmt"

type Decoration uint8

const (
	Quoted Decoration = iota
	Angled
	Intact
)

type Term struct {
	Value string
	Type  Decoration
}

func (term Term) String() string {
	switch term.Type {
	case Quoted:
		return `"` + term.Value + `"`
	case Angled:
		return `<` + term.Value + `>`
	default:
		return term.Value
	}
}

type Facet struct {
	Key  string
	Term Term
}

type Triple struct {
	Subject  Term
	Predicat Term
	Object   Term
	Facets   []Facet
}

func (triple Triple) String() string {
	var facets string
	if len(triple.Facets) > 0 {
		facets = "("
		for i, f := range triple.Facets {
			if i > 0 {
				facets += ", "
			}

			facets += f.Key + "=" + f.Term.String()
		}

		facets += ") "
	}

	return fmt.Sprintf(
		"%s %s %s %s.",
		triple.Subject.String(),
		triple.Predicat.String(),
		triple.Object.String(),
		facets,
	)
}
