package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/affeeal/iu9-databases-coursework/internal/converter"
)

func run(arguments []string, output io.Writer) error {
	var datasetPath, datasetsPath string
	flags := flag.NewFlagSet("converter", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.StringVar(
		&datasetPath,
		"dataset-path",
		"",
		"Path to the root directory of dataset",
	)
	flags.StringVar(
		&datasetsPath,
		"datasets-path",
		"",
		"Path to the directory with the root directories of datasets",
	)
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", flags.Args())
	}
	if (datasetPath == "") == (datasetsPath == "") {
		return errors.New("exactly one of -dataset-path or -datasets-path must be specified")
	}
	if datasetPath != "" {
		return converter.ProcessDataset(datasetPath)
	}
	return converter.ProcessDatasets(datasetsPath)
}

func main() {
	if err := run(os.Args[1:], os.Stderr); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		log.Printf("converter: %v", err)
		os.Exit(1)
	}
}
