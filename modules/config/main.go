package config

import (
	_ "embed"

	"RediDB/modules/structure"
	"log"
)

var cache structure.Config

const configName = "./config.yml"

//go:embed ..\..\config.yml
var defaultConfig string

func init() {
	if !isExits() {
		log.Println("Config not found and will be generated")
		create()
	}
}
