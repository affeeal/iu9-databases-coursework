package converter

import (
	"os"
	"path/filepath"

	"github.com/affeeal/iu9-databases-coursework/internal/rdf"
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

	output, err := os.Create(filepath.Join(datasetPath, outputName))
	if err != nil {
		return err
	}
	defer output.Close()

	sourcesPath := filepath.Join(datasetPath, sourcesName)
	entitiesFacets := make(map[string]entityFacets)

	for _, f := range ds.Files {
		if err = f.process(entitiesFacets, output, sourcesPath); err != nil {
			return errors.Wrap(err, f.Name)
		}
	}

	return nil
}
