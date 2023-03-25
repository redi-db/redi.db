package memcache

import (
	"RediDB/modules/path"
)

func Get(database string, collection string, filter map[string]interface{}) []map[string]interface{} {
	path.Create()

	var result []map[string]interface{}
	var or []interface{}

	if filter != nil && filter["$max"] != nil {
		delete(filter, "$max")
	}

	if filter != nil && filter["$or"] != nil {
		or = filter["$or"].([]interface{})
		delete(filter, "$or")
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

	for _, item := range cache[database][collection] {
		if matchesFilter(item, filter) {
			result = append(result, item)
		}
	}

	if len(or) > 0 && len(result) == 0 {
		for _, orFilter := range or {
			found := false
			for _, item := range cache[database][collection] {
				if matchesFilter(item, orFilter.(map[string]interface{})) {
					found = true
					result = append(result, item)
				}
			}

			if found {
				break
			}
		}
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
