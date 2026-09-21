package converter

import (
	"bufio"
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
	if len(ds.Files) == 0 {
		return fmt.Errorf("dataset must contain at least one source file")
	}

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
	writer := bufio.NewWriter(output)

	for _, f := range ds.Files {
		if err = f.process(entitiesFacets, writer, sourcesPath); err != nil {
			return errors.Wrap(err, f.Name)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("flush temporary RDF output: %w", err)
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
