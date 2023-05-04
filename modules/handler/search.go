package handler

import (
	"RediDB/modules/memcache"
	"RediDB/modules/structure"
	"fmt"
	"reflect"

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

		if data.Filter["$max"] == nil {
			data.Filter["$max"] = 0.0
		}

		if reflect.TypeOf(data.Filter["$max"]).String() != "float64" && reflect.TypeOf(data.Filter["$max"]).String() != "int" {
			return ctx.JSON(fiber.Map{
				"success": false,
				"message": fmt.Sprintf(structure.MUST_BY, "$max", "integer"),
			})
		}

		if data.Filter["$order"] != nil {
			if reflect.TypeOf(data.Filter["$order"]).String() != "map[string]interface {}" {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": fmt.Sprintf(structure.MUST_BY, "$order", "\"type\" and \"by\""),
				})
			}

			orderType, orderTypeOk := data.Filter["$order"].(map[string]interface{})["type"]
			orderBy, orderByOk := data.Filter["$order"].(map[string]interface{})["by"]

			if !orderTypeOk {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": fmt.Sprintf(structure.REQUIRED_FIELD, "$order \"type\""),
				})
			}

			if orderType != "desc" && orderType != "asc" {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": fmt.Sprintf(structure.REQUIRED_INVALID, "$order \"type\"", "\"desc\" and \"asc\""),
				})
			}

			if !orderByOk {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": fmt.Sprintf(structure.REQUIRED_FIELD, "$order \"by\""),
				})
			}

			if reflect.TypeOf(orderBy).String() != "string" {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": fmt.Sprintf(structure.MUST_BY, "\"by\"", "string"),
				})
			}
		}

		if data.Filter["$or"] != nil {
			if reflect.TypeOf(data.Filter["$or"]).String() != "[]interface {}" {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": fmt.Sprintf(structure.MUST_BY, "$or", "array"),
				})
			}

			if len(data.Filter["$or"].([]interface{})) == 0 {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": structure.EMPTY_DATA,
				})
			}

			for i, or := range data.Filter["$or"].([]interface{}) {
				if or == nil || reflect.TypeOf(or).String() != "map[string]interface {}" {
					return ctx.JSON(fiber.Map{
						"success": false,
						"message": fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$or with index %d", i), "object"),
					})
				}
			}
		}

		max := data.Filter["$max"].(float64)
		if int(max) < 0 {
			return ctx.JSON(fiber.Map{
				"success": false,
				"message": fmt.Sprintf(structure.MUST_BY, "$max", ">= 0"),
			})
		}

		found := memcache.Get(data.Database, data.Collection, data.Filter, int(max))
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

	if request.Filter["$max"] == nil {
		request.Filter["$max"] = 0.0
	}

	if reflect.TypeOf(request.Filter["$max"]).String() != "float64" && reflect.TypeOf(request.Filter["$max"]).String() != "int" {
		ws.WriteJSON(structure.WebsocketAnswer{
			Error:   true,
			Message: fmt.Sprintf(structure.MUST_BY, "$max", "integer"),
		})

		return
	}

	if request.Filter["$order"] != nil {
		if reflect.TypeOf(request.Filter["$order"]).String() != "map[string]interface {}" {
			ws.WriteJSON(structure.WebsocketAnswer{
				Error:   true,
				Message: fmt.Sprintf(structure.MUST_BY, "$order", "\"type\" and \"by\""),
			})

			return
		}

		orderType, orderTypeOk := request.Filter["$order"].(map[string]interface{})["type"]
		orderBy, orderByOk := request.Filter["$order"].(map[string]interface{})["by"]

		if !orderTypeOk {
			ws.WriteJSON(structure.WebsocketAnswer{
				Error:   true,
				Message: fmt.Sprintf(structure.REQUIRED_FIELD, "$order \"type\""),
			})

			return
		}

		if orderType != "desc" && orderType != "asc" {
			ws.WriteJSON(structure.WebsocketAnswer{
				Error:   true,
				Message: fmt.Sprintf(structure.REQUIRED_INVALID, "$order \"type\"", "\"desc\" and \"asc\""),
			})

			return
		}

		if !orderByOk {
			ws.WriteJSON(structure.WebsocketAnswer{
				Error:   true,
				Message: fmt.Sprintf(structure.REQUIRED_FIELD, "$order \"by\""),
			})

			return
		}

		if reflect.TypeOf(orderBy).String() != "string" {
			ws.WriteJSON(structure.WebsocketAnswer{
				Error:   true,
				Message: fmt.Sprintf(structure.MUST_BY, "\"by\"", "string"),
			})

			return
		}
	}

	if request.Filter["$or"] != nil {
		if reflect.TypeOf(request.Filter["$or"]).String() != "[]interface {}" {
			ws.WriteJSON(structure.WebsocketAnswer{
				Error:   true,
				Message: fmt.Sprintf(structure.MUST_BY, "$or", "array"),
			})

			return
		}

		if len(request.Filter["$or"].([]interface{})) == 0 {
			ws.WriteJSON(structure.WebsocketAnswer{
				Error:   true,
				Message: structure.EMPTY_DATA,
			})

			return
		}

		for i, or := range request.Filter["$or"].([]interface{}) {
			if or == nil || reflect.TypeOf(or).String() != "map[string]interface {}" {
				ws.WriteJSON(structure.WebsocketAnswer{
					Error:   true,
					Message: fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$or with index %d", i), "object"),
				})

				return
			}
		}
	}

	max := request.Filter["$max"].(float64)
	if int(max) < 0 {
		ws.WriteJSON(structure.WebsocketAnswer{
			Error:   true,
			Message: fmt.Sprintf(structure.MUST_BY, "$max", ">= 0"),
		})

		return
	}

	found := memcache.Get(request.Database, request.Collection, request.Filter, int(max))
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
