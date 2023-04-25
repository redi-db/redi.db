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
				"message": "$max option must be integer",
			})
		}

		if data.Filter["$order"] != nil {
			if reflect.TypeOf(data.Filter["$order"]).String() != "map[string]interface {}" {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": "$order option must be object with \"type\" and \"by\"",
				})
			}

			orderType, orderTypeOk := data.Filter["$order"].(map[string]interface{})["type"]
			orderBy, orderByOk := data.Filter["$order"].(map[string]interface{})["by"]

			if !orderTypeOk {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": "$order parameter \"type\" is required",
				})
			}

			if orderType != "desc" && orderType != "asc" {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": "$order parameter \"type\" must be \"desc\" and \"asc\" only",
				})
			}

			if !orderByOk {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": "$order parameter \"by\" is required",
				})
			}

			if reflect.TypeOf(orderBy).String() != "string" {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": "$order parameter \"by\" must be string",
				})
			}
		}

		if data.Filter["$ew"] != nil {
			if reflect.TypeOf(data.Filter["$ew"]).String() != "map[string]interface {}" {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": "$ew option must be object",
				})
			}

			for i, or := range data.Filter["$ew"].(map[string]any) {
				if reflect.TypeOf(or).String() != "string" {
					return ctx.JSON(fiber.Map{
						"success": false,
						"message": fmt.Sprintf("$ew parametr with index \"%s\" is not string", i),
					})
				}
			}
		}

		if data.Filter["$or"] != nil {
			if reflect.TypeOf(data.Filter["$or"]).String() != "[]interface {}" {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": "$or option must be array",
				})
			}

			if len(data.Filter["$or"].([]interface{})) == 0 {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": "$or option is empty",
				})
			}

			for i, or := range data.Filter["$or"].([]interface{}) {
				if or == nil || reflect.TypeOf(or).String() != "map[string]interface {}" {
					return ctx.JSON(fiber.Map{
						"success": false,
						"message": fmt.Sprintf("$or option with index %d is not object", i),
					})
				}
			}
		}

		max := data.Filter["$max"].(float64)
		if int(max) < 0 {
			return ctx.JSON(fiber.Map{
				"success": false,
				"message": "$max option must be >= 0",
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
			Message: "$max option must be integer",
		})

		return
	}

	if request.Filter["$order"] != nil {
		if reflect.TypeOf(request.Filter["$order"]).String() != "map[string]interface {}" {
			ws.WriteJSON(structure.WebsocketAnswer{
				Error:   true,
				Message: "$order option must be object with \"type\" and \"by\"",
			})

			return
		}

		orderType, orderTypeOk := request.Filter["$order"].(map[string]interface{})["type"]
		orderBy, orderByOk := request.Filter["$order"].(map[string]interface{})["by"]

		if !orderTypeOk {
			ws.WriteJSON(structure.WebsocketAnswer{
				Error:   true,
				Message: "$order parameter \"type\" is required",
			})

			return
		}

		if orderType != "desc" && orderType != "asc" {
			ws.WriteJSON(structure.WebsocketAnswer{
				Error:   true,
				Message: "$order parameter \"type\" must be \"desc\" and \"asc\" only",
			})

			return
		}

		if !orderByOk {
			ws.WriteJSON(structure.WebsocketAnswer{
				Error:   true,
				Message: "$order parameter \"by\" is required",
			})

			return
		}

		if reflect.TypeOf(orderBy).String() != "string" {
			ws.WriteJSON(structure.WebsocketAnswer{
				Error:   true,
				Message: "$order parameter \"by\" must be string",
			})

			return
		}
	}

	if request.Filter["$ew"] != nil {
		if reflect.TypeOf(request.Filter["$ew"]).String() != "map[string]interface {}" {
			ws.WriteJSON(structure.WebsocketAnswer{
				Error:   true,
				Message: "$ew option must be object",
			})

			return
		}

		for i, or := range request.Filter["$ew"].(map[string]any) {
			if reflect.TypeOf(or).String() != "string" {
				ws.WriteJSON(structure.WebsocketAnswer{
					Error:   true,
					Message: fmt.Sprintf("$ew parametr with index \"%s\" is not string", i),
				})

				return
			}
		}
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

	max := request.Filter["$max"].(float64)
	if int(max) < 0 {
		ws.WriteJSON(structure.WebsocketAnswer{
			Error:   true,
			Message: "$max option must be >= 0",
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
