package converter

import (
	"fmt"
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

	output, err := os.CreateTemp(datasetPath, ".output-*.rdf")
	if err != nil {
		return err
	}
	temporaryName := output.Name()
	closed := false
	committed := false
	defer func() {
		if !closed {
			_ = output.Close()
		}
		if !committed {
			_ = os.Remove(temporaryName)
		}
	}()

	sourcesPath := filepath.Join(datasetPath, sourcesName)
	entitiesFacets := make(map[string]entityFacets)

	for _, f := range ds.Files {
		if err = f.process(entitiesFacets, output, sourcesPath); err != nil {
			return errors.Wrap(err, f.Name)
		}
	}

	if err := output.Sync(); err != nil {
		return fmt.Errorf("sync temporary RDF output: %w", err)
	}
	if err := output.Close(); err != nil {
		return fmt.Errorf("close temporary RDF output: %w", err)
	}
	closed = true
	if err := os.Rename(temporaryName, filepath.Join(datasetPath, outputName)); err != nil {
		return fmt.Errorf("replace RDF output: %w", err)
	}
	committed = true
	return nil
}
