package main

import (
	"flag"
	"log"

	"github.com/affeeal/iu9-databases-coursework/internal/converter"
)

var (
	datasetPath  string
	datasetsPath string
)

func init() {
	flag.StringVar(&datasetPath, "dataset-path", "", "Path to the root directory of dataset")
	flag.StringVar(&datasetsPath, "datasets-path", "", "Path to the directory with the root directories of datasets")
}

func main() {
	flag.Parse()

  var err error
  if datasetPath != "" {
    err = converter.ProcessDataset(datasetPath)
  } else if datasetsPath != "" {
    err = converter.ProcessDatasets(datasetsPath)
  } else {
    log.Fatal("Either -dataset-path or -datasets-path must be specified")
  }

  if err != nil {
    log.Fatal(err)
  }
}
