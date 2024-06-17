package converter

import (
	"os"
	"strings"

	"github.com/affeeal/iu9-database-coursework/internal/rdf"
	"github.com/pkg/errors"
)

type dataset struct {
	Files []file `yaml:"files"`
}

type entityFacets map[string]*rdf.Term

func (ds *dataset) process(datasetPath string) error {
	const (
		outputName  = "output.rdf"
		sourcesName = "sources"
	)

	output, err := os.Create(makePath(datasetPath, outputName))
	if err != nil {
		return err
	}
	defer output.Close()

	sourcesPath := makePath(datasetPath, sourcesName)
	entitiesFacets := make(map[string]entityFacets)

	for _, file := range ds.Files {
		if err = file.process(entitiesFacets, output, sourcesPath); err != nil {
			return errors.Wrap(err, file.Name)
		}
	}

	return nil
}

func makePath(names ...string) string {
	return strings.Join(names, "/")
}
