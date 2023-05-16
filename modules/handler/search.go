package handler

import (
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

		filter, err := handleHttpFilter(data.Filter)
		if err != nil {
			return ctx.JSON(err)
		}

		found := memcache.Get(data.Database, data.Collection, filter, filter["$max"].(int))
		if found == nil {
			return ctx.JSON([]interface{}{})
		}

		return ctx.JSON(found)
	})
}

func WSHandleSearch(ws *websocket.Conn, request structure.WebsocketRequest) {
	if request.Filter == nil {
		request.Filter = make(map[string]interface{})
	}

	filter, err := handleWSFilter(request.Filter)
	if err.Error {
		ws.WriteJSON(err)
		return
	}

	found := memcache.Get(request.Database, request.Collection, filter, filter["$max"].(int))
	if found == nil {
		ws.WriteJSON(structure.WebsocketAnswer{
			Data: []interface{}{},
		})

		return
	}

	ws.WriteJSON(structure.WebsocketAnswer{
		Data: found,
	})
}
