package handler

import (
	"RediDB/modules/distributor"
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

			DistributorID string `json:"distributorID"`
			Data          struct {
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

		if data.Data.Update == nil || len(data.Data.Update) == 0 {
			return ctx.JSON(fiber.Map{
				"success": false,
				"message": structure.NOTHING,
			})
		}

		for _, lockkey := range LockedFilters {
			if data.Data.Update[lockkey] != nil {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": fmt.Sprintf(structure.LOCK, lockkey),
				})
			}
		}

		filter, err := handleHttpFilter(data.Data.Filter)
		if err != nil {
			return ctx.JSON(err)
		}

		found := memcache.Get(data.Database, data.Collection, filter, 0)
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

		if len(updated) > _config.Distribute.StartFrom {
			ctx.Status(fiber.StatusPartialContent)

			distributorID := distributor.Set(updated)
			return ctx.JSON(fiber.Map{
				"distribute":    true,
				"distributorID": distributorID,
			})
		}

		return ctx.JSON(updated)
	})
}

func WSHandleInstantUpdate(ws *websocket.Conn, request structure.WebsocketRequest) {
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

	if request.Data == nil || len(request.Data.([]interface{})) == 0 {
		ws.WriteJSON(structure.WebsocketAnswer{
			Error:     true,
			RequestID: request.RequestID,
			Message:   structure.EMPTY_DATA,
		})

		return
	}

	if request.Data.([]interface{})[0] == nil || request.Data.([]interface{})[0].(map[string]interface{}) == nil {
		ws.WriteJSON(structure.WebsocketAnswer{
			Error:     true,
			RequestID: request.RequestID,
			Message:   structure.NOTHING,
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

	for _, lockkey := range LockedFilters {
		if request.Data.([]interface{})[0].(map[string]interface{})[lockkey] != nil {
			ws.WriteJSON(structure.WebsocketAnswer{
				Error:     true,
				RequestID: request.RequestID,
				Message:   fmt.Sprintf(structure.LOCK, lockkey),
			})

			return
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

	if len(updated) > _config.Distribute.StartFrom {
		distributorID = distributor.Set(updated)
		ws.WriteJSON(structure.WebsocketAnswer{
			RequestID:     request.RequestID,
			DistributorID: distributorID,
		})

		return
	}

	ws.WriteJSON(structure.WebsocketAnswer{
		RequestID: request.RequestID,
		Data:      updated,
	})
}
