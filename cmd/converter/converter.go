package main

import (
	"flag"

	"github.com/affeeal/iu9-database-coursework/internal/converter"
)

var datasetsPath string

func init() {
	const (
		USAGE = "path to the datasets"
		VALUE = "datasets"
	)

	flag.StringVar(&datasetsPath, "datasets", VALUE, USAGE)
	flag.StringVar(&datasetsPath, "d", VALUE, USAGE+" (shorthand)")
}

func main() {
	flag.Parse()

	converter.ProcessDatasets(datasetsPath)
}
