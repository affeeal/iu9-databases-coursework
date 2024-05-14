package converter

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/affeeal/iu9-database-coursework/internal/converter/config"
	"github.com/affeeal/iu9-database-coursework/internal/converter/rdf"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type entitiesFacets map[string]map[string]*rdf.Term

const (
	configsName = "configs"
	srcName     = "src"

	configName = "convert.yml"
	outName    = "out.rdf"
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
			wg.Done()
			continue
		}

		go func(datasetPath string) {
			defer wg.Done()

			if err := ProcessDataset(datasetPath); err != nil {
				log.Println(errors.Wrap(err, datasetPath))
				return
			}

			fmt.Println("successfully processed " + datasetPath)
		}(makePath(datasetsPath, entry.Name()))
	}

	wg.Wait()

	return nil
}

func ProcessDataset(datasetPath string) error {
	configFile, err := os.Open(makePath(datasetPath, configsName, configName))
	if err != nil {
		return err
	}
	defer configFile.Close()

	decoder := yaml.NewDecoder(configFile)
	decoder.KnownFields(true)

	var ds config.Dataset
	if err = decoder.Decode(&ds); err != nil {
		return err
	}

	if err = processDataset(&ds, datasetPath); err != nil {
		return err
	}

	return nil
}

func makePath(names ...string) string {
	return strings.Join(names, "/")
}

func processDataset(ds *config.Dataset, datasetPath string) error {
	out, err := os.Create(makePath(datasetPath, outName))
	if err != nil {
		return err
	}
	defer out.Close()

	del, err := validateSymbol(ds.Delimiter, ',')
	if err != nil {
		return err
	}

	com, err := validateSymbol(ds.Comment, 0)
	if err != nil {
		return err
	}

	ef := make(entitiesFacets)
	srcPath := makePath(datasetPath, srcName)

	for _, file := range ds.Files {
		if err = processFile(&file, ef, out, srcPath, del, com); err != nil {
			return errors.Wrap(err, file.Name)
		}
	}

	return nil
}

func validateSymbol(rawSym string, defaultSym rune) (rune, error) {
	if rawSym == "" {
		return defaultSym, nil
	} else if len(rawSym) != 1 {
		return 0, errors.New("special symbol " + rawSym + " must be a single rune")
	}

	sym := rune(rawSym[0])
	if sym == '\r' || sym == '\n' {
		return 0, errors.New("special symbol " + rawSym + ` must not be \r, \n`)
	}

	return sym, nil
}
