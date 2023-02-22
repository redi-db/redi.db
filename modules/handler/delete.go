package handler

import (
	"RediDB/modules/memcache"
	"fmt"
	"log"
	"os"

	fiber "github.com/gofiber/fiber/v2"
)

func handleDelete() {
	App.Delete("/:database/:collection", func(ctx *fiber.Ctx) error {
		var data struct {
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

		found := memcache.Get(ctx.Params("database"), ctx.Params("collection"), data.Filter)
		if found == nil {
			return ctx.JSON([]interface{}{})
		}

		deleted := []interface{}{}
		memcache.Cache.Lock()
		if len(data.Filter) == 0 {
			err := os.RemoveAll(fmt.Sprintf("./data/%s/%s", ctx.Params("database"), ctx.Params("collection")))
			memcache.CacheDelete(ctx.Params("database"), ctx.Params("collection"), "")

			if err != nil {
				for _, document := range found {
					deleted = append(deleted, map[string]interface{}{
						"_id":     document["_id"],
						"deleted": false,
						"reason":  err.Error(),
					})
				}
			} else {
				for _, document := range found {
					deleted = append(deleted, map[string]interface{}{
						"_id":     document["_id"],
						"deleted": true,
					})
				}
			}

		} else {
			for _, document := range found {
				err := os.Remove(fmt.Sprintf("./data/%s/%s/%s.db", ctx.Params("database"), ctx.Params("collection"), document["_id"]))
				memcache.CacheDelete(ctx.Params("database"), ctx.Params("collection"), document["_id"].(string))

				if err != nil {
					deleted = append(deleted, map[string]interface{}{
						"_id":     document["_id"],
						"deleted": false,
						"reason":  err.Error(),
					})
				} else {
					deleted = append(deleted, map[string]interface{}{
						"_id":     document["_id"],
						"deleted": true,
					})
				}
			}

			if len(memcache.CacheGet()[ctx.Params("database")][ctx.Params("collection")]) == 0 {
				delete(memcache.CacheGet()[ctx.Params("database")], ctx.Params("collection"))
				if err := os.Remove(fmt.Sprintf("./data/%s/%s", ctx.Params("database"), ctx.Params("collection"))); err != nil {
					log.Printf("Failed to delete %s/%s collection: %s", ctx.Params("database"), ctx.Params("collection"), err.Error())
				}
			}
		}

		if len(memcache.CacheGet()[ctx.Params("database")]) == 0 {
			delete(memcache.CacheGet(), ctx.Params("database"))
			if err := os.Remove(fmt.Sprintf("./data/%s/", ctx.Params("database"))); err != nil {
				log.Printf("Failed to delete %s database: %s", ctx.Params("database"), err.Error())
			}
		}

		memcache.Cache.Unlock()
		return ctx.JSON(deleted)
	})
}
