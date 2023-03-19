package handler

import (
	"RediDB/modules/memcache"
	"reflect"

	"github.com/gofiber/fiber/v2"
)

func handleSearch() {
	App.Post("/search", func(ctx *fiber.Ctx) error {
		var data struct {
			Database string `json:"database"`
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
