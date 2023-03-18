package handler

import (
	"RediDB/modules/config"
	"RediDB/modules/memcache"

	"github.com/gofiber/fiber/v2"
)

func HandleInfo() {
	App.Get("/list", func(ctx *fiber.Ctx) error {
		login := ctx.FormValue("login")
		password := ctx.FormValue("password")

		if login != config.Get().Database.Login || password != config.Get().Database.Password {
			return ctx.JSON(fiber.Map{
				"success": false,
				"message": "Authorization failed",
			})
		}

		var response []struct {
			Name        string `json:"name"`
			Collections []struct {
				Name  string `json:"name"`
				Count int    `json:"count"`
			} `json:"collections"`
		}

		memcache.Cache.RLock()
		cache := memcache.CacheGet()
		for databaseName := range cache {
			data := struct {
				Name        string `json:"name"`
				Collections []struct {
					Name  string `json:"name"`
					Count int    `json:"count"`
				} `json:"collections"`
			}{
				Name: databaseName,
			}

			for collectionName, collection := range cache[databaseName] {
				data.Collections = append(data.Collections, struct {
					Name  string `json:"name"`
					Count int    `json:"count"`
				}{
					Name:  collectionName,
					Count: len(collection),
				})
			}

			response = append(response, data)
		}

		memcache.Cache.RUnlock()
		return ctx.JSON(response)
	})
}
