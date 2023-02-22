package memcache

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/goccy/go-json"
)

func Load() {
	databases, err := os.ReadDir("./data/")
	if err != nil {
		log.Panicln("Failed to read database data: ", err)
	}

	for _, database := range databases {
		if database.IsDir() {
			collections, err := os.ReadDir(fmt.Sprintf("./data/%s/", database.Name()))
			if err != nil {
				log.Printf("Failed to read database data of %s", database.Name())
				continue
			}

			for _, collection := range collections {
				if collection.IsDir() {
					files, err := os.ReadDir(fmt.Sprintf("./data/%s/%s/", database.Name(), collection.Name()))
					if err != nil {
						log.Printf("Failed to read database data of %s/%s", database.Name(), collection.Name())
						continue
					}

					for _, file := range files {
						if !file.IsDir() && strings.HasSuffix(file.Name(), ".db") {
							data, err := os.ReadFile(fmt.Sprintf("./data/%s/%s/%s", database.Name(), collection.Name(), file.Name()))

							if err != nil {
								log.Printf("Failed to read document of %s/%s/%s: %s", database.Name(), collection.Name(), file.Name(), err.Error())
								continue
							}

							var document map[string]interface{}
							if err := json.Unmarshal(data, &document); err != nil {
								log.Printf("Failed to read document of %s/%s/%s: (FileCorrupt) %s", database.Name(), collection.Name(), file.Name(), err.Error())
								continue
							}

							if document["_id"] == nil {
								log.Printf("The document %s/%s/%s was skipped, because the document without id", database.Name(), collection.Name(), file.Name())
								continue
							}

							if CacheGet()[database.Name()][collection.Name()][document["_id"].(string)] != nil {
								log.Printf("The document %s/%s/%s was skipped, because the document with this id already exists", database.Name(), collection.Name(), file.Name())
								continue
							}

							CacheSet(database.Name(), collection.Name(), document["_id"].(string), document)
						}
					}
				}
			}
		}
	}
}
