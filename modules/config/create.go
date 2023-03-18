package config

import (
	"log"
	"os"
)

func create() {
	file, err := os.Create(configName)
	if err != nil {
		log.Fatalln("Failed to create config file: ", err)
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	file, err = os.OpenFile(configName, os.O_RDWR, 0644)
	if err != nil {
		log.Fatalln("Failed to opening config file: ", err)
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	_, err = file.WriteString(defaultConfig)
	if err != nil {
		log.Fatal("Failed to write default config: ", err)
	}

	err = file.Sync()
	if err != nil {
		log.Fatal("Config sync failed: ", err)
	}
}
