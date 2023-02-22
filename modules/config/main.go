package config

import (
	"RediDB/modules/structure"
	"log"
)

var cache structure.Config

const configName = "./config.yml"
const defaultConfig = `server:
    port: 5000

settings:
    max_threads: 30000

auth:
    login: root
    password: root
`

func init() {
	if !isExits() {
		log.Println("Config not found and will be generated")
		create()
	}
}
