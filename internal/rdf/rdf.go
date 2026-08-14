package rdf

import (
	"fmt"
	"strings"
)

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
		return `"` + escapeString(term.val) + `"`
	case ANGLE_BRACKETS:
		return "<" + term.val + ">"
	}

	return ""
}

func escapeString(value string) string {
	var escaped strings.Builder
	for _, char := range value {
		switch char {
		case '\\':
			escaped.WriteString(`\\`)
		case '"':
			escaped.WriteString(`\"`)
		case '\t':
			escaped.WriteString(`\t`)
		case '\b':
			escaped.WriteString(`\b`)
		case '\n':
			escaped.WriteString(`\n`)
		case '\r':
			escaped.WriteString(`\r`)
		case '\f':
			escaped.WriteString(`\f`)
		default:
			if char < 0x20 || char == 0x7f {
				fmt.Fprintf(&escaped, `\u%04X`, char)
			} else {
				escaped.WriteRune(char)
			}
		}
	}
	return escaped.String()
}

type Facet struct {
	key  string
	term *Term
}

func NewFacet(key string, term *Term) *Facet {
	return &Facet{key: key, term: term}
}

type Rdf struct {
	subject   *Term
	predicate *Term
	object    *Term
	facets    []*Facet
}

func NewRdf(
	subject *Term,
	predicate *Term,
	object *Term,
	facets []*Facet,
) *Rdf {
	return &Rdf{
		subject:   subject,
		predicate: predicate,
		object:    object,
		facets:    facets,
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

	result := fmt.Sprintf(
		"%s %s %s",
		rdf.subject.String(),
		rdf.predicate.String(),
		rdf.object.String(),
	)
	if facets != "" {
		result += " " + strings.TrimSuffix(facets, " ")
	}
	return result + " ."
}

func (rdf *Rdf) Stringln() string {
	return rdf.String() + "\n"
}
