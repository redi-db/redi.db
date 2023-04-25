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
    max_threads: 30000 # Maximum number of branches that will be received from the processor (The higher - the more load will be sustained)
    max_data: 4 # Maximum amount of data in the query in mb
    check_updates: true
    websocket_support: true # It is desirable to use the ws protocol

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
