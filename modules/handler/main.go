package handler

import (
	"RediDB/modules/config"
	"RediDB/modules/structure"
	"RediDB/modules/updates"
	"log"
	"strconv"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
)

var LengthOfID = 30
var App = fiber.New(fiber.Config{
	DisableStartupMessage: true,
	BodyLimit:             config.Get().Settings.MaxData * 1024 * 1024,
	ReduceMemoryUsage:     false,
	UnescapePath:          true,
	JSONEncoder:           json.Marshal,
	JSONDecoder:           json.Unmarshal,
})

func init() {
	if config.Get().Settings.WebSocketAllowed {
		HandleWS()
	}

	App.Use(func(ctx *fiber.Ctx) error {
		if ctx.Method() != "GET" {
			ctx.Request().Header.Set("Content-Type", "application/json")

			var auth structure.Auth
			if err := ctx.BodyParser(&auth); err != nil {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": err.Error(),
				})
			}

			if auth.Login != config.Get().Database.Login || auth.Password != config.Get().Database.Password {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": "Authorization failed",
				})
			}

			if len(auth.Database) == 0 {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": "Database required field",
				})
			} else if len(auth.Collection) == 0 {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": "Collection required field",
				})
			}
		}

		return ctx.Next()
	})

	App.Hooks().OnListen(func() error {
		println()
		log.Println("Served server on port " + strconv.Itoa(config.Get().Web.Port))

		if config.Get().Settings.CheckUpdates {
			version, updateRequired, err := updates.Check()
			if err != nil {
				log.Printf("Failed to check updates: %s", err.Error())
				return nil
			}

			if updateRequired {
				log.Printf("New version is available: v%s (Current v%s)", version, updates.VERSION)
			}
		}
		return nil
	})

	handleSearch()
	handleCreate()
	handleSearchOrCreate()
	handleInstantUpdate()
	handleUpdate()
	handleDelete()
	HandleInfo()

	App.Get("*", func(ctx *fiber.Ctx) error {
		ctx.Context().SetStatusCode(400)
		return ctx.JSON(fiber.Map{
			"success": false,
			"message": "Bad Request",
		})
	})
}
