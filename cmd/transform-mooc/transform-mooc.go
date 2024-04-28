package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/affeeal/iu9-database-coursework/internal/rdf"
)

type actionFacets = map[string]map[string]*rdf.Term

var (
	actionsFilename        string
	actionLabelsFilename   string
	actionFeaturesFilename string
	outputFilename         string
)

func add(m actionFacets, actionId, key string, term *rdf.Term) {
	mm, ok := m[actionId]
	if !ok {
		mm = make(map[string]*rdf.Term)
		m[actionId] = mm
	}

	if _, ok = mm[key]; !ok {
		mm[key] = term
	} else {
		log.Printf("Ignore handeled actionId %s\n", actionId)
	}
}

func convert(m map[string]*rdf.Term) []*rdf.Facet {
	facets := make([]*rdf.Facet, 0, len(m))
	for key, term := range m {
		facets = append(facets, rdf.NewFacet(key, term))
	}
	return facets
}

func fillActionLabels(actionFacets actionFacets, actionLabelsFilename string) {
	const (
		actionIdIdx = iota
		labelIdx
	)

	actionLabelsFile, err := os.Open(actionLabelsFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer actionLabelsFile.Close()

	reader := csv.NewReader(actionLabelsFile)
	reader.Comma = '\t'

	// Skip headers
	record, err := reader.Read()
	if err != nil {
		log.Fatal(err)
	}

	for {
		record, err = reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		actionId := record[actionIdIdx]
		label, err := strconv.ParseBool(record[labelIdx])
		if err != nil {
			log.Fatal(err)
		}

		add(
			actionFacets,
			actionId,
			"label",
			rdf.NewTerm(strconv.FormatBool(label), rdf.None),
		)
	}
}

func fillActionFeatures(
	actionFacets actionFacets,
	actionFeaturesFilename string,
) {
	const (
		actionIdIdx = iota
		feature0Idx

		features = 4
	)

	actionFeaturesFile, err := os.Open(actionFeaturesFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer actionFeaturesFile.Close()

	reader := csv.NewReader(actionFeaturesFile)
	reader.Comma = '\t'

	// Skip headers
	record, err := reader.Read()
	if err != nil {
		log.Fatal(err)
	}

	for {
		record, err = reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		actionId := record[actionIdIdx]
		for i := 0; i < features; i++ {
			add(
				actionFacets,
				actionId,
				"feature"+strconv.Itoa(i),
				rdf.NewTerm(record[feature0Idx+i], rdf.None),
			)
		}
	}
}

func handleActions(
	outputFile *os.File,
	actionFacets actionFacets,
	actionsFilename string,
) {
	const (
		actionIdIdx = iota
		userIdIdx
		targetIdIdx
		timestampIdx
	)

	actionsFile, err := os.Open(actionsFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer actionsFile.Close()

	reader := csv.NewReader(actionsFile)
	reader.Comma = '\t'

	// Skip headers
	record, err := reader.Read()
	if err != nil {
		log.Fatal(err)
	}

	for {
		record, err = reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		actionId := record[actionIdIdx]
		add(
			actionFacets,
			actionId,
			"timestamp",
			rdf.NewTerm(record[timestampIdx], rdf.None),
		)

		blankUserId := rdf.BlankNode("u" + record[userIdIdx])
		blankActionId := rdf.BlankNode("a" + actionId)
		blankTargetId := rdf.BlankNode("t" + record[targetIdIdx])

		performs := rdf.NewTriple(
			rdf.NewTerm(blankUserId, rdf.None),
			rdf.NewTerm("performs", rdf.AngleBrackets),
			rdf.NewTerm(blankActionId, rdf.None),
			convert(actionFacets[actionId]),
		)

		on := rdf.NewTriple(
			rdf.NewTerm(blankActionId, rdf.None),
			rdf.NewTerm("on", rdf.AngleBrackets),
			rdf.NewTerm(blankTargetId, rdf.None),
			[]*rdf.Facet{},
		)

		_, err = outputFile.WriteString(
			fmt.Sprintf("%s\n%s\n", performs.String(), on.String()),
		)
		if err != nil {
			log.Fatal()
		}
	}
}

func init() {
	flag.StringVar(
		&actionsFilename,
		"mooc-actions",
		"datasets/act-mooc/mooc_actions.tsv",
		"path to mooc_actions.tsv",
	)
	flag.StringVar(
		&actionLabelsFilename,
		"mooc-action-labels",
		"datasets/act-mooc/mooc_action_labels.tsv",
		"path to mooc_action_labels.tsv",
	)
	flag.StringVar(
		&actionFeaturesFilename,
		"mooc-action-features",
		"datasets/act-mooc/mooc_action_features.tsv",
		"path to mooc_action_features.tsv",
	)
	flag.StringVar(
		&outputFilename,
		"output",
		"datasets/act-mooc/act_mooc.rdf",
		"rdf output path",
	)
}

func main() {
	flag.Parse()

	actionFacets := make(actionFacets)
	fillActionLabels(actionFacets, actionLabelsFilename)
	fillActionFeatures(actionFacets, actionFeaturesFilename)

	outputFile, err := os.Create(outputFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	handleActions(outputFile, actionFacets, actionsFilename)
}
