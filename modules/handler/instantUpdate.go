package handler

import (
	"RediDB/modules/memcache"
	"fmt"
	"os"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
)

func handleInstantUpdate() {
	App.Put("/", func(ctx *fiber.Ctx) error {
		var data struct {
			Database   string `json:"database"`
			Collection string `json:"collection"`

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

		found := memcache.Get(data.Database, data.Collection, data.Data.Filter)
		if found == nil {
			return ctx.JSON(fiber.Map{
				"success": false,
				"message": "Nothing to update",
			})
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
		return ctx.JSON(updated)
	})
}
