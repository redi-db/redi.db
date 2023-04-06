package handler

import (
	"RediDB/modules/memcache"
	"fmt"
	"reflect"

	"github.com/gofiber/fiber/v2"
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

			for i, or := range data.Filter["$or"].([]interface{}) {
				if reflect.TypeOf(or).String() != "map[string]interface {}" {
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

		found := memcache.Get(data.Database, data.Collection, data.Filter)
		if found == nil {
			return ctx.JSON([]interface{}{})
		}

		if int(max) == 0 {
			return ctx.JSON(found)
		}

		if int(max) > len(found) {
			max = float64(len(found))
		}

		return ctx.JSON(found[:(int(max))])
	})
}
