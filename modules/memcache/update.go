package memcache

import (
	"reflect"
)

func UpdateDocument(document map[string]interface{}, updateData map[string]interface{}) map[string]interface{} {
	for key, val := range updateData {
		if oldVal, ok := document[key]; ok {
			if newVal, ok := val.(map[string]interface{}); ok {
				if oldValMap, ok := oldVal.(map[string]interface{}); ok {
					document[key] = UpdateDocument(oldValMap, newVal)
				} else {
					continue
				}
			} else {
				if !reflect.DeepEqual(oldVal, val) {
					document[key] = val
				}
			}
		} else {
			document[key] = val
		}
	}

	return document
}

func InstantUpdateDocument(document map[string]interface{}, updateData map[string]interface{}) map[string]interface{} {
	updateData["_id"] = document["_id"]
	return updateData
}
