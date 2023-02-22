package handler

import (
	"RediDB/modules/memcache"
	"fmt"
	"os"

	"github.com/goccy/go-json"
	fiber "github.com/gofiber/fiber/v2"
)

func handleSearchOrCreate() {
	App.Post("/:database/:collection/searchOrCreate", func(ctx *fiber.Ctx) error {
		var data struct {
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

		found := memcache.Get(ctx.Params("database"), ctx.Params("collection"), data.Filter)
		result := map[string]interface{}{
			"created": false,
		}

		if found == nil {
			result["created"] = true
			id := generateID(LengthOfID)
			document := data.CreateData
			document["_id"] = id

			if _, err := os.Stat(fmt.Sprintf("./data/%s/%s", ctx.Params("database"), ctx.Params("collection"))); os.IsNotExist(err) {
				err := os.MkdirAll(fmt.Sprintf("./data/%s/%s", ctx.Params("database"), ctx.Params("collection")), os.ModePerm)
				if err != nil {
					return ctx.JSON(fiber.Map{
						"success": false,
						"message": err.Error(),
					})
				}
			}

			file, err := os.Create(fmt.Sprintf("./data/%s/%s/%s.db", ctx.Params("database"), ctx.Params("collection"), id))
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

			file.Close()

			memcache.Cache.Lock()
			memcache.CacheSet(ctx.Params("database"), ctx.Params("collection"), id, document)
			memcache.Cache.Unlock()
			result["data"] = document

			return ctx.JSON(result)
		}

		result["data"] = found[0]
		return ctx.JSON(result)
	})
}
