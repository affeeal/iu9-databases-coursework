package main

import (
	"flag"
	"log"
	"os"

	"github.com/affeeal/iu9-database-coursework/internal/transform"
	"gopkg.in/yaml.v3"
)

var configPath, datasetsPath string

func init() {
	const (
		configUsage   = "path to the YAML configuration file"
		configValue   = "configs/transform.yml"
		datasetsUsage = "path to the datasets"
		datasetsValue = "datasets"
	)

	flag.StringVar(&configPath, "config", configValue, configUsage)
	flag.StringVar(&configPath, "c", configValue, configUsage)

	flag.StringVar(&datasetsPath, "datasets", datasetsValue, datasetsUsage)
	flag.StringVar(&datasetsPath, "d", datasetsValue, datasetsUsage)
}

func main() {
	flag.Parse()

	configFile, err := os.Open(configPath)
	if err != nil {
		log.Fatal(err)
	}
	defer configFile.Close()

	decoder := yaml.NewDecoder(configFile)
	decoder.KnownFields(true)

	var config transform.Config
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal(err)
	}

	transform.Transform(datasetsPath, &config)
}
