package main

import (
	"encoding/csv"
	"flag"
	"io"
	"log"
	"os"
	"strconv"
)

var (
	actionsFilename        string
	actionLabelsFilename   string
	actionFeaturesFilename string
	outputFilename         string
)

func init() {
	flag.StringVar(
		&actionsFilename,
		"mooc-actions",
		"",
		"path to mooc_actions.tsv",
	)
	flag.StringVar(
		&actionLabelsFilename,
		"mooc-action-labels",
		"",
		"path to mooc_action_labels.tsv",
	)
	flag.StringVar(
		&actionFeaturesFilename,
		"mooc-action-features",
		"",
		"path to mooc_action_features.tsv",
	)
	flag.StringVar(
		&outputFilename,
		"output",
		"mooc.rdf",
		"rdf output path",
	)
}

type Action struct {
	label    bool
	features [4]float64
}

type ActionMap = map[uint64]*Action

func fillActionFeatures(am *ActionMap, filename string) {
	const (
		ACTIONID = iota
		FEATURE0

		FEATURES = 4
	)

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	r := csv.NewReader(file)
	r.Comma = '\t'

	// Skip header record
	record, err := r.Read()
	if err != nil {
		log.Fatal(err)
	}

	for {
		record, err = r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		actionId, err := strconv.ParseUint(record[ACTIONID], 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		var features [4]float64
		for i := 0; i < FEATURES; i++ {
			if features[i], err = strconv.ParseFloat(record[FEATURE0+i], 64); err != nil {
				log.Fatal(err)
			}
		}

		if action, ok := (*am)[actionId]; ok {
			action.features = features
		} else {
			(*am)[actionId] = &Action{features: features}
		}
	}
}

func fillActionLabels(am *ActionMap, filename string) {
	const (
		ACTIONID = iota
		LABEL
	)

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	r := csv.NewReader(file)
	r.Comma = '\t'

	// Skip header record
	record, err := r.Read()
	if err != nil {
		log.Fatal(err)
	}

	for {
		record, err = r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		actionId, err := strconv.ParseUint(record[ACTIONID], 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		label, err := strconv.ParseBool(record[LABEL])
		if err != nil {
			log.Fatal(err)
		}

		if action, ok := (*am)[actionId]; ok {
			action.label = label
		} else {
			(*am)[actionId] = &Action{label: label}
		}
	}
}

func handleActions(output *os.File, am *ActionMap, filename string) {
	const (
		ACTIONID = iota
		USERID
		TARGETID
		TIMESTAMP
	)

	actions, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer actions.Close()

	r := csv.NewReader(actions)
	r.Comma = '\t'

	// TODO
}

func main() {
	flag.Parse()

	am := make(ActionMap)

	fillActionFeatures(&am, actionFeaturesFilename)
	fillActionLabels(&am, actionLabelsFilename)

	file, err := os.Create(outputFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	handleActions(file, &am, actionsFilename)
}
