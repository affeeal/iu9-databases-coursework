package converter

import (
	"os"
	"strings"

	"github.com/affeeal/iu9-database-coursework/internal/rdf"
	"github.com/pkg/errors"
)

type Dataset struct {
	Files []File `yaml:"files"`
}

type entityFacets map[string]*rdf.Term

func (ds *Dataset) process(datasetPath string) error {
	const (
		OUTPUT_NAME  = "output.rdf"
		SOURCES_NAME = "sources"
	)

	output, err := os.Create(makePath(datasetPath, OUTPUT_NAME))
	if err != nil {
		return err
	}
	defer output.Close()

	sourcesPath := makePath(datasetPath, SOURCES_NAME)
	entitiesFacets := make(map[string]entityFacets)

	for _, file := range ds.Files {
		if err = file.process(entitiesFacets, output, sourcesPath); err != nil {
			return errors.Wrap(err, "File "+file.Name)
		}
	}

	return nil
}

func makePath(names ...string) string {
	return strings.Join(names, "/")
}
