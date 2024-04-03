package handler

import (
	"RediDB/modules/distributor"
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

			DistributorID string                 `json:"distributorID"`
			Filter        map[string]interface{} `json:"filter"`
		}

		if err := ctx.BodyParser(&data); err != nil {
			return ctx.JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		if len(data.DistributorID) > 0 {
			documents, size, err := distributor.GetData(data.DistributorID)
			if err != nil {
				ctx.Status(fiber.StatusNotFound)
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": err.Error(),
				})
			}

			return ctx.JSON(fiber.Map{
				"residue": size,
				"data":    documents,
			})
		}

		if data.Filter == nil {
			data.Filter = make(map[string]interface{})
		}

		for _, ignoredKey := range IgnoreFilters {
			if data.Filter[ignoredKey] != nil {
				delete(data.Filter, ignoredKey)
			}
		}

		filter, err := handleHttpFilter(data.Filter)
		if len(err) > 0 {
			return ctx.JSON(err)
		}

		found := memcache.Get(data.Database, data.Collection, filter, 0)
		if found == nil {
			return ctx.JSON([]interface{}{})
		}

		var deleted []map[string]interface{}
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
			deleted = make([]map[string]interface{}, 0)
		}

		if len(deleted) > _config.Distribute.StartFrom {
			ctx.Status(fiber.StatusPartialContent)

			distributorID := distributor.Set(deleted)
			return ctx.JSON(fiber.Map{
				"distribute":    true,
				"distributorID": distributorID,
			})
		}

		return ctx.JSON(deleted)
	})
}

func WSHandleDelete(ws *websocket.Conn, request structure.WebsocketRequest) {
	distributorID := request.DistributorID
	if len(distributorID) > 0 {
		documents, size, err := distributor.GetData(distributorID)
		if err != nil {
			ws.WriteJSON(structure.WebsocketAnswer{
				Error:     true,
				RequestID: request.RequestID,
				Message:   err.Error(),
			})

			return
		}

		ws.WriteJSON(structure.WebsocketAnswer{
			RequestID: request.RequestID,

			Residue: size,
			Data:    documents,
		})

		return
	}

	if request.Filter == nil {
		request.Filter = make(map[string]interface{})
	}

	for _, ignoredKey := range IgnoreFilters {
		if request.Filter[ignoredKey] != nil {
			delete(request.Filter, ignoredKey)
		}
	}

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

	var deleted []map[string]interface{}
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
		deleted = make([]map[string]interface{}, 0)
	}

	if len(deleted) > _config.Distribute.StartFrom {
		distributorID = distributor.Set(deleted)
		ws.WriteJSON(structure.WebsocketAnswer{
			RequestID:     request.RequestID,
			DistributorID: distributorID,
		})

		return
	}

	ws.WriteJSON(structure.WebsocketAnswer{
		RequestID: request.RequestID,
		Data:      deleted,
	})
}
