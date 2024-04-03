package handler

import (
	"RediDB/modules/distributor"
	"RediDB/modules/memcache"
	"RediDB/modules/structure"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func handleSearch() {
	App.Post("/search", func(ctx *fiber.Ctx) error {
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

		filter, err := handleHttpFilter(data.Filter)
		if err != nil {
			return ctx.JSON(err)
		}

		found := memcache.Get(data.Database, data.Collection, filter, filter["$max"].(int))
		if found == nil {
			return ctx.JSON([]interface{}{})
		}

		if len(found) > _config.Distribute.StartFrom {
			ctx.Status(fiber.StatusPartialContent)

			distributorID := distributor.Set(found)
			return ctx.JSON(fiber.Map{
				"distribute":    true,
				"distributorID": distributorID,
			})
		}

		return ctx.JSON(found)
	})
}

func WSHandleSearch(ws *websocket.Conn, request structure.WebsocketRequest) {
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

	filter, err := handleWSFilter(request.Filter, request.RequestID)
	if err.Error {
		ws.WriteJSON(err)
		return
	}

	found := memcache.Get(request.Database, request.Collection, filter, filter["$max"].(int))
	if found == nil {
		ws.WriteJSON(structure.WebsocketAnswer{
			RequestID: request.RequestID,
			Data:      []interface{}{},
		})

		return
	}

	if len(found) > _config.Distribute.StartFrom {
		distributorID = distributor.Set(found)
		ws.WriteJSON(structure.WebsocketAnswer{
			RequestID:     request.RequestID,
			DistributorID: distributorID,
		})

		return
	}

	ws.WriteJSON(structure.WebsocketAnswer{
		RequestID:     request.RequestID,
		DistributorID: distributorID,
		Data:          found,
	})
}
