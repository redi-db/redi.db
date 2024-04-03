package memcache

import (
	"RediDB/modules/path"
	"log"
	"reflect"
	"regexp"
	"sort"

	"github.com/mitchellh/copystructure"
)

func Get(database string, collection string, filter map[string]interface{}, max int) []map[string]interface{} {
	var or []interface{}
	var and []interface{}

	if filter != nil && filter["$or"] != nil {
		or = filter["$or"].([]interface{})
		delete(filter, "$or")
	}

	if filter != nil && filter["$and"] != nil {
		and = filter["$and"].([]interface{})
		delete(filter, "$and")
	}

	result := GetDocuments(database, collection, filter, max)
	if len(and) > 0 {
		idMap := make(map[string]bool)
		for _, document := range result {
			idMap[document["_id"].(string)] = true
		}

		for _, andValue := range and {
			data := GetDocuments(database, collection, andValue.(map[string]interface{}), max)

			for _, document := range data {
				id := document["_id"].(string)

				if !idMap[id] {
					idMap[id] = true
					result = append(result, document)
				}
			}
		}
	}

	if len(result) == 0 && len(or) > 0 {
		for _, filterOr := range or {
			if filterOr != nil && filterOr.(map[string]interface{})["$or"] != nil {
				delete(filterOr.(map[string]interface{}), "$or")
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

	var _regex []interface{}
	var _sort map[string]interface{}

	var ne []interface{}

	var only []interface{}
	var omit []interface{}

	var greatThen []interface{}
	var lessThan []interface{}

	if filter != nil {
		if filter["$max"] != nil {
			delete(filter, "$max")
		}

		if filter["$regex"] != nil {
			_regex = filter["$regex"].([]interface{})
			delete(filter, "$regex")
		}

		if filter["$ne"] != nil {
			ne = filter["$ne"].([]interface{})
			delete(filter, "$ne")
		}

		if filter["$order"] != nil {
			_sort = filter["$order"].(map[string]interface{})
			delete(filter, "$order")
		}

		if filter["$only"] != nil {
			only = filter["$only"].([]interface{})
			delete(filter, "$only")
		}

		if filter["$omit"] != nil {
			omit = filter["$omit"].([]interface{})
			delete(filter, "$omit")
		}

		if filter["$gt"] != nil {
			greatThen = filter["$gt"].([]interface{})
			delete(filter, "$gt")
		}

		if filter["$lt"] != nil {
			lessThan = filter["$lt"].([]interface{})
			delete(filter, "$lt")
		}
	}

	Cache.Lock()
	defer Cache.Unlock()

	cache := CacheGet()
	if cache[database] == nil || cache[database][collection] == nil {
		return result
	}

	for _, document := range cache[database][collection] {
		if matchesFilter(document, filter) {
			result = append(result, document)
		}
	}

	if len(ne) > 0 {
		for i := len(result) - 1; i >= 0; i-- {
			document := result[i]
			allConditionsMatch := true
			for _, condition := range ne {
				conditionMap := condition.(map[string]interface{})
				field := conditionMap["by"].(string)
				value := conditionMap["value"]

				fieldValue, contains := getValue(document, field)
				if contains && fieldValue == value {
					allConditionsMatch = false
					break
				}
			}

			if !allConditionsMatch {
				result = RemoveIndex(result, i)
			}
		}
	}

	if len(_regex) > 0 {
		for i := len(result) - 1; i >= 0; i-- {
			document := result[i]
			allConditionsMatch := true

			founded := 0
			for _, filter := range _regex {
				filterMap := filter.(map[string]interface{})
				field := filterMap["by"].(string)
				regexValue := filterMap["value"].(string)

				fieldValue, contains := getValue(document, field)
				if !contains || reflect.TypeOf(fieldValue).String() != "string" {
					allConditionsMatch = false
					break
				}

				regex, err := regexp.Compile(regexValue)
				if err != nil {
					allConditionsMatch = false
					break
				}

				match := regex.MatchString(fieldValue.(string))
				if !match {
					allConditionsMatch = false
					break
				}

				founded++
			}

			if !allConditionsMatch || founded != len(_regex) {
				result = RemoveIndex(result, i)
			}
		}
	}

	if len(greatThen) > 0 {
		for i := len(result) - 1; i >= 0; i-- {
			document := result[i]
			allConditionsMatch := true
			for _, condition := range greatThen {
				conditionMap := condition.(map[string]interface{})
				value, contains := getValue(document, conditionMap["by"].(string))
				if !contains || reflect.TypeOf(value).String() != "float64" || value.(float64) <= conditionMap["value"].(float64) {
					allConditionsMatch = false
					break
				}
			}
			if !allConditionsMatch {
				result = RemoveIndex(result, i)
			}
		}
	}

	if len(lessThan) > 0 {
		for i := len(result) - 1; i >= 0; i-- {
			document := result[i]
			allConditionsMatch := true
			for _, condition := range lessThan {
				conditionMap := condition.(map[string]interface{})
				value, contains := getValue(document, conditionMap["by"].(string))
				if !contains || reflect.TypeOf(value).String() != "float64" || value.(float64) >= conditionMap["value"].(float64) {
					allConditionsMatch = false
					break
				}
			}
			if !allConditionsMatch {
				result = RemoveIndex(result, i)
			}
		}
	}

	if len(_sort) > 0 {
		result = sortData(result, _sort["type"].(string), _sort["by"])
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

	if max != 0 {
		if len(result) > max {
			return result[:max]
		}
	}

	return result
}

func matchesFilter(data map[string]interface{}, filter map[string]interface{}) bool {
	for key, value := range filter {
		dataValue, exists := data[key]
		if !exists || !reflect.DeepEqual(dataValue, value) {
			return false
		}
		if filterMap, ok := value.(map[string]interface{}); ok {
			if dataMap, ok := dataValue.(map[string]interface{}); ok {
				if !matchesFilter(dataMap, filterMap) {
					return false
				}
			} else {
				return false
			}
		}
	}
	return true
}

func sortData(data []map[string]interface{}, sortType string, sortBy interface{}) []map[string]interface{} {
	var sortData []map[string]interface{}
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
	if value, ok := data[key.(string)]; ok {
		return value, true
	}

	for k, v := range data {
		if nestedMap, ok := v.(map[string]interface{}); ok {
			if value, ok := getValue(nestedMap, key); ok {
				return value, true
			}
		}
		if nestedSlice, ok := v.([]interface{}); ok {
			for _, nestedItem := range nestedSlice {
				if nestedMap, ok := nestedItem.(map[string]interface{}); ok {
					if value, ok := getValue(nestedMap, key); ok {
						return value, true
					}
				}
			}
		}
		if k == key {
			return data[k], true
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

func RemoveIndex(s []map[string]interface{}, index int) []map[string]interface{} {
	if index < 0 || index >= len(s) {
		return s
	}
	return append(s[:index], s[index+1:]...)
}
