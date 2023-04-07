package handler

import (
	"RediDB/modules/memcache"
	"fmt"
	"os"
	"reflect"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
)

func handleSearchOrCreate() {
	App.Post("/searchOrCreate", func(ctx *fiber.Ctx) error {
		var data struct {
			Database   string `json:"database"`
			Collection string `json:"collection"`

			Filter     map[string]interface{} `json:"filter"`
			CreateData map[string]interface{} `json:"data"`
		}

		if err := ctx.BodyParser(&data); err != nil {
			return ctx.JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		if data.Filter == nil || len(data.Filter) == 0 || data.CreateData == nil || len(data.CreateData) == 0 {
			return ctx.JSON(fiber.Map{
				"success": false,
				"message": "Nothing to look for or create",
			})
		}

		if data.CreateData != nil {
			delete(data.CreateData, "$or")
			delete(data.CreateData, "$order")
			delete(data.CreateData, "$max")
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
				if reflect.TypeOf(or).String() != "map[string]interface {}" {
					return ctx.JSON(fiber.Map{
						"success": false,
						"message": fmt.Sprintf("$or option with index %d is not object", i),
					})
				}
			}
		}

		if data.Filter != nil {
			delete(data.Filter, "$order")
			delete(data.Filter, "$max")
		}

		found := memcache.Get(data.Database, data.Collection, data.Filter, 0)
		result := map[string]interface{}{
			"created": false,
		}

		if found == nil {
			result["created"] = true
			id := generateID(LengthOfID)
			document := data.CreateData
			document["_id"] = id

			if _, err := os.Stat(fmt.Sprintf("./data/%s/%s", data.Database, data.Collection)); os.IsNotExist(err) {
				err := os.MkdirAll(fmt.Sprintf("./data/%s/%s", data.Database, data.Collection), os.ModePerm)
				if err != nil {
					return ctx.JSON(fiber.Map{
						"success": false,
						"message": err.Error(),
					})
				}
			}

			file, err := os.Create(fmt.Sprintf("./data/%s/%s/%s.db", data.Database, data.Collection, id))
			if err != nil {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": err.Error(),
				})
			}

			encoded, err := json.Marshal(document)
			if err != nil {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": err.Error(),
				})
			}

			if _, err := file.WriteString(string(encoded)); err != nil {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": err.Error(),
				})
			}

			_ = file.Close()

			memcache.Cache.Lock()
			memcache.CacheSet(data.Database, data.Collection, id, document)
			memcache.Cache.Unlock()
			result["data"] = document

			return ctx.JSON(result)
		}

		result["data"] = found[0]
		return ctx.JSON(result)
	})
}
