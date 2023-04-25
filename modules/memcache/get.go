package memcache

import (
	"RediDB/modules/path"
	"reflect"
	"sort"
)

func Get(database string, collection string, filter map[string]interface{}, max int) []map[string]interface{} {
	path.Create()

	var result []map[string]interface{}
	var sort map[string]interface{}
	var or []interface{}

	if filter != nil && filter["$max"] != nil {
		delete(filter, "$max")
	}

	if max == 0 {
		max = -1
	} else {
		max--
	}

	if filter != nil && filter["$or"] != nil {
		or = filter["$or"].([]interface{})
		delete(filter, "$or")
	}

	if filter != nil && filter["$order"] != nil {
		sort = filter["$order"].(map[string]interface{})
		delete(filter, "$order")
	}

	Cache.Lock()
	defer Cache.Unlock()

	var cache = CacheGet()
	if cache[database] == nil {
		return result
	}

	if cache[database][collection] == nil {
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

	if len(or) > 0 && len(result) == 0 {
		for _, orFilter := range or {
			found := false
			for _, document := range cache[database][collection] {
				if max != -1 && len(result) > max {
					found = true
					break
				}

				if matchesFilter(document, orFilter.(map[string]interface{})) {
					found = true
					result = append(result, document)
				}
			}

			if found {
				break
			}
		}
	}

	if len(sort) > 0 {
		result = sortData(result, sort["type"].(string), sort["by"])
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

		if reflect.TypeOf(dataValue).String() == "map[string]interface {}" {
			if res, ok := getValue(dataValue.(map[string]interface{}), key); ok {
				return res, true
			}
		}

		if reflect.TypeOf(dataValue).String() == "[]interface {}" {
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
