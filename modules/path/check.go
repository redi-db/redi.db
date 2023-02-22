package path

import (
	"log"
	"os"
)

func Create() {
	if !isExits() {
		if err := os.Mkdir("./data", os.ModePerm); err != nil {
			log.Panicln("Failed to create database folder: ", err)
		}
	}
}
