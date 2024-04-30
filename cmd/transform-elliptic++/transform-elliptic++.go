package main

import (
	"encoding/csv"
	"flag"
	"io"
	"log"
	"os"

	"github.com/affeeal/iu9-database-coursework/internal/rdf"
)

const (
	datasetName        = "elliptic++"
	datasetDefaultPath = "datasets/" + datasetName + "/"
	sourceDefaultPath  = datasetDefaultPath + "source/"
	outputDefaultPath  = datasetDefaultPath + "transformed/"

	blankTxIdPrefix = "t"
)

var (
	classesFilename  string
	edgelistFilename string
	featuresFilename string
	outputFilename   string
)

func handleClasses(outputFile *os.File, classesFilename string) {
	const (
		txIdIdx = iota
		classIdx
	)

	classesFile, err := os.Open(classesFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer classesFile.Close()

	reader := csv.NewReader(classesFile)
	headers, err := reader.Read()
	if err != nil {
		log.Fatal(err)
	}

	var (
		record []string
		class  *rdf.Triple
	)

	for {
		record, err = reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		class = rdf.NewTriple(
			rdf.NewTerm(
				rdf.BlankNode(blankTxIdPrefix+record[txIdIdx]),
				rdf.None,
			),
			rdf.NewTerm(headers[classIdx], rdf.AngleBrackets),
			rdf.NewTerm(record[classIdx], rdf.Quotes),
			nil,
		)

		if _, err = outputFile.WriteString(class.Stringln()); err != nil {
			log.Fatal(err)
		}
	}
}

func handleEdgelist(outputFile *os.File, edgelistFilename string) {
	const (
		txId1Idx = iota
		txId2Idx
	)

	edgelistFile, err := os.Open(edgelistFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer edgelistFile.Close()

	reader := csv.NewReader(edgelistFile)
	record, err := reader.Read()
	if err != nil {
		log.Fatal(err)
	}

	var successors *rdf.Triple

	for {
		record, err = reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		successors = rdf.NewTriple(
			rdf.NewTerm(
				rdf.BlankNode(blankTxIdPrefix+record[txId1Idx]),
				rdf.None,
			),
			rdf.NewTerm("successors", rdf.AngleBrackets),
			rdf.NewTerm(
				rdf.BlankNode(blankTxIdPrefix+record[txId2Idx]),
				rdf.None,
			),
			nil,
		)

		if _, err = outputFile.WriteString(successors.Stringln()); err != nil {
			log.Fatal(err)
		}
	}
}

func handleFeatures(outputFile *os.File, featuresFilename string) {
	const (
		txIdIdx = iota
		timeStepIdx
		outBtcTotalIdx = 183
	)

	featuresFile, err := os.Open(featuresFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer featuresFile.Close()

	reader := csv.NewReader(featuresFile)
	headers, err := reader.Read()
	if err != nil {
		log.Fatal(err)
	}

	// A predicat name cannot contain whitespaces.
	headers[timeStepIdx] = "Time_step"

	var (
		record   []string
		blankTx  *rdf.Term
		predicat *rdf.Triple
	)

	for {
		record, err = reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		blankTx = rdf.NewTerm(
			rdf.BlankNode(blankTxIdPrefix+record[txIdIdx]),
			rdf.None,
		)

		for i := timeStepIdx; i <= outBtcTotalIdx; i++ {
			if len(record[i]) == 0 {
				continue
			}

			predicat = rdf.NewTriple(
				blankTx, // NOTE: shared ownership
				rdf.NewTerm(headers[i], rdf.AngleBrackets),
				rdf.NewTerm(record[i], rdf.Quotes),
				nil,
			)

			if _, err = outputFile.WriteString(predicat.Stringln()); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func init() {
	flag.StringVar(
		&classesFilename,
		"txs-classes",
		sourceDefaultPath+"txs_classes.csv",
		"path to txc_classes.csv",
	)
	flag.StringVar(
		&edgelistFilename,
		"txs-edgelist",
		sourceDefaultPath+"txs_edgelist.csv",
		"path to txs_edgelist.csv",
	)
	flag.StringVar(
		&featuresFilename,
		"txs-features",
		sourceDefaultPath+"txs_features.csv",
		"path to txs_features.tsv",
	)
	flag.StringVar(
		&outputFilename,
		"output",
		outputDefaultPath+datasetName+".rdf",
		"rdf output path",
	)
}

func main() {
	flag.Parse()

	outputFile, err := os.Create(outputFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	handleClasses(outputFile, classesFilename)
	handleEdgelist(outputFile, edgelistFilename)
	handleFeatures(outputFile, featuresFilename)
}
