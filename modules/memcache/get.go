package memcache

import (
	"RediDB/modules/path"
	"log"
	"reflect"
	"sort"

	"github.com/mitchellh/copystructure"
)

func Get(database string, collection string, filter map[string]interface{}, max int) []map[string]interface{} {
	var or []interface{}
	if filter != nil && filter["$or"] != nil {
		or = filter["$or"].([]interface{})
		delete(filter, "$or")
	}

	result := GetDocuments(database, collection, filter, max)
	if len(result) == 0 && len(or) > 0 {
		for _, filterOr := range or {
			if filterOr != nil && filterOr.(map[string]interface{})["$or"] != nil {
				delete(filterOr.((map[string]interface{})), "$or")
			}

			result = Get(database, collection, filterOr.(map[string]interface{}), max)
			if len(result) > 0 {
				break
			}
		}
	}

	return result
}

func GetDocuments(database string, collection string, filter map[string]interface{}, max int) []map[string]interface{} {
	path.Create()

	var result []map[string]interface{}
	var sort map[string]interface{}

	var only []interface{}
	var omit []interface{}

	if filter != nil && filter["$max"] != nil {
		delete(filter, "$max")
	}

	if max == 0 {
		max = -1
	} else {
		max--
	}

	if filter != nil && filter["$order"] != nil {
		sort = filter["$order"].(map[string]interface{})
		delete(filter, "$order")
	}

	if filter != nil && filter["$only"] != nil {
		only = filter["$only"].([]interface{})
		delete(filter, "$only")
	}

	if filter != nil && filter["$omit"] != nil {
		omit = filter["$omit"].([]interface{})
		delete(filter, "$omit")
	}

	Cache.Lock()
	defer Cache.Unlock()

	cache := CacheGet()
	if cache[database] == nil || cache[database][collection] == nil {
		return result
	}

	for _, document := range cache[database][collection] {
		if max != -1 && len(result) > max {
			break
		}

		if matchesFilter(document, filter) {
			result = append(result, document)
		}
	}

	if len(sort) > 0 {
		result = sortData(result, sort["type"].(string), sort["by"])
	}

	if len(only) > 0 || len(omit) > 0 {
		newResult := make([]map[string]interface{}, len(result))
		for i, document := range result {
			newDocument := make(map[string]interface{})
			for key, value := range document {
				newDocument[key] = value
			}
			newResult[i] = newDocument
		}

		if len(only) > 0 {
			var onlyStrings []string
			for _, v := range only {
				if s, ok := v.(string); ok {
					onlyStrings = append(onlyStrings, s)
				}
			}

			if !contains(onlyStrings, "_id") {
				onlyStrings = append(onlyStrings, "_id")
			}

			for documentIndex, document := range newResult {
				removeDocumentValues(&document, onlyStrings)
				newResult[documentIndex] = document
			}
		}

		for _, omitValue := range omit {
			if omitValue.(string) == "_id" {
				continue
			}

			for documentIndex, document := range newResult {
				_, grabbed := getValue(document, omitValue)
				if grabbed {
					newDocument, err := copystructure.Copy(document)
					if err != nil {
						log.Printf("Omit handler error for {%s}/{%s}/{%s}: {%s}; Document was skipped", database, collection, document["_id"], err.Error())
						continue
					}

					newResult[documentIndex] = removeDocumentValue(newDocument.(map[string]interface{}), omitValue.(string))
				}
			}
		}

		return newResult
	}

	return result
}

func matchesFilter(data map[string]interface{}, filter map[string]interface{}) bool {
	for key, value := range filter {
		if dataValue, ok := data[key]; ok {
			if filterMap, ok := value.(map[string]interface{}); ok {
				if dataMap, ok := dataValue.(map[string]interface{}); ok {
					if !matchesFilter(dataMap, filterMap) {
						return false
					}
				} else {
					return false
				}
			} else {
				if dataValue != value {
					return false
				}
			}
		} else {
			return false
		}
	}
	return true
}

func sortData(data []map[string]interface{}, sortType string, sortBy interface{}) []map[string]interface{} {
	sortData := []map[string]interface{}{}
	for _, document := range data {
		if _, ok := getValue(document, sortBy); ok {
			sortData = append(sortData, document)
		}
	}

	sort.Slice(sortData, func(i, j int) bool {
		val, ok := getValue(sortData[i], sortBy)
		val2, ok2 := getValue(sortData[j], sortBy)

		if ok && ok2 {
			if reflect.TypeOf(val).String() == "float64" && reflect.TypeOf(val2).String() == "float64" {
				if sortType == "asc" {
					return int(val.(float64)) < int(val2.(float64))
				} else {
					return int(val.(float64)) > int(val2.(float64))
				}
			}
		}

		return true
	})

	return sortData
}

func getValue(data map[string]interface{}, key interface{}) (interface{}, bool) {
	for dataKey, dataValue := range data {
		if dataKey == key {
			return dataValue, true
		}

		if dataValue != nil && reflect.TypeOf(dataValue).String() == "map[string]interface {}" {
			if res, ok := getValue(dataValue.(map[string]interface{}), key); ok {
				return res, true
			}
		}

		if dataValue != nil && reflect.TypeOf(dataValue).String() == "[]interface {}" {
			for _, elem := range dataValue.([]interface{}) {
				if reflect.TypeOf(elem).String() == "map[string]interface {}" {
					if res, ok := getValue(elem.(map[string]interface{}), key); ok {
						return res, true
					}
				}
			}
		}
	}

	return nil, false
}

func removeDocumentValues(document *map[string]interface{}, only []string) {
	if len(only) == 0 {
		return
	}

	copyDocument := *document
	for key := range copyDocument {
		if !contains(only, key) {
			delete(copyDocument, key)
		} else if child, ok := copyDocument[key].(map[string]interface{}); ok {
			if contains(only, key) {
				continue
			}
			removeDocumentValues(&child, only)
		}
	}

	*document = copyDocument
}

func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func removeDocumentValue(document map[string]interface{}, key string) map[string]interface{} {
	for k, v := range document {
		if nestedDoc, ok := v.(map[string]interface{}); ok {
			nestedDoc = removeDocumentValue(nestedDoc, key)
			document[k] = nestedDoc
		}
	}

	delete(document, key)
	return document
}
