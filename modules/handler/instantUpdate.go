package handler

import (
	"RediDB/modules/memcache"
	"RediDB/modules/structure"
	"fmt"
	"os"
	"reflect"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func handleInstantUpdate() {
	App.Put("/", func(ctx *fiber.Ctx) error {
		var data struct {
			Database   string `json:"database"`
			Collection string `json:"collection"`

			Data struct {
				Filter map[string]interface{} `json:"filter"`
				Update map[string]interface{} `json:"update"`
			} `json:"data"`
		}

		if err := ctx.BodyParser(&data); err != nil {
			return ctx.JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		if data.Data.Update == nil || len(data.Data.Update) == 0 {
			return ctx.JSON(fiber.Map{
				"success": false,
				"message": "No data to update",
			})
		}

		if data.Data.Update["_id"] != nil {
			return ctx.JSON(fiber.Map{
				"success": false,
				"message": "ID property cannot be changed",
			})
		}

		if data.Data.Update["$max"] != nil {
			return ctx.JSON(fiber.Map{
				"success": false,
				"message": "$max property cannot be changed",
			})
		}

		if data.Data.Update["$order"] != nil {
			return ctx.JSON(fiber.Map{
				"success": false,
				"message": "$order property cannot be changed",
			})
		}

		if data.Data.Filter["$or"] != nil {
			if reflect.TypeOf(data.Data.Filter["$or"]).String() != "[]interface {}" {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": "$or option must be array",
				})
			}

			if len(data.Data.Filter["$or"].([]interface{})) == 0 {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": "$or option is empty",
				})
			}

			for i, or := range data.Data.Filter["$or"].([]interface{}) {
				if or == nil || reflect.TypeOf(or).String() != "map[string]interface {}" {
					return ctx.JSON(fiber.Map{
						"success": false,
						"message": fmt.Sprintf("$or option with index %d is not object", i),
					})
				}
			}
		}

		found := memcache.Get(data.Database, data.Collection, data.Data.Filter, 0)
		if found == nil {
			return ctx.JSON([]map[string]interface{}{})
		}

		var updated []map[string]interface{}

		memcache.Cache.Lock()
		for _, document := range found {
			updatedDocument := memcache.InstantUpdateDocument(document, data.Data.Update)
			encoded, err := json.Marshal(updatedDocument)
			if err != nil {
				updated = append(updated, map[string]interface{}{
					"_id":     document["_id"],
					"created": false,
					"reason":  err.Error(),
				})
				continue
			}

			err = os.WriteFile(fmt.Sprintf("./data/%s/%s/%s.db", data.Database, data.Collection, document["_id"]), encoded, os.ModePerm)
			if err != nil {
				updated = append(updated, map[string]interface{}{
					"_id":     document["_id"],
					"created": false,
					"reason":  err.Error(),
				})
				continue
			}

			memcache.CacheSet(data.Database, data.Collection, document["_id"].(string), updatedDocument)
			updated = append(updated, map[string]interface{}{
				"_id":     document["_id"],
				"updated": true,
			})
		}

		memcache.Cache.Unlock()
		return ctx.JSON(updated)
	})
}

func WSHandleInstantUpdate(ws *websocket.Conn, request structure.WebsocketRequest) {
	if request.Data.([]interface{})[0].(map[string]interface{}) == nil {
		ws.WriteJSON(structure.WebsocketAnswer{
			Error:   true,
			Message: "No data to update",
		})
		return
	}

	if request.Data.([]interface{})[0].(map[string]interface{})["_id"] != nil {
		ws.WriteJSON(structure.WebsocketAnswer{
			Error:   true,
			Message: "ID property cannot be changed",
		})

		return
	}

	if request.Data.([]interface{})[0].(map[string]interface{})["$max"] != nil {
		ws.WriteJSON(structure.WebsocketAnswer{
			Error:   true,
			Message: "$max property cannot be changed",
		})

		return
	}

	if request.Data.([]interface{})[0].(map[string]interface{})["$order"] != nil {
		ws.WriteJSON(structure.WebsocketAnswer{
			Error:   true,
			Message: "$order property cannot be changed",
		})

		return
	}

	if request.Filter["$or"] != nil {
		if reflect.TypeOf(request.Filter["$or"]).String() != "[]interface {}" {
			ws.WriteJSON(structure.WebsocketAnswer{
				Error:   true,
				Message: "$or option must be array",
			})

			return
		}

		if len(request.Filter["$or"].([]interface{})) == 0 {
			ws.WriteJSON(structure.WebsocketAnswer{
				Error:   true,
				Message: "$or option is empty",
			})

			return
		}

		for i, or := range request.Filter["$or"].([]interface{}) {
			if or == nil || reflect.TypeOf(or).String() != "map[string]interface {}" {
				ws.WriteJSON(structure.WebsocketAnswer{
					Error:   true,
					Message: fmt.Sprintf("$or option with index %d is not object", i),
				})

				return
			}
		}
	}

	found := memcache.Get(request.Database, request.Collection, request.Filter, 0)
	if found == nil {
		ws.WriteJSON(structure.WebsocketAnswer{
			Data: []interface{}{},
		})

		return
	}

	var updated []map[string]interface{}

	memcache.Cache.Lock()
	for _, document := range found {
		updatedDocument := memcache.InstantUpdateDocument(document, request.Data.([]interface{})[0].(map[string]interface{}))
		encoded, err := json.Marshal(updatedDocument)
		if err != nil {
			updated = append(updated, map[string]interface{}{
				"_id":     document["_id"],
				"created": false,
				"reason":  err.Error(),
			})
			continue
		}

		err = os.WriteFile(fmt.Sprintf("./data/%s/%s/%s.db", request.Database, request.Collection, document["_id"]), encoded, os.ModePerm)
		if err != nil {
			updated = append(updated, map[string]interface{}{
				"_id":     document["_id"],
				"created": false,
				"reason":  err.Error(),
			})
			continue
		}

		memcache.CacheSet(request.Database, request.Collection, document["_id"].(string), updatedDocument)
		updated = append(updated, map[string]interface{}{
			"_id":     document["_id"],
			"updated": true,
		})
	}

	memcache.Cache.Unlock()
	ws.WriteJSON(structure.WebsocketAnswer{
		Data: updated,
	})
}
