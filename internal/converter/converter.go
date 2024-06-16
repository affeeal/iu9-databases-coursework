package converter

import (
	"log"
	"os"
	"sync"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func ProcessDatasets(datasetsPath string) error {
	entires, err := os.ReadDir(datasetsPath)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(len(entires))

	for _, entry := range entires {
		if !entry.IsDir() {
			log.Println("ignore file " + entry.Name())
			wg.Done()
			continue
		}

		go func(datasetName string) {
			defer wg.Done()

			if err := ProcessDataset(makePath(datasetsPath, datasetName)); err != nil {
				log.Println(errors.Wrap(err, datasetName))
				return
			}

			log.Println("successfully processed " + datasetName)
		}(entry.Name())
	}

	wg.Wait()

	return nil
}

func ProcessDataset(datasetPath string) error {
	const CONFIG_NAME = "convert.yml"

	config, err := os.Open(makePath(datasetPath, CONFIG_NAME))
	if err != nil {
		return err
	}
	defer config.Close()

	decoder := yaml.NewDecoder(config)
	decoder.KnownFields(true)

	var ds Dataset
	if err = decoder.Decode(&ds); err != nil {
		return err
	}

	if err = ds.process(datasetPath); err != nil {
		return err
	}

	return nil
}
