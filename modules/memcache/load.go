package memcache

import (
	"RediDB/modules/config"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/goccy/go-json"
)

func processCollection(database, collection string, taskCounter *int32, wg *sync.WaitGroup) {
	defer wg.Done()

	files, err := os.ReadDir(fmt.Sprintf("./data/%s/%s/", database, collection))
	if err != nil {
		log.Printf("Failed to read database data of %s/%s", database, collection)
		return
	}

	sort.Slice(files, func(i, j int) bool {
		current, err := files[i].Info()
		if err != nil {
			return false
		}

		next, err := files[j].Info()
		if err != nil {
			return false
		}

		return current.Size() < next.Size()
	})

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".db") {
			data, err := os.ReadFile(fmt.Sprintf("./data/%s/%s/%s", database, collection, file.Name()))

			if err != nil {
				log.Printf("Failed to read document of %s/%s/%s: %s", database, collection, file.Name(), err.Error())
				continue
			}

			var document map[string]interface{}
			if err := json.Unmarshal(data, &document); err != nil {
				log.Printf("Failed to read document of %s/%s/%s: (FileCorrupt) %s", database, collection, file.Name(), err.Error())
				continue
			}

			if document["_id"] == nil {
				log.Printf("The document %s/%s/%s was skipped, because the document without id", database, collection, file.Name())
				continue
			}

			Cache.Lock()

			if CacheGet()[database][collection][document["_id"].(string)] != nil {
				log.Printf("The document %s/%s/%s was skipped, because the document with this id already exists", database, collection, file.Name())
			} else {
				CacheSet(database, collection, document["_id"].(string), document)
			}

			Cache.Unlock()
		}
	}

	atomic.AddInt32(taskCounter, -1)
}

func Load() {
	databases, err := os.ReadDir("./data/")
	if err != nil {
		log.Panicln("Failed to read database data: ", err)
	}

	var wg sync.WaitGroup
	var taskCounter int32

	pool := NewWorkerPool(len(databases), config.Get().Settings.TasksCount)
	for _, database := range databases {
		if database.IsDir() {
			collections, err := os.ReadDir(fmt.Sprintf("./data/%s/", database.Name()))
			if err != nil {
				log.Printf("Failed to read database data of %s", database.Name())
				continue
			}

			for _, collection := range collections {
				if collection.IsDir() {
					wg.Add(1)
					atomic.AddInt32(&taskCounter, 1)

					db := database.Name()
					col := collection.Name()

					pool.Submit(func() {
						processCollection(db, col, &taskCounter, &wg)
					})
				}
			}
		}
	}

	wg.Wait()
	pool.Shutdown()
}
