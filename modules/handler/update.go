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

func handleUpdate() {
	App.Patch("/", func(ctx *fiber.Ctx) error {
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
				"message": structure.EMPTY_DATA,
			})
		}

		if data.Data.Update["_id"] != nil {
			return ctx.JSON(fiber.Map{
				"success": false,
				"message": fmt.Sprintf(structure.LOCK, "_id"),
			})
		}

		if data.Data.Filter != nil {
			delete(data.Data.Filter, "$order")
			delete(data.Data.Filter, "$only")
			delete(data.Data.Filter, "$omit")
			delete(data.Data.Filter, "$max")
		}

		filter, err := handleHttpFilter(data.Data.Filter)
		if len(err) > 0 {
			return ctx.JSON(err)
		}

		found := memcache.Get(data.Database, data.Collection, filter, 0)
		if found == nil {
			return ctx.JSON([]map[string]interface{}{})
		}

		var updated []map[string]interface{}
		memcache.Cache.Lock()
		for _, document := range found {
			updatedDocument := memcache.UpdateDocument(document, data.Data.Update)
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

func WSHandleUpdate(ws *websocket.Conn, request structure.WebsocketRequest) {
	if request.Data == nil || len(request.Data.([]interface{})) == 0 {
		ws.WriteJSON(structure.WebsocketAnswer{
			Error:     true,
			RequestID: request.RequestID,
			Message:   structure.EMPTY_DATA,
		})

		return
	}

	if request.Data.([]interface{})[0] == nil || reflect.TypeOf(request.Data.([]interface{})[0]).String() != "map[string]interface {}" {
		ws.WriteJSON(structure.WebsocketAnswer{
			Error:     true,
			RequestID: request.RequestID,
			Message:   structure.INVALID_STRUCTURE,
		})

		return
	}

	if request.Data.([]interface{})[0].(map[string]interface{})["_id"] != nil {
		ws.WriteJSON(structure.WebsocketAnswer{
			Error:     true,
			RequestID: request.RequestID,
			Message:   fmt.Sprintf(structure.LOCK, "_id"),
		})

		return
	}

	if request.Filter != nil {
		delete(request.Filter, "$order")
		delete(request.Filter, "$only")
		delete(request.Filter, "$omit")
		delete(request.Filter, "$max")
	}

	filter, err := handleWSFilter(request.Filter, request.RequestID)
	if err.Error {
		ws.WriteJSON(err)
		return
	}

	if request.Data.([]interface{})[0].(map[string]interface{})["$max"] != nil {
		ws.WriteJSON(structure.WebsocketAnswer{
			Error:     true,
			RequestID: request.RequestID,
			Message:   fmt.Sprintf(structure.LOCK, "$max"),
		})

		return
	}

	if request.Data.([]interface{})[0].(map[string]interface{})["$order"] != nil {
		ws.WriteJSON(structure.WebsocketAnswer{
			Error:     true,
			RequestID: request.RequestID,
			Message:   fmt.Sprintf(structure.LOCK, "$order"),
		})

		return
	}

	if request.Data.([]interface{})[0].(map[string]interface{})["$only"] != nil {
		ws.WriteJSON(structure.WebsocketAnswer{
			Error:     true,
			RequestID: request.RequestID,
			Message:   fmt.Sprintf(structure.LOCK, "$only"),
		})

		return
	}

	if request.Data.([]interface{})[0].(map[string]interface{})["$omit"] != nil {
		ws.WriteJSON(structure.WebsocketAnswer{
			Error:     true,
			RequestID: request.RequestID,
			Message:   fmt.Sprintf(structure.LOCK, "$omit"),
		})

		return
	}

	found := memcache.Get(request.Database, request.Collection, filter, 0)
	if found == nil {
		ws.WriteJSON(structure.WebsocketAnswer{
			RequestID: request.RequestID,
			Data:      []interface{}{},
		})

		return
	}

	var updated []map[string]interface{}
	memcache.Cache.Lock()
	for _, document := range found {
		updatedDocument := memcache.UpdateDocument(document, request.Data.([]interface{})[0].(map[string]interface{}))
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
		RequestID: request.RequestID,
		Data:      updated,
	})
}
