package memcache

import (
	"sync"
)

type cache struct {
	sync.RWMutex
	Data map[string]map[string]map[string]map[string]interface{}
}

var Cache = &cache{Data: make(map[string]map[string]map[string]map[string]interface{})}

func CacheGet() map[string]map[string]map[string]map[string]interface{} {
	return Cache.Data
}

func CacheSet(db string, collection string, documentId string, document map[string]interface{}) {
	if Cache.Data[db] == nil {
		Cache.Data[db] = make(map[string]map[string]map[string]interface{})
	}

	if Cache.Data[db][collection] == nil {
		Cache.Data[db][collection] = make(map[string]map[string]interface{})
	}

	Cache.Data[db][collection][documentId] = document
}

func CacheDelete(db string, collection string, documentId string) {
	if collection != "" && documentId == "" {
		delete(Cache.Data[db], collection)
	} else {
		delete(Cache.Data[db][collection], documentId)
	}
}
