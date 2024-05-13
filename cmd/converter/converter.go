package main

import (
	"flag"

	"github.com/affeeal/iu9-database-coursework/internal/converter"
)

var datasetsPath string

func init() {
	const (
		usage = "path to the datasets"
		value = "datasets"
	)

	flag.StringVar(&datasetsPath, "datasets", value, usage)
	flag.StringVar(&datasetsPath, "d", value, usage+" (shorthand)")
}

func main() {
	flag.Parse()

	converter.ProcessDatasets(datasetsPath)
}
