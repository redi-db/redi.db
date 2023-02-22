package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

func load() {
	yamlFile, err := os.ReadFile(configName)

	if err != nil {
		log.Panicln(err)
	}

	err = yaml.Unmarshal(yamlFile, &cache)
	if err != nil {
		log.Fatalln("Failed to read config as yml: ", err)
	}

	cache.Init()
}
