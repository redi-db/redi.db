package handler

import (
	"RediDB/modules/memcache"
	"RediDB/modules/structure"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func handleDelete() {
	App.Delete("/", func(ctx *fiber.Ctx) error {
		var data struct {
			Database   string `json:"database"`
			Collection string `json:"collection"`

			Filter map[string]interface{} `json:"filter"`
		}

		if err := ctx.BodyParser(&data); err != nil {
			return ctx.JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		if data.Filter == nil {
			data.Filter = make(map[string]interface{})
		}

		delete(data.Filter, "$order")
		delete(data.Filter, "$only")
		delete(data.Filter, "$omit")
		delete(data.Filter, "$max")

		filter, err := handleHttpFilter(data.Filter)
		if len(err) > 0 {
			return ctx.JSON(err)
		}

		found := memcache.Get(data.Database, data.Collection, filter, 0)
		if found == nil {
			return ctx.JSON([]interface{}{})
		}

		var deleted []interface{}
		memcache.Cache.Lock()
		if len(data.Filter) == 0 {
			err := os.RemoveAll(fmt.Sprintf("./data/%s/%s", data.Database, data.Collection))
			memcache.CacheDelete(data.Database, data.Collection, "")

			if err != nil {
				for _, document := range found {
					deleted = append(deleted, map[string]interface{}{
						"_id":     document["_id"],
						"deleted": false,
						"reason":  err.Error(),
					})
				}
			} else {
				for _, document := range found {
					deleted = append(deleted, map[string]interface{}{
						"_id":     document["_id"],
						"deleted": true,
					})
				}
			}

		} else {
			for _, document := range found {
				err := os.Remove(fmt.Sprintf("./data/%s/%s/%s.db", data.Database, data.Collection, document["_id"]))
				memcache.CacheDelete(data.Database, data.Collection, document["_id"].(string))

				if err != nil {
					deleted = append(deleted, map[string]interface{}{
						"_id":     document["_id"],
						"deleted": false,
						"reason":  err.Error(),
					})
				} else {
					deleted = append(deleted, map[string]interface{}{
						"_id":     document["_id"],
						"deleted": true,
					})
				}
			}

			if len(memcache.CacheGet()[data.Database][data.Collection]) == 0 {
				delete(memcache.CacheGet()[data.Database], data.Collection)
				if err := os.Remove(fmt.Sprintf("./data/%s/%s", data.Database, data.Collection)); err != nil {
					log.Printf("Failed to delete %s/%s collection: %s", data.Database, data.Collection, err.Error())
				}
			}
		}

		if len(memcache.CacheGet()[data.Database]) == 0 {
			delete(memcache.CacheGet(), data.Database)
			if err := os.Remove(fmt.Sprintf("./data/%s/", data.Database)); err != nil {
				log.Printf("Failed to delete %s database: %s", data.Database, err.Error())
			}
		}

		memcache.Cache.Unlock()

		if deleted == nil {
			deleted = make([]interface{}, 0)
		}

		return ctx.JSON(deleted)
	})
}

func WSHandleDelete(ws *websocket.Conn, request structure.WebsocketRequest) {
	if request.Filter == nil {
		request.Filter = make(map[string]interface{})
	}

	delete(request.Filter, "$order")
	delete(request.Filter, "$only")
	delete(request.Filter, "$omit")
	delete(request.Filter, "$max")

	filter, err := handleWSFilter(request.Filter, request.RequestID)
	if err.Error {
		ws.WriteJSON(err)
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

	var deleted []interface{}
	memcache.Cache.Lock()
	if len(request.Filter) == 0 {
		err := os.RemoveAll(fmt.Sprintf("./data/%s/%s", request.Database, request.Collection))
		memcache.CacheDelete(request.Database, request.Collection, "")

		if err != nil {
			for _, document := range found {
				deleted = append(deleted, map[string]interface{}{
					"_id":     document["_id"],
					"deleted": false,
					"reason":  err.Error(),
				})
			}
		} else {
			for _, document := range found {
				deleted = append(deleted, map[string]interface{}{
					"_id":     document["_id"],
					"deleted": true,
				})
			}
		}

	} else {
		for _, document := range found {
			err := os.Remove(fmt.Sprintf("./data/%s/%s/%s.db", request.Database, request.Collection, document["_id"]))
			memcache.CacheDelete(request.Database, request.Collection, document["_id"].(string))

			if err != nil {
				deleted = append(deleted, map[string]interface{}{
					"_id":     document["_id"],
					"deleted": false,
					"reason":  err.Error(),
				})
			} else {
				deleted = append(deleted, map[string]interface{}{
					"_id":     document["_id"],
					"deleted": true,
				})
			}
		}

		if len(memcache.CacheGet()[request.Database][request.Collection]) == 0 {
			delete(memcache.CacheGet()[request.Database], request.Collection)
			if err := os.Remove(fmt.Sprintf("./data/%s/%s", request.Database, request.Collection)); err != nil {
				log.Printf("Failed to delete %s/%s collection: %s", request.Database, request.Collection, err.Error())
			}
		}
	}

	if len(memcache.CacheGet()[request.Database]) == 0 {
		delete(memcache.CacheGet(), request.Database)
		if err := os.Remove(fmt.Sprintf("./data/%s/", request.Database)); err != nil {
			log.Printf("Failed to delete %s database: %s", request.Database, err.Error())
		}
	}

	memcache.Cache.Unlock()

	if deleted == nil {
		deleted = make([]interface{}, 0)
	}

	ws.WriteJSON(structure.WebsocketAnswer{
		RequestID: request.RequestID,
		Data:      deleted,
	})
}
