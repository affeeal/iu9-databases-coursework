package converter

import (
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

func ProcessDatasets(datasetsPath string) error {
	entires, err := os.ReadDir(datasetsPath)
	if err != nil {
		return err
	}

	g := new(errgroup.Group)
	for _, entry := range entires {
		if !entry.IsDir() {
			continue
		}

		datasetName := entry.Name()
		g.Go(func() error {
			err := errors.Wrap(
				ProcessDataset(filepath.Join(datasetsPath, datasetName)),
				datasetName,
			)
			if err != nil {
				log.Println("Error while processing dataset", err)
			} else {
				log.Println("Successfully processed dataset", datasetName)
			}
			return err
		})
	}

	return g.Wait()
}

func ProcessDataset(datasetPath string) error {
	const configName = "convert.yml"

	config, err := os.Open(filepath.Join(datasetPath, configName))
	if err != nil {
		return err
	}
	defer config.Close()

	decoder := yaml.NewDecoder(config)
	decoder.KnownFields(true)

	var ds dataset
	if err = decoder.Decode(&ds); err != nil {
		return err
	}

	return ds.process(datasetPath)
}
