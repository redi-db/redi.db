package handler

import (
	"RediDB/modules/memcache"
	"fmt"
	"os"

	"github.com/goccy/go-json"
	fiber "github.com/gofiber/fiber/v2"
)

func handleUpdate() {
	App.Put("/:database/:collection", func(ctx *fiber.Ctx) error {
		var data struct {
			Data struct {
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

		if data.Data.Update == nil || len(data.Data.Update) == 0 {
			return ctx.JSON(fiber.Map{
				"success": false,
				"message": "No data to update",
			})
		}

		if data.Data.Update["_id"] != nil {
			return ctx.JSON(fiber.Map{
				"success": false,
				"message": "ID property cannot be changed",
			})
		}

		found := memcache.Get(ctx.Params("database"), ctx.Params("collection"), data.Data.Filter)
		if found == nil {
			return ctx.JSON(fiber.Map{
				"success": false,
				"message": "Nothing to update",
			})
		}

		updated := []map[string]interface{}{}

		memcache.Cache.Lock()
		for _, document := range found {
			updatedDocument := memcache.UpdateDocument(document, data.Data.Update)
			encoded, err := json.Marshal(updatedDocument)
			if err != nil {
				updated = append(updated, map[string]interface{}{
					"_id":     document["_id"],
					"created": false,
					"reason":  err.Error(),
				})
				continue
			}

			err = os.WriteFile(fmt.Sprintf("./data/%s/%s/%s.db", ctx.Params("database"), ctx.Params("collection"), document["_id"]), encoded, os.ModePerm)
			if err != nil {
				updated = append(updated, map[string]interface{}{
					"_id":     document["_id"],
					"created": false,
					"reason":  err.Error(),
				})
				continue
			}

			memcache.CacheSet(ctx.Params("database"), ctx.Params("collection"), document["_id"].(string), memcache.UpdateDocument(document, data.Data.Update))
			updated = append(updated, map[string]interface{}{
				"_id":     document["_id"],
				"updated": true,
			})
		}

		memcache.Cache.Unlock()
		return ctx.JSON(updated)
	})
}
